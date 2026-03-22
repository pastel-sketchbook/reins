# Go — Specific Rules

These rules apply when modifying `.go` files in this project.
Loaded via `rules/INDEX.yaml` trigger: `**/*.go`.

---

## Language

- **S-GO-01** — MUST use the Go toolchain version declared in `go.mod`.
- **S-GO-02** — MUST use generics for reusable algorithms and types where they make code clearer.
- **S-GO-03** — MUST wrap errors with `%w` and check with `errors.Is` / `errors.As`. Use `errors.Join` when composing multiple errors.
- **S-GO-04** — MUST pass `context.Context` explicitly for cancellation and timeouts. Never store it in a struct.
- **S-GO-05** — MUST prefer small, well-named helper functions over inline repetition. Reuse existing `min`/`max` helpers or standard-library equivalents (`math.Min`/`math.Max` for floats) before adding new ones.
- **S-GO-06** — MUST use the appropriate loop form: `for i := 0; i < n; i++` when the index is needed; `for _, v := range collection` when iterating elements.
- **S-GO-07** — MUST favor standard-library helpers (`time.Now().UnixMilli()`, `net/url.URL.Redacted()`, etc.) over reimplementing common behavior.
- **S-GO-08** — MUST format all code with `gofmt` and organize imports with `goimports`.
- **S-GO-09** — MUST pass `golangci-lint` with zero warnings before committing.

## Verification

- **S-GO-10** — MUST run `task check:all` (format, lint, unit tests) before every commit.
- **S-GO-11** — MUST fix all issues reported by `gofmt`, `goimports`, and `golangci-lint` before committing.
