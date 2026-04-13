package helpers

import (
	"os"
	"path/filepath"
	"strings"
)

// -----------------------------------------------------------------------------
// GetDefaultLogPath returns the default log file path:
// ./log/<executable_name>.log
func GetDefaultLogPath() string {
	exePath, err := os.Executable()
	if err != nil {
		// Fallback if we can't get executable path
		return "./logs/application.log"
	}

	exeName := filepath.Base(exePath)

	// Remove extension from exeName (e.g. main.exe -> main)
	ext := filepath.Ext(exeName)
	if ext != "" {
		exeName = strings.TrimSuffix(exeName, ext)
	}

	// Use Current Working Directory (CWD) for logs instead of EXE dir.
	// This avoids permission issues when the EXE is in a protected folder (e.g. Program Files).
	wd, err := os.Getwd()
	if err != nil {
		wd = "."
	}
	logDir := filepath.Join(wd, "logs")

	// Ensure log directory exists
	_ = os.MkdirAll(logDir, 0755)

	return filepath.Join(logDir, exeName+".log")
}
