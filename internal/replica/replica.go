// Package replica owns the event loop and wires together Clock, Transport,
// and StateMachine.
//
// The event loop is Replica.Run. There is no separate "loop" package.
// In MVP, the standalone binary calls Run on a goroutine; the embedded
// API does the same. A future DST simulator will call Tick directly.
package replica

import (
	"context"
	"time"

	"github.com/radicaio/radica/internal/clock"
	"github.com/radicaio/radica/internal/sm"
	"github.com/radicaio/radica/internal/transport"
)

// Replica is the running coordination state machine.
type Replica struct {
	clock     clock.Clock
	transport *transport.Transport
	sm        *sm.StateMachine

	// tickPeriod is how long Run sleeps between iterations when nothing is
	// happening. Keeps the loop from busy-spinning.
	tickPeriod time.Duration
}

// New wires the three dependencies into a Replica. The handler is
// registered with the transport here, so callers only need to provide the
// dependencies and then call Run.
func New(c clock.Clock, t *transport.Transport, m *sm.StateMachine) *Replica {
	r := &Replica{
		clock:      c,
		transport:  t,
		sm:         m,
		tickPeriod: time.Millisecond,
	}
	t.SetHandler(r.onRequest)
	return r
}

// Run is the production event loop. Blocks until ctx is cancelled.
// Per the style guide, library code does not spawn its own goroutine;
// the caller invokes Run on a goroutine they own (or, in the standalone
// binary, on main).
func (r *Replica) Run(ctx context.Context) error {
	for ctx.Err() == nil {
		r.Tick()
		// Small sleep to avoid pegging a core when idle. A real
		// production loop will replace this with a blocking
		// transport.Tick that wakes on I/O readiness.
		time.Sleep(r.tickPeriod)
	}
	return ctx.Err()
}

// Tick performs one iteration of the loop. Exposed for DST.
func (r *Replica) Tick() {
	r.transport.Tick()
	r.clock.Tick()
	// Replica has no deferred internal work yet; left for future use
	// (timer wheel, queued retries, VSR timeouts).
}

// onRequest is the receive handler registered with the transport.
// Runs on the loop goroutine.
func (r *Replica) onRequest(req transport.Request) {
	reply := r.sm.Apply(req.Op, r.clock.Now())
	r.transport.Send(transport.Reply{RequestID: req.ID, Reply: reply})
}
