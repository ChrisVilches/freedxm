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

func notifZenity(isWarn bool, title string, msgs ...string) []string {
	level := "--info"
	if isWarn {
		level = "--warning"
	}
	return []string{
		"zenity", level, "--title", title, "--text", strings.Join(msgs, " ")}
}

// TODO: Why does this print a (U)???
func notifNotifySend(
	isWarn bool,
	useDunstify bool,
	title string,
	msgs ...string,
) []string {
	level := "normal"
	if isWarn {
		level = "critical"
	}

	program := "notify-send"
	if useDunstify {
		program = "dunstify"
	}

	return []string{program, "-u", level, title, strings.Join(msgs, " ")}
}

func getExecArgs(isWarn bool, notifier, title string, msgs ...string) []string {
	switch notifier {
	case "notify-send":
		return notifNotifySend(isWarn, false, title, msgs...)
	case "dunstify":
		return notifNotifySend(isWarn, true, title, msgs...)
	case "zenity":
		return notifZenity(isWarn, title, msgs...)
	default:
		return nil
	}
}

func notifyAux(isWarn bool, title string, msgs ...string) {
	conf, err := config.GetConfig()

	if err != nil {
		log.Println(err)
		return
	}

	notifier := conf.Options.Notifier

	if notifier == "" {
		return
	}

	args := getExecArgs(isWarn, notifier, title, msgs...)

	if args == nil {
		log.Println("notifier is wrong:", notifier)
		return
	}

	go handleCmd(exec.Command(args[0], args[1:]...))
}

func Notify(title string, msgs ...string) {
	notifyAux(false, title, msgs...)
}

func NotifyWarn(title string, msgs ...string) {
	notifyAux(true, title, msgs...)
}
