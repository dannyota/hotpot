//go:build !windows

package atlascfg

import (
	"os/exec"
	"strings"
)

// configPipe returns a file URI and setup function that pipes config to the
// command via stdin. On Unix, this simply uses /dev/stdin.
func configPipe(config string) (uri string, setupCmd func(*exec.Cmd), cleanup func(), err error) {
	uri = "file:///dev/stdin"
	setupCmd = func(cmd *exec.Cmd) {
		cmd.Stdin = strings.NewReader(config)
	}
	cleanup = func() {}
	return
}
