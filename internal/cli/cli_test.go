package cli

import (
	"bytes"
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// newTestLogger configures slog to write to a buffer and returns it.
// The previous default logger is restored via t.Cleanup.
func newTestLogger(t *testing.T) *bytes.Buffer {
	t.Helper()
	var buf bytes.Buffer
	h := slog.NewTextHandler(&buf, &slog.HandlerOptions{
		ReplaceAttr: func(_ []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.Attr{}
			}
			return a
		},
	})
	old := slog.Default()
	slog.SetDefault(slog.New(h))
	t.Cleanup(func() { slog.SetDefault(old) })
	return &buf
}

// suppressOutput discards both stdout and slog output while fn runs.
func suppressOutput(t *testing.T, fn func()) {
	t.Helper()

	// Discard slog output.
	oldLogger := slog.Default()
	slog.SetDefault(slog.New(slog.NewTextHandler(io.Discard, nil)))

	// Discard stdout.
	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("opening devnull: %v", err)
	}
	defer devNull.Close()

	oldOut := os.Stdout
	os.Stdout = devNull

	fn()

	os.Stdout = oldOut
	slog.SetDefault(oldLogger)
}

func TestRunInit_CreatesExpectedFiles(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := runInit(ctx)
	if code != 0 {
		t.Fatalf("runInit() = %d, want 0", code)
	}

	wantFiles := []string{
		// Managed files
		".reins/AGENTS.md",
		".reins/agents/rule-guard.md",
		".reins/rules/principles/quality.md",
		".reins/VERSION",
		".reins/templates/specifics/go.md",
		".reins/templates/specifics/echo.md",
		".reins/templates/specifics/badgerdb.md",
		// Scaffold files
		"CLAUDE.md",
		"Taskfile.yml",
		"rules/INDEX.yaml",
	}

	for _, f := range wantFiles {
		if _, err := os.Stat(f); err != nil {
			t.Errorf("expected %s to exist: %v", f, err)
		}
	}

	wantDirs := []string{
		"rules/principles",
		"rules/concerns",
		"rules/specifics",
	}

	for _, d := range wantDirs {
		info, err := os.Stat(d)
		if err != nil {
			t.Errorf("expected directory %s to exist: %v", d, err)
			continue
		}
		if !info.IsDir() {
			t.Errorf("expected %s to be a directory", d)
		}
	}
}

func TestRunInit_RejectsDoubleInit(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	if code := runInit(ctx); code != 0 {
		t.Fatalf("first runInit() = %d, want 0", code)
	}

	code := runInit(ctx)
	if code != 1 {
		t.Errorf("second runInit() = %d, want 1", code)
	}
}

func TestRunInit_WritesVersion(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	SetVersion("1.2.3")
	t.Cleanup(func() { SetVersion("dev") })

	if code := runInit(ctx); code != 0 {
		t.Fatalf("runInit() = %d, want 0", code)
	}

	data, err := os.ReadFile(".reins/VERSION")
	if err != nil {
		t.Fatalf("failed to read VERSION: %v", err)
	}
	if got := string(data); got != "1.2.3\n" {
		t.Errorf("VERSION = %q, want %q", got, "1.2.3\n")
	}
}

func TestRunUpdate_RefreshesManagedFiles(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	SetVersion("1.0.0")

	if code := runInit(ctx); code != 0 {
		t.Fatalf("runInit() = %d, want 0", code)
	}

	// Tamper with a managed file.
	agentsPath := filepath.Join(".reins", "AGENTS.md")
	if err := os.WriteFile(agentsPath, []byte("tampered"), 0o644); err != nil {
		t.Fatalf("failed to tamper with AGENTS.md: %v", err)
	}

	// Simulate a newer version of reins so update proceeds.
	SetVersion("1.1.0")
	t.Cleanup(func() { SetVersion("dev") })

	code := runUpdate(ctx)
	if code != 0 {
		t.Fatalf("runUpdate() = %d, want 0", code)
	}

	data, err := os.ReadFile(agentsPath)
	if err != nil {
		t.Fatalf("failed to read AGENTS.md after update: %v", err)
	}
	if string(data) == "tampered" {
		t.Error("AGENTS.md was not refreshed by update")
	}
}

