package scaffold

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type SourceOptions struct {
	TemplateRoot string
	Source       string
}

type TemplateSource struct {
	Dir     string
	Display string
}

func ResolveTemplateSource(ctx context.Context, tpl Template, opts SourceOptions) (TemplateSource, func(), error) {
	if opts.Source != "" {
		return resolveExplicitSource(ctx, opts.Source)
	}

	if opts.TemplateRoot != "" {
		candidate := filepath.Join(opts.TemplateRoot, tpl.LocalName)
		if isDir(candidate) {
			return localSource(candidate), nil, nil
		}
		return TemplateSource{}, nil, fmt.Errorf("template %s not found under %s", tpl.LocalName, opts.TemplateRoot)
	}

	return cloneTemplate(ctx, tpl.RemoteURL)
}

func resolveExplicitSource(ctx context.Context, source string) (TemplateSource, func(), error) {
	if isGitURL(source) {
		return cloneTemplate(ctx, source)
	}

	if isDir(source) {
		return localSource(source), nil, nil
	}

	return TemplateSource{}, nil, fmt.Errorf("template source %q is not a directory or git URL", source)
}

func cloneTemplate(ctx context.Context, remoteURL string) (TemplateSource, func(), error) {
	tempDir, err := os.MkdirTemp("", "harness-template-*")
	if err != nil {
		return TemplateSource{}, nil, fmt.Errorf("create temporary template directory: %w", err)
	}

	cleanup := func() {
		_ = os.RemoveAll(tempDir)
	}

	cmd := exec.CommandContext(ctx, "git", "clone", "--depth=1", remoteURL, tempDir)
	if output, err := cmd.CombinedOutput(); err != nil {
		cleanup()
		return TemplateSource{}, nil, fmt.Errorf("clone template from %s: %w\n%s", remoteURL, err, strings.TrimSpace(string(output)))
	}

	return TemplateSource{Dir: tempDir, Display: remoteURL}, cleanup, nil
}

func localSource(dir string) TemplateSource {
	abs, err := filepath.Abs(dir)
	if err != nil {
		abs = dir
	}
	return TemplateSource{Dir: abs, Display: abs}
}

func isDir(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}

func isGitURL(value string) bool {
	if strings.HasPrefix(value, "git@") || strings.HasPrefix(value, "ssh://") {
		return true
	}

	parsed, err := url.Parse(value)
	if err != nil {
		return false
	}
	return parsed.Scheme == "http" || parsed.Scheme == "https" || parsed.Scheme == "git"
}
