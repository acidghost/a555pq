package github

import (
	"sort"

	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/acidghost/a555pq/internal/github"
	"github.com/spf13/cobra"
)

var useRest bool

var versionsCmd = &cobra.Command{
	Use:   "versions <owner/repo>",
	Short: "Show all versions of a repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		repoName := args[0]

		client := github.NewClient(useRest)
		versions, err := client.GetVersions(repoName)
		if err != nil {
			return err
		}

		sort.Slice(versions, func(i, j int) bool {
			return versions[i].UploadDate > versions[j].UploadDate
		})

		output := &formatter.VersionsOutput{
			Package:  repoName,
			Versions: versions,
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
	versionsCmd.Flags().BoolVar(&useRest, "rest", false, "Use REST API instead of GraphQL (for unauthenticated requests)")
	Cmd.AddCommand(versionsCmd)
}
