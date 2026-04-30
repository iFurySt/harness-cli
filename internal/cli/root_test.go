package cli

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRootCommandAcceptsPositionalTarget(t *testing.T) {
	source := testTemplateSource(t)
	target := filepath.Join(t.TempDir(), "project")

	_, _, err := executeRootCommand(t,
		target,
		"--language", "en",
		"--source", source,
		"--git=false",
	)
	if err != nil {
		t.Fatalf("execute root command: %v", err)
	}

	assertFileContent(t, filepath.Join(target, "README.md"), "hello\n")
}

func TestInitCommandAcceptsPositionalTarget(t *testing.T) {
	source := testTemplateSource(t)
	target := filepath.Join(t.TempDir(), "project")

	_, _, err := executeRootCommand(t,
		"init",
		target,
		"--language", "en",
		"--source", source,
		"--git=false",
	)
	if err != nil {
		t.Fatalf("execute init command: %v", err)
	}

	assertFileContent(t, filepath.Join(target, "README.md"), "hello\n")
}

func TestPositionalTargetConflictsWithTargetFlag(t *testing.T) {
	_, _, err := executeRootCommand(t, "project", "--target", "other")
	if err == nil {
		t.Fatal("expected positional target and --target conflict")
	}

	if got, want := err.Error(), "target specified both as positional argument"; !strings.Contains(got, want) {
		t.Fatalf("error = %q, want substring %q", got, want)
	}
}

func executeRootCommand(t *testing.T, args ...string) (string, string, error) {
	t.Helper()

	var out bytes.Buffer
	var errOut bytes.Buffer
	cmd := NewRootCommand("test", bytes.NewBuffer(nil), &out, &errOut)
	cmd.SetArgs(args)

	err := cmd.ExecuteContext(context.Background())
	return out.String(), errOut.String(), err
}

func testTemplateSource(t *testing.T) string {
	t.Helper()

	source := t.TempDir()
	if err := os.WriteFile(filepath.Join(source, "README.md"), []byte("hello\n"), 0o644); err != nil {
		t.Fatalf("write template file: %v", err)
	}
	return source
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if string(got) != want {
		t.Fatalf("%s = %q, want %q", path, string(got), want)
	}
}
