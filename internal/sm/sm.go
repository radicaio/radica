// Package sm holds the deterministic state machine.
//
// StateMachine.Apply is pure with respect to its inputs: for the same
// starting state and the same op, it produces the same (state', reply)
// byte-for-byte. This property is what makes DST and future VSR replication
// work.
//
// The state machine owns the KV. It never touches the network, the clock,
// or any external resource. Timestamps are passed in.
package sm

import (
	"github.com/radicaio/radica/internal/clock"
	"github.com/radicaio/radica/internal/kv"
	"github.com/radicaio/radica/internal/wire"
)

// StateMachine is a concrete struct, not an interface — no interface
// dispatch on the hot path. The swap point is the KV inside it.
type StateMachine struct {
	kv kv.KV
}

// New returns a state machine backed by the given KV.
func New(store kv.KV) *StateMachine {
	return &StateMachine{kv: store}
}

// Apply executes one op against the state. The now parameter is unused in
// MVP; it is present so the signature matches the future TTL/expiry-aware
// version without a breaking change.
func (sm *StateMachine) Apply(op wire.Op, _ clock.Time) wire.Reply {
	switch op.Code {
	case wire.OpGet:
		v, ok := sm.kv.Get(op.Key)
		if !ok {
			return wire.Reply{Status: wire.StatusMiss}
		}
		return wire.Reply{Status: wire.StatusOK, Value: v}

	case wire.OpSet:
		sm.kv.Set(op.Key, op.Value)
		return wire.Reply{Status: wire.StatusOK}

	case wire.OpDelete:
		existed := sm.kv.Delete(op.Key)
		if !existed {
			return wire.Reply{Status: wire.StatusMiss}
		}
		return wire.Reply{Status: wire.StatusOK}

	default:
		return wire.Reply{Status: wire.StatusInvalid}
	}
}

// Snapshot is sketched for future VSR state transfer and persistence.
// Panics in MVP.
func (sm *StateMachine) Snapshot() []byte {
	panic("sm.Snapshot: not implemented in MVP")
}

// Restore is sketched for future VSR state transfer and persistence.
// Panics in MVP.
func (sm *StateMachine) Restore(_ []byte) error {
	panic("sm.Restore: not implemented in MVP")
}
