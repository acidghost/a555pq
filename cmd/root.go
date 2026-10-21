package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime/pprof"

	"github.com/acidghost/a555pq/cmd/container"
	"github.com/acidghost/a555pq/cmd/github"
	"github.com/acidghost/a555pq/cmd/registry"
	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/spf13/cobra"
)

var (
	buildVersion = "dev"
	buildCommit  = "unknown"
	buildDate    = "unknown"
)

var RootCmd = &cobra.Command{
	Use:   "a555pq",
	Short: "Query different package indices",
	Long:  "A CLI tool to query package information from various package managers like PyPI, npm, container registries, and more.",
}

func SetBuildInfo(version, commit, date string) {
	buildVersion = version
	buildCommit = commit
	buildDate = date
}

func Execute() {
	if err := executeCommand(RootCmd); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// executeCommand keeps cleanup outside Cobra's post-run hooks: those hooks are
// skipped when a command or pre-run hook returns an error.
func executeCommand(root *cobra.Command) (err error) {
	var profile *os.File
	previousPreRunE := root.PersistentPreRunE
	previousPreRun := root.PersistentPreRun
	previousTraverse := cobra.EnableTraverseRunHooks
	// A child's persistent hook must not shadow the root's profiler hook.
	cobra.EnableTraverseRunHooks = true
	root.PersistentPreRunE = func(command *cobra.Command, args []string) error {
		path, err := command.Flags().GetString("pprof")
		if err != nil {
			return err
		}
		if path != "" {
			file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0600) // #nosec G304 -- --pprof explicitly chooses the path; O_EXCL prevents overwrites.
			if err != nil {
				return fmt.Errorf("create CPU profile %q: %w", path, err)
			}
			if err := pprof.StartCPUProfile(file); err != nil {
				return errors.Join(fmt.Errorf("start CPU profile: %w", err), file.Close())
			}
			profile = file
		}
		if previousPreRunE != nil {
			return previousPreRunE(command, args)
		}
		if previousPreRun != nil {
			previousPreRun(command, args)
		}
		return nil
	}
	defer func() {
		root.PersistentPreRunE = previousPreRunE
		cobra.EnableTraverseRunHooks = previousTraverse
		if profile != nil {
			pprof.StopCPUProfile()
			if closeErr := profile.Close(); closeErr != nil {
				err = errors.Join(err, fmt.Errorf("close CPU profile: %w", closeErr))
			}
		}
	}()
	return root.Execute()
}

func init() {
	RootCmd.PersistentFlags().VarP(&shared.OutputFormat, "output", "o", "Output format (table|json)")
	RootCmd.PersistentFlags().String("pprof", "", "Write a Go CPU profile to a new file (refuses to overwrite)")

	RootCmd.AddCommand(container.Cmd)
	RootCmd.AddCommand(github.Cmd)
	registry.RegisterCommands(RootCmd)

	RootCmd.AddCommand(versionCmd)

}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print version information",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf("Version: %s\n", buildVersion)
		fmt.Printf("Commit:  %s\n", buildCommit)
		fmt.Printf("Date:    %s\n", buildDate)
	},
}
