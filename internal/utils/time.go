package utils

import "time"

func DurationString(from time.Time) string {
	return time.Since(from).String()
}
