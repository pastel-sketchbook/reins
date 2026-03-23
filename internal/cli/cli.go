// Package cli implements the reins CLI commands.
package cli

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/pastel-sketchbook/reins/content"
)

const (
	managedDir    = ".reins"
	versionFile   = ".reins/VERSION"
	skillRelPath  = ".agents/skills/reins/SKILL.md"
	skillEmbedded = "skill/SKILL.md"
)

// version is set via ldflags at build time.
var version = "dev"

// stdin is the reader used by interactive prompts.
// Tests replace this to inject input without blocking.
var stdin io.Reader = os.Stdin

// openTTY opens the controlling terminal for interactive fallback
// when stdin is a pipe (e.g. curl | sh). Tests replace this to
// prevent blocking on /dev/tty.
var openTTY = func() (io.ReadCloser, error) {
	return os.Open("/dev/tty")
}

// SetVersion sets the embedded version string (called from main).
func SetVersion(v string) {
	version = v
}

// Run dispatches the CLI command.
func Run(ctx context.Context, args []string) int {
	if len(args) < 2 {
		printUsage()
		return 0
	}

	switch args[1] {
	case "init":
		return runInit(ctx, args[2:])
	case "update":
		return runUpdate(ctx)
	case "list":
		return runList(ctx)
	case "skill":
		return runSkill(ctx)
	case "version":
		fmt.Println(version)
		return 0
	case "help", "-h", "--help":
		printUsage()
		return 0
	default:
		slog.ErrorContext(ctx, "unknown command", "command", args[1])
		printUsage()
		return 1
	}
}

func printUsage() {
	fmt.Print(`Usage: reins <command>

Commands:
  init [--lang <name>]  Bootstrap reins in the current project
  update                Refresh managed files (.reins/) to latest version
  list                  List available language/framework templates
  skill                 Install the reins skill for AI tool discovery
  version               Print reins version

Language presets (--lang):
  go      Go project (gofmt, go vet, staticcheck, go test)

Run 'reins init' from your project root to get started.
Run 'reins init --lang go' for a preconfigured Go project.
`)
}

// runInit creates .reins/ with managed files and copies templates to the
// project root. Templates are never overwritten; managed files are only
// written if .reins/ doesn't exist yet. When args contains --lang <name>,
// a language preset is applied after the generic scaffold.
func runInit(ctx context.Context, args []string) int {
	lang := parseLangFlag(args)

	// Validate the preset exists before doing any work.
	if lang != "" {
		if err := validatePreset(lang); err != nil {
			slog.ErrorContext(ctx, "unknown language preset", "lang", lang, "err", err)
			return 1
		}
	}

	if _, err := os.Stat(managedDir); err == nil {
		slog.ErrorContext(ctx, "already initialized, use 'reins update' to refresh", "path", managedDir+"/")
		return 1
	}

	// Warn if not running from a project root (no .git directory).
	if _, err := os.Stat(".git"); errors.Is(err, fs.ErrNotExist) {
		slog.WarnContext(ctx, "no .git directory found, are you in the project root?")
	}

	slog.InfoContext(ctx, "initializing")

	// 1. Copy managed files → .reins/
	if err := copyEmbeddedDir(ctx, "managed", managedDir, true); err != nil {
		slog.ErrorContext(ctx, "failed to write managed files", "err", err)
		return 1
	}

	// 2. Write version marker.
	if err := os.WriteFile(versionFile, []byte(version+"\n"), 0o644); err != nil {
		slog.ErrorContext(ctx, "failed to write version file", "err", err)
		return 1
	}

	// 3. Copy scaffold → project root (skip existing).
	if err := copyEmbeddedDir(ctx, "scaffold", ".", false); err != nil {
		slog.ErrorContext(ctx, "failed to write scaffold files", "err", err)
		return 1
	}

	// 4. Copy language templates → .reins/templates/ for manual use.
	if err := copyEmbeddedDir(ctx, "templates", filepath.Join(managedDir, "templates"), true); err != nil {
		slog.ErrorContext(ctx, "failed to write language templates", "err", err)
		return 1
	}

	// 5. Create local rule directories.
	for _, dir := range []string{"rules/principles", "rules/concerns", "rules/specifics"} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			slog.ErrorContext(ctx, "failed to create directory", "path", dir, "err", err)
			return 1
		}
	}

	// 6. Apply language preset if specified.
	if lang != "" {
		if code := applyPreset(ctx, lang); code != 0 {
			return code
		}
	}

	// 7. Prompt for skill installation.
	choice := promptSkillLocation()
	switch choice {
	case "g":
		if code := installSkill(ctx, true); code != 0 {
			return code
		}
	case "l":
		if code := installSkill(ctx, false); code != 0 {
			return code
		}
	}

	fmt.Println()
	if lang != "" {
		printPresetNextSteps(lang)
	} else {
		printGenericNextSteps()
	}

	return 0
}

