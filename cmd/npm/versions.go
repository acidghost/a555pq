package npm

import (
	"sort"

	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/acidghost/a555pq/internal/npm"
	"github.com/spf13/cobra"
)

var versionsCmd = &cobra.Command{
	Use:   "versions <package>",
	Short: "Show all versions of a package",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		packageName := args[0]

		client := npm.NewClient()
		versions, err := client.GetVersions(packageName)
		if err != nil {
			return err
		}

		sort.Slice(versions, func(i, j int) bool {
			return versions[i].UploadDate > versions[j].UploadDate
		})

		output := &formatter.VersionsOutput{
			Package:  packageName,
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
	Cmd.AddCommand(versionsCmd)
}
