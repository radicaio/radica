# Session Note — 2026-04-22 — dst-hello demo

## Participants

- Inel
- Claude (build mode; wrote code)

## What happened

Built a tiny pedagogical demo of the core/runtime split under
`experiments/dst-hello/` before starting the actual P0 work. Goal was
to feel the shape of DST discipline at the smallest possible scale —
a single-node app that receives an HTTP POST, sleeps, writes to disk.

## Scope

- **Included:** pure `Step(state, event) -> (state', []effects)` core;
  sealed Event and Effect sum types (via `isEvent()` / `isEffect()`
  unexported methods); Clock / Disk / HTTPServer / Random / EventSource
  interfaces; real runtime using `net/http` + `time.AfterFunc` +
  `os.WriteFile`; sim runtime with min-heap scheduler + in-memory
  filesystem + ChaCha8-seeded PRNG with named sub-streams (workload,
  fault, scheduler); trace recorder with FNV-64 hash oracle;
  disk-write fault injection; `dst-hello-sim` CLI with --seed/--ops/
  --fault-prob flags.
- **Not included:** multiple nodes, replication, consensus, packet
  simulator, seed minimization, two-phase run, layered checkers
  beyond the trace oracle, CI.

## Deliverables

```
experiments/dst-hello/
├── go.mod                              (module path under inelpandzic/radica/experiments)
├── README.md                           (how to run both modes + discipline notes)
├── core/
│   ├── core.go                         (State, Event, Effect, Step — ~150 lines)
│   └── core_test.go                    (5 unit tests, no harness)
├── runtime/
│   ├── interfaces.go                   (Clock, Disk, HTTPServer, Random, EventSource)
│   ├── loop.go                         (the shared runtime loop)
│   ├── trace.go                        (Trace + FNV-64 Hash)
│   ├── real.go                         (real impls)
│   └── sim.go                          (sim impls + SimScheduler min-heap)
├── cmd/
│   ├── dst-hello/main.go               (real runtime entrypoint)
│   └── dst-hello-sim/main.go           (sim runtime entrypoint)
└── dst_test.go                         (4 DST oracle tests)
```

## Verification

- `go build ./...` — clean
- `go vet ./...` — clean
- `go test ./...` — 9/9 passing
  - Determinism across 6 seeds (same seed -> same hash + responses + files)
  - Different seeds -> different hashes (PRNG plumbing sanity)
  - Fault injection reproducible across 3 seeds at 30% fault prob
  - Liveness: no pending requests at sim end under 0%, 50%, 90% fault prob
- Manual: 4x sim runs showed identical trace hashes for matched seeds
  (`6f76c70b38d8eb45` x2 clean, `d93818e13e39d92d` x2 with faults)
- Manual: real app smoke-tested with `curl` POST writing to disk

## Lessons from this exercise

1. **The core stays genuinely tiny.** `core.go` is ~170 lines. Everything
   substantial is in the runtime. This is the expected shape and it
   scales — radica's real core will still be small relative to its
   runtime machinery.

2. **Effects need to be real types, not method calls.** Initially
   tempting to give the core a "runtime handle" with methods like
   `runtime.WriteFile(path, bytes)`. Doing so would:
   - Couple core to the runtime interface.
   - Make effect traces hard to record.
   - Hide non-determinism sources (you can't grep for them).

   Explicit typed sum type (StartSleep / WriteFile / RespondHTTP) is
   the right call. Confirmed our prior decision.

3. **Named PRNG sub-streams are cheap insurance.** Using three
   sub-streams (workload, fault, scheduler) each seeded from a master
   ChaCha8 means I can add a new sub-stream later without shifting
   existing outputs. Lifted straight from TigerBeetle.

4. **The real runtime's `select` is fine.** Real runtime merges event
   sources via a channel. A `select` there is not deterministic, but
   that's correct: the real runtime is not trying to be deterministic.
   The discipline is that the CORE never has non-determinism; the
   runtime is allowed the world's mess.

5. **Pure unit tests are almost free.** `core_test.go` has 5 tests, no
   mocks, no harness. You just call `Step` with plain values. One of
   the biggest wins of the split.

6. **The demo refuted nothing.** Every design call we'd made on paper
   (effect model, sub-streams, scheduler next-event policy) survived
   contact with code. Modest update to confidence for the actual spike.

## Not part of Radica v0.1

`experiments/dst-hello/` is explicitly outside v0.1 scope. It is a
learning artifact. Candidates for its future:
- Delete before `v0.1.0` tag.
- Keep as an appendix to the DST-in-Go blog post.
- Graduate into the dedicated `radica-spikes` repo when P0-1 creates it.

## Next session candidates

Still the same list as the previous session's, with one revision:

1. **DST harness design doc (P0-T1 prep)** — still the highest-leverage
   next move. We now have a concrete reference (dst-hello) to point to
   in the doc, which simplifies it. *Recommended.*
2. **Scaffold `radica-spikes` repo** (P0-1).
3. **Draft Raft-vs-VSR writeup** (P0-4).
4. **Move dst-hello over to radica-spikes** once that repo exists, then
   delete from the main repo.

## Decisions log

No new entries needed — the demo applies decisions already in
DECISIONS.md, doesn't introduce any. The three parked design questions
(fault atlas, effect model, scheduler policy) should still get formal
DECISIONS entries when the harness design doc is written.

## Open items carried forward

- Parked design questions still not formally logged to DECISIONS.md
  (leanings have been acted on in code but not recorded).
- Go module path, license, CI provider — still undecided. The demo
  dodged these by being under the existing repo with no CI of its own.
