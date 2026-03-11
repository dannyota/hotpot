package httputil

import (
	"fmt"
	"io"
	"log/slog"
	"time"
)

// ProgressReader wraps an io.Reader and periodically reports download progress
// via slog and an optional callback.
type ProgressReader struct {
	reader     io.Reader
	total      int64 // from Content-Length; 0 if unknown
	read       int64
	label      string
	interval   time.Duration
	lastReport time.Time
	onProgress func(string)
}

// NewProgressReader creates a reader that reports progress every interval.
// total is the expected byte count (0 if unknown). onProgress is called with
// a human-readable status string; it may be nil.
func NewProgressReader(r io.Reader, total int64, label string, interval time.Duration, onProgress func(string)) *ProgressReader {
	return &ProgressReader{
		reader:     r,
		total:      total,
		label:      label,
		interval:   interval,
		lastReport: time.Now(),
		onProgress: onProgress,
	}
}

func (pr *ProgressReader) Read(p []byte) (int, error) {
	n, err := pr.reader.Read(p)
	pr.read += int64(n)

	if time.Since(pr.lastReport) >= pr.interval {
		pr.lastReport = time.Now()
		pr.report()
	}

	return n, err
}

func (pr *ProgressReader) report() {
	mb := float64(pr.read) / 1024 / 1024

	if pr.total > 0 {
		pct := float64(pr.read) / float64(pr.total) * 100
		totalMB := float64(pr.total) / 1024 / 1024
		msg := fmt.Sprintf("%s: %.1f/%.1f MB (%.0f%%)", pr.label, mb, totalMB, pct)
		slog.Info("Download progress", "label", pr.label, "progress", fmt.Sprintf("%.0f%%", pct), "bytes", pr.read, "total", pr.total)
		if pr.onProgress != nil {
			pr.onProgress(msg)
		}
	} else {
		msg := fmt.Sprintf("%s: %.1f MB downloaded", pr.label, mb)
		slog.Info("Download progress", "label", pr.label, "bytes", pr.read)
		if pr.onProgress != nil {
			pr.onProgress(msg)
		}
	}
}
