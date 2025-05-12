package util

import (
	"testing"
)

func TestSecondsToHHMMSS(t *testing.T) {
	tests := []struct {
		input    int64
		expected string
	}{
		{0, "0:00:00"},
		{59, "0:00:59"},
		{60, "0:01:00"},
		{3599, "0:59:59"},
		{3600, "1:00:00"},
		{3661, "1:01:01"},
		{86399, "23:59:59"},
		{86400, "24:00:00"},
		{987654, "274:20:54"},
		{360000, "100:00:00"},
		{360001, "100:00:01"},
		{-0, "0:00:00"},
		{-1, "-0:00:01"},
		{-60, "-0:01:00"},
		{-3661, "-1:01:01"},
		{-987654, "-274:20:54"},
		{-360000, "-100:00:00"},
	}

	for _, tt := range tests {
		got := SecondsToHHMMSS(tt.input)
		if got != tt.expected {
			t.Errorf("SecondsToHHMMSS(%d) = %q; want %q", tt.input, got, tt.expected)
		}
	}
}
