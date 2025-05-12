package util

import (
	"fmt"
)

const (
	secondsPerMinute = 60
	secondsPerHour   = 60 * secondsPerMinute
)

func SecondsToHHMMSS(totalSeconds int32) string {
	sign := ""
	if totalSeconds < 0 {
		sign = "-"
		totalSeconds = -totalSeconds
	}

	hours := totalSeconds / secondsPerHour
	minutes := (totalSeconds % secondsPerHour) / secondsPerMinute
	seconds := totalSeconds % secondsPerMinute

	return fmt.Sprintf("%s%d:%02d:%02d", sign, hours, minutes, seconds)
}
