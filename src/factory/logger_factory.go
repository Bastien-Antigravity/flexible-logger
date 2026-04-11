package factory

import (
	"os"
	"path/filepath"

	"github.com/Bastien-Antigravity/flexible-logger/src/engine"
	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

// -----------------------------------------------------------------------------
// CreateLogEngine creates a new fully configured LogEngine instance.
func CreateLogEngine(name string, level models.Level, sink interfaces.Sink, collectCallerInfo bool, samplingRate float64) interfaces.Logger {
	hostname, _ := os.Hostname()
	return &engine.LogEngine{
		Name:              name,
		Level:             level,
		Sink:              sink,
		Hostname:          hostname,
		ProcessID:         os.Getpid(),
		ProcessName:       filepath.Base(os.Args[0]),
		CollectCallerInfo: collectCallerInfo,
		SamplingRate:      samplingRate,
	}
}
