// Package radica is the embedded-mode entry point.
//
// Open constructs a Radica with the MVP dependencies (real clock,
// in-process transport, map-backed state machine). The caller runs the
// event loop on a goroutine they own:
//
//	r := radica.Open()
//	go r.Run(ctx)
//	val, ok, err := r.Get(ctx, []byte("k"))
//
// This package is intentionally thin. All real work lives in
// internal/replica, internal/sm, internal/transport, internal/kv.
package radica

import (
	"context"
	"errors"

	"github.com/radicaio/radica/internal/clock"
	"github.com/radicaio/radica/internal/kv"
	"github.com/radicaio/radica/internal/replica"
	"github.com/radicaio/radica/internal/sm"
	"github.com/radicaio/radica/internal/transport"
	"github.com/radicaio/radica/internal/wire"
)

// Radica is the embedded coordination cache handle.
type Radica struct {
	replica   *replica.Replica
	transport *transport.Transport
}

// Open constructs a new Radica with MVP defaults.
// The caller must call Run on a goroutine to start the event loop.
func Open() *Radica {
	tp := transport.New()
	rp := replica.New(clock.NewReal(), tp, sm.New(kv.NewMap()))
	return &Radica{replica: rp, transport: tp}
}

// Run drives the event loop. Blocks until ctx is cancelled.
func (r *Radica) Run(ctx context.Context) error {
	return r.replica.Run(ctx)
}

// Transport exposes the in-process transport so the HTTP listener can
// submit ops directly. Returned value's Submit method is safe for
// concurrent use.
func (r *Radica) Transport() *transport.Transport { return r.transport }

// Get returns the value for key. ok is false if the key is absent.
func (r *Radica) Get(ctx context.Context, key []byte) (value []byte, ok bool, err error) {
	reply, err := r.submit(ctx, wire.Op{Code: wire.OpGet, Key: key})
	if err != nil {
		return nil, false, err
	}
	switch reply.Status {
	case wire.StatusOK:
		return reply.Value, true, nil
	case wire.StatusMiss:
		return nil, false, nil
	default:
		return nil, false, errors.New("radica: unexpected status")
	}
}

// Set stores value under key.
func (r *Radica) Set(ctx context.Context, key, value []byte) error {
	reply, err := r.submit(ctx, wire.Op{Code: wire.OpSet, Key: key, Value: value})
	if err != nil {
		return err
	}
	if reply.Status != wire.StatusOK {
		return errors.New("radica: set failed")
	}
	return nil
}

// Delete removes key. existed reports whether the key was present.
func (r *Radica) Delete(ctx context.Context, key []byte) (existed bool, err error) {
	reply, err := r.submit(ctx, wire.Op{Code: wire.OpDelete, Key: key})
	if err != nil {
		return false, err
	}
	switch reply.Status {
	case wire.StatusOK:
		return true, nil
	case wire.StatusMiss:
		return false, nil
	default:
		return false, errors.New("radica: unexpected status")
	}
}

// submit pushes an op through the transport and waits for the reply or
// for ctx to be cancelled.
func (r *Radica) submit(ctx context.Context, op wire.Op) (wire.Reply, error) {
	ch := r.transport.Submit(op)
	select {
	case reply := <-ch:
		return reply, nil
	case <-ctx.Done():
		return wire.Reply{}, ctx.Err()
	}
}
