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
		log.Println("failed to send notification:", err)
	}
}

// TODO: Why does this print a (U)???
func notifyAux(isWarn bool, title string, msgs ...string) {
	conf, err := config.GetConfig()

	if err != nil {
		log.Println(err)
		return
	}

	args := conf.Notification.Normal

	if isWarn {
		args = conf.Notification.Warning
	}

	if args == nil || len(args) == 0 {
		return
	}

	fullMessage := strings.Join(msgs, " ")

	for i := range args {
		args[i] = strings.ReplaceAll(args[i], "%title", title)
		args[i] = strings.ReplaceAll(args[i], "%message", fullMessage)
	}

	go handleCmd(exec.Command(args[0], args[1:]...))
}

func Notify(title string, msgs ...string) {
	notifyAux(false, title, msgs...)
}

func NotifyWarn(title string, msgs ...string) {
	notifyAux(true, title, msgs...)
}
