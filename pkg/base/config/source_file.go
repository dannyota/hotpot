package config

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/fsnotify/fsnotify"
)

// FileSource reads config from a JSON file.
type FileSource struct {
	path string
}

// NewFileSource creates a file-based config source.
func NewFileSource(path string) *FileSource {
	return &FileSource{path: path}
}

// Load reads config from the JSON file.
func (f *FileSource) Load(ctx context.Context) (*Config, error) {
	data, err := os.ReadFile(f.path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("parse config file: %w", err)
	}

	return &cfg, nil
}

// Watch watches the config file for changes using fsnotify.
func (f *FileSource) Watch(ctx context.Context, onChange func()) (func(), error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("create watcher: %w", err)
	}

	if err := watcher.Add(f.path); err != nil {
		watcher.Close()
		return nil, fmt.Errorf("watch file: %w", err)
	}

	done := make(chan struct{})

	go func() {
		defer watcher.Close()
		for {
			select {
			case <-ctx.Done():
				return
			case <-done:
				return
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				// Trigger on write or create (some editors delete and recreate)
				if event.Op&(fsnotify.Write|fsnotify.Create) != 0 {
					onChange()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				// Log but don't stop watching
				fmt.Printf("config watch error: %v\n", err)
			}
		}
	}()

	return func() { close(done) }, nil
}

// Type returns "file".
func (f *FileSource) Type() string {
	return "file"
}
