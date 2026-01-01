package factory

import (
	"os"

	"github.com/Bastien-Antigravity/flexible-logger/src/engine"
	"github.com/Bastien-Antigravity/flexible-logger/src/interfaces"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

// -----------------------------------------------------------------------------
// CreateLogEngine creates a new fully configured LogEngine instance.
func CreateLogEngine(name string, level models.Level, sink interfaces.Sink) interfaces.Logger {
	hostname, _ := os.Hostname()
	return &engine.LogEngine{
		Name:     name,
		Level:    level,
		Sink:     sink,
		Hostname: hostname,
	}
}
