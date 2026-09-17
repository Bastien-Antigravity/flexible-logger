package factory

// =============================================================================
// ESSENTIAL PROCESS: Factory module for instantiating fully configured LogEngine instances with system metadata.
//
// DATA FLOW:
//   1. Discovers host OS parameters (hostname, process ID, executable name).
//   2. Configures LogEngine with designated sink, severity level, and sampling.
//   3. Returns initialized interfaces.Logger instance.
//
// KEY PARAMETERS:
//   - name: Application or logger identifier.
//   - level: Filtering threshold for log emission.
//   - sink: Root destination sink.
//   - collectCallerInfo: Flag controlling stack frame resolution.
//   - samplingRate: Frequency ratio for sampling low-severity logs.
// =============================================================================

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
