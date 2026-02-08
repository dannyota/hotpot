//go:build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"unsafe"

	"golang.org/x/sys/windows"
)

// configPipe creates a Windows named pipe that serves the config string to
// Atlas. The config flows through kernel memory only — never touches disk or
// CLI args.
func configPipe(config string) (uri string, setupCmd func(*exec.Cmd), cleanup func(), err error) {
	pipeName := fmt.Sprintf(`\\.\pipe\hotpot-atlas-%d`, os.Getpid())
	pipeNamePtr, err := windows.UTF16PtrFromString(pipeName)
	if err != nil {
		return "", nil, nil, fmt.Errorf("pipe name: %w", err)
	}

	handle, err := windows.CreateNamedPipe(
		pipeNamePtr,
		windows.PIPE_ACCESS_OUTBOUND,
		windows.PIPE_TYPE_BYTE|windows.PIPE_WAIT,
		1,    // max instances
		4096, // out buffer
		4096, // in buffer
		0,    // default timeout
		nil,  // default security
	)
	if err != nil {
		return "", nil, nil, fmt.Errorf("create named pipe: %w", err)
	}

	// Atlas parses "file:////./pipe/hotpot-atlas-<pid>" via url.Parse, then
	// filepath.Join(u.Host, u.Path) which yields "\\.\pipe\hotpot-atlas-<pid>"
	// on Windows.
	uri = fmt.Sprintf("file:////./pipe/hotpot-atlas-%d", os.Getpid())

	setupCmd = func(cmd *exec.Cmd) {}

	cleanup = func() {
		windows.CloseHandle(handle)
	}

	// Serve the config in a goroutine — ConnectNamedPipe blocks until Atlas
	// opens the pipe, then we write the config and close our end.
	go func() {
		err := windows.ConnectNamedPipe(handle, nil)
		if err != nil {
			return
		}
		data := []byte(config)
		var written uint32
		_ = windows.WriteFile(handle, data, &written, nil)
		// Signal EOF by disconnecting
		disconnectNamedPipe(handle)
	}()

	return
}

// disconnectNamedPipe calls DisconnectNamedPipe to signal EOF to the reader.
func disconnectNamedPipe(handle windows.Handle) {
	procDisconnectNamedPipe.Call(uintptr(handle))
}

var (
	modkernel32            = windows.NewLazySystemDLL("kernel32.dll")
	procDisconnectNamedPipe = modkernel32.NewProc("DisconnectNamedPipe")
)

// Ensure unsafe is used (required for windows.Handle → uintptr conversion).
var _ = unsafe.Pointer(nil)
