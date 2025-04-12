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

func notifI3nagbar(isWarn bool, msgs ...string) []string {
	level := "warning"
	if isWarn {
		level = "error"
	}
	return []string{"i3-nagbar", "-t", level, "-m", strings.Join(msgs, " ")}
}

// TODO: Why does this print a (U)???
func notifNotifySend(isWarn bool, useDunstify bool, msgs ...string) []string {
	level := "normal"
	if isWarn {
		level = "critical"
	}

	program := "notify-send"
	if useDunstify {
		program = "dunstify"
	}

	return []string{program, "-u", level, strings.Join(msgs, " ")}
}

func getExecArgs(isWarn bool, notifier string, msgs ...string) []string {
	switch notifier {
	case "notify-send":
		return notifNotifySend(isWarn, false, msgs...)
	case "i3-nagbar":
		return notifI3nagbar(isWarn, msgs...)
	case "dunstify":
		return notifNotifySend(isWarn, true, msgs...)
	default:
		return nil
	}
}

func notifyAux(isWarn bool, msgs ...string) {
	conf, err := config.GetConfig()

	if err != nil {
		log.Println(err)
		return
	}

	notifier := conf.Options.Notifier

	if notifier == "" {
		return
	}

	args := getExecArgs(isWarn, notifier, msgs...)

	if args == nil {
		log.Println("notifier is wrong:", notifier)
		return
	}

	go handleCmd(exec.Command(args[0], args[1:]...))
}

func Notify(msgs ...string) {
	notifyAux(false, msgs...)
}

func NotifyWarn(msgs ...string) {
	notifyAux(true, msgs...)
}
