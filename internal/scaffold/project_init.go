package scaffold

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const ProjectInitScript = "scripts/init-project.sh"

func RunProjectInitScript(ctx context.Context, targetDir, projectName string, dryRun bool, out io.Writer) (bool, error) {
	scriptPath := filepath.Join(targetDir, ProjectInitScript)
	info, err := os.Stat(scriptPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, fmt.Errorf("inspect project initialization script: %w", err)
	}
	if !info.Mode().IsRegular() {
		return false, fmt.Errorf("project initialization script is not a regular file: %s", scriptPath)
	}

	if dryRun {
		return true, nil
	}

	cmd := exec.CommandContext(ctx, scriptPath, projectName)
	cmd.Dir = targetDir
	output, err := cmd.CombinedOutput()
	if len(output) > 0 {
		fmt.Fprint(out, string(output))
	}
	if err != nil {
		return false, fmt.Errorf("run project initialization script: %w\n%s", err, strings.TrimSpace(string(output)))
	}

	return true, nil
}
