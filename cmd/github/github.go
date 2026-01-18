package github

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "github",
	Short: "Query GitHub repositories",
	Long:  "Query repository information from GitHub.",
}
