package pypi

import (
	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/acidghost/a555pq/internal/pypi"
	"github.com/spf13/cobra"
)

var latestCmd = &cobra.Command{
	Use:   "latest <package>",
	Short: "Show latest version of a package",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		packageName := args[0]

		client := pypi.NewClient()
		version, err := client.GetLatestVersion(packageName)
		if err != nil {
			return err
		}

		output := &formatter.LatestOutput{
			Package: packageName,
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
