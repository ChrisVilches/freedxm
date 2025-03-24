package process

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
)

var statRegex = regexp.MustCompile(`^\d+\s\(([^)]+)\)`)
var requiredMatchElements = 2

func GetProcessNames() ([]string, error) {
	procDirs, err := filepath.Glob("/proc/[0-9]*/stat")
	if err != nil {
		return nil, err
	}

	names := map[string]struct{}{}

	for _, procFile := range procDirs {
		data, err := os.ReadFile(procFile)
		if err != nil {
			continue
		}

		match := statRegex.FindStringSubmatch(string(data))

		if len(match) < requiredMatchElements {
			return nil, fmt.Errorf(
				"failed to parse /proc/[pid]/stat file: unexpected format",
			)
		}

		names[match[1]] = struct{}{}
	}

	result := []string{}

	for name := range names {
		result = append(result, name)
	}

	return result, nil
}