// runUpdate refreshes managed files in .reins/ without touching project-owned
// files (AGENTS.md, Taskfile.yml, rules/INDEX.yaml).
func runUpdate(ctx context.Context) int {
	if _, err := os.Stat(managedDir); errors.Is(err, fs.ErrNotExist) {
		slog.ErrorContext(ctx, "not initialized, run 'reins init' first")
		return 1
	}

	// Check if already at the current version.
	if installed, err := os.ReadFile(versionFile); err == nil {
		if strings.TrimSpace(string(installed)) == version {
			slog.InfoContext(ctx, "managed files already current", "version", version)
			return 0
		}
	}

	slog.InfoContext(ctx, "updating managed files")

	if err := copyEmbeddedDir(ctx, "managed", managedDir, true); err != nil {
		slog.ErrorContext(ctx, "failed to update managed files", "err", err)
		return 1
	}

	// Update version marker.
	if err := os.WriteFile(versionFile, []byte(version+"\n"), 0o644); err != nil {
		slog.ErrorContext(ctx, "failed to write version file", "err", err)
		return 1
	}

	// Also refresh template copies inside .reins/ so users can diff
	// against the latest versions and manually copy language templates.
	if err := copyEmbeddedDir(ctx, "scaffold", filepath.Join(managedDir, "scaffold"), true); err != nil {
		slog.ErrorContext(ctx, "failed to update scaffold templates", "err", err)
		return 1
	}

	if err := copyEmbeddedDir(ctx, "templates", filepath.Join(managedDir, "templates"), true); err != nil {
		slog.ErrorContext(ctx, "failed to update language templates", "err", err)
		return 1
	}

	// Refresh skill if previously installed (global or local).
	refreshSkill(ctx)

	slog.InfoContext(ctx, "update complete", "version", version)

	fmt.Println()
	fmt.Println("Project-owned files were not modified (AGENTS.md, .editorconfig, Taskfile.yml, rules/INDEX.yaml, AUTOPILOT.md).")
	fmt.Println("To check for scaffold changes:")
	fmt.Println("  diff AGENTS.md .reins/scaffold/AGENTS.md")
	fmt.Println("  diff rules/INDEX.yaml .reins/scaffold/rules/INDEX.yaml")
	fmt.Println("  diff AUTOPILOT.md .reins/scaffold/AUTOPILOT.md")
	fmt.Println()

	return 0
}

// runList prints available language/framework templates.
func runList(ctx context.Context) int {
	fmt.Println("Available language/framework templates:")
	fmt.Println()

	err := fs.WalkDir(content.FS, "templates/specifics", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() || !strings.HasSuffix(path, ".md") {
			return nil
		}

		name := strings.TrimSuffix(filepath.Base(path), ".md")
		fmt.Printf("  %-16s  rules/specifics/%s.md\n", name, name)

		return nil
	})
	if err != nil {
		slog.ErrorContext(ctx, "failed to list templates", "err", err)
		return 1
	}

	fmt.Println()
	fmt.Println("To use a template:")
	fmt.Println("  cp .reins/templates/specifics/<lang>.md rules/specifics/<lang>.md")
	fmt.Println()
	fmt.Println("Then uncomment the matching trigger in rules/INDEX.yaml.")
	fmt.Println()

	return 0
}

// runSkill is the standalone `reins skill` command. It prompts for
// the installation location and copies the embedded SKILL.md.
func runSkill(ctx context.Context) int {
	choice := promptSkillLocation()
	switch choice {
	case "g":
		return installSkill(ctx, true)
	case "l":
		return installSkill(ctx, false)
	default:
		fmt.Println("Skipped.")
		return 0
	}
}

