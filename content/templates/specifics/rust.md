# Rust — Specific Rules

These rules apply when modifying `.rs` files in this project.
Loaded via `rules/INDEX.yaml` trigger: `**/*.rs`.

---

## Language

- **S-RS-01** — MUST use the Rust edition declared in `Cargo.toml`.
- **S-RS-02** — MUST use `Result<T, E>` for fallible operations. Propagate errors with `?`; do not `.unwrap()` or `.expect()` outside of tests.
- **S-RS-03** — MUST use `thiserror` for library error types and `anyhow` for application error types. Do not implement `std::error::Error` manually unless justified.
- **S-RS-04** — MUST prefer borrowing (`&T`, `&mut T`) over cloning. Clone only when ownership transfer is required.
- **S-RS-05** — MUST use `clippy` at the `pedantic` lint level. Suppress individual lints with `#[allow(...)]` and a justification comment.
- **S-RS-06** — MUST format all code with `rustfmt`. Do not use `#[rustfmt::skip]` without justification.
- **S-RS-07** — MUST use `unsafe` only when strictly necessary. Each `unsafe` block MUST have a `// SAFETY:` comment explaining the invariant.
- **S-RS-08** — MUST prefer iterator combinators (`.map`, `.filter`, `.collect`) over manual loops when they improve clarity.
- **S-RS-09** — MUST use `#[must_use]` on functions whose return value should not be silently discarded.

## Verification

- **S-RS-10** — MUST run `task check:all` (fmt, clippy, test) before every commit.
- **S-RS-11** — MUST fix all warnings reported by `cargo clippy -- -D warnings` before committing.
