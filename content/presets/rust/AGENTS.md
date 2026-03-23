# Project Instructions

Read and follow all instructions in `.reins/METHODOLOGY.md`. That file is
your core methodology — TDD, Tidy First, commit rules, and quality standards.

## Tech Stack

- Language: Rust 1.94+, edition 2024 (see `Cargo.toml` for MSRV)
- Async runtime: `tokio`
- Web framework: `axum` (with `tower` middleware)
- TUI framework: [ratatui](https://ratatui.rs) with `crossterm` backend
- Error handling: `anyhow` for application, `thiserror` for library crates
- Observability: `tracing` + `tracing-subscriber` (with `EnvFilter`)
- Build system: [Task](https://taskfile.dev) (see `Taskfile.yml`)

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
  0001_ratatui-over-cursive.md
  0002_crossterm-over-termion.md
  ...
```

Use the `000n_<slug>.md` naming convention. Each record documents the
context, decision, and consequences.

## Verification

- Run `task check:all` before every commit.
- The rule-guard agent is defined in `.reins/agents/rule-guard.md`.
  Invoke it for independent verification — do not self-review.
