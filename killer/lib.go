package killer

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

// TODO: should it be read/write mutex or some variation??
var mu sync.Mutex
var active = make(map[string]struct{})

// TODO: double-check this function. It should set to active and return true if it was set,
// or false if it was already set. This is like a CompareAndSwap (atomic bool) but for a map and
// protected by a mutex.
func toggleActive(processName string, val bool) bool {
	mu.Lock()
	defer mu.Unlock()

	if val {
		_, exists := active[processName]
		if exists {
			return false
		}
		active[processName] = struct{}{}
	} else {
		delete(active, processName)
	}
	return true
}

func tryKill(processName string) {
	if !toggleActive(processName, true) {
		fmt.Println("(being killed already) couldn't kill", processName)
		return
	}

	defer toggleActive(processName, false)

	flags := []string{"-TERM", "-9"}

	for _, flag := range flags {
		cmd := exec.Command("killall", flag, processName)
		if err := cmd.Run(); err == nil {
			fmt.Printf("Killed %s (%s)\n", processName, flag)
			return
		}
	}

	fmt.Fprintf(os.Stderr, "failed to kill process %s", processName)
}

// TODO: This may not work if the process name is different. Verify that.
func KillAll(processName string) {
	go tryKill(processName)
}
