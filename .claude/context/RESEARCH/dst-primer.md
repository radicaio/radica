# Deterministic Simulation Testing (DST) — Primer

> Research notes for the Radica DST harness spike (P0-T1 / Phase 1).

## What DST is

DST runs the system inside a **simulated universe** the test controls completely:

- **Single-threaded event loop** — no real goroutines, no OS scheduler
- **Seeded PRNG** drives every non-deterministic choice
- **Simulated clock, network, disk** — no wall-clock, no real I/O
- **Same seed → bit-exact identical execution, every time**

Consequences:

1. Bugs are reproducible. Crash reports are a seed number, not a stack trace.
2. Millions/billions of simulated "hours" per CPU-hour (no real I/O).
3. Seed minimization / shrinking reduces a 10M-op failing trace to a tiny repro.
4. Fault injection (drop, reorder, duplicate, crash, skew) is free and deterministic.
5. Catches concurrency and partition bugs that are statistically impossible to
   hit in normal testing.

Tradeoff: the **entire system must be written to be deterministic-under-simulation**.
This is architectural, not a testing bolt-on.

## Core disciplines

1. All I/O goes through interfaces with a sim impl and a real impl.
2. No wall-clock time in logic — only logical/simulated clocks.
3. No direct goroutines in core — a deterministic scheduler picks next event.
4. PRNG with named sub-streams (network, fault, time — independent streams).
5. Determinism oracle: hash the event trace; same seed → same hash, always.
6. No unordered map iteration in logic that influences output.

## Go-specific hazards

Go has DST-hostile defaults. Code review / lint rules to enforce:

- **`map` iteration order is randomized** → use sorted iteration or ordered maps
- **`select` with multiple ready channels is randomized** → avoid `select` in core
- **Goroutines and the scheduler** are non-deterministic → single-threaded core
- **`time.Now()`, `time.After`, `time.Timer`** → banned in core; use `Clock` interface
- **`sync.Map`, `sync.Pool`** → internal non-determinism; avoid in core
- **`runtime.GC`, finalizers, goroutine IDs** → all non-deterministic
- **`math/rand` global state** → always use injected `*rand.Rand`
- **Network and file I/O** → must go through interfaces

These become the first entries in `TIGER_STYLE_GO.md` (P2-8).

## Two-layer architecture

```
                  ┌──────────────────────────────────────────────┐
                  │             CORE (deterministic)             │
                  │                                              │
                  │   ┌────────────┐      ┌──────────────────┐   │
   events in ───▶ │   │  Node FSM  │◀────▶│  KV / VSR / etc. │   │
                  │   └────────────┘      └──────────────────┘   │
                  │         │                                    │
                  │         ▼  effects out                       │
                  └─────────┼────────────────────────────────────┘
                            │ (Send, ScheduleTimer, DiskWrite, …)
                            │
         ┌──────────────────┴───────────────────┐
         │                                      │
         ▼                                      ▼
┌────────────────────┐                ┌──────────────────────┐
│   PROD RUNTIME     │                │    SIM RUNTIME       │
│                    │                │                      │
│ • net.Conn         │                │ • in-mem msg bus     │
│ • os.File / fsync  │                │ • in-mem disk        │
│ • time.Now         │                │ • logical clock      │
│ • goroutines       │                │ • seeded PRNG        │
│                    │                │ • fault scheduler    │
└────────────────────┘                │ • trace recorder     │
                                      │ • determinism oracle │
                                      └──────────┬───────────┘
                                                 │
                                                 ▼
                                       ┌──────────────────┐
                                       │  radica-sim CLI  │
                                       │  run/replay/min  │
                                       └──────────────────┘
```

## Simulated run loop

```
  seed ──▶ PRNG ──▶ fault schedule + message ordering + timing
                         │
                         ▼
            ┌────────────────────────────┐
            │   single-threaded loop     │
            │                            │
            │   while events remain:     │
            │     e = pick_next(prng)    │
            │     deliver(e) ──▶ core    │
            │     collect(effects)       │
            │     record(trace)          │
            │     assert(invariants)     │
            └────────────┬───────────────┘
                         │
                         ▼
                   hash(trace) ──▶ oracle
                         │
           same seed ────┴──▶ same hash, always
```

## Layer contract table

| Layer | Contains | Never contains |
|---|---|---|
| **Core** | FSM, KV, consensus, business logic | `time.Now`, `go`, real I/O, `sync.Map`, global `rand` |
| **Interfaces** | `Clock`, `Network`, `Disk`, `Random` | implementations |
| **Prod runtime** | real impls of the interfaces | test assertions, determinism hooks |
| **Sim runtime** | fake impls + scheduler + fault injector + trace recorder | any real I/O |
| **Oracle** | invariant checks, trace hashing, seed replay | logic under test |

## Mental model (one line)

> *The core is a pure function `(state, event) → (state', effects)`.
> The simulator is a loop that feeds it events and lies about the world.*

## Reading list

### Tier 1 — must-read

- **Will Wilson, "Testing Distributed Systems w/ Deterministic Simulation"**
  (Strange Loop 2014) — https://www.youtube.com/watch?v=4fFDFbi3toc
  **Watch this first.** ~45 min. Changes how you think.
- **TigerBeetle VOPR docs** — https://docs.tigerbeetle.com/about/vopr/
  (Also note: TigerBeetle uses VSR, same as Radica.)
- **TigerBeetle source** — `src/vopr.zig`, `src/testing/*` — a working reference.
  Local copy used during research: `/Users/inelpandzic/Dev/Workspace/IP/tigerbeetle`
- **Antithesis blog** — https://antithesis.com/blog/ — especially
  https://antithesis.com/blog/is_something_bugging_you/

### Tier 2 — high value

- **FoundationDB architecture** — https://apple.github.io/foundationdb/
- **Flow (the actor language FoundationDB built for DST)** —
  https://github.com/apple/foundationdb/blob/main/flow/README.md
- **Resonate HQ** (Go-native DST project) —
  https://github.com/resonatehq/resonate
- **Jepsen** (not DST, but complementary adversarial-testing mindset) —
  https://jepsen.io/ and https://aphyr.com/tags/jepsen

### Tier 3 — background / theory

- **"Why Is Random Testing Effective for Partition Tolerance Bugs?"**
  Majumdar & Niksic, POPL 2018 — https://dl.acm.org/doi/10.1145/3158134
- **VSR paper** — Liskov & Cowling, "Viewstamped Replication Revisited" (2012)
  https://pmg.csail.mit.edu/papers/vr-revisited.pdf
- **Martin Kleppmann**, *Designing Data-Intensive Applications*, ch. 8–9

## Suggested research cadence (1–2 weeks, read-only)

| Day | Activity |
|---|---|
| 1 | Will Wilson's Strange Loop talk. Watch twice if needed. |
| 2 | TigerBeetle VOPR docs + skim `src/testing/` and `src/vsr/` |
| 3 | Antithesis blog — everything tagged simulation/determinism |
| 4 | FoundationDB architecture + Flow README |
| 5 | Resonate's simulator code |
| 6 | VSR paper (deep read — overlaps with VSR spike prep) |
| 7 | Majumdar & Niksic POPL paper |
| 8–10 | Sketch Radica's harness architecture as a design doc (NOT code) |
| 11–14 | First-draft outline of the DST-in-Go blog post |
