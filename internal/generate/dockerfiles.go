package main

import (
	"regexp"
	"strings"
)

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
