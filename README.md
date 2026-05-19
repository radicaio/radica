# Radica

DST-verified coordination cache for Go.

Built for distributed rate limiting, idempotency keys, and distributed
locks — without running a separate coordination cluster.

## Status

Pre-v0.1. Under active development.

## Design

- **DST-first.** The core is a pure `(state, event) → (state', effects)`
  function; all I/O is interfaced. The architecture supports
  Deterministic Simulation Testing — bugs reproducible from a seed.
  The DST harness itself is the next milestone.
- **Go Tiger-style.** Strict assertions, predictable tail latency, no
  `interface{}` in core. See [docs/GO_TIGER_STYLE.md](docs/GO_TIGER_STYLE.md).
- **Two deployment modes.** Embedded as a Go library, or standalone as a
  static binary with no external dependencies.
- **Viewstamped Replication consensus (planned).** Memory-only
  replication in v0.1; persistence and VSR are post-MVP.

See [docs/architecture.md](docs/architecture.md) for the current shape
and what's deferred.
