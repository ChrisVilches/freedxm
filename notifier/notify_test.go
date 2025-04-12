package notifier

import (
	"reflect"
	"testing"
)

func TestGetExecArgs(t *testing.T) {
	msg := "test message"
	tests := []struct {
		isWarn   bool
		notifier string
		msgs     []string
		expected []string
	}{
		{false, "notify-send", []string{msg}, []string{"notify-send", "-u", "normal", msg}},
		{true, "notify-send", []string{msg}, []string{"notify-send", "-u", "critical", msg}},
		{false, "i3-nagbar", []string{msg}, []string{"i3-nagbar", "-t", "warning", "-m", msg}},
		{true, "i3-nagbar", []string{msg}, []string{"i3-nagbar", "-t", "error", "-m", msg}},
		{false, "dunstify", []string{msg}, []string{"dunstify", "-u", "normal", msg}},
		{true, "dunstify", []string{msg}, []string{"dunstify", "-u", "critical", msg}},
		{false, "unknown", []string{msg}, nil},
	}

	for _, test := range tests {
		result := getExecArgs(test.isWarn, test.notifier, test.msgs...)
		if !reflect.DeepEqual(result, test.expected) {
			t.Errorf("getExecArgs(%v, %s, %v) = %v; want %v", test.isWarn, test.notifier, test.msgs, result, test.expected)
		}
	}
}
