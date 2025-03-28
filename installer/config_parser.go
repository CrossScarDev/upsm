package main

import "strings"

func ParseConfig(content string) map[string]string {
	config := make(map[string]string)

	for _, line := range strings.Split(content, "\n") {
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			continue
		}

		config[key] = value
	}

	return config
}
