---
name: reins
version: 0.4.0
description: |
  Use the reins CLI to bootstrap and maintain AI-assisted development
  frameworks in any project. Covers init, update, rule system, analysis
  lenses, autopilot, and verification workflow. Use when initializing
  reins in a new project, updating managed files, managing language rule
  templates, generating analysis-lens concerns, or running the reins
  development workflow (TDD, Tidy First, layered rules).
allowed-tools:
  - Bash
  - Read
  - Grep
  - Glob
  - Task
  - TodoWrite
---

# reins

Reins is a framework for consistent, high-quality AI-assisted development.
It shapes how AI coding agents collaborate with developers by codifying
TDD methodology, layered code rules, and independent verification.

---

## CLI Reference

### Commands

| Command | Description |
|---------|-------------|
| `reins init [--lang <name>]` | Bootstrap reins in the current project. With `--lang`, apply a language preset |
| `reins update` | Refresh managed files in `.reins/` to the latest version |
| `reins list` | List available language/framework rule templates |
| `reins lens [--preset <name>] [--lens <alias>...]` | Generate analysis-lens concern templates. No flags lists available lenses |
| `reins skill` | Install the reins skill for AI tool discovery |
| `reins version` | Print installed reins version |

### `reins init`

Run from the project root. Creates:

- `.reins/` -- managed directory (methodology, quality principles, rule-guard agent, templates)
- `AGENTS.md` -- bridge file pointing AI agents to `.reins/METHODOLOGY.md`
- `Taskfile.yml` -- task automation skeleton with placeholder commands
- `rules/INDEX.yaml` -- rule trigger mapping
- `AUTOPILOT.md` -- autonomous session template
- `.editorconfig` -- editor settings

Scaffold files (AGENTS.md, Taskfile.yml, etc.) are created once and never
overwritten. They are project-owned. Managed files in `.reins/` are owned
by reins and refreshed on `reins update`.

#### Language presets (`--lang`)

The `--lang` flag replaces generic placeholders with a fully configured
scaffold for a specific language. Presets only configure reins artifacts --
they do not generate application code.

| Preset | Flag | What it configures |
|--------|------|--------------------|
| Go | `--lang go` | Taskfile with gofmt/vet/staticcheck/test, Go-specific AGENTS.md, `**/*.go` trigger, `rules/specifics/go.md`, `docs/rationale/` |
| Rust | `--lang rust` | Taskfile with cargo fmt/clippy/test/build, Rust TUI AGENTS.md (ratatui + crossterm), `**/*.rs` trigger, `rules/specifics/rust.md`, `docs/rationale/` |
| Zig | `--lang zig` | Taskfile with zig fmt/build/test, Zig AGENTS.md (allocators + std.log), `**/*.zig` trigger, `rules/specifics/zig.md`, `docs/rationale/` |

Adding a new preset requires only a `content/presets/<lang>/` directory and
an entry in the `presetRuleTemplates` map.

### `reins update`

Refreshes managed files in `.reins/` when a newer version of reins is
installed. Never touches project-owned files. Also refreshes the skill
file if one was previously installed (global or local).

### `reins list`

Lists available language/framework rule templates that can be copied from
`.reins/templates/specifics/` into `rules/specifics/`.

### `reins lens`

Generates analysis-lens concern templates that instruct the agent to apply
structured research lenses during codebase review. The generated file lands
in `rules/concerns/analysis-lenses.md` by default and is loaded via
`INDEX.yaml` triggers like any other concern.

With no flags, prints available lenses and presets.

#### Flags

| Flag | Description |
|------|-------------|
| `--preset <name>` | Select a preset bundle: `quick`, `dd`, `strat`, `all` |
| `--lens <alias>` | Select an individual lens (repeatable): `synth`, `stake`, `time`, `evidence`, `contra`, `assume`, `weak`, `frame`, `impl`, `quest` |
| `--output <path>` | Custom output path (default: `rules/concerns/analysis-lenses.md`) |

#### Presets

