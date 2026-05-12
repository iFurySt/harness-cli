package scaffold

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
)

const (
	InitialCommitSubject = "Init Commit"
	InitialCommitBody    = "Initialized with harness-cli: https://github.com/iFurySt/harness-cli"
)

type InitialCommitStatus int

const (
	InitialCommitCreated InitialCommitStatus = iota
	InitialCommitWouldCreate
	InitialCommitSkippedHasCommits
	InitialCommitSkippedNoChanges
)

func EnsureGitRepository(ctx context.Context, targetDir string, dryRun bool) (bool, error) {
	if isGitRepositoryRoot(ctx, targetDir) {
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

func CreateInitialCommit(ctx context.Context, targetDir string, dryRun bool) (InitialCommitStatus, error) {
	if dryRun {
		return InitialCommitWouldCreate, nil
	}

	if !isGitRepositoryRoot(ctx, targetDir) {
		return InitialCommitSkippedNoChanges, fmt.Errorf("target is not a git repository root: %s", targetDir)
	}

	hasCommits, err := repositoryHasCommits(ctx, targetDir)
	if err != nil {
		return InitialCommitSkippedNoChanges, err
	}
	if hasCommits {
		return InitialCommitSkippedHasCommits, nil
	}

	add := exec.CommandContext(ctx, "git", "-C", targetDir, "add", "-A")
	if output, err := add.CombinedOutput(); err != nil {
		return InitialCommitSkippedNoChanges, fmt.Errorf("stage initial files: %w\n%s", err, strings.TrimSpace(string(output)))
	}

	hasChanges, err := hasStagedChanges(ctx, targetDir)
	if err != nil {
		return InitialCommitSkippedNoChanges, err
	}
	if !hasChanges {
		return InitialCommitSkippedNoChanges, nil
	}

	args := gitCommitArgs(ctx, targetDir, "-m", InitialCommitSubject, "-m", InitialCommitBody)
	commit := exec.CommandContext(ctx, "git", args...)
	if output, err := commit.CombinedOutput(); err != nil {
		return InitialCommitSkippedNoChanges, fmt.Errorf("create initial git commit: %w\n%s", err, strings.TrimSpace(string(output)))
	}

	return InitialCommitCreated, nil
}

func isGitRepositoryRoot(ctx context.Context, targetDir string) bool {
	root, err := gitTopLevel(ctx, targetDir)
	if err != nil {
		return false
	}

	targetAbs, err := physicalPath(targetDir)
	if err != nil {
		return false
	}

	rootAbs, err := physicalPath(root)
	if err != nil {
		return false
	}

	return filepath.Clean(rootAbs) == filepath.Clean(targetAbs)
}

func gitTopLevel(ctx context.Context, targetDir string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", targetDir, "rev-parse", "--show-toplevel")
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func repositoryHasCommits(ctx context.Context, targetDir string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", targetDir, "rev-parse", "--verify", "HEAD")
	if err := cmd.Run(); err == nil {
		return true, nil
	} else if isExitError(err) {
		return false, nil
	} else {
		return false, fmt.Errorf("check git history: %w", err)
	}
}

func hasStagedChanges(ctx context.Context, targetDir string) (bool, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", targetDir, "diff", "--cached", "--quiet", "--exit-code")
	if err := cmd.Run(); err == nil {
		return false, nil
	} else if isExitError(err) {
		return true, nil
	} else {
		return false, fmt.Errorf("check staged changes: %w", err)
	}
}

func gitCommitArgs(ctx context.Context, targetDir string, args ...string) []string {
	gitArgs := []string{"-C", targetDir}
	if !gitConfigExists(ctx, targetDir, "user.name") {
		gitArgs = append(gitArgs, "-c", "user.name=harness-cli")
	}
	if !gitConfigExists(ctx, targetDir, "user.email") {
		gitArgs = append(gitArgs, "-c", "user.email=harness-cli@users.noreply.github.com")
	}
	gitArgs = append(gitArgs, "commit")
	gitArgs = append(gitArgs, args...)
	return gitArgs
}

func gitConfigExists(ctx context.Context, targetDir, key string) bool {
	cmd := exec.CommandContext(ctx, "git", "-C", targetDir, "config", "--get", key)
	output, err := cmd.Output()
	return err == nil && strings.TrimSpace(string(output)) != ""
}

func isExitError(err error) bool {
	var exitErr *exec.ExitError
	return errors.As(err, &exitErr)
}

func physicalPath(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", err
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return "", err
	}
	return resolved, nil
}
