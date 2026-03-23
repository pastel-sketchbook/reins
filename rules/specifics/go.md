# Go — Specific Rules (reins CLI)

These rules apply when modifying `.go` files in the reins repository.
Loaded via `rules/INDEX.yaml` trigger: `**/*.go`.

**Minimum version: Go 1.26+.** The `go.mod` directive is the single
source of truth. All code MUST compile and pass `go vet` with the
declared toolchain version.

---

## Language

- **S-GO-01** — MUST use the Go toolchain version declared in `go.mod`.
- **S-GO-02** — MUST wrap errors with `%w` and check with `errors.Is` / `errors.As`. Prefer `errors.AsType[T]()` (Go 1.26) over `errors.As` when the target is a concrete type.
- **S-GO-03** — MUST prefer small, well-named helper functions over inline repetition.
- **S-GO-04** — MUST favor standard-library helpers over reimplementing common behavior.
- **S-GO-05** — MUST format all code with `gofmt`.
- **S-GO-06** — MUST pass `go vet` and `staticcheck` with zero warnings before committing.
- **S-GO-07** — MUST NOT introduce external dependencies. The reins CLI uses stdlib only.

## Modernization (Go 1.22–1.26)

These rules enforce idiomatic patterns available in the minimum Go
version. Run `go fix ./...` (Go 1.26 modernizer) to auto-apply many
of these.

- **S-GO-20** — MUST run `go fix ./...` as part of the quality gate. The
  Go 1.26 modernizer rewrites legacy patterns automatically.
- **S-GO-21** — MUST use range-over-integer (`for i := range n`) instead
  of C-style `for i := 0; i < n; i++` when the loop body does not
  modify the index variable.  *(Go 1.22)*
- **S-GO-22** — MUST use `errors.AsType[T]()` for type-safe error
  unwrapping instead of declaring a variable and calling `errors.As`.
  *(Go 1.26)*
- **S-GO-23** — MUST use `new(expr)` to create a pointer to a computed
  or literal value (e.g. `new(42)`, `new("default")`) instead of
  helper functions or address-of-local patterns.  *(Go 1.26)*
- **S-GO-24** — MUST use iterator-based helpers from `slices`, `maps`,
  and `strings` packages (`slices.All`, `slices.Collect`,
  `strings.Lines`, `strings.SplitSeq`, etc.) instead of
  hand-rolled loops where they improve clarity.  *(Go 1.23–1.24)*
- **S-GO-25** — MUST use `tool` directives in `go.mod` for executable
  dependencies instead of the `tools.go` blank-import pattern.
  *(Go 1.24)*
- **S-GO-26** — MUST use `testing.B.Loop()` for benchmarks instead of
  `for i := 0; i < b.N; i++` or `for range b.N`.  *(Go 1.24)*
- **S-GO-27** — MUST NOT use deprecated stdlib APIs when a modern
  replacement exists. Examples:
  - `math/rand` → `math/rand/v2`  *(Go 1.22)*
  - `runtime.SetFinalizer` → `runtime.AddCleanup`  *(Go 1.24)*
  - `io/ioutil` → `io` / `os` equivalents  *(deprecated since Go 1.16)*

## Testing

- **S-GO-08** — MUST use `t.Chdir(t.TempDir())` for tests that interact with the filesystem.
- **S-GO-09** — MUST use `t.Cleanup` (not `defer`) for test teardown that must survive subtests.

## Verification

- **S-GO-10** — MUST run `task check:all` (format, fix, vet, test) before every commit.
