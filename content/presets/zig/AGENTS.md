# Project Instructions

Read and follow all instructions in `.reins/METHODOLOGY.md`. That file is
your core methodology — TDD, Tidy First, commit rules, and quality standards.

## Tech Stack

- Language: Zig 0.15.2+ (see `build.zig.zon` for version)
- Build system: `build.zig` (Zig's native build system)
- Logging: `std.log` (stdlib only)
- Allocators: explicit allocation (arena, GPA, page) — no global allocator
- Task runner: [Task](https://taskfile.dev) (see `Taskfile.yml`)
- Lines of code: [tokei](https://github.com/XAMPPRocky/tokei) (`task loc`)

## Rule System

Rules are loaded on demand based on which files you modify.

- **Principles** (always loaded): `.reins/rules/principles/`
- **Project-specific rules**: `rules/specifics/` and `rules/concerns/` (local)
- **Trigger mapping**: `rules/INDEX.yaml` (local — references both `.reins/`
  and local rule files)

Consult `rules/INDEX.yaml` at the start of every task (see METHODOLOGY.md L-01).

## Architecture Decision Records

Decision records are stored in `docs/rationale/` with sequentially numbered
filenames:

```
docs/rationale/
  0001_gpa-over-arena-for-server.md
  0002_c-interop-strategy.md
  ...
```

Use the `000n_<slug>.md` naming convention. Each record documents the
context, decision, and consequences.

## Verification

- Run `task check:all` before every commit.
- The rule-guard agent is defined in `.reins/agents/rule-guard.md`.
  Invoke it for independent verification — do not self-review.

## Audit

Run `task audit` periodically to check for compiler warnings and
deprecated patterns. Use `zig build test` to run the full test suite.

### Memory Management

Zig requires explicit memory management. Follow these conventions:

- **Allocator discipline**: Pass allocators explicitly to all functions
  that allocate. Choose the right allocator for the scope:
  - `std.heap.ArenaAllocator` for request-scoped / batch work
  - `std.heap.GeneralPurposeAllocator` for long-lived allocations
  - `std.heap.page_allocator` for large, infrequent blocks
- **Cleanup with `defer` / `errdefer`**: Every allocation must have a
  corresponding `defer allocator.free(...)` or `errdefer` for partial
  initialization cleanup. Never rely on manual free at every return point.
- **No memory leaks in error paths**: When building a struct that
  allocates multiple fields, use `errdefer` to free partially
  initialized fields on error.

### Safety

- `@ptrCast` and `@intToPtr` must have a `// SAFETY:` comment.
- `@setRuntimeSafety(false)` requires a performance justification comment.
- Prefer slices over raw pointers. Use pointer casts only for C interop.

See `rules/specifics/zig.md` for the complete rule set.
