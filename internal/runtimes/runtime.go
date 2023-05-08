package runtimes

import (
	"path/filepath"

	"github.com/containers/podman/v4/pkg/specgen"
	"github.com/opencontainers/runtime-spec/specs-go"
)

// Define config for an image
type Image struct {
	FullyQualifiedName string
	ShortName          string
	// key is a single directory name, value is the full path container side
	Mounts map[string]string
	EnvVars []*EnvVar
}

// Define config for a runtime
type Runtime struct {
	Name   string
	Images map[string]Image
}

// Get all mount names, that is the keys of Image.Mounts for each image
// No duplicates
func (r Runtime) MountNames() []string {
	names := make(map[string]struct{})
	for _, image := range r.Images {
		for name := range image.Mounts {
			names[name] = struct{}{}
		}
	}

	keys := make([]string, 0, len(names))
	for k := range names {
		keys = append(keys, k)
	}
	return keys
}

func (i Image) ToContainerSpec(basePath string, inputEnvVar map[string]string) (*specgen.SpecGenerator, error) {
	spec := specgen.NewSpecGenerator(i.FullyQualifiedName, false)
	spec.Terminal = true
	spec.Env = make(map[string]string)

	for name, path := range i.Mounts {
		// May break if two mounts have the same directory name
		spec.Mounts = append(spec.Mounts, specs.Mount{
			Destination: path,
			Type:        "bind",
			Source:      filepath.Join(basePath, name),
		})
	}

	// Firstly add all env vars from user input
	for name, value := range inputEnvVar {
		spec.Env[name] = value
	}

	// Set env vars from runtime, using input env vars as modifier parameters
	// If no input env vars are provided, use the default value
	for _, env := range i.EnvVars {
		input, exists := inputEnvVar[env.Name]

		var envValue string
		var err error

		if exists {
			envValue, err = env.ApplyModifiersWithInput(&input)
		} else {
			envValue, err = env.ApplyModifiers()
		}

		if err != nil {
			return nil, err
		}

		spec.Env[env.Name] = envValue
	}

	return spec, nil
}
