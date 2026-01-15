package npm

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "npm",
	Short: "Query npm (Node Package Manager)",
	Long:  "Query package information from the npm registry.",
}
