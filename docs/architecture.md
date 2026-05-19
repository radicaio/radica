# Architecture

> MVP, single-node, no VSR, no persistence. Generic KV store with
> `Get`/`Set`/`Delete`. This document will change.

## Core idea

Radica is an event-loop replica with three injected dependencies: `Clock`,
`Transport`, and `StateMachine`. The same replica code runs in production
(real time, real TCP) and under Deterministic Simulation Testing (virtual
time, in-memory transport). Determinism is the test discipline; it falls out
of keeping I/O behind interfaces and never blocking in core.

## Components

### Clock

Time abstraction. Production impl uses monotonic `time.Now`. Simulator impl
exposes a virtual clock advanced by the DST scheduler. Core code never calls
`time.Now` directly.

### IO

Byte-level network primitive: submit reads/writes on a connection, get
completions back. Production impl uses stdlib `net` with helper goroutines
that block on `Read`/`Write` and post completions to a channel drained by the
replica. Simulator impl is fully in-memory.

The replica never imports `IO` directly. It is the substrate `Transport` is
built on.

### Transport

Framing and per-client routing on top of `IO`. Owns:

- Parsing the fixed-layout binary header.
- Reading the variable payload.
- Tracking which connection a request arrived on, so the reply goes back to
  the same client.
- Sending replies.

The replica calls `Transport.Send(client, msg)` and registers a receive
handler. Connections and frames are invisible to the replica.

### StateMachine

The deterministic, pure core. Contract:

- `Apply(op, now) → reply` — pure function. Same starting state plus same op
  yields the same `(state', reply)` byte-for-byte.
- `Snapshot() → bytes` and `Restore(bytes)` — sketched for future VSR state
  transfer and persistence. MVP impls panic.

What it owns: the user-visible data (the KV map in MVP).

What it does not own: network, framing, time (timestamps are passed in),
client session bookkeeping, allocation policy.

In code the `StateMachine` is a concrete struct, not an interface — no
interface dispatch on the hot path. The **KV** sits behind an interface
inside it as the swap point. For MVP the KV is a plain `map[string][]byte`;
it will be replaced with a real KV engine later without touching the
replica.

### Replica

Owns the event loop and the three dependencies. Exposes `Tick()` (one step
of work) and `Run(ctx)` (the production loop calling `Tick` forever).

### DST scheduler

Drives `Tick()` deterministically. Owns: virtual clock, event priority
queue, seeded PRNG, fault injection (drop, reorder, partition, delay),
workload generator, invariant checker, seed/replay harness.

Lives under `internal/dst/`. Runs as a regular `go test`.

## The event loop

The event loop is the `Replica.Run(ctx)` method. There is no separate "loop"
package or type.

```go
func (r *Replica) Run(ctx context.Context) error {
    for ctx.Err() == nil {
        r.transport.Tick()   // drain network completions, fire handlers
        r.clock.Tick()       // advance timers (real or virtual)
        r.tick()             // do any deferred internal work
    }
    return ctx.Err()
}
```

The loop never blocks. `Transport.Tick` returns immediately after draining
ready completions. Any I/O that would block is delegated to helper
goroutines at the I/O boundary; they post completions back to the loop.

Three drivers call the same `Tick` methods:

```
prod:    main → Replica.Run → (Transport.Tick, Clock.Tick, Replica.tick)
embed:   caller's goroutine → Replica.Run → same
DST:     scheduler.Step → (Transport.Tick, Clock.Tick, Replica.tick)
```

Prod and embedded use `Run`. DST drives `Tick` directly, advancing virtual
time between calls. Same component, same method bodies.

## Request flow

A `Set` request, single-node, end to end:

1. Bytes arrive on a TCP connection. Helper goroutine reads them and posts a
   completion to the loop.
2. `Replica.Run` calls `Transport.Tick`. The completion fires the registered
   receive handler with the parsed `Message`.
3. The handler decodes the op (`Set`, key, value) and calls
   `StateMachine.Apply(op, now)`.
4. `Apply` mutates the KV, returns a reply.
5. The handler calls `Transport.Send(client, reply)`. Transport queues the
   bytes; the writer helper goroutine sends them.
6. Loop continues. Other requests are interleaved freely; the replica never
   blocked.

In DST the helper goroutines do not exist. The scheduler synthesizes
completions directly at deterministic virtual times.

## Wire protocol

### MVP

HTTP, temporarily. The standalone binary exposes:

- `GET /healthz` — liveness probe.
- `GET /kv/{key}` — returns the value, or 404.
- `PUT /kv/{key}` — request body is the value.
- `DELETE /kv/{key}` — 204 or 404.

HTTP is here to make the system testable from `curl` and unit tests
without writing a client first. It is **not** the long-term data plane.

Embedded mode has no wire protocol at all. The Go API exposes
`Get`/`Set`/`Delete` directly; calls submit ops to the in-proc transport
and wait on completion.

### Planned

A custom binary protocol over TCP will replace HTTP on the data path:
fixed 32-byte header (magic, version, opcode, `client_id` uint128,
`request_id` uint64, `key_len`, `value_len`, flags, checksum) followed by
variable payload. Defaults: `max_key_size = 1 KiB`,
`max_value_size = 64 KiB`. HTTP will narrow to a control plane
(`/healthz`, `/metrics`, pprof) on a separate port.

## DST

Runs as standard `go test ./internal/dst/...`. The scheduler advances
virtual time, injects faults via seeded PRNG, generates workload, and
checks invariants after each step.

Failures dump seed + tick + a one-line replay command. Reproducibility from
a seed is a hard requirement.

Smoke tests run on every PR. The full burn-in (1M ticks × 100 seeds) runs
nightly and on release tags.

## Deployment modes

- **Embedded.** `radica.Open(cfg)` returns a `*Radica`. The caller spawns a
  goroutine and calls `r.Run(ctx)` on it. No wire protocol, no
  serialization.
- **Standalone.** `cmd/radicad` runs the same replica behind a TCP
  listener. One static binary, no external dependencies.

Same engine, same code path; only the construction site differs.

## Project layout

```
radica/
├── radica.go                       public embedded-mode API
├── docs/
│   ├── GO_TIGER_STYLE.md
│   └── architecture.md
├── internal/
│   ├── assert/                     radica/assert primitives
│   ├── clock/                      Clock interface + real + sim
│   ├── wire/                       opcodes, Op, Reply
│   ├── transport/                  in-process request queue + reply routing
│   ├── kv/                         KV interface + map impl
│   ├── sm/                         StateMachine
│   ├── replica/                    Replica struct, Tick, Run
│   └── http/                       HTTP listener (/healthz, /kv)
├── cmd/
│   └── radicad/                    standalone binary
└── experiments/                    spike code, not shipped
```

`client/` (binary-protocol Go SDK), `internal/io/`, and `internal/dst/`
are deferred; they will appear when the binary protocol and DST harness
land.

## What's deferred

- VSR consensus and multi-node replication.
- Persistence (memory-only in v0.1).
- TTL, CAS, Incr, and any op beyond `Get`/`Set`/`Delete`.
- Pre-allocation pools and arenas.
- Multi-language clients (Go client only in MVP).
- Use-case convenience helpers (rate-limit, idempotency, lock sugar) — they
  layer on top of the KV API in client SDKs, not in the engine.
- Linux-native async I/O (`io_uring`, raw `epoll`); MVP uses stdlib `net`
  with helper goroutines.