func TestRunUpdate_DoesNotOverwriteProjectFiles(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	SetVersion("1.0.0")

	if code := runInit(ctx); code != 0 {
		t.Fatalf("runInit() = %d, want 0", code)
	}

	// Tamper with a project-owned file.
	customContent := []byte("# My custom CLAUDE.md\n")
	if err := os.WriteFile("CLAUDE.md", customContent, 0o644); err != nil {
		t.Fatalf("failed to write custom CLAUDE.md: %v", err)
	}

	// Simulate a newer version so update proceeds.
	SetVersion("1.1.0")
	t.Cleanup(func() { SetVersion("dev") })

	code := runUpdate(ctx)
	if code != 0 {
		t.Fatalf("runUpdate() = %d, want 0", code)
	}

	// Project file should be untouched.
	data, err := os.ReadFile("CLAUDE.md")
	if err != nil {
		t.Fatalf("failed to read CLAUDE.md after update: %v", err)
	}
	if string(data) != string(customContent) {
		t.Errorf("CLAUDE.md was modified by update:\ngot:  %q\nwant: %q", string(data), string(customContent))
	}
}

func TestRunUpdate_FailsWithoutInit(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := runUpdate(ctx)
	if code != 1 {
		t.Errorf("runUpdate() without init = %d, want 1", code)
	}
}

func TestRunList_Succeeds(t *testing.T) {
	ctx := context.Background()

	code := runList(ctx)
	if code != 0 {
		t.Errorf("runList() = %d, want 0", code)
	}
}

func TestRun_Version(t *testing.T) {
	ctx := context.Background()

	code := Run(ctx, []string{"reins", "version"})
	if code != 0 {
		t.Errorf("Run version = %d, want 0", code)
	}
}

func TestRun_Help(t *testing.T) {
	ctx := context.Background()

	code := Run(ctx, []string{"reins", "help"})
	if code != 0 {
		t.Errorf("Run help = %d, want 0", code)
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	ctx := context.Background()

	code := Run(ctx, []string{"reins", "bogus"})
	if code != 1 {
		t.Errorf("Run bogus = %d, want 1", code)
	}
}

func TestRun_NoArgs(t *testing.T) {
	ctx := context.Background()

	code := Run(ctx, []string{"reins"})
	if code != 0 {
		t.Errorf("Run no args = %d, want 0", code)
	}
}

func TestRunUpdate_SkipsWhenAlreadyCurrent(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	SetVersion("1.0.0")
	t.Cleanup(func() { SetVersion("dev") })

	suppressOutput(t, func() {
		if code := runInit(ctx); code != 0 {
			t.Fatalf("runInit() = %d, want 0", code)
		}
	})

	// Update with the same version should succeed and report already current.
	logBuf := newTestLogger(t)
	code := runUpdate(ctx)
	if code != 0 {
		t.Errorf("runUpdate() same version = %d, want 0", code)
	}

	if !strings.Contains(logBuf.String(), "already current") {
		t.Errorf("expected 'already current' in log output, got:\n%s", logBuf.String())
	}

	// VERSION file should still contain the same version.
	data, err := os.ReadFile(".reins/VERSION")
	if err != nil {
		t.Fatalf("failed to read VERSION: %v", err)
	}
	if got := string(data); got != "1.0.0\n" {
		t.Errorf("VERSION = %q, want %q", got, "1.0.0\n")
	}
}

func TestRunUpdate_ProceedsWhenVersionDiffers(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	SetVersion("1.0.0")

	if code := runInit(ctx); code != 0 {
		t.Fatalf("runInit() = %d, want 0", code)
	}

	// Simulate a newer version of reins.
	SetVersion("2.0.0")
	t.Cleanup(func() { SetVersion("dev") })

	code := runUpdate(ctx)
	if code != 0 {
		t.Errorf("runUpdate() newer version = %d, want 0", code)
	}

	// VERSION file should reflect the new version.
	data, err := os.ReadFile(".reins/VERSION")
	if err != nil {
		t.Fatalf("failed to read VERSION: %v", err)
	}
	if got := string(data); got != "2.0.0\n" {
		t.Errorf("VERSION = %q, want %q", got, "2.0.0\n")
	}
}

func TestRunInit_WarnsWithoutGitDir(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	// No .git directory — init should still succeed (return 0) but warn.
	logBuf := newTestLogger(t)
	code := runInit(ctx)
	if code != 0 {
		t.Errorf("runInit() without .git = %d, want 0 (with warning)", code)
	}
	if !strings.Contains(logBuf.String(), "WARN") {
		t.Errorf("expected WARN in log output without .git, got:\n%s", logBuf.String())
	}
}

func TestRunInit_NoWarningWithGitDir(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	ctx := context.Background()

	// Create .git directory to simulate a project root.
	if err := os.Mkdir(".git", 0o755); err != nil {
		t.Fatalf("failed to create .git: %v", err)
	}

	logBuf := newTestLogger(t)
	code := runInit(ctx)
	if code != 0 {
		t.Errorf("runInit() with .git = %d, want 0", code)
	}
	if strings.Contains(logBuf.String(), "WARN") {
		t.Errorf("unexpected WARN in log output with .git present:\n%s", logBuf.String())
	}
}
