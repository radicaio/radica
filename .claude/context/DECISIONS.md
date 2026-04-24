# Radica — Decisions Log

> ADR-lite. Append-only. Each entry: date, decision, context, status.
> Reverse chronological (newest first).

---

## 2026-04-22 — DST-driven architecture: two-layer split

**Decision:** Core logic is a pure, deterministic `(state, event) → (state', effects)`
function. Runtime (prod or sim) is swappable behind interfaces (`Clock`, `Network`,
`Disk`, `Random`). Core emits typed effect descriptors the runtime executes.

**Context:** Required for DST. Same binary runs under real I/O in prod and under
simulator in tests. Modelled on TigerBeetle's split (see `RESEARCH/tigerbeetle-map.md`).

**Status:** Accepted. To be implemented starting P0-T1 (DST spike).

---

## 2026-04-22 — Spike ordering: DST → KV → VSR → integration

**Decision:** Run the three planned spikes in this fixed order, plus an implicit
fourth integration spike before the P0 gate.

**Context:** VSR without DST is how people ship subtly broken consensus. KV without
DST has no correctness oracle. Therefore DST must be first. KV is tractable and
gives VSR a real backing store. VSR benefits from both.

**Alternatives considered:** Parallel spikes (rejected — too much context switching
solo). VSR first (rejected — no way to verify it).

**Status:** Accepted.

---

## 2026-04-22 — Blog publishing strategy: hybrid

**Decision:** Spike-phase engineering posts go on `inelpandzic.com` first.
After P0 gate passes, stand up `radica.io` and start cross-posting with
`rel="canonical"` pointing to `radica.io/blog`. From P1 onward, new posts go
to `radica.io/blog` first, cross-posted/summarized on inelpandzic.com.

**Context:** Existing personal-blog audience provides distribution today. Product
domain has zero SEO and audience. Also de-risks P0 bail scenario (don't burn
product domain on abandoned work).

**Alternatives considered:** radica.io/blog from day one (rejected — commits
product domain before feasibility is proven). blog.radica.io subdomain (rejected —
worse SEO than subpath).

**Status:** Accepted.

---

## 2026-04-22 — Consensus protocol: VSR, not Raft

**Decision:** Use View Stamped Replication (VSR) as the consensus protocol for
Radica.

**Context:** VSR has a simpler state machine and cleaner view change semantics than
Raft. TigerBeetle uses VSR successfully. Better fit for DST verification due to
cleaner protocol edges. Reference: Liskov & Cowling, "Viewstamped Replication
Revisited" (2012).

**Alternatives considered:** Raft (more reference material and community mindshare,
but more edge cases in the protocol itself).

**Status:** Accepted. Formal Raft-vs-VSR writeup remains a P0 deliverable
(P0-T3 / P0-4) for public documentation and launch narrative, even though
the decision is already made.

---

## Template for new entries

```
## YYYY-MM-DD — <short title>

**Decision:** <what was decided>

**Context:** <why>

**Alternatives considered:** <what else, why rejected>

**Status:** Accepted | Superseded by <ref> | Reverted
```
