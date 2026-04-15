package container

import (
	"fmt"

	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/container"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/spf13/cobra"
)

var browseCmd = &cobra.Command{
	Use:   "browse <image>",
	Short: "Open container image page in browser",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		imageName := args[0]

		client := container.NewClient()
		url, err := client.GetBrowseURL(imageName)
		if err != nil {
			return err
		}

		opened := shared.OpenBrowser(url)

		if !opened {
			fmt.Println(url)
			return nil
		}

		output := &formatter.BrowseOutput{
			Package: imageName,
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
