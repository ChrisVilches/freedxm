package notifier

import (
	"reflect"
	"testing"
)

func TestGetExecArgs(t *testing.T) {
	msgs := []string{"test", "message"}
	msg := "test message"

	title := "Test Title"
	tests := []struct {
		isWarn   bool
		notifier string
		msgs     []string
		expected []string
	}{
		{false, "notify-send", msgs, []string{"notify-send", "-u", "normal", title, msg}},
		{true, "notify-send", msgs, []string{"notify-send", "-u", "critical", title, msg}},
		{false, "dunstify", msgs, []string{"dunstify", "-u", "normal", title, msg}},
		{true, "dunstify", msgs, []string{"dunstify", "-u", "critical", title, msg}},
		{false, "zenity", msgs, []string{"zenity", "--info", "--title", title, "--text", msg}},
		{true, "zenity", msgs, []string{"zenity", "--warning", "--title", title, "--text", msg}},
		{false, "unknown", msgs, nil},
	}

	for _, test := range tests {
		result := getExecArgs(test.isWarn, test.notifier, title, test.msgs...)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("getExecArgs(%v, %s, %v) = %v; want %v", test.isWarn, test.notifier, test.msgs, result, test.expected)
		}
	}
}
