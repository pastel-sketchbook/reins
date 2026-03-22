# Go — Specific Rules (reins CLI)

These rules apply when modifying `.go` files in the reins repository.
Loaded via `rules/INDEX.yaml` trigger: `**/*.go`.

---

## Language

- **S-GO-01** — MUST use the Go toolchain version declared in `go.mod`.
- **S-GO-02** — MUST wrap errors with `%w` and check with `errors.Is` / `errors.As`.
- **S-GO-03** — MUST prefer small, well-named helper functions over inline repetition.
- **S-GO-04** — MUST favor standard-library helpers over reimplementing common behavior.
- **S-GO-05** — MUST format all code with `gofmt`.
- **S-GO-06** — MUST pass `go vet` and `staticcheck` with zero warnings before committing.
- **S-GO-07** — MUST NOT introduce external dependencies. The reins CLI uses stdlib only.

## Testing

- **S-GO-08** — MUST use `t.Chdir(t.TempDir())` for tests that interact with the filesystem.
- **S-GO-09** — MUST use `t.Cleanup` (not `defer`) for test teardown that must survive subtests.

## Verification

- **S-GO-10** — MUST run `task check:all` (format, vet, test) before every commit.
