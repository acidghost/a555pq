package container

import (
	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/container"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/spf13/cobra"
)

var versionsCmd = &cobra.Command{
	Use:   "versions <image>",
	Short: "Show all tags of a container image",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		imageName := args[0]

		client := container.NewClient()
		tags, err := client.GetTags(imageName)
		if err != nil {
			return err
		}

		var versionItems []formatter.VersionItem
		for _, tag := range tags {
			versionItems = append(versionItems, formatter.VersionItem{
				Version:    tag.Name,
				UploadDate: tag.CreatedAt,
			})
		}

		output := &formatter.VersionsOutput{
			Package:  imageName,
			Versions: versionItems,
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
