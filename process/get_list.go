package process

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
)

type process struct {
	PID  int
	Name string
}

// TODO: Does this include all possible process names? I think so, but verify
var statRegex = regexp.MustCompile(`^(\d+)\s\(([^)]+)\)`)

// TODO: Note, one thing I had in mind is to do a not so frequent polling but increase the polling
// if the user is active (moving mouse or keyboard), that would put some weight on the mouse hooks though,
// but I don't know. It's an idea to explore. But I think this would only happen during a blocking session,
// not at all times, even if I run this process as a long-running service.

// TODO: write a fucking comment
// TODO: Works, but it should state that the names are unique. Maybe make it configurable by parameters or change the function name.
func GetProcessList() ([]process, error) {
	procDirs, err := filepath.Glob("/proc/[0-9]*/stat")
	if err != nil {
		return nil, err
	}

	result := []process{}
	// TODO: This "set" is trash. Is there a better way???
	addedNames := map[string]struct{}{}

	for _, procFile := range procDirs {
		data, err := os.ReadFile(procFile)
		if err != nil {
			// TODO: This "ignore errors" is a bit suspicious.
			continue // Ignore errors (process may have exited)
		}

		match := statRegex.FindStringSubmatch(string(data))

		if len(match) < 3 {
			// TODO: I think it should return error here (assuming all lines should have the desired format).
			// but perhaps not all lines have the expected format, in that case it'd be better to ignore it
			// and log the error.
			return nil, fmt.Errorf("no match")
		}

		// TODO: Verify this is doing something. It should make them unique (by name).
		if _, exists := addedNames[match[2]]; exists {
			continue
		}

		pid, err := strconv.Atoi(match[1])

		if err != nil {
			return nil, fmt.Errorf("cannot convert PID (%v)", err)
		}

		result = append(result, process{PID: pid, Name: match[2]})
		addedNames[match[2]] = struct{}{}
	}

	return result, nil
}
