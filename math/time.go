package math

import "time"

// TimeInMillis returns the amount of milliseconds in d as a float64.
func TimeInMillis(d time.Duration) float64 {
	return float64(d.Nanoseconds()) / (float64(time.Millisecond) / float64(time.Nanosecond))
}
