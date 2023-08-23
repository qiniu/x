package shell

import (
	"os/exec"
	"runtime"
)

// Open opens the specified URL in the default browser of the user.
// See https://stackoverflow.com/questions/39320371/how-start-web-server-to-open-page-in-browser-in-golang
func Open(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default: // "linux", "freebsd", "openbsd", "netbsd"
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
