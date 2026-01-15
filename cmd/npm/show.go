package npm

import (
	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/acidghost/a555pq/internal/npm"
	"github.com/spf13/cobra"
)

var rawOutput bool

var showCmd = &cobra.Command{
	Use:   "show <package>",
	Short: "Show all info of a package",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		packageName := args[0]

		client := npm.NewClient()
		info, err := client.GetPackageInfo(packageName)
		if err != nil {
			return err
		}

		if rawOutput {
			f := formatter.NewJSONFormatter()
			return f.Format(info)
		}

		latestVersion := info.DistTags["latest"]

		latestInfo := info.Versions[latestVersion]

		dependencies := make([]string, 0, len(latestInfo.Dependencies))
		for dep := range latestInfo.Dependencies {
			dependencies = append(dependencies, dep)
		}

		var author, authorEmail string
		if info.Author.Name != "" {
			author = info.Author.Name
			authorEmail = info.Author.Email
		} else if latestInfo.Author.Name != "" {
			author = latestInfo.Author.Name
			authorEmail = latestInfo.Author.Email
		}

		output := &formatter.ShowOutput{
			Name:         info.Name,
			Version:      latestVersion,
			Description:  info.Readme,
			Author:       author,
			AuthorEmail:  authorEmail,
			License:      latestInfo.License,
			HomePage:     latestInfo.Homepage,
			Dependencies: dependencies,
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
	showCmd.Flags().BoolVar(&rawOutput, "raw", false, "Output raw JSON from npm")
	Cmd.AddCommand(showCmd)
}
