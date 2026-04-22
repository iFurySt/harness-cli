package scaffold

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
)

type IO struct {
	In  io.Reader
	Out io.Writer
}

type Initializer struct {
	io IO
}

type InitOptions struct {
	Language     string
	TargetDir    string
	TemplateRoot string
	Source       string
	Force        bool
	DryRun       bool
	InitGit      bool
}

func NewInitializer(streams IO) *Initializer {
	if streams.Out == nil {
		streams.Out = io.Discard
	}
	return &Initializer{io: streams}
}

func (i *Initializer) Run(ctx context.Context, opts InitOptions) error {
	targetDir := opts.TargetDir
	if targetDir == "" {
		targetDir = "."
	}

	targetAbs, err := filepath.Abs(targetDir)
	if err != nil {
		return fmt.Errorf("resolve target directory: %w", err)
	}

	language, err := i.resolveLanguage(opts.Language)
	if err != nil {
		return err
	}

	tpl, err := TemplateFor(language)
	if err != nil {
		return err
	}

	source, cleanup, err := ResolveTemplateSource(ctx, tpl, SourceOptions{
		TemplateRoot: opts.TemplateRoot,
		Source:       opts.Source,
	})
	if cleanup != nil {
		defer cleanup()
	}
	if err != nil {
		return err
	}

	fmt.Fprintf(i.io.Out, "Using %s template from %s\n", tpl.Label, source.Display)

	report, err := CopyTemplate(CopyOptions{
		SourceDir: source.Dir,
		TargetDir: targetAbs,
		Force:     opts.Force,
		DryRun:    opts.DryRun,
	})
	if err != nil {
		return err
	}

	printCopyReport(i.io.Out, report, opts.DryRun)

	if opts.InitGit {
		initialized, err := EnsureGitRepository(ctx, targetAbs, opts.DryRun)
		if err != nil {
			return err
		}
		switch {
		case opts.DryRun && initialized:
			fmt.Fprintln(i.io.Out, "Would initialize git repository")
		case initialized:
			fmt.Fprintln(i.io.Out, "Initialized git repository")
		default:
			fmt.Fprintln(i.io.Out, "Git repository already present")
		}
	}

	return nil
}

func (i *Initializer) resolveLanguage(value string) (Language, error) {
	if value != "" {
		return ParseLanguage(value)
	}
	return promptLanguage(i.io.In, i.io.Out)
}

func printCopyReport(out io.Writer, report CopyReport, dryRun bool) {
	prefix := ""
	if dryRun {
		prefix = "Would "
	}

	fmt.Fprintf(out, "%scopy %d file(s)\n", prefix, report.Copied)
	if report.Updated > 0 {
		fmt.Fprintf(out, "%supdate %d file(s)\n", prefix, report.Updated)
	}
	if report.Unchanged > 0 {
		fmt.Fprintf(out, "Left %d unchanged file(s)\n", report.Unchanged)
	}
}
