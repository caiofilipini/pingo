package pinger

import "time"

// clock is a type capable of returning the current time.
type clock interface {
	// Now returns the current time.
	Now() time.Time
}

// defaultClock is the default clock implementation. It relies
// on the time package to return the current system time.
type defaultClock struct{}

// Now returns the current time as returned by time.Now().
func (defaultClock) Now() time.Time {
	return time.Now()
}

// fakeClock is a fake clock implementation to be used in tests.
type fakeClock struct {
	fakeTime time.Time
}

// Now returns the fakeTime configured for this fakeClock.
func (f fakeClock) Now() time.Time {
	return f.fakeTime
}
