package pypi

import (
	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/acidghost/a555pq/internal/pypi"
	"github.com/spf13/cobra"
)

var rawOutput bool

var showCmd = &cobra.Command{
	Use:   "show <package>",
	Short: "Show all info of a package",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		packageName := args[0]

		client := pypi.NewClient()
		info, err := client.GetPackageInfo(packageName)
		if err != nil {
			return err
		}

		if rawOutput {
			f := formatter.NewJSONFormatter()
			return f.Format(info)
		}

		output := &formatter.ShowOutput{
			Name:         info.Info.Name,
			Version:      info.Info.Version,
			Description:  truncateString(info.Info.Description, 200),
			Author:       info.Info.Author,
			AuthorEmail:  info.Info.AuthorEmail,
			License:      info.Info.License,
			HomePage:     info.Info.HomePage,
			Dependencies: info.Info.RequiresDist,
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
	showCmd.Flags().BoolVar(&rawOutput, "raw", false, "Output raw JSON from PyPI")
	Cmd.AddCommand(showCmd)
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}
