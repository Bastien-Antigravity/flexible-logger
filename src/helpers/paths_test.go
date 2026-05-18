package helpers

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGetDefaultLogPath(t *testing.T) {
	path := GetDefaultLogPath()

	if !strings.HasSuffix(path, ".log") {
		t.Errorf("Expected path to end with .log, got %s", path)
	}

	logDir := filepath.Dir(path)
	if _, err := os.Stat(logDir); os.IsNotExist(err) {
		t.Errorf("Expected log directory %s to exist", logDir)
	}

	// Check if base name matches executable name
	exePath, _ := os.Executable()
	exeName := filepath.Base(exePath)
	ext := filepath.Ext(exeName)
	if ext != "" {
		exeName = strings.TrimSuffix(exeName, ext)
	}

	if !strings.Contains(path, exeName) {
		t.Errorf("Expected path %s to contain executable name %s", path, exeName)
	}
}