// promptSkillLocation asks the user where to install the reins skill
// and returns "g" (global), "l" (local), or "n" (skip).
//
// When stdin is a pipe or closed (e.g. curl | sh), the initial scan
// hits EOF. On Unix systems we fall back to /dev/tty so the user can
// still interact. If /dev/tty is unavailable (Windows, CI), we skip.
func promptSkillLocation() string {
	fmt.Println()
	fmt.Println("Install reins skill for AI tool discovery?")
	fmt.Println()
	fmt.Println("  [g] Global  ~/.agents/skills/reins/  (shared across all projects)")
	fmt.Println("  [l] Local   .agents/skills/reins/    (this project only)")
	fmt.Println("  [n] Skip")
	fmt.Println()
	fmt.Print("Choice [g/l/n]: ")

	scanner := bufio.NewScanner(stdin)
	if scanner.Scan() {
		return strings.ToLower(strings.TrimSpace(scanner.Text()))
	}

	// stdin was a pipe or hit EOF — try /dev/tty as fallback.
	tty, err := openTTY()
	if err != nil {
		return "n"
	}
	defer tty.Close()
	ttyScanner := bufio.NewScanner(tty)
	if ttyScanner.Scan() {
		return strings.ToLower(strings.TrimSpace(ttyScanner.Text()))
	}
	return "n"
}

// installSkill copies the embedded SKILL.md to the chosen location.
// When global is true, it installs to ~/.agents/skills/reins/SKILL.md.
// When false, it installs to .agents/skills/reins/SKILL.md in the cwd.
func installSkill(ctx context.Context, global bool) int {
	data, err := fs.ReadFile(content.FS, skillEmbedded)
	if err != nil {
		slog.ErrorContext(ctx, "failed to read embedded skill", "err", err)
		return 1
	}

	dst := localSkillPath()
	if global {
		dst = globalSkillPath()
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		slog.ErrorContext(ctx, "failed to create skill directory", "path", dst, "err", err)
		return 1
	}

	if err := os.WriteFile(dst, data, 0o644); err != nil {
		slog.ErrorContext(ctx, "failed to write skill file", "path", dst, "err", err)
		return 1
	}

	if global {
		slog.InfoContext(ctx, "installed global skill", "path", dst)
	} else {
		slog.InfoContext(ctx, "installed local skill", "path", dst)
	}
	return 0
}

// refreshSkill updates the SKILL.md in whichever location(s) it already
// exists. Called during `reins update`.
func refreshSkill(ctx context.Context) {
	data, err := fs.ReadFile(content.FS, skillEmbedded)
	if err != nil {
		slog.WarnContext(ctx, "failed to read embedded skill for refresh", "err", err)
		return
	}

	for _, dst := range []string{globalSkillPath(), localSkillPath()} {
		if _, statErr := os.Stat(dst); statErr != nil {
			continue
		}
		if writeErr := os.WriteFile(dst, data, 0o644); writeErr != nil {
			slog.WarnContext(ctx, "failed to refresh skill", "path", dst, "err", writeErr)
			continue
		}
		slog.InfoContext(ctx, "refreshed skill", "path", dst)
	}
}

// globalSkillPath returns ~/.agents/skills/reins/SKILL.md.
func globalSkillPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		home = os.Getenv("HOME")
	}
	return filepath.Join(home, skillRelPath)
}

// localSkillPath returns .agents/skills/reins/SKILL.md relative to cwd.
func localSkillPath() string {
	return skillRelPath
}

// copyEmbeddedDir walks srcRoot inside the embedded FS and writes files to
// dstRoot on disk. When overwrite is false, existing files are skipped.
func copyEmbeddedDir(ctx context.Context, srcRoot, dstRoot string, overwrite bool) error {
	return fs.WalkDir(content.FS, srcRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		rel, err := filepath.Rel(srcRoot, path)
		if err != nil {
			return fmt.Errorf("computing relative path for %s: %w", path, err)
		}

		dst := filepath.Join(dstRoot, rel)

		if d.IsDir() {
			if err := os.MkdirAll(dst, 0o755); err != nil {
				return fmt.Errorf("creating directory %s: %w", dst, err)
			}
			return nil
		}

		if !overwrite {
			if _, statErr := os.Stat(dst); statErr == nil {
				slog.InfoContext(ctx, "skip", "path", dst, "reason", "already exists")
				return nil
			}
		}

		data, readErr := fs.ReadFile(content.FS, path)
		if readErr != nil {
			return fmt.Errorf("reading embedded %s: %w", path, readErr)
		}

		if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
			return fmt.Errorf("creating parent directory for %s: %w", dst, err)
		}

		slog.InfoContext(ctx, "create", "path", dst)

		if err := os.WriteFile(dst, data, 0o644); err != nil {
			return fmt.Errorf("writing %s: %w", dst, err)
		}
		return nil
	})
}

