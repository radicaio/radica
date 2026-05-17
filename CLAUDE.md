# CLAUDE.md — Entry point for Claude Code sessions

For every answer keep it short, I'll ask for details if I need them.

> This file is always-on context. Keep it short. Deeper material lives under
> `.claude/context/`. Human-facing documentation (e.g. the style guide) lives
> under `docs/`.

## Project: Radica

DST-verified, tiger-style distributed coordination cache for Go.
Lead use case: distributed rate limiting. Secondary: idempotency keys,
distributed locks. Consensus: VSR. 

## Always read these first

4. `docs/GO_TIGER_STYLE.md` — coding standard; all code must conform

## Read when relevant

- `.claude/context/RESEARCH/dst-primer.md` — DST methodology, reading list,
  Go-specific hazards
- `.claude/context/RESEARCH/tigerbeetle-map.md` — TigerBeetle → Radica lessons,
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
- **Append to `.claude/context/DECISIONS.md`** whenever a new non-trivial
  decision is made. Append-only, newest first.
- **End each working session with a new file** under
  `.claude/context/SESSIONS/YYYY-MM-DD-<topic>.md` summarizing what happened
  and what's next.
