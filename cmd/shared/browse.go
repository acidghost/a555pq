package shared

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func OpenBrowser(url string) bool {
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
		fmt.Fprintf(os.Stderr, "Warning: could not open browser: %v\n", err)
		return false
	}
	return true
}
