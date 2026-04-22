# Session Note — 2026-04-22 — DST Kickoff & Project Context Bootstrap

> Distilled notes from the first working session on Radica.
> Purpose: lets a future Claude session (or you) pick up where we left off
> without re-reading a long chat.

## Participants

- Inel (solo founder/engineer on Radica)
- Claude (assistant; mostly read-only plan-mode during this session)

## What happened, in order

1. **Shared the v0.1 MVP plan.** The full plan is now persisted at
   `docs/context/PLAN.md`.

2. **Discussed value proposition.**
   - Lead use case (distributed rate limiting) is well-chosen.
   - Secondary use cases (idempotency keys, distributed locks) are solid.
   - Not a general cache. Not a k8s datastore.
   - The DST story + tiger-style-Go discipline are the real differentiators.
   - Community value (DST-in-Go as a published discipline) likely exceeds
     user value at v0.1. The DST-in-Go blog post is the highest-leverage
     artifact of the whole project.

3. **Discussed competition with etcd.**
   - v0.1 Radica is an **etcd-avoider**, not an etcd competitor. Correct
     framing: "you don't need to run a coordination cluster."
   - Head-on etcd-in-k8s competition is unwinnable and off-limits.
   - Medium-term (v0.3–v0.4 with watch + persistence + dynamic membership),
     Radica could genuinely take etcd's non-Kubernetes coordination footprint.

4. **Confirmed spike plan.** Three spikes, blog each:
   - DST-driven app design + simulator showcase
   - GC-free concurrent KV in-memory store
   - View Stamped Replication
   - **Plus an implicit fourth integration spike** before P0 gate.
   - **Order is fixed: DST → KV → VSR → integration.**

5. **Decided publishing strategy: hybrid.**
   - Spike posts → `inelpandzic.com` first.
   - After P0 gate passes: stand up `radica.io`, cross-post with canonical
     URLs pointing to `radica.io/blog`.
   - From P1 onward: radica.io/blog first.
   - Subpath (`/blog`), not subdomain, for SEO.

6. **Consensus choice confirmed: VSR, not Raft.**
   - Simpler state machine, cleaner view changes, better for DST.
   - TigerBeetle uses VSR successfully.
   - Raft-vs-VSR writeup remains a P0 deliverable for launch narrative.

7. **Did DST research & architecture overview.**
   - Full primer persisted at `docs/context/RESEARCH/dst-primer.md`.
   - Reading list included.
   - Go-specific hazards catalogued.

8. **Inspected TigerBeetle source as reference.**
   - Local copy at `/Users/inelpandzic/Dev/Workspace/IP/tigerbeetle`.
   - Full architecture lessons persisted at
     `docs/context/RESEARCH/tigerbeetle-map.md`, including:
     - Layout to steal
     - Key source files to study (in order)
     - Lessons (per-link state, fault atlas, three exit codes, two-phase run,
       layered checkers, sub-PRNGs, typed effects)
     - Zig → Go idiom mapping
     - Refinements to the MVP plan

9. **Bootstrapped `docs/context/` structure.** This note + PROJECT.md,
   PLAN.md, DECISIONS.md, RESEARCH/dst-primer.md, RESEARCH/tigerbeetle-map.md.

## Where we are in the plan

```
Phase 0 — Research & Feasibility
├── P0-1  radica-spikes repo scaffold         ← NOT STARTED
├── P0-2  DST harness spike (5 days)          ← RESEARCH DONE, architecture sketched
├── P0-3  Tiger-style-Go discipline probe     ← NOT STARTED
├── P0-4  Raft-vs-VSR decision doc            ← DECISION MADE (VSR), writeup pending
├── P0-5  Arena GC-free benchmark             ← NOT STARTED
├── P0-6  Launch-reviewer list                ← NOT STARTED
└── P0-7  v0.1-scope.md gate doc              ← NOT STARTED
```

## Immediate next steps (candidates)

Pick one to kick off the next session:

1. **Start the DST spike for real.** Execute the 5-day plan from
   `RESEARCH/dst-primer.md`:
   - Day 1: Watch Will Wilson's Strange Loop talk.
   - Day 2: Read TigerBeetle's `src/testing/time.zig` and `packet_simulator.zig`.
   - Day 3: Read Antithesis blog posts.
   - Day 4: FoundationDB architecture.
   - Day 5+: Sketch Radica's harness design doc (`docs/design/dst-harness.md`).
2. **Scaffold `radica-spikes` repo.** Separate from `radica` main repo per the
   MVP plan (P0-1). Go module skeleton, lint config, CI starter.
3. **Draft Raft-vs-VSR writeup** (P0-4) — decision is already made but the
   public writeup is still owed. Good candidate for first blog post since it
   requires less engineering.
4. **Draft DST harness design doc** — before any spike code is written, commit
   to: interface shapes (`Clock`, `Network`, `Disk`, `Random`), typed
   `CoreEffect` sum type, scheduler algorithm, oracle contract, trace format.

## Open design questions parked for the spike

(Also in `RESEARCH/tigerbeetle-map.md`.)

1. **Fault atlas scope in v0.1.** Current lean: minimal atlas ("never all N
   replicas faulty on same region").
2. **Effect model.** Explicit typed sum type vs implicit via interface methods.
   Current lean: explicit sum type.
3. **Scheduler next-event policy.** Current lean: strict `ready_at` ordering,
   PRNG for ties.

## Things explicitly not decided yet

- Name of the Go module (presumably `github.com/inelpandzic/radica`, TBD).
- License (Apache 2.0, MIT, BSL — TBD).
- CI provider (GitHub Actions assumed, not confirmed).
- Blog engine for radica.io (TBD — decide at P0 gate, not before).

## How to resume this context in a future session

Paste into Claude at session start:

> I'm working on Radica. Read `docs/context/PROJECT.md`, `docs/context/PLAN.md`,
> `docs/context/DECISIONS.md`, then the most recent file in
> `docs/context/SESSIONS/`. If relevant to my question, also read files under
> `docs/context/RESEARCH/`.

Or, if using Claude Code with this repo as workspace, a future `CLAUDE.md` at
the repo root (not yet created) will auto-load the right entry points.
