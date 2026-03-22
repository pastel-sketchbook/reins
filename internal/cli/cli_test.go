package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// captureStdout calls fn while redirecting os.Stdout to a buffer.
// Returns the captured output and any error from pipe plumbing.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	old := os.Stdout
	os.Stdout = w

	fn()

	os.Stdout = old
	if err := w.Close(); err != nil {
		t.Fatalf("closing pipe writer: %v", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("reading captured output: %v", err)
	}
	return buf.String()
}

// captureStderr calls fn while redirecting os.Stderr to a buffer.
func captureStderr(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe: %v", err)
	}

	old := os.Stderr
	os.Stderr = w

	fn()

	os.Stderr = old
	if err := w.Close(); err != nil {
		t.Fatalf("closing pipe writer: %v", err)
	}

	var buf bytes.Buffer
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("reading captured output: %v", err)
	}
	return buf.String()
}

// suppressOutput calls fn while discarding both stdout and stderr.
func suppressOutput(t *testing.T, fn func()) {
	t.Helper()

	devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
	if err != nil {
		t.Fatalf("opening devnull: %v", err)
	}
	defer devNull.Close()

	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout = devNull
	os.Stderr = devNull

	fn()

	os.Stdout = oldOut
	os.Stderr = oldErr
}

func TestRunInit_CreatesExpectedFiles(t *testing.T) {
	t.Chdir(t.TempDir())

	code := runInit()
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

	if code := runInit(); code != 0 {
		t.Fatalf("first runInit() = %d, want 0", code)
	}

	code := runInit()
	if code != 1 {
		t.Errorf("second runInit() = %d, want 1", code)
	}
}

func TestRunInit_WritesVersion(t *testing.T) {
	t.Chdir(t.TempDir())

	SetVersion("1.2.3")
	t.Cleanup(func() { SetVersion("dev") })

	if code := runInit(); code != 0 {
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

	SetVersion("1.0.0")

	if code := runInit(); code != 0 {
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

	code := runUpdate()
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

	SetVersion("1.0.0")

	if code := runInit(); code != 0 {
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

	code := runUpdate()
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

	code := runUpdate()
	if code != 1 {
		t.Errorf("runUpdate() without init = %d, want 1", code)
	}
}

func TestRunList_Succeeds(t *testing.T) {
	code := runList()
	if code != 0 {
		t.Errorf("runList() = %d, want 0", code)
	}
}

func TestRun_Version(t *testing.T) {
	code := Run([]string{"reins", "version"})
	if code != 0 {
		t.Errorf("Run version = %d, want 0", code)
	}
}

func TestRun_Help(t *testing.T) {
	code := Run([]string{"reins", "help"})
	if code != 0 {
		t.Errorf("Run help = %d, want 0", code)
	}
}

func TestRun_UnknownCommand(t *testing.T) {
	code := Run([]string{"reins", "bogus"})
	if code != 1 {
		t.Errorf("Run bogus = %d, want 1", code)
	}
}

func TestRun_NoArgs(t *testing.T) {
	code := Run([]string{"reins"})
	if code != 0 {
		t.Errorf("Run no args = %d, want 0", code)
	}
}

func TestRunUpdate_SkipsWhenAlreadyCurrent(t *testing.T) {
	t.Chdir(t.TempDir())

	SetVersion("1.0.0")
	t.Cleanup(func() { SetVersion("dev") })

	suppressOutput(t, func() {
		if code := runInit(); code != 0 {
			t.Fatalf("runInit() = %d, want 0", code)
		}
	})

	// Update with the same version should succeed and report already current.
	var code int
	out := captureStdout(t, func() {
		code = runUpdate()
	})
	if code != 0 {
		t.Errorf("runUpdate() same version = %d, want 0", code)
	}

	if !strings.Contains(out, "already at version") {
		t.Errorf("expected 'already at version' in output, got:\n%s", out)
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

	SetVersion("1.0.0")

	if code := runInit(); code != 0 {
		t.Fatalf("runInit() = %d, want 0", code)
	}

	// Simulate a newer version of reins.
	SetVersion("2.0.0")
	t.Cleanup(func() { SetVersion("dev") })

	code := runUpdate()
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
	// No .git directory — init should still succeed (return 0) but warn.
	var code int
	stderr := captureStderr(t, func() {
		captureStdout(t, func() {
			code = runInit()
		})
	})
	if code != 0 {
		t.Errorf("runInit() without .git = %d, want 0 (with warning)", code)
	}
	if !strings.Contains(stderr, "warning") {
		t.Errorf("expected warning on stderr without .git, got:\n%q", stderr)
	}
}

func TestRunInit_NoWarningWithGitDir(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)

	// Create .git directory to simulate a project root.
	if err := os.Mkdir(".git", 0o755); err != nil {
		t.Fatalf("failed to create .git: %v", err)
	}

	var code int
	stderr := captureStderr(t, func() {
		captureStdout(t, func() {
			code = runInit()
		})
	})
	if code != 0 {
		t.Errorf("runInit() with .git = %d, want 0", code)
	}
	if strings.Contains(stderr, "warning") {
		t.Errorf("unexpected warning on stderr with .git present:\n%q", stderr)
	}
}
