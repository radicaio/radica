# Radica — Project Context

> Always-on context for LLM-assisted sessions. Keep this file short and stable.
> Deeper material lives in `DECISIONS.md`, `PLAN.md`, `RESEARCH/`, and `SESSIONS/`.

## One-line identity

DST-verified, tiger-style distributed coordination cache for Go.

## Lead use case

Distributed rate limiting.

## Secondary use cases

- Idempotency keys
- Distributed locks / leases

## What Radica is NOT (v0.1)

- Not a general cache (Ristretto / Redis win there)
- Not durable — no WAL or disk persistence until v0.2
- Not a Kubernetes datastore, not an etcd replacement
- Not a pub/sub, queue, search, or analytics system
- Not accessible from non-Go languages (no SDKs in v0.1)
- No watch API, no rich txn API, no dynamic membership

## Positioning

An **embeddable Go library** that lets teams *avoid needing* etcd/Redis for
coordination workloads. Competes by making coordination clusters unnecessary
for the workloads it targets, not by attacking incumbents head-on.

## Differentiation (the bet)

1. **DST-first** — deterministic simulation testing as the primary correctness
   strategy. Bugs are reproducible from a seed number, not a stack trace.
2. **Tiger-style Go** — zero-alloc hot paths, no `interface{}` in core,
   predictable tail latency, no GC surprises.

If these two hold up and are *publicly legible*, Radica has a real wedge.

## Budget & scope

- **Solo, 10–14 months** to v0.1.
- **Consensus:** View Stamped Replication (VSR), not Raft.
- **No persistence in v0.1.** Memory-only, replicated.
- Full MVP plan: see `PLAN.md`.

## Cross-cutting practices (from Phase 1 onward)

- Every public function has pre/post/invariant assertions
- Every PR includes a DST run at its scope level
- Nothing merges without seed/replay reproducibility
- Tiger-style lint checks in CI (no `interface{}` in hot path, no hidden allocs)

## Audience (who adopts)

- Go shops with correctness-sensitive workloads (fintech, infra, auth, API gateways)
- Platform teams tired of running Redis/etcd clusters for coordination-only workloads
- Correctness-curious engineers — the TigerBeetle / FoundationDB / Antithesis audience

## Publishing / narrative

- Spike-phase engineering posts: hybrid — write on **inelpandzic.com** first,
  cross-post to **radica.io/blog** once radica.io has a real landing page
  (target: after P0 gate passes).
- At launch, radica.io/blog is the primary home, with canonical URLs pointing there.
- The DST-in-Go writeup is the single highest-leverage artifact of the project.

## Key files to always load

When starting a new Claude session about Radica, load:

1. `.claude/context/PROJECT.md` — this file
2. `.claude/context/PLAN.md` — MVP plan v0.1
3. `.claude/context/DECISIONS.md` — decisions log
4. `docs/GO_TIGER_STYLE.md` — coding standard
5. Any relevant `.claude/context/RESEARCH/*.md`
6. Latest `.claude/context/SESSIONS/*.md` for in-flight work
