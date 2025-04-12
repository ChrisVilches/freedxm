package notifier

import (
	"log"
	"os/exec"
	"strings"

	"github.com/ChrisVilches/freedxm/config"
)

func handleCmd(cmd *exec.Cmd) {
	err := cmd.Run()
	if err != nil {
		log.Println("Failed to send notification:", err)
	}
}

func notifI3nagbar(msgs ...string) {
	handleCmd(exec.Command("i3-nagbar", "-m", strings.Join(msgs, " ")))
}

// TODO: Why does this print a (U)???
func notifNotifySend(msgs ...string) {
	handleCmd(exec.Command("notify-send", strings.Join(msgs, " ")))
}

var notifierFunctions = map[string]func(...string){
	"notify-send": notifNotifySend,
	"i3-nagbar":   notifI3nagbar,
}

func Notify(msgs ...string) {
	conf, err := config.GetConfig()

	if err != nil {
		log.Println(err)
		return
	}

	notifier := conf.Options.Notifier

	if notifier == "" {
		return
	}

	if notifyFunc, exists := notifierFunctions[notifier]; exists {
		go notifyFunc(msgs...)
	} else {
		log.Println("notifier is wrong:", notifier)
	}
}
