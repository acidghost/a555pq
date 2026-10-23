package container

import (
	"time"

	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/container"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/spf13/cobra"
)

var latestMinReleaseAge time.Duration

var latestCmd = &cobra.Command{
	Use:   "latest <image>",
	Short: "Show the latest tag of a container image",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		imageName := args[0]

		client := container.NewClient()
		options := container.Options{MinReleaseAge: latestMinReleaseAge}
		tag, err := client.GetLatestTag(imageName, options)
		if err != nil {
			return err
		}

		output := &formatter.ContainerLatestOutput{
			Image:   imageName,
			Version: tag.Name,
			Digest:  tag.Digest,
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
	latestCmd.Flags().VarP(
		shared.NewTimespanValue(&latestMinReleaseAge),
		"min-release-age", "",
		"require Docker Hub tags last pushed at least this long ago (e.g. 7d, 6mo, 1y); unknown push times are excluded",
	)
	Cmd.AddCommand(latestCmd)
}
