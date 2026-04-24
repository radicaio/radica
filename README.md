# Radica

DST-verified coordination cache for Go.

Built for distributed rate limiting, idempotency keys, and distributed locks
— without running a separate coordination cluster.

## Status

Pre-v0.1. Under active development.

## Design

- **DST-first.** DST (Deterministic Simulation Testing): the core is a pure `(state, event) → (state', effects)` function; all I/O is interfaced. Bugs are reproducible from a seed.
- **Go Tiger-style.** Zero-alloc hot paths, no `interface{}` in core, predictable tail latency. See [docs/GO_TIGER_STYLE.md](docs/GO_TIGER_STYLE.md).
- **GC-free data plane.** KV pairs live in pre-allocated arenas, not the Go heap. The GC never scans them.
- **Viewstamped Replication consensus.** Memory-only replication in v0.1.
