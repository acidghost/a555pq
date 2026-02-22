package container

import (
	"github.com/spf13/cobra"
)

var Cmd = &cobra.Command{
	Use:   "container",
	Short: "Query container registries",
	Long:  "Query container image information from various registries like Docker Hub, GHCR, GCR, ACR, ECR, and Quay.",
}
