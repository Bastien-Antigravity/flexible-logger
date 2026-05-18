package factory

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Bastien-Antigravity/flexible-logger/src/engine"
	"github.com/Bastien-Antigravity/flexible-logger/src/models"
)

type DummySink struct{}

func (d *DummySink) Write(entry *models.LogEntry) error { return nil }
func (d *DummySink) Close() error                       { return nil }

func TestCreateLogEngine_PopulatesMetadata(t *testing.T) {
	name := "TestApp"
	level := models.LevelInfo
	sink := &DummySink{}

	logger := CreateLogEngine(name, level, sink, true, 1.0)

	eng, ok := logger.(*engine.LogEngine)
	if !ok {
		t.Fatal("Expected *engine.LogEngine")
	}

	if eng.Name != name {
		t.Errorf("Expected name %s, got %s", name, eng.Name)
	}

	hostname, _ := os.Hostname()
	if eng.Hostname != hostname {
		t.Errorf("Expected hostname %s, got %s", hostname, eng.Hostname)
	}

	if eng.ProcessID != os.Getpid() {
		t.Errorf("Expected PID %d, got %d", os.Getpid(), eng.ProcessID)
	}

	procName := filepath.Base(os.Args[0])
	if eng.ProcessName != procName {
		t.Errorf("Expected process name %s, got %s", procName, eng.ProcessName)
	}
}
