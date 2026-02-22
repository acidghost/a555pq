package cmd

import (
	"fmt"
	"os"

	"github.com/acidghost/a555pq/cmd/container"
	"github.com/acidghost/a555pq/cmd/github"
	"github.com/acidghost/a555pq/cmd/npm"
	"github.com/acidghost/a555pq/cmd/pypi"
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
	if err := RootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func init() {
	RootCmd.PersistentFlags().VarP(&shared.OutputFormat, "output", "o", "Output format (table|json)")

	RootCmd.AddCommand(container.Cmd)
	RootCmd.AddCommand(github.Cmd)
	RootCmd.AddCommand(npm.Cmd)
	RootCmd.AddCommand(pypi.Cmd)

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
