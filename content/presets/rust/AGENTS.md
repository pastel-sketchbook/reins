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

## Audit

Run `task audit` periodically to check for vulnerable dependencies
(`cargo audit`). Use `task fmt:check` in CI for non-modifying format
verification.

### Modernization (Rust 1.65–1.94 / edition 2024)

When modifying existing code, apply edition 2024 modernization patterns.
Flag and replace deprecated crates and idioms:

- `lazy_static!` → `std::sync::LazyLock`  *(1.80)*
- `once_cell` crate → `std::sync::OnceLock` / `std::cell::OnceCell`  *(1.70)*
- `#[async_trait]` → native `async fn` in traits  *(1.75)*
- `match` + single-arm + return → `let ... else`  *(1.65)*
- `Box<dyn Future>` → `impl Future` or `async fn` in trait methods

### Framework conventions

- **anyhow**: Use `.context()` on every `?`, `bail!` for early returns,
  `ensure!` for preconditions. See rules S-RS-30 through S-RS-33.
- **tracing**: No `println!`/`eprintln!`/`log` crate. Use `#[instrument]`
  on public functions, structured fields in events. See S-RS-40 through S-RS-43.
- **tokio**: No blocking in async context. Use `tokio::sync` when lock
  guards cross `.await`. Wire graceful shutdown. See S-RS-50 through S-RS-54.
- **axum**: Extractor pattern, `State<Arc<AppState>>`, `IntoResponse` for
  errors, `tower` middleware layers. See S-RS-60 through S-RS-64.
- **ratatui**: Crossterm backend, MVU architecture, terminal restore on
  panic, `Layout` with `Constraint`, pure render functions. See S-RS-70
  through S-RS-74.

See `rules/specifics/rust.md` for the complete rule set.
