# TigerBeetle → Radica: DST Architecture Lessons

> Grounded reference. TigerBeetle is a production DST-first database in Zig that
> uses VSR. Radica's DST harness will be heavily modelled on TigerBeetle's
> approach (adapted to Go idioms).
>
> Local source inspected: `/Users/inelpandzic/Dev/Workspace/IP/tigerbeetle`
> (owned by the user; available for offline reference during P0/P1).

## Proposed Radica layout (inspired by TigerBeetle's `src/testing/`)

```
radica/
├── src/
│   ├── core/               ← pure, deterministic
│   │   ├── vsr/            ← consensus state machine
│   │   ├── kv/             ← arena storage, FSM
│   │   └── clock.go        ← Clock interface only
│   ├── io/
│   │   ├── io_real.go      ← prod: net, disk, time
│   │   └── io_sim.go       ← sim: in-mem everything
│   └── testing/            ← ≈ tigerbeetle/src/testing
│       ├── cluster.go      ← wires N sim replicas together
│       ├── network.go      ← packet simulator (drop/delay/reorder/partition)
│       ├── storage.go      ← in-mem disk w/ latency + fault atlas (P2)
│       ├── time.go         ← TimeSim: ticks, drift, skew
│       ├── state_checker.go← invariant oracle
│       ├── fault_atlas.go  ← "at least one replica can recover" budget (P2)
│       └── fuzz.go         ← seed + PRNG plumbing
└── cmd/
    └── radica-sim/         ← = vopr.zig: run, replay, minimize
```

## Key source files to study (in this order)

| # | File | Lines | Why |
|---|------|-------|-----|
| 1 | `src/testing/time.zig` | 98 | Smallest piece. Full picture of clock sim. Start here. |
| 2 | `src/testing/packet_simulator.zig` | 533 | Core of network sim. Per-link priority queues, filters, clogs. |
| 3 | `src/testing/cluster/state_checker.zig` | 331 | What "an oracle" looks like. Tracks `commit_mins[]` per replica. |
| 4 | `src/testing/cluster/network.zig` | 412 | How packet sim plugs into a cluster. |
| 5 | `src/vopr.zig` | 1809 | CLI/driver tying it all together. Skim only. |
| 6 | `src/testing/storage.zig` | 1224 | Disk sim + fault atlas. Defer to P2 prep. |

Skip for v0.1: `grid_checker.zig`, `manifest_checker.zig`, `journal_checker.zig`
(LSM-tree-specific, not relevant to Radica v0.1).

## Lessons to lift

### 1. Clock is tiny — start here

TigerBeetle's `testing/time.zig` is 98 lines. It's a tick counter + a drift
model (linear / periodic / step / non-ideal). The "simplest piece that proves
the pattern" — good first deliverable for the DST spike.

### 2. Packet simulator is generic over message type

`PacketSimulatorType(Packet)` — it doesn't know VSR, it only knows
"deliver this opaque thing between nodes." This separation lets the simulator
ship independently of consensus. Radica should do the same: make the Go sim
generic (via generics or `any`-with-discipline) over the message type.

### 3. Per-link state, not global network state

Each `(source, target)` pair has its own priority queue, filter, and clog
timer. Partitions are computed into link filters, not tracked as a separate
"network mode." Models reality well (links, not networks, fail) and makes
asymmetric partitions natural.

### 4. Fault atlas pattern

TB only injects faults the protocol is *supposed* to survive. E.g., read/write
faults in WAL are distributed across replicas so at least one copy remains
valid. This keeps the simulator honest about what it's actually testing.

For Radica: defer full atlas to P2 when storage shape exists. In v0.1, start
with a minimal atlas: "never all N replicas faulty on the same WAL region."

### 5. Three exit codes

- `crash (127)` — assertion fired
- `liveness (128)` — no progress for N ticks
- `correctness (129)` — replica divergence / invariant violated

Different bugs need different reactions; shrink/minimize differently.
**Add from day one** — cheap now, painful to retrofit.

### 6. Two-phase run

Each simulation has:
1. **Request phase** (`ticks_max_requests`, with load)
2. **Convergence phase** (`ticks_max_convergence`, no new work, must quiesce)

Many bugs only surface during quiescence (stuck replicas, pending timers,
unfinished view changes). Bake this into the CLI.

### 7. Layered checkers

Storage checker, state checker, manifest checker — separate modules, each
watches one invariant class. When one fires, the layer of the bug is obvious.

For v0.1 Radica, minimum two:
- `state_checker` — linearizability, commit monotonicity, no divergence
- `network_checker` — basic sanity (no message to dead peer, no infinite queue)

### 8. Sub-PRNGs per concern

Time has its own PRNG, network has its own, each replica has its own. Prevents
"fix a bug in fault scheduling, now all network traces change" churn — your
minimal repros stay stable as you fix unrelated things.

### 9. Typed effects (the Go-specific refinement)

TB encodes effects implicitly via MessagePool + IO. In Go, encoding them
**explicitly as a typed sum type** (`CoreEffect` interface + concrete variants
`SendMsg`, `SetTimer`, `DiskWrite`, ...) fits the language better and gives the
type checker something to enforce.

## Mapping Zig idioms → Go equivalents

| Zig (TigerBeetle) | Go (Radica) |
|---|---|
| `comptime StateMachineType: anytype` | Go generics: `[SM StateMachine]` |
| `*anyopaque + vtable` | interface type |
| `std.PriorityQueue` | `container/heap` |
| `stdx.PRNG` (seeded) | `math/rand/v2.Rand` with `ChaCha8` source |
| `assert(cond)` (comptime-disabled in release) | custom `assert` package, build-tag-gated |
| `std.Instant`, `Duration` | custom `Tick`, `TickDelta` types (don't reuse `time.Time`) |
| `MessagePool` arenas | arena allocator (P2 work) |
| `union(enum)` | sealed interface + type switch |

## Adjustments to the Radica MVP plan

Small but real refinements, rooted in TB observations:

- **P0-T1 (DST spike, 5 days):** Implement clock + single-goroutine scheduler
  + generic packet bus + one toy replicated counter + seed replay.
  **Don't try to do storage sim in the spike.** TB's storage sim alone is
  1200+ lines — that's P2 territory.
- **P1-4 (Simulated network):** Model on `packet_simulator.zig`. Per-link
  priority queues, filters for partitions, clog timers. Generic over message type.
- **P1-5 (Simulated disk):** Model on `testing/storage.zig`. Defer the fault
  atlas to Phase 2 when WAL/storage shape exists.
- **P1-9 (Assertion framework):** Split into **layered checkers** like TB does.
  At minimum: `state_checker` and `network_checker` by end of P1.
- **P1-8 (radica-sim CLI):** Add the **two-phase run** pattern and **three
  exit codes** from day one.
- **New micro-item for P1:** a typed `CoreEffect` sum type — core emits typed
  effects, runtime executes them. Go equivalent of TB's implicit effect model.

## Open v0.1 design questions

1. **Fault atlas scope:** (a) skip, only benign faults; (b) minimal atlas
   (never all N replicas faulty on same region); (c) full TB-style atlas.
   Leaning (b) — cheap insurance.
2. **Effect model:** explicit typed sum type vs implicit via interface methods.
   Leaning explicit (better for Go type system, easier to serialize for traces).
3. **Scheduler next-event policy:** strict `ready_at` ordering vs PRNG-weighted
   among ready events. TB uses per-link queues keyed by `ready_at`. Leaning
   strict-ordering + PRNG for ties.
