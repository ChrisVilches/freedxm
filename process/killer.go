package process

import (
	"log"
	"os/exec"
	"sync"
)

var mu sync.Mutex
var active = make(map[string]struct{})

func toggleActive(processName string, val bool) bool {
	mu.Lock()
	defer mu.Unlock()

	if !val {
		delete(active, processName)
		return true
	}

	_, exists := active[processName]
	if !exists {
		active[processName] = struct{}{}
	}
	return !exists
}

func tryKill(processName string) {
	if !toggleActive(processName, true) {
		log.Println("(being killed already) couldn't kill", processName)
		return
	}

	defer toggleActive(processName, false)

	flags := []string{"-TERM", "-9"}

	for _, flag := range flags {
		cmd := exec.Command("killall", flag, processName)
		if err := cmd.Run(); err == nil {
			log.Printf("killed %s (%s)", processName, flag)
			return
		}
	}

	log.Printf("failed to kill process %s", processName)
}

func KillAll(processName string) {
	go tryKill(processName)
}
