# Radica v0.1 — MVP Project Plan

**Framing:** DST-verified, tiger-style distributed coordination cache for Go.
Lead use case: rate limiting. Secondary: idempotency keys, distributed locks.
**Budget:** 10–14 months solo. **Consensus:** decided in P0-T4 (current: VSR).

## Phase 0 — Research & Feasibility (2 weeks)

- **P0-1** Scaffold `radica-spikes` repo
- **P0-2** Task 1: DST harness spike (Go) — 5 days
- **P0-3** Task 2: Tiger-style-Go discipline probe (arena or ring buffer) — 2 days
- **P0-4** Task 3: Prior-art teardown + Raft-vs-VSR decision doc — 3 days
- **P0-5** Task 4: Arena GC-free benchmark vs sync.Map / Ristretto — 2 days
- **P0-6** Task 5: Named launch-reviewer list (10–20 people) — 1 day
- **P0-7** Author `v0.1-scope.md` gate doc — 2 days
- **P0-GATE** Phase 0 review; bail if any spike fails pass criteria

## Phase 1 — DST Harness (M1–M2)

- **P1-1** Repo scaffold `radica` (module, layout, license, CI skeleton)
- **P1-2** Logical clock + seeded PRNG with named sub-streams
- **P1-3** Single-goroutine event loop (no `go`, no `select`, no wall time)
- **P1-4** Simulated network (point-to-point bus, drop/delay/reorder/duplicate)
- **P1-5** Simulated disk stub (interface + latency injection)
- **P1-6** Trace recorder + xxhash determinism oracle
- **P1-7** Fault schedule (JSON-scripted + seeded random)
- **P1-8** `radica-sim` CLI: `run`, `replay`, `minimize` (seed shrinking)
- **P1-9** Assertion framework (pre/post/invariant, fail-fast with trace dump)
- **P1-10** Property-test harness + 10M-op smoke run on toy protocol
- **P1-GATE** Bit-exact replay proven; fault reproducibility proven

## Phase 2 — Storage Layer (M3–M4)

- **P2-1** `Storage` interface (persistence-ready shape: `Put`/`Get`/`Delete`/`Scan`/`Snapshot`/`Truncate`)
- **P2-2** `MapStorage` test fake
- **P2-3** Size-classed arena allocator with generation counters
- **P2-4** Packed key+value slot layout (ABA-safe)
- **P2-5** `ArenaStorage` real implementation
- **P2-6** Op-log abstraction (future WAL shape, memory-only now)
- **P2-7** Storage-layer DST suite (fuzz + fault injection)
- **P2-8** `TIGER_STYLE_GO.md` finalized and enforced via lint/CI checks
- **P2-GATE** 0 allocs/op hot path; storage survives 100M DST ops

## Phase 3 — Single-Node KV (M4–M5)

- **P3-1** Public Go API (`New`, `Set`, `Get`, `Delete`, `Close`) — ~5 line quickstart
- **P3-2** TTL support with deterministic expiration under sim clock
- **P3-3** TinyLFU eviction (deterministic variant)
- **P3-4** Metrics hooks (counters only, no runtime deps)
- **P3-5** Single-node DST suite (API-level properties)
- **P3-GATE** Single-node correct under 100M ops; p99 < 1ms steady

## Phase 4 — Replication (M5–M10, biggest block)

- **P4-1** Consensus skeleton (VSR per P0-T3): views, op-numbers, log, state machine
- **P4-2** Leader (primary) election / view change under simulator
- **P4-3** Log replication + commit number
- **P4-4** Snapshot / log truncation
- **P4-5** Static 3-node membership (no dynamic changes)
- **P4-6** Client request routing (primary redirect, retry semantics)
- **P4-7** Linearizable reads (read index or lease-based, DST-verified)
- **P4-8** Replication DST suite: partitions, slow nodes, clock skew, message loss
- **P4-9** ≥1B simulated ops milestone with injected faults, zero assertion failures
- **P4-10** Replay-from-seed CLI proven on a real recovered bug
- **P4-GATE** 3-node cluster survives 1 node failure + 1 partition under DST

## Phase 5 — Hardening & v0.1 Release (M11–M12)

- **P5-1** Minimal HTTP stub (GET/SET/DELETE + health)
- **P5-2** Reference 3-node deployment (docker-compose + systemd unit)
- **P5-3** Rate-limiter reference integration (token bucket on Radica)
- **P5-4** Idempotency-key reference integration
- **P5-5** Distributed-lock reference integration
- **P5-6** Benchmark suite (throughput, p99/p99.9/p99.99, under partition)
- **P5-7** Public docs site (concepts, quickstart, DST design, tiger-style rules)
- **P5-8** Design doc: "How Radica uses DST" (blog post / paper-ish)
- **P5-9** Launch announcement to the P0-T5 reviewer list
- **P5-10** Tag `v0.1.0`

## Buffer (M13–M14)

- **P6-1** Slippage absorption
- **P6-2** External review cycle (2–3 engineers from reviewer list)
- **P6-3** Post-launch bug triage + v0.1.x patches

## v0.2+ Backlog (explicitly deferred)

Persistence (WAL + disk) · dynamic membership · observability (OpenTelemetry)
· gRPC · semantic lookup · vector index · auth/TLS · non-Go SDKs · WAN replication
· admin UI · watch API · rich txn API

## Spike strategy (pre-P0-GATE)

Three primary spikes, executed in order:

1. **DST harness + simulator showcase** (the highest-risk, highest-leverage piece)
2. **GC-free concurrent KV in-memory store** (arenas, tiger-style discipline)
3. **View Stamped Replication** (consensus)

Plus an implicit **fourth integration spike**: wire all three together on a toy
KV before declaring P0 done (this is what P0-GATE is really gating on).

Each spike produces a blog post. See `SESSIONS/` for detailed spike plans.
