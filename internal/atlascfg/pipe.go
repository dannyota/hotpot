// Package atlascfg provides utilities for piping Atlas HCL configuration
// to the atlas CLI without exposing credentials in CLI args or on disk.
package atlascfg

import "os/exec"

// ConfigPipe returns a file URI pointing to the config content, a function to
// set up the command (e.g. attach stdin), and a cleanup function. The
// implementation is platform-specific (Unix: /dev/stdin, Windows: named pipe).
//
// Usage:
//
//	uri, setup, cleanup, err := ConfigPipe(config)
//	defer cleanup()
//	cmd := exec.Command("atlas", "--config", uri, ...)
//	setup(cmd)
//	cmd.Run()
func ConfigPipe(config string) (uri string, setupCmd func(*exec.Cmd), cleanup func(), err error) {
	return configPipe(config)
}
