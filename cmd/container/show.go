package container

import (
	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/container"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/spf13/cobra"
)

var rawOutput bool

var showCmd = &cobra.Command{
	Use:   "show <image>",
	Short: "Show detailed information about a container image",
	Long:  "Show detailed information about a container image. Tag can be included in image reference (e.g., nginx:latest). If tag is not specified, shows latest tag.",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		imageName := args[0]

		client := container.NewClient()
		info, err := client.GetImageInfo(imageName)
		if err != nil {
			return err
		}

		if rawOutput {
			f := formatter.NewJSONFormatter()
			return f.Format(info)
		}

		var digest string
		if info.Manifest != nil {
			digest = info.Manifest.Digest
		}

		output := &formatter.ContainerShowOutput{
			Name:         info.Name,
			Description:  info.Description,
			Tag:          info.LatestTag,
			TagDate:      info.TagDate,
			TagSize:      info.Size,
			Digest:       digest,
			Registry:     info.Registry,
			FullImageRef: info.FullImageRef,
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
	showCmd.Flags().BoolVar(&rawOutput, "raw", false, "Output raw JSON from registry")
	Cmd.AddCommand(showCmd)
}
