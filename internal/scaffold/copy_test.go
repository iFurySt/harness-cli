package scaffold

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCopyTemplateCopiesFilesAndSkipsLocalMetadata(t *testing.T) {
	source := t.TempDir()
	target := t.TempDir()

	writeFile(t, filepath.Join(source, "README.md"), "hello\n")
	writeFile(t, filepath.Join(source, ".github", "workflows", "ci.yml"), "name: ci\n")
	writeFile(t, filepath.Join(source, ".git", "config"), "secret\n")
	writeFile(t, filepath.Join(source, ".idea", "workspace.xml"), "local\n")
	writeFile(t, filepath.Join(source, ".DS_Store"), "local\n")

	report, err := CopyTemplate(CopyOptions{SourceDir: source, TargetDir: target})
	if err != nil {
		t.Fatalf("CopyTemplate returned error: %v", err)
	}
	if report.Copied != 2 {
		t.Fatalf("Copied = %d, want 2", report.Copied)
	}

	assertFile(t, filepath.Join(target, "README.md"), "hello\n")
	assertFile(t, filepath.Join(target, ".github", "workflows", "ci.yml"), "name: ci\n")
	assertMissing(t, filepath.Join(target, ".git", "config"))
	assertMissing(t, filepath.Join(target, ".idea", "workspace.xml"))
	assertMissing(t, filepath.Join(target, ".DS_Store"))
}

func TestCopyTemplateConflictsUnlessForced(t *testing.T) {
	source := t.TempDir()
	target := t.TempDir()

	writeFile(t, filepath.Join(source, "README.md"), "from template\n")
	writeFile(t, filepath.Join(target, "README.md"), "from project\n")

	_, err := CopyTemplate(CopyOptions{SourceDir: source, TargetDir: target})
	var conflictErr ConflictError
	if !errors.As(err, &conflictErr) {
		t.Fatalf("CopyTemplate error = %v, want ConflictError", err)
	}

	report, err := CopyTemplate(CopyOptions{SourceDir: source, TargetDir: target, Force: true})
	if err != nil {
		t.Fatalf("CopyTemplate force returned error: %v", err)
	}
	if report.Updated != 1 {
		t.Fatalf("Updated = %d, want 1", report.Updated)
	}
	assertFile(t, filepath.Join(target, "README.md"), "from template\n")
}

func TestCopyTemplateDryRunDoesNotWrite(t *testing.T) {
	source := t.TempDir()
	target := t.TempDir()

	writeFile(t, filepath.Join(source, "README.md"), "hello\n")

	report, err := CopyTemplate(CopyOptions{SourceDir: source, TargetDir: target, DryRun: true})
	if err != nil {
		t.Fatalf("CopyTemplate returned error: %v", err)
	}
	if report.Copied != 1 {
		t.Fatalf("Copied = %d, want 1", report.Copied)
	}
	assertMissing(t, filepath.Join(target, "README.md"))
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", path, err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func assertFile(t *testing.T, path, want string) {
	t.Helper()
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile(%s): %v", path, err)
	}
	if string(got) != want {
		t.Fatalf("ReadFile(%s) = %q, want %q", path, got, want)
	}
}

func assertMissing(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("Stat(%s) error = %v, want not exist", path, err)
	}
}
