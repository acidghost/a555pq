package pypi

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/spf13/cobra"
)

var browseCmd = &cobra.Command{
	Use:   "browse <package>",
	Short: "Open package web page on index in browser",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		packageName := args[0]
		url := fmt.Sprintf("https://pypi.org/project/%s/", packageName)

		var opened bool
		var err error

		switch runtime.GOOS {
		case "darwin":
			//nolint:gosec
			err = exec.Command("open", url).Start()
		case "linux":
			//nolint:gosec
			err = exec.Command("xdg-open", url).Start()
		case "windows":
			//nolint:gosec
			err = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
		default:
			err = fmt.Errorf("unsupported platform")
		}

		if err != nil {
			opened = false
			fmt.Fprintf(os.Stderr, "Warning: could not open browser: %v\n", err)
		} else {
			opened = true
		}

		if !opened {
			fmt.Println(url)
			return nil
		}

		output := &formatter.BrowseOutput{
			Package: packageName,
			URL:     url,
			Opened:  opened,
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
	Cmd.AddCommand(browseCmd)
}
