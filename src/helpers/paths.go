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
		return "./log/application.log"
	}

	exeDir := filepath.Dir(exePath)
	exeName := filepath.Base(exePath)

	// Remove extension from exeName (e.g. main.exe -> main)
	ext := filepath.Ext(exeName)
	if ext != "" {
		exeName = strings.TrimSuffix(exeName, ext)
	}

	logDir := filepath.Join(exeDir, "log")

	// Ensure log directory exists
	_ = os.MkdirAll(logDir, 0755)

	return filepath.Join(logDir, exeName+".log")
}
