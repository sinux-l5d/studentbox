package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/sinux-l5d/studentbox/internal/runtimes"
)

func getImageConfigFromFile(path string) (*runtimes.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	// extract content of file
	contentBytes, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}

	runtime := filepath.Base(filepath.Dir(path))
	shortName := strings.TrimSuffix(filepath.Base(path), ".containerfile")
	image := &runtimes.Image{
		ShortName:          shortName,
		FullyQualifiedName: "ghcr.io/sinux-l5d/studentbox/runtime/" + runtime + "." + shortName,
		Mounts:             make(map[string]string),
		EnvVar:             make([]*runtimes.EnvVar, 0),
	}
	// extract labels from content
	labels, err := extractLabelsFromDockerfile(string(contentBytes))
	die(err)
	for name, value := range labels {
		switch name {
		case "studentbox.config.mounts":
			// split mounts
			mounts := strings.Split(value, ",")
			// split each mount into key and value
			for _, mount := range mounts {
				mountSplit := strings.Split(mount, ":")
				image.Mounts[mountSplit[0]] = mountSplit[1]
			}
		case "studentbox.config.envs":
			// get ENV instructions for default values
			defaultsValues, err := extractEnvFromDockerfile(string(contentBytes))
			die(err)
			envvars, err := parseEnvConfig(value, defaultsValues)
			die(err)
			image.EnvVar = envvars
		}

	}
	return image, nil
}

func extractLabelsFromDockerfile(dockerfile string) (map[string]string, error) {
	labelRegex := regexp.MustCompile(`LABEL\s+((?:(?:(?:"[^"]+")|(?:[^\s]+))=(?:(?:"[^"]+")|(?:[^\s]+))\s*)+)`)
	keyValueRegex := regexp.MustCompile(`(?:(?:(?:"([^"]+)")|(?:([^\s]+)))=(?:(?:"([^"]+)")|(?:([^\s]+))))`)

	labels := make(map[string]string)

	lines := strings.Split(dockerfile, "\n")
	for _, line := range lines {
		matches := labelRegex.FindStringSubmatch(line)
		if len(matches) > 0 {
			keyValuePairs := keyValueRegex.FindAllStringSubmatch(matches[1], -1)
			for _, pair := range keyValuePairs {
				key := pair[1]
				if key == "" {
					key = pair[2]
				}
				value := pair[3]
				if value == "" {
					value = pair[4]
				}
				labels[key] = value
			}
		}
	}
	return labels, nil
}

// won't work if env value is containing a reference to another variable
func extractEnvFromDockerfile(dockerfile string) (map[string]string, error) {
	envRegex := regexp.MustCompile(`ENV\s+((?:(?:(?:"[^"]+")|(?:[^\s]+))=(?:(?:"[^"]+")|(?:[^\s]+))\s*)+)`)
	keyValueRegex := regexp.MustCompile(`(?:(?:(?:"([^"]+)")|(?:([^\s]+)))=(?:(?:"([^"]+)")|(?:([^\s]+))))`)

	env := make(map[string]string)

	lines := strings.Split(dockerfile, "\n")
	for _, line := range lines {
		matches := envRegex.FindStringSubmatch(line)
		if len(matches) > 0 {
			keyValuePairs := keyValueRegex.FindAllStringSubmatch(matches[1], -1)
			for _, pair := range keyValuePairs {
				key := pair[1]
				if key == "" {
					key = pair[2]
				}
				value := pair[3]
				if value == "" {
					value = pair[4]
				}
				env[key] = value
			}
		}
	}
	return env, nil
}

// Parse a studentbox.config.envs string into a slice of EnvVar
// e.g. : "MYSQL_DATABASE,MYSQL_USER,MYSQL_PASSWORD:password(10),MYSQL_ROOT_PASSWORD:password(30)"
func parseEnvConfig(envConfig string, defaultValues map[string]string) ([]*runtimes.EnvVar, error) {
	rawVars := strings.Split(envConfig, ",")
	envVars := make([]*runtimes.EnvVar, len(rawVars))

	for i, rawVar := range rawVars {
		envVar := &runtimes.EnvVar{}
		envVarName, params, found := strings.Cut(rawVar, ":")

		envVar.Name = envVarName
		envVar.DefaultValue = defaultValues[envVarName]

		if found {
			modifiers, err := parseModifiers(params)
			if err != nil {
				return nil, err
			}
			envVar.Modifiers = modifiers
		}
		envVars[i] = envVar
	}
	return envVars, nil
}

// Parse modifiers of a env var
// e.g. : "password(10):failempty"
func parseModifiers(modifiers string) ([]runtimes.EnvModifierParams, error) {
	rawModifiers := strings.Split(modifiers, ":")
	envModifiers := make([]runtimes.EnvModifierParams, len(rawModifiers))

	singleModifierRegex := regexp.MustCompile(`^(?P<name>[a-zA-Z]+)(\((?P<params>[a-zA-Z0-9]+(?:,[a-zA-Z0-9]+)*)\))?$`)

	for i, rawModifier := range rawModifiers {
		envModifier := runtimes.EnvModifierParams{}

		matcher := singleModifierRegex.FindStringSubmatch(rawModifier)
		matches := make(map[string]string)
		for i, name := range singleModifierRegex.SubexpNames() {
			if i != 0 && name != "" {
				matches[name] = matcher[i]
			}
		}

		if len(matches) == 0 {
			return nil, fmt.Errorf("invalid modifier %s", rawModifier)
		}

		envModifier.Name = matches["name"]
		if len(matches["params"]) != 0 {
			envModifier.Params = strings.Split(matches["params"], ",")
		}

		envModifiers[i] = envModifier
	}
	return envModifiers, nil
}
