package clock

// Sim is a virtual clock advanced explicitly by the DST scheduler.
// Tick is a no-op; the scheduler calls Advance between events.
type Sim struct {
	now Time
}

// NewSim returns a virtual clock starting at t=0.
func NewSim() *Sim {
	return &Sim{}
}

// Now returns the current virtual time.
func (c *Sim) Now() Time {
	return c.now
}

// Tick is a no-op. Sim time only moves on Advance.
func (c *Sim) Tick() {}

// Advance moves virtual time forward by d.
// Called only by the DST scheduler.
func (c *Sim) Advance(d Duration) {
	c.now += Time(d)
}

// Set jumps virtual time to t. Used to advance to the next scheduled event.
func (c *Sim) Set(t Time) {
	if t < c.now {
		panic("clock.Sim.Set: time cannot go backwards")
	}
	c.now = t
}
