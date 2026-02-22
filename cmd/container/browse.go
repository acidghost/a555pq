package container

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/acidghost/a555pq/cmd/shared"
	"github.com/acidghost/a555pq/internal/container"
	"github.com/acidghost/a555pq/internal/formatter"
	"github.com/spf13/cobra"
)

var browseCmd = &cobra.Command{
	Use:   "browse <image>",
	Short: "Open container image page in browser",
	Args:  cobra.ExactArgs(1),
	RunE: func(_ *cobra.Command, args []string) error {
		imageName := args[0]

		client := container.NewClient()
		url, err := client.GetBrowseURL(imageName)
		if err != nil {
			return err
		}

		var opened bool
		var openErr error

		switch runtime.GOOS {
		case "darwin":
			//nolint:gosec
			openErr = exec.Command("open", url).Start()
		case "linux":
			//nolint:gosec
			openErr = exec.Command("xdg-open", url).Start()
		case "windows":
			//nolint:gosec
			openErr = exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
		default:
			openErr = fmt.Errorf("unsupported platform")
		}

		if openErr != nil {
			opened = false
			fmt.Fprintf(os.Stderr, "Warning: could not open browser: %v\n", openErr)
		} else {
			opened = true
		}

		if !opened {
			fmt.Println(url)
			return nil
		}

		output := &formatter.BrowseOutput{
			Package: imageName,
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
