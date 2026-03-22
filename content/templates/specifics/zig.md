# Zig — Specific Rules

These rules apply when modifying `.zig` files in this project.
Loaded via `rules/INDEX.yaml` trigger: `**/*.zig`.

---

## Language

- **S-ZIG-01** — MUST use the Zig version declared in `build.zig.zon` or the project README.
- **S-ZIG-02** — MUST handle all error unions explicitly. Use `try` for propagation, `catch` for recovery. Do not discard errors with `_ = expr`.
- **S-ZIG-03** — MUST use `defer` and `errdefer` for resource cleanup. Never rely on manual cleanup at every return point.
- **S-ZIG-04** — MUST prefer slices and iterators over raw pointer arithmetic. Use `@ptrCast` and `@alignCast` only when interfacing with C.
- **S-ZIG-05** — MUST use `comptime` for compile-time validation and generic programming. Do not use runtime checks for invariants that can be verified at compile time.
- **S-ZIG-06** — MUST choose the appropriate allocator for the context (arena for request-scoped, GPA for long-lived, page allocator for large blocks). Do not hardcode a single allocator.
- **S-ZIG-07** — MUST format all code with `zig fmt`. Do not disable formatting directives without justification.
- **S-ZIG-08** — MUST use `std.log` for diagnostic output. Do not use `std.debug.print` in production code.
- **S-ZIG-09** — MUST prefer tagged unions over separate type + value fields for variant data.
- **S-ZIG-10** — MUST use `std.testing` for tests. Place tests in the same file as the code they test using `test` blocks.

## Verification

- **S-ZIG-11** — MUST run `task check:all` (fmt, build, test) before every commit.
- **S-ZIG-12** — MUST fix all warnings reported by the Zig compiler before committing.
