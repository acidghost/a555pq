package pypi

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "pypi",
	Short: "Query PyPI (Python Package Index)",
	Long:  "Query package information from the Python Package Index (PyPI).",
}
