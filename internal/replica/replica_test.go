package replica

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/radicaio/radica/internal/clock"
	"github.com/radicaio/radica/internal/kv"
	"github.com/radicaio/radica/internal/sm"
	"github.com/radicaio/radica/internal/transport"
	"github.com/radicaio/radica/internal/wire"
)

// TestReplica_EndToEnd drives Set then Get through the full stack
// (transport → replica → state machine) and waits for the replies via
// the transport. This is the smallest possible end-to-end check.
func TestReplica_EndToEnd(t *testing.T) {
	tp := transport.New()
	r := New(clock.NewReal(), tp, sm.New(kv.NewMap()))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	done := make(chan error, 1)
	go func() { done <- r.Run(ctx) }()

	// Set k=v.
	setReply := <-tp.Submit(wire.Op{Code: wire.OpSet, Key: []byte("k"), Value: []byte("v")})
	if setReply.Status != wire.StatusOK {
		t.Fatalf("set: status=%d", setReply.Status)
	}

	// Get k.
	getReply := <-tp.Submit(wire.Op{Code: wire.OpGet, Key: []byte("k")})
	if getReply.Status != wire.StatusOK || !bytes.Equal(getReply.Value, []byte("v")) {
		t.Fatalf("get: status=%d value=%q", getReply.Status, getReply.Value)
	}

	// Delete k.
	delReply := <-tp.Submit(wire.Op{Code: wire.OpDelete, Key: []byte("k")})
	if delReply.Status != wire.StatusOK {
		t.Fatalf("delete: status=%d", delReply.Status)
	}

	// Get again → miss.
	missReply := <-tp.Submit(wire.Op{Code: wire.OpGet, Key: []byte("k")})
	if missReply.Status != wire.StatusMiss {
		t.Fatalf("get-after-delete: status=%d", missReply.Status)
	}

	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("replica did not exit within 1s of cancel")
	}
}
