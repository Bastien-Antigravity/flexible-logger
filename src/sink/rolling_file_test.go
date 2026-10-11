package sink

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func TestRollingFileWriter_Rotation(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "test.log")

	// Limit to 50 bytes with 3 backups
	writer, err := NewRollingFileWriter(logPath, 50, 3)
	if err != nil {
		t.Fatalf("Failed to create RollingFileWriter: %v", err)
	}
	defer writer.Close()

	// Write 30 bytes (no rotation yet)
	msg1 := []byte("123456789012345678901234567890") // 30 bytes
	if _, err := writer.Write(msg1); err != nil {
		t.Fatalf("Write 1 failed: %v", err)
	}

	if _, err := os.Stat(fmt.Sprintf("%s.0", logPath)); err == nil {
		t.Errorf("Backup file .0 should not exist yet")
	}

	// Write another 30 bytes (exceeds 50 bytes -> triggers rotation)
	msg2 := []byte("abcdefghijklmnopqrstuvwxyz1234") // 30 bytes
	if _, err := writer.Write(msg2); err != nil {
		t.Fatalf("Write 2 failed: %v", err)
	}

	// Now .0 should exist with msg1, active file with msg2
	if _, err := os.Stat(fmt.Sprintf("%s.0", logPath)); err != nil {
		t.Errorf("Backup file .0 should exist after rotation: %v", err)
	}

	activeContent, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read active log: %v", err)
	}
	if string(activeContent) != string(msg2) {
		t.Errorf("Active file content mismatch. Got: %s, Want: %s", string(activeContent), string(msg2))
	}

	backupContent, err := os.ReadFile(fmt.Sprintf("%s.0", logPath))
	if err != nil {
		t.Fatalf("Failed to read backup log: %v", err)
	}
	if string(backupContent) != string(msg1) {
		t.Errorf("Backup file content mismatch. Got: %s, Want: %s", string(backupContent), string(msg1))
	}
}
