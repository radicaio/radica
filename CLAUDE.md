# CLAUDE.md — Entry point for Claude Code sessions

> This file is always-on context. Keep it short. Deeper material lives under
> `docs/context/`.

## Project: Radica

DST-verified, tiger-style distributed coordination cache for Go.
Lead use case: distributed rate limiting. Secondary: idempotency keys,
distributed locks. Consensus: VSR. Budget: 10–14 months solo.

## Always read these first

1. `docs/context/PROJECT.md` — stable project identity, non-goals, positioning
2. `docs/context/PLAN.md` — v0.1 MVP plan (all phases)
3. `docs/context/DECISIONS.md` — decisions log (ADR-lite)
4. The latest file under `docs/context/SESSIONS/` — in-flight work & next steps

## Read when relevant

- `docs/context/RESEARCH/dst-primer.md` — DST methodology, reading list,
  Go-specific hazards
- `docs/context/RESEARCH/tigerbeetle-map.md` — TigerBeetle → Radica lessons,
  Zig→Go idiom map, layered-checker design

## Operating rules

- **Cross-cutting practices** (once coding starts):
  - Every public function has pre/post/invariant assertions.
  - Every PR includes a DST run at its scope level.
  - Nothing merges without seed/replay reproducibility.
  - Tiger-style lint checks in CI (no `interface{}` in hot path, no hidden allocs).
- **Writing style:** concise, factually precise, no cheerleading. Push back
  when the user is wrong.
- **When uncertain, ask.** Especially on architecture direction and scope.
- **Append to `docs/context/DECISIONS.md`** whenever a new non-trivial
  decision is made. Append-only, newest first.
- **End each working session with a new file** under
  `docs/context/SESSIONS/YYYY-MM-DD-<topic>.md` summarizing what happened
  and what's next.
