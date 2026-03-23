package cli

import (
	"bytes"
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestMain disables /dev/tty fallback so tests never block on terminal input.
func TestMain(m *testing.M) {
	openTTY = func() (io.ReadCloser, error) {
		return nil, errors.New("no tty in tests")
	}
	os.Exit(m.Run())
}

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

	code := runInit(ctx, nil)
	if code != 0 {
		t.Fatalf("runInit() = %d, want 0", code)
	}

	wantFiles := []string{
		// Managed files
		".reins/METHODOLOGY.md",
		".reins/agents/rule-guard.md",
		".reins/rules/principles/quality.md",
		".reins/VERSION",
		".reins/templates/specifics/go.md",
		".reins/templates/specifics/echo.md",
		".reins/templates/specifics/badgerdb.md",
		".reins/templates/specifics/python.md",
		".reins/templates/specifics/typescript.md",
		".reins/templates/specifics/rust.md",
		".reins/templates/specifics/zig.md",
		// Scaffold files
		".editorconfig",
		"AGENTS.md",
		"Taskfile.yml",
		"rules/INDEX.yaml",
		"AUTOPILOT.md",
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

	if code := runInit(ctx, nil); code != 0 {
		t.Fatalf("first runInit() = %d, want 0", code)
	}

	code := runInit(ctx, nil)
	if code != 1 {
		t.Errorf("second runInit() = %d, want 1", code)
	}
}

func TestRunInit_WritesVersion(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	SetVersion("1.2.3")
	t.Cleanup(func() { SetVersion("dev") })

	if code := runInit(ctx, nil); code != 0 {
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

	if code := runInit(ctx, nil); code != 0 {
		t.Fatalf("runInit() = %d, want 0", code)
	}

	// Tamper with a managed file.
	methodologyPath := filepath.Join(".reins", "METHODOLOGY.md")
	if err := os.WriteFile(methodologyPath, []byte("tampered"), 0o644); err != nil {
		t.Fatalf("failed to tamper with METHODOLOGY.md: %v", err)
	}

	// Simulate a newer version of reins so update proceeds.
	SetVersion("1.1.0")
	t.Cleanup(func() { SetVersion("dev") })

	code := runUpdate(ctx)
	if code != 0 {
		t.Fatalf("runUpdate() = %d, want 0", code)
	}

	data, err := os.ReadFile(methodologyPath)
	if err != nil {
		t.Fatalf("failed to read METHODOLOGY.md after update: %v", err)
	}
	if string(data) == "tampered" {
		t.Error("METHODOLOGY.md was not refreshed by update")
	}
}

func TestRunUpdate_DoesNotOverwriteProjectFiles(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	SetVersion("1.0.0")

	if code := runInit(ctx, nil); code != 0 {
		t.Fatalf("runInit() = %d, want 0", code)
	}

	// Tamper with a project-owned file.
	customContent := []byte("# My custom AGENTS.md\n")
	if err := os.WriteFile("AGENTS.md", customContent, 0o644); err != nil {
		t.Fatalf("failed to write custom AGENTS.md: %v", err)
	}

	// Simulate a newer version so update proceeds.
	SetVersion("1.1.0")
	t.Cleanup(func() { SetVersion("dev") })

	code := runUpdate(ctx)
	if code != 0 {
		t.Fatalf("runUpdate() = %d, want 0", code)
	}

	// Project file should be untouched.
	data, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("failed to read AGENTS.md after update: %v", err)
	}
	if string(data) != string(customContent) {
		t.Errorf("AGENTS.md was modified by update:\ngot:  %q\nwant: %q", string(data), string(customContent))
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
		if code := runInit(ctx, nil); code != 0 {
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

	if code := runInit(ctx, nil); code != 0 {
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
	code := runInit(ctx, nil)
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

	// Skip skill prompt during init.
	setStdin(t, "n\n")

	code := runInit(ctx, nil)
	if code != 0 {
		t.Errorf("runInit() with .git = %d, want 0", code)
	}
	if strings.Contains(logBuf.String(), "WARN") {
		t.Errorf("unexpected WARN in log output with .git present:\n%s", logBuf.String())
	}
}

// ---------------------------------------------------------------------------
// Skill installation tests
// ---------------------------------------------------------------------------

// setStdin replaces stdin with the given content and restores it on cleanup.
func setStdin(t *testing.T, content string) {
	t.Helper()
	old := stdin
	stdin = strings.NewReader(content)
	t.Cleanup(func() { stdin = old })
}

func TestInstallSkill_Global(t *testing.T) {
	t.Chdir(t.TempDir())

	// Use a temp dir as the fake home so we don't touch the real one.
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)

	ctx := context.Background()
	code := installSkill(ctx, true)
	if code != 0 {
		t.Fatalf("installSkill(global) = %d, want 0", code)
	}

	skillPath := filepath.Join(fakeHome, ".agents", "skills", "reins", "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("skill file not found at %s: %v", skillPath, err)
	}
	if !strings.Contains(string(data), "name: reins") {
		t.Error("skill file missing expected frontmatter")
	}
}

func TestInstallSkill_Local(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := installSkill(ctx, false)
	if code != 0 {
		t.Fatalf("installSkill(local) = %d, want 0", code)
	}

	skillPath := filepath.Join(".agents", "skills", "reins", "SKILL.md")
	data, err := os.ReadFile(skillPath)
	if err != nil {
		t.Fatalf("skill file not found at %s: %v", skillPath, err)
	}
	if !strings.Contains(string(data), "name: reins") {
		t.Error("skill file missing expected frontmatter")
	}
}

func TestPromptSkillLocation_Global(t *testing.T) {
	setStdin(t, "g\n")
	got := promptSkillLocation()
	if got != "g" {
		t.Errorf("promptSkillLocation() = %q, want %q", got, "g")
	}
}

func TestPromptSkillLocation_Local(t *testing.T) {
	setStdin(t, "l\n")
	got := promptSkillLocation()
	if got != "l" {
		t.Errorf("promptSkillLocation() = %q, want %q", got, "l")
	}
}

func TestPromptSkillLocation_Skip(t *testing.T) {
	setStdin(t, "n\n")
	got := promptSkillLocation()
	if got != "n" {
		t.Errorf("promptSkillLocation() = %q, want %q", got, "n")
	}
}

func TestPromptSkillLocation_TTYFallback(t *testing.T) {
	// Simulate stdin being a closed pipe (EOF), with /dev/tty providing input.
	old := stdin
	oldTTY := openTTY
	stdin = strings.NewReader("") // empty — triggers fallback
	openTTY = func() (io.ReadCloser, error) {
		return io.NopCloser(strings.NewReader("g\n")), nil
	}
	t.Cleanup(func() { stdin = old; openTTY = oldTTY })

	got := promptSkillLocation()
	if got != "g" {
		t.Errorf("promptSkillLocation() TTY fallback = %q, want %q", got, "g")
	}
}

func TestPromptSkillLocation_NoTTYSkips(t *testing.T) {
	// When stdin is EOF and /dev/tty is unavailable, default to "n".
	old := stdin
	stdin = strings.NewReader("")
	t.Cleanup(func() { stdin = old })
	// openTTY already returns error from TestMain.

	got := promptSkillLocation()
	if got != "n" {
		t.Errorf("promptSkillLocation() no TTY = %q, want %q", got, "n")
	}
}

func TestRunInit_InstallsSkillWhenChosen(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	// Choose local skill installation.
	setStdin(t, "l\n")

	code := runInit(ctx, nil)
	if code != 0 {
		t.Fatalf("runInit() = %d, want 0", code)
	}

	skillPath := filepath.Join(".agents", "skills", "reins", "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Errorf("expected local skill at %s: %v", skillPath, err)
	}
}

func TestRunInit_SkipsSkillWhenDeclined(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	setStdin(t, "n\n")

	code := runInit(ctx, nil)
	if code != 0 {
		t.Fatalf("runInit() = %d, want 0", code)
	}

	skillPath := filepath.Join(".agents", "skills", "reins", "SKILL.md")
	if _, err := os.Stat(skillPath); err == nil {
		t.Error("skill file should not exist when user chose 'n'")
	}
}

func TestRunUpdate_RefreshesGlobalSkill(t *testing.T) {
	t.Chdir(t.TempDir())
	fakeHome := t.TempDir()
	t.Setenv("HOME", fakeHome)
	ctx := context.Background()

	SetVersion("1.0.0")

	// Init with skip-skill to avoid prompt, then manually place a global skill.
	setStdin(t, "n\n")
	if code := runInit(ctx, nil); code != 0 {
		t.Fatalf("runInit() = %d, want 0", code)
	}

	globalSkill := filepath.Join(fakeHome, ".agents", "skills", "reins", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(globalSkill), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(globalSkill, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	SetVersion("1.1.0")
	t.Cleanup(func() { SetVersion("dev") })

	code := runUpdate(ctx)
	if code != 0 {
		t.Fatalf("runUpdate() = %d, want 0", code)
	}

	data, err := os.ReadFile(globalSkill)
	if err != nil {
		t.Fatalf("global skill not found after update: %v", err)
	}
	if string(data) == "old" {
		t.Error("global skill was not refreshed by update")
	}
}

func TestRunUpdate_RefreshesLocalSkill(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	SetVersion("1.0.0")

	setStdin(t, "n\n")
	if code := runInit(ctx, nil); code != 0 {
		t.Fatalf("runInit() = %d, want 0", code)
	}

	localSkill := filepath.Join(".agents", "skills", "reins", "SKILL.md")
	if err := os.MkdirAll(filepath.Dir(localSkill), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(localSkill, []byte("old"), 0o644); err != nil {
		t.Fatal(err)
	}

	SetVersion("1.1.0")
	t.Cleanup(func() { SetVersion("dev") })

	code := runUpdate(ctx)
	if code != 0 {
		t.Fatalf("runUpdate() = %d, want 0", code)
	}

	data, err := os.ReadFile(localSkill)
	if err != nil {
		t.Fatalf("local skill not found after update: %v", err)
	}
	if string(data) == "old" {
		t.Error("local skill was not refreshed by update")
	}
}

func TestRun_SkillCommand(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	// Choose local.
	setStdin(t, "l\n")

	code := Run(ctx, []string{"reins", "skill"})
	if code != 0 {
		t.Errorf("Run skill = %d, want 0", code)
	}

	skillPath := filepath.Join(".agents", "skills", "reins", "SKILL.md")
	if _, err := os.Stat(skillPath); err != nil {
		t.Errorf("expected skill at %s: %v", skillPath, err)
	}
}

// ---------------------------------------------------------------------------
// Language preset tests (--lang)
// ---------------------------------------------------------------------------

func TestRunInit_GoPresetCreatesGoTaskfile(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := runInit(ctx, []string{"--lang", "go"})
	if code != 0 {
		t.Fatalf("runInit(--lang go) = %d, want 0", code)
	}

	data, err := os.ReadFile("Taskfile.yml")
	if err != nil {
		t.Fatalf("failed to read Taskfile.yml: %v", err)
	}

	content := string(data)

	// Go preset Taskfile must NOT contain placeholder exit-1 errors.
	if strings.Contains(content, "exit 1") {
		t.Error("Go preset Taskfile.yml still contains placeholder 'exit 1'")
	}

	// Must contain real Go commands.
	for _, want := range []string{"gofmt", "go vet", "go test", "go fix"} {
		if !strings.Contains(content, want) {
			t.Errorf("Go preset Taskfile.yml missing %q", want)
		}
	}
}

func TestRunInit_GoPresetCreatesGoIndexYaml(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := runInit(ctx, []string{"--lang", "go"})
	if code != 0 {
		t.Fatalf("runInit(--lang go) = %d, want 0", code)
	}

	data, err := os.ReadFile("rules/INDEX.yaml")
	if err != nil {
		t.Fatalf("failed to read INDEX.yaml: %v", err)
	}

	content := string(data)

	// Go preset INDEX.yaml must have an uncommented Go trigger.
	if !strings.Contains(content, "trigger: \"**/*.go\"") {
		t.Error("Go preset INDEX.yaml missing uncommented Go trigger")
	}
	if !strings.Contains(content, "rules/specifics/go.md") {
		t.Error("Go preset INDEX.yaml missing go.md rule reference")
	}
}

func TestRunInit_GoPresetCopiesGoRules(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := runInit(ctx, []string{"--lang", "go"})
	if code != 0 {
		t.Fatalf("runInit(--lang go) = %d, want 0", code)
	}

	// Go rules should be auto-copied from templates.
	data, err := os.ReadFile("rules/specifics/go.md")
	if err != nil {
		t.Fatalf("expected rules/specifics/go.md to exist: %v", err)
	}

	if !strings.Contains(string(data), "S-GO-01") {
		t.Error("rules/specifics/go.md missing expected Go rule content")
	}
}

func TestRunInit_GoPresetCreatesADRDirectory(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := runInit(ctx, []string{"--lang", "go"})
	if code != 0 {
		t.Fatalf("runInit(--lang go) = %d, want 0", code)
	}

	info, err := os.Stat("docs/rationale")
	if err != nil {
		t.Fatalf("expected docs/rationale/ to exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("docs/rationale should be a directory")
	}
}

func TestRunInit_GoPresetAgentsMdHasTechStack(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := runInit(ctx, []string{"--lang", "go"})
	if code != 0 {
		t.Fatalf("runInit(--lang go) = %d, want 0", code)
	}

	data, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("failed to read AGENTS.md: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "Tech Stack") {
		t.Error("Go preset AGENTS.md missing 'Tech Stack' section")
	}
	if !strings.Contains(content, "Architecture Decision Records") {
		t.Error("Go preset AGENTS.md missing 'Architecture Decision Records' section")
	}
}

// ---------------------------------------------------------------------------
// Rust preset tests (--lang rust)
// ---------------------------------------------------------------------------

func TestRunInit_RustPresetCreatesRustTaskfile(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := runInit(ctx, []string{"--lang", "rust"})
	if code != 0 {
		t.Fatalf("runInit(--lang rust) = %d, want 0", code)
	}

	data, err := os.ReadFile("Taskfile.yml")
	if err != nil {
		t.Fatalf("failed to read Taskfile.yml: %v", err)
	}

	content := string(data)

	// Rust preset Taskfile must NOT contain placeholder exit-1 errors.
	if strings.Contains(content, "exit 1") {
		t.Error("Rust preset Taskfile.yml still contains placeholder 'exit 1'")
	}

	// Must contain real Rust/cargo commands.
	for _, want := range []string{"cargo fmt", "cargo clippy", "cargo test", "cargo build"} {
		if !strings.Contains(content, want) {
			t.Errorf("Rust preset Taskfile.yml missing %q", want)
		}
	}
}

func TestRunInit_RustPresetCreatesRustIndexYaml(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := runInit(ctx, []string{"--lang", "rust"})
	if code != 0 {
		t.Fatalf("runInit(--lang rust) = %d, want 0", code)
	}

	data, err := os.ReadFile("rules/INDEX.yaml")
	if err != nil {
		t.Fatalf("failed to read INDEX.yaml: %v", err)
	}

	content := string(data)

	// Rust preset INDEX.yaml must have an uncommented Rust trigger.
	if !strings.Contains(content, "trigger: \"**/*.rs\"") {
		t.Error("Rust preset INDEX.yaml missing uncommented Rust trigger")
	}
	if !strings.Contains(content, "rules/specifics/rust.md") {
		t.Error("Rust preset INDEX.yaml missing rust.md rule reference")
	}
}

func TestRunInit_RustPresetCopiesRustRules(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := runInit(ctx, []string{"--lang", "rust"})
	if code != 0 {
		t.Fatalf("runInit(--lang rust) = %d, want 0", code)
	}

	// Rust rules should be auto-copied from templates.
	data, err := os.ReadFile("rules/specifics/rust.md")
	if err != nil {
		t.Fatalf("expected rules/specifics/rust.md to exist: %v", err)
	}

	if !strings.Contains(string(data), "S-RS-01") {
		t.Error("rules/specifics/rust.md missing expected Rust rule content")
	}
}

func TestRunInit_RustPresetCreatesADRDirectory(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := runInit(ctx, []string{"--lang", "rust"})
	if code != 0 {
		t.Fatalf("runInit(--lang rust) = %d, want 0", code)
	}

	info, err := os.Stat("docs/rationale")
	if err != nil {
		t.Fatalf("expected docs/rationale/ to exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("docs/rationale should be a directory")
	}
}

func TestRunInit_RustPresetAgentsMdHasTechStack(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := runInit(ctx, []string{"--lang", "rust"})
	if code != 0 {
		t.Fatalf("runInit(--lang rust) = %d, want 0", code)
	}

	data, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("failed to read AGENTS.md: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "Tech Stack") {
		t.Error("Rust preset AGENTS.md missing 'Tech Stack' section")
	}
	if !strings.Contains(content, "ratatui") {
		t.Error("Rust preset AGENTS.md missing 'ratatui' in tech stack")
	}
	if !strings.Contains(content, "Architecture Decision Records") {
		t.Error("Rust preset AGENTS.md missing 'Architecture Decision Records' section")
	}
}

func TestRunInit_RustPresetViaRunDispatcher(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := Run(ctx, []string{"reins", "init", "--lang", "rust"})
	if code != 0 {
		t.Errorf("Run init --lang rust = %d, want 0", code)
	}

	data, err := os.ReadFile("Taskfile.yml")
	if err != nil {
		t.Fatalf("failed to read Taskfile.yml: %v", err)
	}
	if strings.Contains(string(data), "exit 1") {
		t.Error("Taskfile.yml should be the Rust preset, not the generic skeleton")
	}
}

func TestRunInit_InvalidLangFails(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	logBuf := newTestLogger(t)
	code := runInit(ctx, []string{"--lang", "cobol"})
	if code != 1 {
		t.Errorf("runInit(--lang cobol) = %d, want 1", code)
	}

	if !strings.Contains(logBuf.String(), "unknown language preset") {
		t.Errorf("expected 'unknown language preset' in log, got:\n%s", logBuf.String())
	}
}

func TestRunInit_NoLangStillWorks(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := runInit(ctx, nil)
	if code != 0 {
		t.Fatalf("runInit(no lang) = %d, want 0", code)
	}

	// Should have the generic skeleton Taskfile (with exit 1).
	data, err := os.ReadFile("Taskfile.yml")
	if err != nil {
		t.Fatalf("failed to read Taskfile.yml: %v", err)
	}
	if !strings.Contains(string(data), "exit 1") {
		t.Error("generic init should have placeholder 'exit 1' in Taskfile.yml")
	}

	// Should NOT have docs/rationale (that's preset-only).
	if _, err := os.Stat("docs/rationale"); err == nil {
		t.Error("generic init should not create docs/rationale/")
	}

	// Should NOT have rules/specifics/go.md auto-copied.
	if _, err := os.Stat("rules/specifics/go.md"); err == nil {
		t.Error("generic init should not auto-copy rules/specifics/go.md")
	}
}

func TestRunInit_LangViaRunDispatcher(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := Run(ctx, []string{"reins", "init", "--lang", "go"})
	if code != 0 {
		t.Errorf("Run init --lang go = %d, want 0", code)
	}

	data, err := os.ReadFile("Taskfile.yml")
	if err != nil {
		t.Fatalf("failed to read Taskfile.yml: %v", err)
	}
	if strings.Contains(string(data), "exit 1") {
		t.Error("Taskfile.yml should be the Go preset, not the generic skeleton")
	}
}

// ---------------------------------------------------------------------------
// Zig preset tests (--lang zig)
// ---------------------------------------------------------------------------

func TestRunInit_ZigPresetCreatesZigTaskfile(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := runInit(ctx, []string{"--lang", "zig"})
	if code != 0 {
		t.Fatalf("runInit(--lang zig) = %d, want 0", code)
	}

	data, err := os.ReadFile("Taskfile.yml")
	if err != nil {
		t.Fatalf("failed to read Taskfile.yml: %v", err)
	}

	content := string(data)

	// Zig preset Taskfile must NOT contain placeholder exit-1 errors.
	if strings.Contains(content, "exit 1") {
		t.Error("Zig preset Taskfile.yml still contains placeholder 'exit 1'")
	}

	// Must contain real Zig commands.
	for _, want := range []string{"zig fmt", "zig build", "zig build test"} {
		if !strings.Contains(content, want) {
			t.Errorf("Zig preset Taskfile.yml missing %q", want)
		}
	}
}

func TestRunInit_ZigPresetCreatesZigIndexYaml(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := runInit(ctx, []string{"--lang", "zig"})
	if code != 0 {
		t.Fatalf("runInit(--lang zig) = %d, want 0", code)
	}

	data, err := os.ReadFile("rules/INDEX.yaml")
	if err != nil {
		t.Fatalf("failed to read INDEX.yaml: %v", err)
	}

	content := string(data)

	// Zig preset INDEX.yaml must have an uncommented Zig trigger.
	if !strings.Contains(content, "trigger: \"**/*.zig\"") {
		t.Error("Zig preset INDEX.yaml missing uncommented Zig trigger")
	}
	if !strings.Contains(content, "rules/specifics/zig.md") {
		t.Error("Zig preset INDEX.yaml missing zig.md rule reference")
	}
}

func TestRunInit_ZigPresetCopiesZigRules(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := runInit(ctx, []string{"--lang", "zig"})
	if code != 0 {
		t.Fatalf("runInit(--lang zig) = %d, want 0", code)
	}

	// Zig rules should be auto-copied from templates.
	data, err := os.ReadFile("rules/specifics/zig.md")
	if err != nil {
		t.Fatalf("expected rules/specifics/zig.md to exist: %v", err)
	}

	if !strings.Contains(string(data), "S-ZIG-01") {
		t.Error("rules/specifics/zig.md missing expected Zig rule content")
	}
}

func TestRunInit_ZigPresetCreatesADRDirectory(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := runInit(ctx, []string{"--lang", "zig"})
	if code != 0 {
		t.Fatalf("runInit(--lang zig) = %d, want 0", code)
	}

	info, err := os.Stat("docs/rationale")
	if err != nil {
		t.Fatalf("expected docs/rationale/ to exist: %v", err)
	}
	if !info.IsDir() {
		t.Error("docs/rationale should be a directory")
	}
}

func TestRunInit_ZigPresetAgentsMdHasTechStack(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := runInit(ctx, []string{"--lang", "zig"})
	if code != 0 {
		t.Fatalf("runInit(--lang zig) = %d, want 0", code)
	}

	data, err := os.ReadFile("AGENTS.md")
	if err != nil {
		t.Fatalf("failed to read AGENTS.md: %v", err)
	}

	content := string(data)

	if !strings.Contains(content, "Tech Stack") {
		t.Error("Zig preset AGENTS.md missing 'Tech Stack' section")
	}
	if !strings.Contains(content, "build.zig") {
		t.Error("Zig preset AGENTS.md missing 'build.zig' in tech stack")
	}
	if !strings.Contains(content, "Architecture Decision Records") {
		t.Error("Zig preset AGENTS.md missing 'Architecture Decision Records' section")
	}
}

func TestRunInit_ZigPresetViaRunDispatcher(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	code := Run(ctx, []string{"reins", "init", "--lang", "zig"})
	if code != 0 {
		t.Errorf("Run init --lang zig = %d, want 0", code)
	}

	data, err := os.ReadFile("Taskfile.yml")
	if err != nil {
		t.Fatalf("failed to read Taskfile.yml: %v", err)
	}
	if strings.Contains(string(data), "exit 1") {
		t.Error("Taskfile.yml should be the Zig preset, not the generic skeleton")
	}
}

// ---------------------------------------------------------------------------
// Lens subcommand tests (reins lens)
// ---------------------------------------------------------------------------

// captureStdout runs fn and returns whatever it wrote to os.Stdout.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating pipe: %v", err)
	}
	defer r.Close()

	old := os.Stdout
	os.Stdout = w

	fn()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	if _, err := io.Copy(&buf, r); err != nil {
		t.Fatalf("reading pipe: %v", err)
	}
	return buf.String()
}

func TestRunLens_NoArgs_PrintsList(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	got := captureStdout(t, func() {
		code := Run(ctx, []string{"reins", "lens"})
		if code != 0 {
			t.Errorf("Run lens (no args) = %d, want 0", code)
		}
	})

	// Should print the lens listing, not write a file.
	if !strings.Contains(got, "Available analysis lenses") {
		t.Errorf("expected lens listing in stdout, got:\n%s", got)
	}
	if !strings.Contains(got, "Presets:") {
		t.Errorf("expected preset listing in stdout, got:\n%s", got)
	}
}

func TestRunLens_RequiresInit(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	// No .reins/ directory — should fail.
	logBuf := newTestLogger(t)
	code := Run(ctx, []string{"reins", "lens", "--preset", "quick"})
	if code != 1 {
		t.Errorf("Run lens without init = %d, want 1", code)
	}
	if !strings.Contains(logBuf.String(), "not initialized") {
		t.Errorf("expected 'not initialized' in log, got:\n%s", logBuf.String())
	}
}

func TestRunLens_WritesFile(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	// Init first so .reins/ exists and rules/concerns/ is created.
	suppressOutput(t, func() {
		if code := runInit(ctx, nil); code != 0 {
			t.Fatalf("runInit() = %d, want 0", code)
		}
	})

	suppressOutput(t, func() {
		code := Run(ctx, []string{"reins", "lens", "--preset", "dd"})
		if code != 0 {
			t.Fatalf("Run lens --preset dd = %d, want 0", code)
		}
	})

	data, err := os.ReadFile("rules/concerns/analysis-lenses.md")
	if err != nil {
		t.Fatalf("expected analysis-lenses.md to exist: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "# Analysis Lenses") {
		t.Error("missing header in generated file")
	}
	if !strings.Contains(content, "C-LENS-04") {
		t.Error("missing Evidence Mapper rule ID (C-LENS-04)")
	}
	if !strings.Contains(content, "C-LENS-07") {
		t.Error("missing Weakness Spotter rule ID (C-LENS-07)")
	}
}

func TestRunLens_CustomOutput(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	suppressOutput(t, func() {
		if code := runInit(ctx, nil); code != 0 {
			t.Fatalf("runInit() = %d, want 0", code)
		}
	})

	customPath := "my/custom/lenses.md"
	suppressOutput(t, func() {
		code := Run(ctx, []string{"reins", "lens", "--preset", "quick", "--output", customPath})
		if code != 0 {
			t.Fatalf("Run lens --output custom = %d, want 0", code)
		}
	})

	data, err := os.ReadFile(customPath)
	if err != nil {
		t.Fatalf("expected %s to exist: %v", customPath, err)
	}
	if !strings.Contains(string(data), "# Analysis Lenses") {
		t.Error("custom output file missing header")
	}
}

func TestRunLens_IndividualLenses(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	suppressOutput(t, func() {
		if code := runInit(ctx, nil); code != 0 {
			t.Fatalf("runInit() = %d, want 0", code)
		}
	})

	suppressOutput(t, func() {
		code := Run(ctx, []string{"reins", "lens", "--lens", "synth", "--lens", "weak"})
		if code != 0 {
			t.Fatalf("Run lens --lens synth --lens weak = %d, want 0", code)
		}
	})

	data, err := os.ReadFile("rules/concerns/analysis-lenses.md")
	if err != nil {
		t.Fatalf("expected analysis-lenses.md: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "C-LENS-01") {
		t.Error("missing Expert Synthesizer (C-LENS-01)")
	}
	if !strings.Contains(content, "C-LENS-07") {
		t.Error("missing Weakness Spotter (C-LENS-07)")
	}
}

func TestRunLens_InvalidPreset(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	suppressOutput(t, func() {
		if code := runInit(ctx, nil); code != 0 {
			t.Fatalf("runInit() = %d, want 0", code)
		}
	})

	logBuf := newTestLogger(t)
	code := Run(ctx, []string{"reins", "lens", "--preset", "bogus"})
	if code != 1 {
		t.Errorf("Run lens --preset bogus = %d, want 1", code)
	}
	if !strings.Contains(logBuf.String(), "unknown preset") {
		t.Errorf("expected 'unknown preset' in log, got:\n%s", logBuf.String())
	}
}

func TestRunLens_RegistersInIndexYaml(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	suppressOutput(t, func() {
		if code := runInit(ctx, nil); code != 0 {
			t.Fatalf("runInit() = %d, want 0", code)
		}
	})

	suppressOutput(t, func() {
		code := Run(ctx, []string{"reins", "lens", "--preset", "dd"})
		if code != 0 {
			t.Fatalf("Run lens --preset dd = %d, want 0", code)
		}
	})

	data, err := os.ReadFile("rules/INDEX.yaml")
	if err != nil {
		t.Fatalf("failed to read INDEX.yaml: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, "rules/concerns/analysis-lenses.md") {
		t.Errorf("INDEX.yaml should contain lens file path after reins lens, got:\n%s", content)
	}
}

func TestRunLens_SkipsDuplicateRegistration(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	suppressOutput(t, func() {
		if code := runInit(ctx, nil); code != 0 {
			t.Fatalf("runInit() = %d, want 0", code)
		}
	})

	// Run lens twice.
	for range 2 {
		suppressOutput(t, func() {
			code := Run(ctx, []string{"reins", "lens", "--preset", "dd"})
			if code != 0 {
				t.Fatalf("Run lens --preset dd = %d, want 0", code)
			}
		})
	}

	data, err := os.ReadFile("rules/INDEX.yaml")
	if err != nil {
		t.Fatalf("failed to read INDEX.yaml: %v", err)
	}

	content := string(data)
	count := strings.Count(content, "rules/concerns/analysis-lenses.md")
	if count != 1 {
		t.Errorf("INDEX.yaml has %d occurrences of lens path, want exactly 1:\n%s", count, content)
	}
}

func TestRunLens_CustomOutput_RegistersCustomPath(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	suppressOutput(t, func() {
		if code := runInit(ctx, nil); code != 0 {
			t.Fatalf("runInit() = %d, want 0", code)
		}
	})

	customPath := "my/custom/lenses.md"
	suppressOutput(t, func() {
		code := Run(ctx, []string{"reins", "lens", "--preset", "quick", "--output", customPath})
		if code != 0 {
			t.Fatalf("Run lens --output custom = %d, want 0", code)
		}
	})

	data, err := os.ReadFile("rules/INDEX.yaml")
	if err != nil {
		t.Fatalf("failed to read INDEX.yaml: %v", err)
	}

	content := string(data)
	if !strings.Contains(content, customPath) {
		t.Errorf("INDEX.yaml should contain custom path %q, got:\n%s", customPath, content)
	}
	// Default path should NOT be registered.
	if strings.Contains(content, defaultLensOutput) {
		t.Errorf("INDEX.yaml should not contain default path when custom output was used:\n%s", content)
	}
}

func TestRunLens_InvalidLens(t *testing.T) {
	t.Chdir(t.TempDir())
	ctx := context.Background()

	suppressOutput(t, func() {
		if code := runInit(ctx, nil); code != 0 {
			t.Fatalf("runInit() = %d, want 0", code)
		}
	})

	logBuf := newTestLogger(t)
	code := Run(ctx, []string{"reins", "lens", "--lens", "bogus"})
	if code != 1 {
		t.Errorf("Run lens --lens bogus = %d, want 1", code)
	}
	if !strings.Contains(logBuf.String(), "unknown lens") {
		t.Errorf("expected 'unknown lens' in log, got:\n%s", logBuf.String())
	}
}