| Preset | Alias | Lenses |
|--------|-------|--------|
| `quick-synthesis` | `quick` | Expert Synthesizer + Implementation Blueprint |
| `due-diligence` | `dd` | Evidence Mapper, Contradiction Hunter, Assumption Excavator, Weakness Spotter |
| `strategic-planning` | `strat` | Expert Synthesizer, Framework Builder, Implementation Blueprint, Question Generator |
| `full-protocol` | `all` | All 10 lenses |

Presets and individual lenses can be combined. Duplicates are merged and
results are always in canonical order (Synthesizers, Auditors, Architects).
When 3+ lenses are active, a cross-lens synthesis section is appended.

---

## File Ownership Model

| Location | Owned by | Overwritten on update? |
|----------|----------|----------------------|
| `.reins/METHODOLOGY.md` | Reins CLI | Yes |
| `.reins/agents/rule-guard.md` | Reins CLI | Yes |
| `.reins/rules/principles/quality.md` | Reins CLI | Yes |
| `.reins/templates/` | Reins CLI | Yes |
| `.reins/VERSION` | Reins CLI | Yes |
| `AGENTS.md` | Project | No |
| `Taskfile.yml` | Project | No |
| `rules/INDEX.yaml` | Project | No |
| `rules/specifics/` | Project | No |
| `rules/concerns/` | Project | No |
| `AUTOPILOT.md` | Project | No |

---

## Rule System

### Three-tier hierarchy

| Tier | Scope | Loaded when |
|------|-------|-------------|
| **Principles** | Universal quality rules | Always |
| **Concerns** | Cross-cutting constraints | File path + optional content pattern match |
| **Specifics** | Language/framework rules | File path match |

### Loading protocol

At the start of every task:

1. Open `rules/INDEX.yaml`
2. Load all files under `principles:` unconditionally
3. For each file being modified, match against `trigger:` globs in `concerns:` and `specifics:`
4. For concerns with `content_pattern:`, scan file content and load only if the regex matches
5. Read each loaded rule file in full before writing code

### Conflict resolution

Most-specific wins: **Specifics > Concerns > Principles**.

### Adding rules

Copy a template and register in INDEX.yaml:

```bash
reins list
cp .reins/templates/specifics/go.md rules/specifics/go.md
```

Then add the trigger in `rules/INDEX.yaml`:

```yaml
specifics:
  - trigger: "**/*.go"
    rules:
      - rules/specifics/go.md
```

---

## Core Workflow

Every task follows this sequence:

1. **Explore** -- Read relevant source files before planning
2. **Plan** -- State approach before writing code
3. **Implement** -- Write code following loaded rules
4. **Verify** -- Run `task check:all`. Fix failures before committing
5. **Commit** -- One logical unit per commit, Conventional Commits format

### TDD cycle

Red (failing test) -> Green (minimum code to pass) -> Refactor (only when green).

### Tidy First

Separate structural changes from behavioral changes. Never mix in the same
commit. Make structural changes first when both are needed.

### Verification

Run `task check:all` before every commit. This runs format, lint, vet, and
unit tests. The rule-guard agent (`.reins/agents/rule-guard.md`) performs
independent verification -- do not self-review against rules.

---

## Autopilot

`AUTOPILOT.md` defines goal, constraints, and iteration protocol for
autonomous agent sessions. Each iteration:

1. **Hypothesize** -- State what you will change and the expected effect
2. **Implement** -- Make the smallest change that tests the hypothesis
3. **Verify** -- Run `task check:all`. Fix or revert if it fails
4. **Evaluate** -- Did the change move toward the goal?
5. **Decide** -- Continue or stop

The "What to Do Next" section in `AGENTS.md` (or `.reins/METHODOLOGY.md`)
is a persistent backlog that survives across sessions.

---

## Setup Checklist (after `reins init`)

> **Tip:** If you used `reins init --lang <name>`, steps 1-3 are already done.
> Skip to step 4.

1. Edit `Taskfile.yml` -- replace placeholder commands with your toolchain
2. Edit `rules/INDEX.yaml` -- add triggers for your language/framework
3. Copy language templates: `cp .reins/templates/specifics/<lang>.md rules/specifics/`
4. (Optional) Edit `AUTOPILOT.md` -- define an autonomous session goal
5. Commit: `git add .reins AGENTS.md rules/ Taskfile.yml && git commit -m "chore: init reins"`
