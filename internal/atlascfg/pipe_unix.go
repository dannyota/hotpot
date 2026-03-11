//go:build !windows

package atlascfg

import (
	"fmt"
	"os"
	"os/exec"
)

// configPipe returns a file URI and setup function that delivers config to
// atlas. Uses a temp file so that atlas can still read stdin for interactive
// confirmation prompts (e.g. destructive migrations).
func configPipe(config string) (uri string, setupCmd func(*exec.Cmd), cleanup func(), err error) {
	f, err := os.CreateTemp("", "atlas-*.hcl")
	if err != nil {
		return "", nil, nil, fmt.Errorf("create temp config: %w", err)
	}
	if _, err := f.WriteString(config); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", nil, nil, fmt.Errorf("write temp config: %w", err)
	}
	f.Close()

	uri = "file://" + f.Name()
	setupCmd = func(cmd *exec.Cmd) {}
	cleanup = func() { os.Remove(f.Name()) }
	return
}
