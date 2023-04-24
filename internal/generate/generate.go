// this file is used to generate various files for the project
package main

//go:generate go run . runtimes

import (
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/sinux-l5d/studentbox/internal/runtimes"
)

func die(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	args := os.Args[1:]
	switch args[0] {
	case "runtimes":
		generateRuntimes()
	default:
		die(errors.New("unknown generator"))
	}
}

var runtimesTemplate = `// Code generated by go generate; DO NOT EDIT.
package runtimes

var OfficialRuntimes = map[string]Runtime{
	{{- range $key, $value := .}}
	"{{ $key }}": {
		{{- with $value }}
		Name: "{{ .Name }}",
		Images: map[string]Image{
			{{- range .Images }}
			"{{ .ShortName }}": {
				FullyQualifiedName: "{{ .FullyQualifiedName }}",
				ShortName:          "{{ .ShortName }}",
				Mounts: map[string]string{
					{{- range $key, $value := .Mounts }}
					"{{ $key }}": "{{ $value }}",
					{{- end }}
				},
			},
			{{- end }}
		{{- end }}
		},
	},
	{{- end }}
}
`

func generateRuntimes() {
	tmpl := template.Must(template.New("runtimes").Parse(runtimesTemplate))

	// get dir names in runtimes dir
	runtimesDir, err := os.Open("../../runtimes")
	die(err)

	runtimesName, err := runtimesDir.Readdirnames(0)
	runtimesDir.Close()
	die(err)

	forTemplate := make(map[string]runtimes.Runtime, len(runtimesName))
	for _, runtimeName := range runtimesName {
		// get dir names in runtimes/<runtimeName> dir
		runtimeDir, err := os.Open("../../runtimes/" + runtimeName)
		die(err)
		defer runtimeDir.Close()

		forTemplate[runtimeName] = runtimes.Runtime{Name: runtimeName, Images: make(map[string]runtimes.Image)}

		// Get all file in runtimes/<runtimeName> dir but keep only the ones that end with .containerfile
		runtimeFiles, err := runtimeDir.Readdirnames(0)
		die(err)
		for _, filename := range runtimeFiles {
			if strings.HasSuffix(filename, ".containerfile") {
				imgConfig, err := getImageConfigFromFile("../../runtimes/" + runtimeName + "/" + filename)
				die(err)
				forTemplate[runtimeName].Images[imgConfig.ShortName] = *imgConfig
			}
		}
	}

	f, err := os.Create("../runtimes/generated.go")
	die(err)
	defer f.Close()
	tmpl.Execute(f, forTemplate)
}

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
	}
	// extract labels from content
	labels, err := extractLabelsFromDockerfile(string(contentBytes))
	die(err)
	for k, v := range labels {
		switch k {
		case "studentbox.config.mounts":
			// split mounts
			mounts := strings.Split(v, ",")
			// split each mount into key and value
			for _, mount := range mounts {
				mountSplit := strings.Split(mount, ":")
				image.Mounts[mountSplit[0]] = mountSplit[1]
			}
		}
	}
	return image, nil
}
