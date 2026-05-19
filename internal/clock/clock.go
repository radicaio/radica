// Package clock provides the time abstraction used by the replica.
//
// Production code uses Real, which reads monotonic time.
// DST code uses Sim, which exposes a virtual clock advanced by the scheduler.
// Core code never calls time.Now directly.
package clock

import "time"

// Time is a monotonic instant. Opaque outside this package.
type Time int64

// Duration is the difference between two Times, in nanoseconds.
type Duration = time.Duration

// Clock returns the current time and advances internal state on Tick.
type Clock interface {
	// Now returns the current monotonic time.
	Now() Time

	// Tick is called once per event-loop iteration. In production it is a
	// no-op; in simulation the scheduler advances virtual time elsewhere.
	Tick()
}
