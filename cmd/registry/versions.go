package registry

import (
	"context"

	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/acidghost/a555pq/internal/registry"
	"github.com/spf13/cobra"
)

func newVersionsCmd(ecosystem string) *cobra.Command {
	return &cobra.Command{
		Use:   "versions <package>",
		Short: "Show all versions of a package",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			client, err := registry.New(resolveEcosystem(ecosystem))
			if err != nil {
				return err
			}

			output, err := client.Versions(context.Background(), args[0])
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
}
