package scaffold

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

type CopyOptions struct {
	SourceDir string
	TargetDir string
	Force     bool
	DryRun    bool
}

type CopyReport struct {
	Copied    int
	Updated   int
	Unchanged int
}

type Conflict struct {
	Path   string
	Reason string
}

type ConflictError struct {
	Conflicts []Conflict
}

func (e ConflictError) Error() string {
	if len(e.Conflicts) == 0 {
		return "template conflicts with existing files"
	}

	message := "template conflicts with existing files:"
	for _, conflict := range e.Conflicts {
		message += fmt.Sprintf("\n  - %s: %s", conflict.Path, conflict.Reason)
	}
	message += "\nRun again with --force to overwrite conflicting files."
	return message
}

func CopyTemplate(opts CopyOptions) (CopyReport, error) {
	sourceAbs, err := filepath.Abs(opts.SourceDir)
	if err != nil {
		return CopyReport{}, fmt.Errorf("resolve source directory: %w", err)
	}

	targetAbs, err := filepath.Abs(opts.TargetDir)
	if err != nil {
		return CopyReport{}, fmt.Errorf("resolve target directory: %w", err)
	}

	if sourceAbs == targetAbs {
		return CopyReport{}, fmt.Errorf("source and target are the same directory: %s", sourceAbs)
	}

	var report CopyReport
	var conflicts []Conflict

	walkErr := filepath.WalkDir(sourceAbs, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}

		rel, err := filepath.Rel(sourceAbs, path)
		if err != nil {
			return err
		}
		if rel == "." {
			return nil
		}

		if shouldSkip(rel, entry) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		targetPath := filepath.Join(targetAbs, rel)

		info, err := entry.Info()
		if err != nil {
			return err
		}

		switch {
		case entry.IsDir():
			if opts.DryRun {
				return nil
			}
			return os.MkdirAll(targetPath, info.Mode().Perm())
		case info.Mode().Type() == 0:
			action, err := copyRegularFile(path, targetPath, info.Mode().Perm(), opts)
			if err != nil {
				var conflict Conflict
				if errors.As(err, &conflict) {
					conflicts = append(conflicts, conflict)
					return nil
				}
				return err
			}
			applyAction(&report, action)
			return nil
		case info.Mode()&os.ModeSymlink != 0:
			action, err := copySymlink(path, targetPath, opts)
			if err != nil {
				var conflict Conflict
				if errors.As(err, &conflict) {
					conflicts = append(conflicts, conflict)
					return nil
				}
				return err
			}
			applyAction(&report, action)
			return nil
		default:
			return nil
		}
	})
	if walkErr != nil {
		return report, walkErr
	}

	if len(conflicts) > 0 {
		return report, ConflictError{Conflicts: conflicts}
	}

	return report, nil
}

type copyAction int

const (
	actionNone copyAction = iota
	actionCopied
	actionUpdated
	actionUnchanged
)

func applyAction(report *CopyReport, action copyAction) {
	switch action {
	case actionCopied:
		report.Copied++
	case actionUpdated:
		report.Updated++
	case actionUnchanged:
		report.Unchanged++
	}
}

func copyRegularFile(sourcePath, targetPath string, mode fs.FileMode, opts CopyOptions) (copyAction, error) {
	exists, isSame, err := compareExistingFile(sourcePath, targetPath)
	if err != nil {
		return actionNone, err
	}
	if exists && isSame {
		return actionUnchanged, nil
	}
	if exists && !opts.Force {
		return actionNone, Conflict{Path: targetPath, Reason: "file already exists with different content"}
	}

	if opts.DryRun {
		if exists {
			return actionUpdated, nil
		}
		return actionCopied, nil
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return actionNone, err
	}

	source, err := os.Open(sourcePath)
	if err != nil {
		return actionNone, err
	}
	defer source.Close()

	target, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, mode)
	if err != nil {
		return actionNone, err
	}
	if _, err := io.Copy(target, source); err != nil {
		_ = target.Close()
		return actionNone, err
	}
	if err := target.Close(); err != nil {
		return actionNone, err
	}
	if err := os.Chmod(targetPath, mode); err != nil {
		return actionNone, err
	}

	if exists {
		return actionUpdated, nil
	}
	return actionCopied, nil
}

func compareExistingFile(sourcePath, targetPath string) (bool, bool, error) {
	targetInfo, err := os.Lstat(targetPath)
	if errors.Is(err, os.ErrNotExist) {
		return false, false, nil
	}
	if err != nil {
		return false, false, err
	}
	if !targetInfo.Mode().IsRegular() {
		return true, false, Conflict{Path: targetPath, Reason: "existing path is not a regular file"}
	}

	sourceBytes, err := os.ReadFile(sourcePath)
	if err != nil {
		return true, false, err
	}
	targetBytes, err := os.ReadFile(targetPath)
	if err != nil {
		return true, false, err
	}

	return true, bytes.Equal(sourceBytes, targetBytes), nil
}

func copySymlink(sourcePath, targetPath string, opts CopyOptions) (copyAction, error) {
	linkTarget, err := os.Readlink(sourcePath)
	if err != nil {
		return actionNone, err
	}

	exists := false
	existingTarget, err := os.Readlink(targetPath)
	if err == nil {
		exists = true
		if existingTarget == linkTarget {
			return actionUnchanged, nil
		}
		if !opts.Force {
			return actionNone, Conflict{Path: targetPath, Reason: "symlink already exists with different target"}
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		if _, statErr := os.Lstat(targetPath); statErr == nil {
			exists = true
		}
		if !opts.Force {
			return actionNone, Conflict{Path: targetPath, Reason: "existing path is not a symlink"}
		}
	} else {
		if _, statErr := os.Lstat(targetPath); statErr == nil {
			exists = true
			if !opts.Force {
				return actionNone, Conflict{Path: targetPath, Reason: "existing path is not a symlink"}
			}
		} else if !errors.Is(statErr, os.ErrNotExist) {
			return actionNone, statErr
		}
	}

	if opts.DryRun {
		if exists {
			return actionUpdated, nil
		}
		return actionCopied, nil
	}

	if err := os.MkdirAll(filepath.Dir(targetPath), 0o755); err != nil {
		return actionNone, err
	}
	if opts.Force {
		_ = os.Remove(targetPath)
	}
	if err := os.Symlink(linkTarget, targetPath); err != nil {
		return actionNone, err
	}
	if exists {
		return actionUpdated, nil
	}
	return actionCopied, nil
}

func shouldSkip(rel string, entry fs.DirEntry) bool {
	name := entry.Name()
	switch name {
	case ".git", ".idea", ".DS_Store":
		return true
	default:
		return false
	}
}

func (c Conflict) Error() string {
	return c.Reason
}
