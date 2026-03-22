// Package cli implements the reins CLI commands.
package cli

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/pastel-sketchbook/reins/content"
)

const (
	managedDir  = ".reins"
	versionFile = ".reins/VERSION"
)

// version is set via ldflags at build time.
var version = "dev"

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
		return runInit(ctx)
	case "update":
		return runUpdate(ctx)
	case "list":
		return runList(ctx)
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
  init      Bootstrap reins in the current project
  update    Refresh managed files (.reins/) to latest version
  list      List available language/framework templates
  version   Print reins version

Run 'reins init' from your project root to get started.
`)
}

// runInit creates .reins/ with managed files and copies templates to the
// project root. Templates are never overwritten; managed files are only
// written if .reins/ doesn't exist yet.
func runInit(ctx context.Context) int {
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

	fmt.Println()
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
