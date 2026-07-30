package registry

import (
	"context"
	"time"

	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/acidghost/a555pq/internal/registry"
	"github.com/spf13/cobra"
)

func newLatestCmd(ecosystem string) *cobra.Command {
	var minReleaseAge time.Duration

	cmd := &cobra.Command{
		Use:   "latest <package>",
		Short: "Show latest version of a package",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			client, err := registry.New(resolveEcosystem(ecosystem))
			if err != nil {
				return err
			}

			options := registry.Options{MinReleaseAge: minReleaseAge}
			output, err := client.Latest(context.Background(), args[0], options)
			if err != nil {
				return err
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

	cmd.Flags().VarP(
		shared.NewTimespanValue(&minReleaseAge),
		"min-release-age",
		"",
		"ignore versions released within this timespan when selecting the latest (e.g. 7d, 6mo, 1y); ecosystems without release timestamps are unaffected",
	)
	return cmd
}
