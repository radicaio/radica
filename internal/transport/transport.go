// Package transport routes requests from the outside world to the
// replica loop and routes replies back to the originating caller.
//
// MVP ships one implementation: an in-process queue. Producers (the HTTP
// listener, the embedded API) call Submit and wait on the returned
// channel. The replica loop calls Tick to drain pending requests and
// Send to deliver replies.
//
// A real network transport will replace or join this one later. The
// shape (Submit / Tick / Send / SetHandler) stays the same.
package transport

import (
	"sync"

	"github.com/radicaio/radica/internal/wire"
)

// RequestID identifies one request in flight. Replies are routed back
// to the same RequestID. One Submit allocates one fresh RequestID; IDs
// are never reused.
type RequestID uint64

// Request is one inbound op plus the request identifier it was assigned.
type Request struct {
	ID RequestID
	Op wire.Op
}

// Reply is one reply plus the request it answers.
type Reply struct {
	RequestID RequestID
	Reply     wire.Reply
}

// Handler is the replica-side receive callback. The transport calls it
// for each pending Request, on the loop goroutine, during Tick.
type Handler func(Request)

// Transport is the in-process request/reply queue.
//
// Submit is safe to call from many goroutines. Tick and Send run on the
// replica loop goroutine.
type Transport struct {
	handler Handler

	mu      sync.Mutex
	pending []request
	replies map[RequestID]chan wire.Reply
	nextID  RequestID
}

type request struct {
	id RequestID
	op wire.Op
}

// New returns an empty Transport.
func New() *Transport {
	return &Transport{
		replies: make(map[RequestID]chan wire.Reply),
	}
}

// SetHandler registers the replica's receive callback.
func (t *Transport) SetHandler(h Handler) { t.handler = h }

// Submit enqueues an op and returns a channel that will receive the
// reply. Safe to call from any goroutine.
func (t *Transport) Submit(op wire.Op) <-chan wire.Reply {
	ch := make(chan wire.Reply, 1)
	t.mu.Lock()
	t.nextID++
	id := t.nextID
	t.replies[id] = ch
	t.pending = append(t.pending, request{id: id, op: op})
	t.mu.Unlock()
	return ch
}

// Send delivers a reply to the waiting caller. Called by the replica on
// the loop goroutine.
func (t *Transport) Send(out Reply) {
	t.mu.Lock()
	ch, ok := t.replies[out.RequestID]
	if ok {
		delete(t.replies, out.RequestID)
	}
	t.mu.Unlock()
	if ok {
		ch <- out.Reply
	}
}

// Tick drains the request queue and fires the handler for each pending
// request. Runs on the loop goroutine. Returns without blocking.
func (t *Transport) Tick() {
	t.mu.Lock()
	batch := t.pending
	t.pending = nil
	t.mu.Unlock()

	for _, r := range batch {
		if t.handler != nil {
			t.handler(Request{ID: r.id, Op: r.op})
		}
	}
}
