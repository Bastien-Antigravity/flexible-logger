package sink

// =============================================================================
// ESSENTIAL PROCESS: Thread-safe rolling file writer rotating logs by file size with backup retention.
//
// DATA FLOW:
//   1. Receives byte slice to write.
//   2. Checks if writing payload exceeds maxBytes threshold.
//   3. Triggers rotation of historical backups (file.N -> file.N+1) if threshold exceeded.
//   4. Appends payload to active file handle.
//
// KEY PARAMETERS:
//   - filePath: Path to target active log file.
//   - maxBytes: Maximum size in bytes before file rotation is triggered.
//   - maxBackups: Number of rotated backup files to retain.
// =============================================================================

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
)

var _ io.WriteCloser = (*RollingFileWriter)(nil)

// RollingFileWriter implements io.WriteCloser with size-based rotation.
type RollingFileWriter struct {
	mu         sync.Mutex
	filePath   string
	maxBytes   int64
	maxBackups int
	file       *os.File
	curSize    int64
	closed     bool
}

// -----------------------------------------------------------------------------

// NewRollingFileWriter initializes a new RollingFileWriter.
func NewRollingFileWriter(filePath string, maxBytes int64, maxBackups int) (*RollingFileWriter, error) {
	if maxBytes <= 0 {
		maxBytes = 10 * 1024 * 1024 // 10 MB default
	}
	if maxBackups <= 0 {
		maxBackups = 5 // 5 backups default
	}

	dir := filepath.Dir(filePath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}

	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, err
	}

	return &RollingFileWriter{
		filePath:   filePath,
		maxBytes:   maxBytes,
		maxBackups: maxBackups,
		file:       f,
		curSize:    info.Size(),
	}, nil
}

// -----------------------------------------------------------------------------

// Write writes data to the current file, rotating beforehand if size exceeds maxBytes.
func (w *RollingFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return 0, os.ErrClosed
	}

	writeLen := int64(len(p))
	if w.curSize+writeLen > w.maxBytes && w.curSize > 0 {
		if err := w.rotate(); err != nil {
			fmt.Fprintf(os.Stderr, "RollingFileWriter: rotation failed: %v\n", err)
		}
	}

	n, err := w.file.Write(p)
	w.curSize += int64(n)
	return n, err
}

// -----------------------------------------------------------------------------

func (w *RollingFileWriter) rotate() error {
	if w.file != nil {
		_ = w.file.Close()
		w.file = nil
	}

	// Rotate backups: file.N -> file.N+1
	for i := w.maxBackups - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", w.filePath, i-1)
		newPath := fmt.Sprintf("%s.%d", w.filePath, i)
		if _, err := os.Stat(oldPath); err == nil {
			_ = os.Rename(oldPath, newPath)
		}
	}

	// Rename active file to .0
	_ = os.Rename(w.filePath, fmt.Sprintf("%s.0", w.filePath))

	// Reopen active file truncated
	f, err := os.OpenFile(w.filePath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	w.file = f
	w.curSize = 0
	return nil
}

// -----------------------------------------------------------------------------

// Close closes the underlying file handle.
func (w *RollingFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.closed {
		return nil
	}
	w.closed = true
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}
