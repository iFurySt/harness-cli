package cli

import (
	"context"
	"io"

	"github.com/iFurySt/harness-cli/internal/scaffold"
	"github.com/spf13/cobra"
)

type initOptions struct {
	language     string
	targetDir    string
	templateRoot string
	source       string
	force        bool
	dryRun       bool
	initGit      bool
}

// Execute runs the harness-cli command tree.
func Execute(ctx context.Context, version string, in io.Reader, out, errOut io.Writer) error {
	cmd := NewRootCommand(version, in, out, errOut)
	return cmd.ExecuteContext(ctx)
}

// NewRootCommand builds the command tree. The root command intentionally runs
// the init flow so `harness-cli` is enough for the common path.
func NewRootCommand(version string, in io.Reader, out, errOut io.Writer) *cobra.Command {
	rootOpts := initOptions{initGit: true}

	rootCmd := &cobra.Command{
		Use:           "harness-cli",
		Short:         "Initialize agent-first project repositories from Harness templates",
		Version:       version,
		SilenceUsage:  true,
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd.Context(), rootOpts, in, out)
		},
	}
	rootCmd.SetIn(in)
	rootCmd.SetOut(out)
	rootCmd.SetErr(errOut)
	bindInitFlags(rootCmd, &rootOpts)

	initCmd := newInitCommand(in, out)
	rootCmd.AddCommand(initCmd)

	return rootCmd
}

func newInitCommand(in io.Reader, out io.Writer) *cobra.Command {
	opts := initOptions{initGit: true}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the current repository from a Harness template",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd.Context(), opts, in, out)
		},
	}
	bindInitFlags(cmd, &opts)
	return cmd
}

func bindInitFlags(cmd *cobra.Command, opts *initOptions) {
	flags := cmd.Flags()
	flags.StringVarP(&opts.language, "language", "l", "", "template language: en or zh")
	flags.StringVarP(&opts.targetDir, "target", "t", ".", "directory to initialize")
	flags.StringVar(&opts.templateRoot, "template-root", "", "local directory containing harness-template and harness-template-cn")
	flags.StringVar(&opts.source, "source", "", "specific local template directory or git URL to use")
	flags.BoolVarP(&opts.force, "force", "f", false, "overwrite files that already exist with different content")
	flags.BoolVar(&opts.dryRun, "dry-run", false, "show what would change without writing files")
	flags.BoolVar(&opts.initGit, "git", true, "initialize git if the target is not already inside a work tree")
}

func runInit(ctx context.Context, opts initOptions, in io.Reader, out io.Writer) error {
	initializer := scaffold.NewInitializer(scaffold.IO{
		In:  in,
		Out: out,
	})

	return initializer.Run(ctx, scaffold.InitOptions{
		Language:     opts.language,
		TargetDir:    opts.targetDir,
		TemplateRoot: opts.templateRoot,
		Source:       opts.source,
		Force:        opts.force,
		DryRun:       opts.dryRun,
		InitGit:      opts.initGit,
	})
}
