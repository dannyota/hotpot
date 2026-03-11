package logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"sync"
	"time"
)

// ANSI color codes.
const (
	reset   = "\033[0m"
	dim     = "\033[2m"
	yellow  = "\033[33m"
	cyan    = "\033[36m"
	boldRed = "\033[1;31m"
)

// Handler is a colored console handler for slog.
type Handler struct {
	level slog.Leveler
	attrs []slog.Attr
	w     io.Writer
	mu    *sync.Mutex
}

// New creates a slog.Logger with colored console output.
func New(level slog.Level) *slog.Logger {
	return slog.New(&Handler{
		level: level,
		w:     os.Stderr,
		mu:    &sync.Mutex{},
	})
}

func (h *Handler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level.Level()
}

func (h *Handler) Handle(_ context.Context, r slog.Record) error {
	var color string
	var label string
	switch {
	case r.Level >= slog.LevelError:
		color = boldRed
		label = "ERROR"
	case r.Level >= slog.LevelWarn:
		color = yellow
		label = "WARN "
	case r.Level >= slog.LevelInfo:
		color = cyan
		label = "INFO "
	default:
		color = dim
		label = "DEBUG"
	}

	ts := r.Time.Format(time.TimeOnly)

	// Build key-value string from pre-set attrs + record attrs.
	var kvs string
	writeAttr := func(a slog.Attr) {
		if a.Key == "" {
			return
		}
		kvs += fmt.Sprintf(" %s%s%s=%v", dim, a.Key, reset, a.Value)
	}
	for _, a := range h.attrs {
		writeAttr(a)
	}
	r.Attrs(func(a slog.Attr) bool {
		writeAttr(a)
		return true
	})

	line := fmt.Sprintf("%s%s%s %s%s%s %s%s\n",
		dim, ts, reset,
		color, label, reset,
		r.Message, kvs,
	)

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write([]byte(line))
	return err
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		level: h.level,
		attrs: append(h.attrs, attrs...),
		w:     h.w,
		mu:    h.mu,
	}
}

func (h *Handler) WithGroup(_ string) slog.Handler {
	return h // groups not needed for console output
}
