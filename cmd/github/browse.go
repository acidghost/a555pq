package github

import (
	"fmt"

	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/spf13/cobra"
)

var browseCmd = &cobra.Command{
	Use:   "browse <owner/repo>",
	Short: "Open repository page on GitHub in browser",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		repoName := args[0]
		url := fmt.Sprintf("https://github.com/%s", repoName)

		opened := shared.OpenBrowser(url)

		if !opened {
			fmt.Println(url)
			return nil
		}

		output := &formatter.BrowseOutput{
			Package: repoName,
			URL:     url,
			Opened:  opened,
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
	Cmd.AddCommand(browseCmd)
}