// parseLangFlag extracts the --lang value from args. Returns "" if not present.
func parseLangFlag(args []string) string {
	for i, arg := range args {
		if arg == "--lang" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}

// validatePreset checks that a language preset directory exists in the embedded FS.
func validatePreset(lang string) error {
	presetDir := filepath.Join("presets", lang)
	_, err := fs.Stat(content.FS, presetDir)
	if err != nil {
		return fmt.Errorf("preset %q not found: %w", lang, err)
	}
	return nil
}

// presetRuleTemplates maps language presets to the rule template files that
// should be auto-copied into rules/specifics/. Each entry is a filename
// relative to templates/specifics/ in the embedded FS.
var presetRuleTemplates = map[string][]string{
	"go": {"go.md"},
}

// applyPreset overwrites generic scaffold files with language-specific
// preset versions, copies rule templates, and creates preset directories.
func applyPreset(ctx context.Context, lang string) int {
	presetDir := filepath.Join("presets", lang)
	slog.InfoContext(ctx, "applying language preset", "lang", lang)

	// Overwrite scaffold files with preset versions.
	if err := copyEmbeddedDir(ctx, presetDir, ".", true); err != nil {
		slog.ErrorContext(ctx, "failed to apply preset", "lang", lang, "err", err)
		return 1
	}

	// Copy rule templates to active rules directory.
	for _, tmpl := range presetRuleTemplates[lang] {
		src := filepath.Join("templates", "specifics", tmpl)
		dst := filepath.Join("rules", "specifics", tmpl)

		data, err := fs.ReadFile(content.FS, src)
		if err != nil {
			slog.ErrorContext(ctx, "failed to read rule template", "template", src, "err", err)
			return 1
		}

		slog.InfoContext(ctx, "create", "path", dst)
		if err := os.WriteFile(dst, data, 0o644); err != nil {
			slog.ErrorContext(ctx, "failed to write rule file", "path", dst, "err", err)
			return 1
		}
	}

	// Create ADR directory.
	adrDir := filepath.Join("docs", "rationale")
	if err := os.MkdirAll(adrDir, 0o755); err != nil {
		slog.ErrorContext(ctx, "failed to create ADR directory", "path", adrDir, "err", err)
		return 1
	}

	return 0
}

// printGenericNextSteps prints post-init instructions for generic (no-lang) init.
func printGenericNextSteps() {
	fmt.Println("Done. Next steps:")
	fmt.Println()
	fmt.Println("  1. Edit Taskfile.yml — replace TODO placeholders with your toolchain commands.")
	fmt.Println("  2. Edit rules/INDEX.yaml — uncomment or add specifics for your language.")
	fmt.Println("  3. Copy a language template if needed:")
	fmt.Println("       reins list              # see available templates")
	fmt.Println("       cp .reins/templates/specifics/go.md rules/specifics/go.md")
	fmt.Println("  4. (Optional) Edit AUTOPILOT.md — define a goal for autonomous agent sessions.")
	fmt.Println("  5. Commit:")
	fmt.Println("       git add .reins .editorconfig AGENTS.md rules/ Taskfile.yml AUTOPILOT.md")
	fmt.Println("       git commit -m 'chore: init reins framework'")
	fmt.Println()
}

// printPresetNextSteps prints post-init instructions when a language preset was applied.
func printPresetNextSteps(lang string) {
	fmt.Println("Done. Language preset applied: " + lang)
	fmt.Println()
	fmt.Println("  Taskfile.yml, AGENTS.md, and rules/ have been preconfigured.")
	fmt.Println()
	fmt.Println("  Next steps:")
	fmt.Println()
	fmt.Println("  1. Review Taskfile.yml — adjust tasks for your project's toolchain.")
	fmt.Println("  2. Review rules/INDEX.yaml — uncomment additional triggers as needed.")
	fmt.Println("  3. (Optional) Edit AUTOPILOT.md — define a goal for autonomous agent sessions.")
	fmt.Println("  4. Commit:")
	fmt.Println("       git add .reins .editorconfig AGENTS.md rules/ Taskfile.yml AUTOPILOT.md docs/")
	fmt.Println("       git commit -m 'chore: init reins framework (" + lang + " preset)'")
	fmt.Println()
}
