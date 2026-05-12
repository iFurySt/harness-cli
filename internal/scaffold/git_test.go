package scaffold

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestEnsureGitRepositoryInitializesTargetInsideParentWorkTree(t *testing.T) {
	ctx := context.Background()
	parent := t.TempDir()
	runGitCommand(t, parent, "init")

	target := filepath.Join(parent, "project")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", target, err)
	}

	initialized, err := EnsureGitRepository(ctx, target, false)
	if err != nil {
		t.Fatalf("EnsureGitRepository returned error: %v", err)
	}
	if !initialized {
		t.Fatal("EnsureGitRepository initialized = false, want true")
	}

	root, err := gitTopLevel(ctx, target)
	if err != nil {
		t.Fatalf("gitTopLevel returned error: %v", err)
	}
	targetPhysical, err := physicalPath(target)
	if err != nil {
		t.Fatalf("physicalPath(%s): %v", target, err)
	}
	rootPhysical, err := physicalPath(root)
	if err != nil {
		t.Fatalf("physicalPath(%s): %v", root, err)
	}
	if rootPhysical != targetPhysical {
		t.Fatalf("git root = %q, want %q", rootPhysical, targetPhysical)
	}
}

func runGitCommand(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmdArgs := append([]string{"-C", dir}, args...)
	output, err := exec.Command("git", cmdArgs...).CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", cmdArgs, err, output)
	}
}
