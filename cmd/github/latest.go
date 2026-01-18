package github

import (
	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/acidghost/a555pq/internal/github"
	"github.com/spf13/cobra"
)

var latestCmd = &cobra.Command{
	Use:   "latest <owner/repo>",
	Short: "Show latest version of a repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		repoName := args[0]

		client := github.NewClient(false)
		version, err := client.GetLatestVersion(repoName)
		if err != nil {
			return err
		}

		output := &formatter.LatestOutput{
			Package: repoName,
			Version: version,
		}

		var f formatter.OutputFormatter
		if shared.OutputFormat == shared.JSON {
			f = formatter.NewJSONFormatter()
		} else {
			f = formatter.NewTableFormatter()
		}

		return f.Format(output)
	},
}

func init() {
	Cmd.AddCommand(latestCmd)
}
