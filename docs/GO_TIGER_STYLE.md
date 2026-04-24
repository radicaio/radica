# GoTigerStyle

> Based on TigerBeetle's
> [TIGER_STYLE.md](https://github.com/tigerbeetle/tigerbeetle/blob/main/docs/TIGER_STYLE.md).
> Adapted for Go and the needs of Radica: DST-verified, tiger-style distributed
> coordination. Where TigerStyle and idiomatic Go disagree, we pick the choice
> that best serves safety, performance, and developer experience — in that
> order.

## Why Have Style

Style is design. Design is how it works. Our goals, ranked:

1. **Safety** — correctness under adversarial conditions (faults, concurrency,
   partial failure).
2. **Performance** — we design for it up front; it is not a later concern.
3. **Developer experience** — the code must be a pleasure to read, review, and
   change.

Simplicity is not a compromise between these — it is how we achieve all three
at once. Simplicity is the hardest revision, not the first attempt.

## Zero Technical Debt

Problems are cheapest to fix in design, more expensive in implementation, and
most expensive in production. We do it right the first time. If a showstopper
surfaces — an unbounded allocation, an O(n²) on the hot path, a race — we stop
and solve it. We don't ship and schedule a follow-up.

---

## Safety

### The Power of Ten, Go Edition

1. **Simple control flow.** No recursion. No `goto`. No clever closures on the
   hot path. If bounded termination is not obvious at a glance, it is a bug.
2. **Every loop has a fixed upper bound.** Static if possible, asserted
   otherwise. Event loops that must run forever assert that they cannot exit.
3. **No heap allocation on the hot path after startup.** Allocation belongs to
   initialization and the control plane. The data plane reuses pre-allocated
   storage: sized slices, `sync.Pool`, per-worker arenas.
4. **Function length: soft ≤70 lines, hard ≤100.** Go's error-return pattern
   inflates LOC; 100 is the ceiling. Past 70, justify it.
5. **Assertion density ≥ 2 per function.** Preconditions, postconditions,
   invariants. See [Assertions](#assertions).
6. **Smallest possible scope for every variable.** Declare at use site. Prefer
   `if x, err := f(); err != nil { … }`.
7. **Every return value is checked. Every parameter is validated.** No `_ =`
   on an error. A function that ignores an error is a function that lies.
8. **No reflection or codegen on the hot path.** No `encoding/json`, no
   `fmt.Sprintf`, no `reflect`. Hand-rolled encode/decode for wire formats.
9. **Restrict the dangerous parts of Go.** `unsafe`, `cgo`, `any`/`interface{}`
   in hot paths, `init()` side effects, and goroutines spawned inside library
   functions are all PR-gated and cited.
10. **Warnings are errors.** `go vet`, `staticcheck`, `errcheck`, and our
    linter bundle fail CI on any finding.

### Assertions

Go has no built-in `assert`. We provide `radica/assert`:

```go
assert.True(cond, "cache must be non-empty")
assert.Eq(got, want, "sequence number")
assert.Nil(err)
assert.Lt(index, length)
assert.Implies(a, b)
```

Rules:

- **Always on.** Assertion failures are catastrophic-bug indicators. Panic.
  Crash. Do not compile them out in release. Tiger-style treats a wrong
  program as more dangerous than a dead one.
- **Density ≥ 2 per function.** Pre/post/invariant. A function that operates
  blindly on its inputs is broken.
- **Pair assertions.** Assert validity at the write site *and* at the read
  site. Drift between the two is where bugs live.
- **Assert positive and negative space.** Assert what you expect. Also assert
  what must not happen. Tests follow the same rule.
- **Split compound assertions.** `assert.True(a); assert.True(b)` beats
  `assert.True(a && b)`: more precise failure messages, easier to read.
- **Split compound conditions.** Nested `if/else` beats `if a && b && c`. The
  branches make the cases legible; the positive and negative spaces stay
  honest.
- **Compile-time assertions.** Use generics, `const _ = unsafe.Sizeof(T{}) -
  <expected>`, and build-time checks to fail before the program runs.
- **State invariants positively.** `if index < length { … }` beats `if index
  >= length { … }`.
- **Never rely on the fuzzer alone.** A fuzzer proves presence of bugs, not
  absence. Build the mental model first, encode it as assertions, *then* let
  DST hunt the gaps.

### Memory Discipline

Go has a GC. We cannot forbid allocation; we *can* forbid it where it matters.

- **Startup may allocate freely.** The control plane may allocate at the
  cost of lower frequency. The data plane may not allocate at all after warm-up.
- **`sync.Pool` and arenas** absorb churn for request-scoped buffers.
- **`testing.AllocsPerRun` budgets** live next to every `//radica:hot`
  function. CI fails if the number rises.
- **Escape-analysis gates.** Hot-path changes run `go build -gcflags=-m=2`;
  unintended escapes block the PR.
- **Preallocate slices and maps** with capacity hints. `make([]T, 0, n)`
  beats `append` into a zero-cap slice.

### Control Flow

- **No recursion.** Anywhere. Convert to explicit stacks with fixed bounds.
- **Bounded loops.** Every `for` has a statically provable or asserted bound.
- **Push ifs up, push fors down.** Centralize branching in parent functions;
  keep leaf functions linear and pure.
- **Handle every error.** Return it, wrap it with `%w`, or — in truly
  exceptional cases — panic through `radica/assert`. Never discard.
- **Indent error flow.** Happy path left-aligned; error path early-returns.
- **Panic is reserved for assertion failures.** Business errors are values.

### Concurrency

- **Every goroutine has a named owner and a documented exit condition.**
  Unowned goroutines are leaks in waiting.
- **No goroutines spawned inside library functions.** Concurrency is the
  caller's decision.
- **Synchronous APIs by default.** It is trivial to add concurrency at the
  call site; it is hard to remove it.
- **Bounded channels.** Every channel has a capacity rationale. The closer is
  documented at the channel's declaration.
- **`context.Context` is the first parameter of any function that might
  block, do I/O, or call one that does.** It never lives in a struct.
- **Race detector is mandatory in CI.** `go test -race` on every package.

### Determinism (for DST)

DST requires deterministic replay. Violations here are not style — they are
correctness bugs.

- **Injectable clock.** Library code takes a `Clock` interface. No direct
  `time.Now()`, no `time.Sleep` outside wrapped primitives.
- **Injectable RNG.** No bare `math/rand`. Tests and production code both
  receive a seeded source. Keys use `crypto/rand`.
- **No map iteration on serialized paths.** Map order is non-deterministic.
  Sort, or use a slice.
- **No `init()` side effects** beyond fixed pool allocation. No network, no
  disk, no env reads, no globals that depend on wall time.
- **No `os.Getenv` in library code.** Configuration is passed in.

### Types

- **Fixed-width integers at every boundary.** `uint64` on the wire and on
  disk. `int`/`uint` are for local arithmetic only.
- **Named types for `Index`, `Count`, `Size`.** They look alike and convert
  like crazy. Make the compiler help:

  ```go
  type SequenceNumber uint64
  type ReplicaCount   uint32
  type BytesOffset    uint64
  ```

- **No `any`/`interface{}` in hot paths.** Use concrete types or generics.
- **Return-type ladder.** Prefer, in order: `()` → `bool` → `(T, bool)` →
  `(T, error)`. Don't invent errors that cannot happen.
- **Struct field order: largest → smallest.** Minimize padding. Assert
  `unsafe.Sizeof` for wire-format structs.
- **Keep the zero value useful** (Effective Go agrees).

---

## Performance

> "The lack of back-of-the-envelope performance sketches is the root of all
> evil." — Rivacindela Hudsoni

- **Design for performance before implementation.** The 1000× wins come from
  the shape of the system, not the inner loop.
- **Back-of-envelope sketches across four resources × two characteristics:**
  network, disk, memory, CPU × bandwidth, latency. Be "roughly right" first.
- **Optimize the slowest resource first,** weighted by how often it is used.
  A cache miss called a million times can dominate a disk fsync.
- **Separate control plane from data plane.** Control plane is correctness
  and flexibility; data plane is throughput. Assertion-heavy code in both;
  allocation-free code in the latter.
- **Batch.** At every layer that touches a slow resource. Events arrive at
  their pace; the system runs at its own.
- **Mechanical sympathy.**
  - Extract hot loops into standalone functions with primitive arguments
    (no `self`). The compiler optimizes more, the reader sees more.
  - No interface dispatch in hot loops. Concrete types or generics.
  - Preallocate. Reuse. Keep hot data contiguous.
- **Profiling gates.** Performance-sensitive PRs include `benchstat` output.
  Hot-path functions carry an `allocs/op` budget enforced in CI.

---

## Developer Experience

### Naming

Go convention wins on capitalization; TigerStyle semantics win on meaning.

- **`MixedCaps` / `mixedCaps`.** Underscores only in filenames where Go
  tooling requires it.
- **No abbreviations.** `source`/`target`, not `src`/`dst`. `allocator`, not
  `alloc`. Exception: idiomatic short receivers (`c`, `r`, `n`).
- **Units last, in descending significance.** `latencyMsMax`, not
  `maxLatencyMs`. Related variables sort and align naturally.
- **Equal-length related names** when possible. `source` and `target` share
  six letters; `sourceOffset` and `targetOffset` line up.
- **Callee prefixed with caller.** `readSector` and `readSectorCallback` show
  the call graph in the name.
- **Nouns over participles.** `replica.Pipeline` documents better than
  `replica.Preparing`.
- **Initialisms keep case.** `URL`, `HTTP`, `ID`, `VSR`. `ServeHTTP`, not
  `ServeHttp`. `appID`, not `appId`.
- **Receiver names are one or two letters**, consistent across every method
  on a type. Never `me`, `this`, `self`.
- **Infuse names with meaning.** `gpa Allocator` and `arena Allocator` beat
  two `allocator Allocator`s.
- **Package names are short, lowercase, singular, evocative.** No `util`,
  `common`, `misc`, `types`, `helpers`.

### Ordering

Readers scan top-down. Order matters even when semantics don't care.

- `main` first in `main` packages.
- Struct declaration order: **fields → nested types → methods**.
- Important things near the top of the file. When in doubt, sort
  alphabetically — big-endian names make that work.

### Cache Invalidation

- **Don't alias variables.** State out of sync is the root of a bug family.
- **Shrink scope.** Fewer variables in scope, fewer ways to use the wrong one.
- **Calculate near use** to avoid place-of-check-to-place-of-use gaps.
- **In-place init for large structs.** Pass `*T` and fill it; avoid
  construct-then-move.

### Off-by-One

- **`Index`, `Count`, `Size` are distinct named types.** Converting between
  them is a deliberate act, not a copy-paste.
- **Show division intent.** `divExact`, `divFloor`, `divCeil`. Silent
  integer division is a trap.

### Style by the Numbers

- **`gofmt`/`goimports` is mandatory.** Non-negotiable.
- **Tabs for indentation.** Displayed at width 4.
- **100-column hard limit.** Two copies side by side, just like TigerStyle.
- **Always brace `if`/`for`.** Defense against "goto fail" bugs, even when
  Go's syntax makes them rarer.

### Comments and Commits

- **Comments are sentences** — capital letter, period, proper English.
- **Always say why.** Code is the *what*; comments are the *why*. Without
  the rationale, the next reader cannot tell which parts are load-bearing.
- **Say how for tests.** A short header at the top of a test case saves the
  next debugger an hour.
- **Commit messages are documentation.** They live in `git blame` forever. A
  PR description does not. Write commit messages for the reader five years
  from now.

### Errors

- **Wrap with `%w`.** `fmt.Errorf("read sector %d: %w", id, err)`. Check
  with `errors.Is` / `errors.As`.
- **Sentinel errors are typed.** `var ErrClosed = errors.New("closed")`.
- **Error strings: lowercase, no trailing punctuation.** They compose into
  larger log lines.
- **Never `panic` for expected failures.** Panic means a bug in *our* code.

### Testing

- **Useful failure messages.** `got = X; want Y; input = Z`. Assume the
  debugger is a tired stranger.
- **Table-driven where it fits.** Named subtests, so failures localize.
- **DST seed/replay reproducibility** for every PR at its scope level.
  Nothing merges without it.
- **Property and fuzz tests** on protocol code and state machines. They find
  what we would not think to test.

### Dependencies and Tooling

- **Stdlib only.** Exceptions require a `DECISIONS.md` entry explaining the
  dependency, its maintenance posture, and why we cannot do without it.
- **Scripts in Go.** `scripts/*.go` run via `go run`. No bash unless the task
  genuinely is a shell task.
- **One toolbox.** `gofmt`, `go vet`, `staticcheck`, `errcheck`,
  `ineffassign`, `funlen`, `gocyclo`, `gosec`, plus `radicalint` (below).

---

## The `//radica:hot` Marker

Hot-path functions carry a comment marker:

```go
//radica:hot
func applyBatch(…) { … }
```

CI enforces stricter rules on marked functions:

- No `any` / `interface{}`.
- No reflection, no `fmt.Sprintf`, no `encoding/json`.
- No heap escapes (checked via `-gcflags=-m=2`).
- No recursion.
- No unbounded `for`.
- An explicit `testing.AllocsPerRun` budget test must exist.

---

## Enforcement

Every rule maps to a mechanism. If a rule has no mechanism, it is
aspirational; we add one.

| Rule | Mechanism | Blocking? |
|------|-----------|-----------|
| Formatting | `gofmt` / `goimports` | Yes |
| Imports ordering, unused identifiers | `goimports`, `go vet` | Yes |
| Lint baseline | `staticcheck` | Yes |
| Ignored errors | `errcheck` | Yes |
| Dead assignments | `ineffassign` | Yes |
| Function length (≤100) | `funlen` | Yes |
| Cyclomatic complexity | `gocyclo` | Yes |
| Security smells | `gosec` | Yes |
| Data races | `go test -race` | Yes |
| Allocation budgets | `testing.AllocsPerRun` tests | Yes |
| Escape analysis on hot paths | `go build -gcflags=-m=2` + `radicalint` | Yes |
| `//radica:hot` constraints | `radicalint` | Yes |
| Recursion ban | `radicalint` | Yes |
| Unbounded `for` | `radicalint` | Yes |
| `_ =` on errors | `radicalint` | Yes |
| `time.Now` / bare `math/rand` in library code | `radicalint` | Yes |
| `init()` side effects | `radicalint` | Yes |
| DST seed/replay reproducibility | CI harness | Yes |
| Dependency additions | PR review + `DECISIONS.md` entry | Yes |
| Assertion density ≥ 2/fn | Review checklist | No (soft) |
| Commit message quality | Review checklist | No (soft) |
| Naming semantics (units last, etc.) | Review checklist | No (soft) |

---

## The Last Stage

Radica is small on purpose. The rules feel tight because they are — that is
how we ship something we can trust with correctness-critical workloads. Keep
it small, keep it honest, keep it fun.
