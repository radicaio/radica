package sm

import (
	"bytes"
	"testing"

	"github.com/radicaio/radica/internal/kv"
	"github.com/radicaio/radica/internal/wire"
)

// TestApply_GetSetDelete verifies each op's reply contract end-to-end
// against the MVP map-backed KV.
func TestApply_GetSetDelete(t *testing.T) {
	sm := New(kv.NewMap())

	// Get on empty: Miss.
	r := sm.Apply(wire.Op{Code: wire.OpGet, Key: []byte("k")}, 0)
	if r.Status != wire.StatusMiss {
		t.Fatalf("Get miss: status=%d", r.Status)
	}

	// Set: OK.
	r = sm.Apply(wire.Op{Code: wire.OpSet, Key: []byte("k"), Value: []byte("v")}, 0)
	if r.Status != wire.StatusOK {
		t.Fatalf("Set: status=%d", r.Status)
	}

	// Get present: OK + value.
	r = sm.Apply(wire.Op{Code: wire.OpGet, Key: []byte("k")}, 0)
	if r.Status != wire.StatusOK || !bytes.Equal(r.Value, []byte("v")) {
		t.Fatalf("Get hit: status=%d value=%q", r.Status, r.Value)
	}

	// Delete present: OK.
	r = sm.Apply(wire.Op{Code: wire.OpDelete, Key: []byte("k")}, 0)
	if r.Status != wire.StatusOK {
		t.Fatalf("Delete present: status=%d", r.Status)
	}

	// Delete missing: Miss.
	r = sm.Apply(wire.Op{Code: wire.OpDelete, Key: []byte("k")}, 0)
	if r.Status != wire.StatusMiss {
		t.Fatalf("Delete missing: status=%d", r.Status)
	}
}

// TestApply_InvalidOp verifies the catch-all returns StatusInvalid rather
// than panicking, so a malformed op doesn't take the replica down.
func TestApply_InvalidOp(t *testing.T) {
	sm := New(kv.NewMap())
	r := sm.Apply(wire.Op{Code: wire.OpCode(99)}, 0)
	if r.Status != wire.StatusInvalid {
		t.Fatalf("invalid op: status=%d", r.Status)
	}
}
