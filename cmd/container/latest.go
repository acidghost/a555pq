package container

import (
	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/container"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/spf13/cobra"
)

var latestCmd = &cobra.Command{
	Use:   "latest <image>",
	Short: "Show the latest tag of a container image",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		imageName := args[0]

		client := container.NewClient()
		tag, err := client.GetLatestTag(imageName)
		if err != nil {
			return err
		}

		output := &formatter.ContainerLatestOutput{
			Image:   imageName,
			Version: tag,
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
