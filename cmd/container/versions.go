package container

import (
	"fmt"
	"time"

	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/container"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/spf13/cobra"
)

var versionsMinReleaseAge time.Duration
var versionsMaxPages int

var versionsCmd = &cobra.Command{
	Use:   "versions <image>",
	Short: "Show all tags of a container image",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		imageName := args[0]

		client := container.NewClient()
		tags, err := client.GetTags(imageName, container.Options{MinReleaseAge: versionsMinReleaseAge, MaxPages: versionsMaxPages})
		if err != nil {
			return err
		}

		var versionItems []formatter.VersionItem
		for _, tag := range tags.Tags {
			versionItems = append(versionItems, formatter.VersionItem{
				Version:    tag.Name,
				UploadDate: tag.CreatedAt,
			})
		}

		output := &formatter.VersionsOutput{
			Package:   imageName,
			Versions:  versionItems,
			Truncated: tags.Truncated,
		}

		var f formatter.OutputFormatter
		if shared.OutputFormat == shared.JSON {
			f = formatter.NewJSONFormatter()
		} else {
			if tags.Truncated {
				fmt.Fprintf(cmd.ErrOrStderr(), "Warning: results truncated by --max-pages=%d; additional registry pages were not retrieved.\n", versionsMaxPages)
			}
			f = formatter.NewTableFormatter()
		}

		return f.Format(output)
	},
}

func init() {
	versionsCmd.Flags().IntVar(&versionsMaxPages, "max-pages", 0, "Limit Docker Hub metadata pages before age filtering (100 tags per page; 0 = unlimited)")
	versionsCmd.Flags().VarP(
		shared.NewTimespanValue(&versionsMinReleaseAge),
		"min-release-age", "",
		"require Docker Hub tags last pushed at least this long ago (e.g. 7d, 6mo, 1y); unknown push times are excluded",
	)
	Cmd.AddCommand(versionsCmd)
}
