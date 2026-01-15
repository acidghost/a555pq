package main

import (
	"github.com/acidghost/a555pq/cmd"
)

var (
	buildVersion = "dev"
	buildCommit  = "unknown"
	buildDate    = "unknown"
)

func main() {
	cmd.SetBuildInfo(buildVersion, buildCommit, buildDate)
	cmd.Execute()
}
