package clock

import "time"

// Real is the production Clock. It reads the monotonic clock.
type Real struct {
	start time.Time
}

// NewReal returns a production clock.
func NewReal() *Real {
	return &Real{start: time.Now()}
}

// Now returns the monotonic time since process start, in nanoseconds.
func (c *Real) Now() Time {
	return Time(time.Since(c.start))
}

// Tick is a no-op in production. Real time advances on its own.
func (c *Real) Tick() {}

