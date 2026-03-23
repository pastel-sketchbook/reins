# Project Instructions

Read and follow all instructions in `.reins/METHODOLOGY.md`. That file is
your core methodology — TDD, Tidy First, commit rules, and quality standards.

## Tech Stack

- Language: Go 1.26+ (see `go.mod` for version)
- CLI framework: `cobra` (command structure, flags, help generation)
- Structured logging: `log/slog` (stdlib only)
- Context: `context.Context` threaded explicitly through all functions
- Build system: [Task](https://taskfile.dev) (see `Taskfile.yml`)
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
  0001_chose-echo-over-gin.md
  0002_slog-over-zerolog.md
  ...
```

Use the `000n_<slug>.md` naming convention. Each record documents the
context, decision, and consequences.

## Verification

- Run `task check:all` before every commit.
- The rule-guard agent is defined in `.reins/agents/rule-guard.md`.
  Invoke it for independent verification — do not self-review.

## Audit

Run `task audit` periodically to check for vulnerable dependencies
(`govulncheck`). Use `task test:race` to detect race conditions.

### Modernization (Go 1.22–1.26)

When modifying existing code, apply Go 1.26 modernization patterns.
Run `go fix ./...` (the Go 1.26 modernizer) to auto-rewrite legacy
patterns. Key patterns to enforce:

- `for i := range n` over C-style loops  *(1.22)*
- `"GET /users/{id}"` ServeMux patterns over manual method checks  *(1.22)*
- `errors.AsType[T]()` over `var target *T; errors.As(err, &target)`  *(1.26)*
- `new(expr)` for pointer-to-literal  *(1.26)*
- `slices.All`, `strings.Lines` over hand-rolled loops  *(1.23–1.24)*
- `tool` directives in `go.mod` over `tools.go` blank-imports  *(1.24)*
- `sync.WaitGroup.Go()` over manual `wg.Add(1)` + `go func()`  *(1.25)*
- `testing.B.Loop()` over `for i := 0; i < b.N; i++`  *(1.24)*

See `rules/specifics/go.md` for the complete rule set (S-GO-20 through S-GO-31).
