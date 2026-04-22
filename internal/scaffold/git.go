package scaffold

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

func EnsureGitRepository(ctx context.Context, targetDir string, dryRun bool) (bool, error) {
	if insideGitWorkTree(ctx, targetDir) {
		return false, nil
	}

	if dryRun {
		return true, nil
	}

	cmd := exec.CommandContext(ctx, "git", "-C", targetDir, "init")
	if output, err := cmd.CombinedOutput(); err != nil {
		return false, fmt.Errorf("initialize git repository: %w\n%s", err, strings.TrimSpace(string(output)))
	}

	return true, nil
}

func insideGitWorkTree(ctx context.Context, targetDir string) bool {
	cmd := exec.CommandContext(ctx, "git", "-C", targetDir, "rev-parse", "--is-inside-work-tree")
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) == "true"
}
