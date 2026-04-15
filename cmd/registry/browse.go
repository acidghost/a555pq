package registry

import (
	"fmt"

	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/acidghost/a555pq/internal/registry"
	"github.com/spf13/cobra"
)

func newBrowseCmd(ecosystem string) *cobra.Command {
	return &cobra.Command{
		Use:   "browse <package>",
		Short: "Open package web page on registry in browser",
		Args:  cobra.ExactArgs(1),
		RunE: func(_ *cobra.Command, args []string) error {
			packageName := args[0]

			client, err := registry.New(resolveEcosystem(ecosystem))
			if err != nil {
				return err
			}

			url := client.BrowseURL(packageName)
			if url == "" {
				return fmt.Errorf("browse URL not available for %s", ecosystem)
			}

			opened := shared.OpenBrowser(url)

			if !opened {
				fmt.Println(url)
				return nil
			}

			output := &formatter.BrowseOutput{
				Package: packageName,
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
}
