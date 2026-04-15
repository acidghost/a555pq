package registry

import (
	"fmt"

	"github.com/acidghost/a555pq/internal/registry"
	"github.com/spf13/cobra"
)

var aliases = map[string]string{
	"go":        "golang",
	"rubygems":  "gem",
	"ruby":      "gem",
	"packagist": "composer",
	"php":       "composer",
	"python":    "pypi",
	"rust":      "cargo",
	"brew":      "homebrew",
}

func resolveEcosystem(s string) string {
	if aliased, ok := aliases[s]; ok {
		return aliased
	}
	return s
}

func RegisterCommands(root *cobra.Command) {
	for _, eco := range registry.SupportedEcosystems() {
		ecoCmd := &cobra.Command{
			Use:   eco,
			Short: fmt.Sprintf("Query %s registry", eco),
		}
		ecoCmd.AddCommand(newShowCmd(eco))
		ecoCmd.AddCommand(newLatestCmd(eco))
		ecoCmd.AddCommand(newVersionsCmd(eco))
		ecoCmd.AddCommand(newBrowseCmd(eco))
		root.AddCommand(ecoCmd)
	}
}
