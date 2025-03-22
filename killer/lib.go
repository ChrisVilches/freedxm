package killer

import (
	"fmt"
	"os"
	"os/exec"
)

// TODO: This may not work if the process name is different. Verify that.
func KillAll(processName string) {
	cmd := exec.Command("killall", processName)

	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error running killall: %v\n", err)
	} else {
		fmt.Printf("Killall command executed successfully (%s)\n", processName)
	}
}
