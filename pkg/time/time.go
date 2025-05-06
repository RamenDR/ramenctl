package time

import "time"

// A Time represents an instant in time with nanosecond precision. This is an
// alias of time.Time to make it possible to control the time during tests.
type Time = time.Time

// Now returns the current local time. It can be set to fake time function
// during tests.
var Now func() Time

// Since returns the time elapsed since t. It is shorthand for time.Now().Sub(t).
func Since(t Time) time.Duration {
	return Now().Sub(t)
}

func now() Time {
	return time.Now()
}

func init() {
	Now = now
}
