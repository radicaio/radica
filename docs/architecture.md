# Architecture

> MVP, single-node, no VSR, no persistence. Generic KV store with
> `Get`/`Set`/`Delete` over HTTP. This document will change.

## Core idea

Radica is an event-loop replica with three injected dependencies: `Clock`,
`Transport`, and `StateMachine`. The replica logic is pure with respect to
its inputs, so the same code can run in production (real time, HTTP) and
under Deterministic Simulation Testing (virtual time, scripted workload).
Determinism falls out of keeping side effects behind interfaces and never
blocking in core.

## Components

### Clock

Time abstraction. Production impl reads monotonic `time.Now`. Sim impl
exposes a virtual clock advanced explicitly. Core code never calls
`time.Now` directly.

### Transport

Request queue and reply router.

- `Submit(op) → chan reply` is called by listeners (today: HTTP) and the
  embedded API. It enqueues the op, assigns a `ClientID`, and returns a
  buffered channel the caller waits on.
- `Tick()` is called by the replica loop. It drains the pending queue and
  invokes the registered handler for each request.
- `Send(reply)` is called by the replica after `Apply`. It pushes the
  reply onto the channel the producer is waiting on.

Submit is safe to call from many goroutines. Tick and Send run on the
loop goroutine. Nothing blocks the loop.

### StateMachine

The deterministic, pure core.

- `Apply(op, now) → reply` — pure function. Same starting state plus same
  op yields the same `(state', reply)`.
- `Snapshot()` / `Restore(bytes)` — sketched for future VSR state transfer
  and persistence. MVP impls panic.

Owns the user-visible data. Does not own network, framing, time
(timestamps are passed in), or allocation policy.

In code the `StateMachine` is a concrete struct, not an interface — no
interface dispatch on the hot path. The **KV** sits behind an interface
inside it as the swap point. MVP KV is a plain `map[string][]byte`; it
will be replaced by a real KV engine later without touching the replica.

### Replica

Owns the event loop and the three dependencies. Registers its `onRequest`
method as the transport's handler at construction:

```
replica.New(...) {
    ...
    transport.SetHandler(r.onRequest)
}
```

Exposes `Tick()` (one step of work) and `Run(ctx)` (the production loop).

### HTTP listener

The MVP front door. Translates HTTP requests into `wire.Op` values and
calls `transport.Submit`. Will eventually be joined or replaced by a
binary TCP protocol; for now it is the only way in.

### DST scheduler

TODO

## The event loop

The event loop is the `Replica.Run(ctx)` method. There is no separate
"loop" package.

```go
func (r *Replica) Run(ctx context.Context) error {
    for ctx.Err() == nil {
        r.Tick()
        time.Sleep(r.tickPeriod) // ~1ms, prevents busy-spin when idle
    }
    return ctx.Err()
}

func (r *Replica) Tick() {
    r.transport.Tick()  // drain pending requests, fire handler per request
    r.clock.Tick()      // no-op in prod
}
```

`Tick` does bounded work and returns. `transport.Tick` drains whatever is
in the queue at the moment the mutex is taken; requests submitted after
that land in the next batch. `transport.Send` writes to a buffered
channel (cap=1), so it never blocks.

Two drivers will exist when DST is built:

- **Production / embedded**: `Run(ctx)` calls `Tick` on a real-time loop.
- **DST**: the scheduler calls `Tick` directly, advancing virtual time
  between calls. `Run` is not used.

## Request flow

A `PUT /kv/foo` with body `bar`, end to end:

1. `net/http` dispatches the request to the HTTP listener's `kv` handler.
2. The handler builds `wire.Op{Set, "foo", "bar"}` and calls
   `transport.Submit(op)`. Submit enqueues it, assigns a `ClientID`,
   registers a reply channel, and returns the channel.
3. The HTTP goroutine parks on `<-ch`.
4. On its next `Tick`, the replica calls `transport.Tick`, which fires
   the handler (`replica.onRequest`) with the parsed request.
5. `onRequest` calls `sm.Apply(op, clock.Now())`. `Apply` calls
   `kv.Set("foo", "bar")` and returns `Reply{StatusOK}`.
6. `onRequest` calls `transport.Send(reply)`. Send looks up the channel
   by `ClientID`, removes the routing entry, and pushes the reply.
7. The HTTP goroutine wakes, writes `204 No Content`, and returns.

Many requests can be in flight at once; each has its own `ClientID` and
its own reply channel.

## What's deterministic, what's not

Deterministic today:
- `sm.Apply` is pure.
- `kv.Map` is deterministic for a fixed op sequence.
- `transport` shuffles Go values; no encoding, no network.
- `Replica.Tick` runs the same body in any caller.

Non-deterministic at the edges (production only):
- Wall clock (`clock.Real`).
- Arrival order and batching of `Submit` calls (depends on goroutine
  scheduling).
- Real-time sleeps between Ticks.

DST will replace each of those with deterministic equivalents
(`clock.Sim`, a scripted workload, scheduler-driven Ticks) without
changing the replica.

## Wire protocol

### MVP

HTTP. The standalone binary exposes:

- `GET /healthz` — liveness probe.
- `GET /kv/{key}` — returns the value, or 404.
- `PUT /kv/{key}` — request body is the value, 204 on success.
- `DELETE /kv/{key}` — 204 or 404.

Embedded mode has no wire protocol. The Go API calls `transport.Submit`
directly.

### Planned

A custom binary protocol over TCP will join HTTP on a separate port:
fixed-layout header (magic, version, opcode, `client_id`, `request_id`,
`key_len`, `value_len`, flags, checksum) plus variable payload. HTTP will
narrow toward control-plane use (`/healthz`, `/metrics`, pprof).

## Deployment modes

- **Embedded.** `radica.Open()` returns a `*Radica`. The host application
  calls `r.Run(ctx)` on a goroutine it owns and uses `r.Get/Set/Delete`.
  No wire protocol.
- **Standalone.** `cmd/radicad` runs the same replica behind the HTTP
  listener. One static binary, no external dependencies.

Same engine, same code path; only the construction site differs.

## Project layout

```
radica/
├── radica.go                  public embedded-mode API
├── docs/
│   ├── GO_TIGER_STYLE.md
│   └── architecture.md
├── internal/
│   ├── assert/                radica/assert primitives
│   ├── clock/                 Clock interface + real + sim
│   ├── wire/                  Op, Reply, opcodes
│   ├── transport/             in-process request queue + reply routing
│   ├── kv/                    KV interface + map impl
│   ├── sm/                    StateMachine
│   ├── replica/               Replica struct, Tick, Run
│   └── http/                  HTTP listener (/healthz, /kv)
├── cmd/
│   └── radicad/               standalone binary
└── experiments/               spike code, not shipped
```

`internal/dst/` and a binary-protocol `client/` SDK will appear when
their respective features land.

## What's deferred

- DST scheduler (the harness itself; the architecture already supports it).
- VSR consensus and multi-node replication.
- Persistence (memory-only).
- TTL, CAS, Incr, and any op beyond `Get`/`Set`/`Delete`.
- Pre-allocation pools and arenas.
- Binary protocol and Go client SDK.
- Multi-language clients.
- Use-case convenience helpers (rate-limit, idempotency, lock sugar) —
  these layer on top of the KV API in client SDKs, not in the engine.
