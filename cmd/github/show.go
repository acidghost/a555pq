package github

import (
	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/acidghost/a555pq/internal/github"
	"github.com/spf13/cobra"
)

var rawOutput bool

var showCmd = &cobra.Command{
	Use:   "show <owner/repo>",
	Short: "Show all info of a repository",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		repoName := args[0]

		client := github.NewClient(false)
		repo, err := client.GetPackageInfo(repoName)
		if err != nil {
			return err
		}

		if rawOutput {
			f := formatter.NewJSONFormatter()
			return f.Format(repo)
		}

		var license string
		if repo.License != nil {
			license = repo.License.Name
		}

		latestVersion, err := client.GetLatestVersion(repoName)
		if err != nil {
			latestVersion = repo.DefaultBranch
		}

		dependencies := []string{}

		output := &formatter.ShowOutput{
			Name:         repo.FullName,
			Version:      latestVersion,
			Description:  repo.Description,
			Author:       repo.Owner.Login,
			AuthorEmail:  "",
			License:      license,
			HomePage:     repo.Homepage,
			Dependencies: dependencies,
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
	showCmd.Flags().BoolVar(&rawOutput, "raw", false, "Output raw JSON from GitHub")
	Cmd.AddCommand(showCmd)
}
