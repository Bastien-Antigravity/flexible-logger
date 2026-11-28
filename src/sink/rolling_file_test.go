package sink

// =============================================================================
// ESSENTIAL PROCESS: Unit tests verifying RollingFileWriter size-based log rotation and rename-failure preservation.
//
// DATA FLOW:
//   1. Writes log payloads exceeding maxBytes threshold.
//   2. Verifies active and rotated backup file contents and rename-failure safety.
//
// KEY PARAMETERS:
//   - t: Testing harness.
// =============================================================================

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// -----------------------------------------------------------------------------

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

// -----------------------------------------------------------------------------

func TestRollingFileWriter_RotateRenameFailurePreservesActiveLog(t *testing.T) {
	tempDir := t.TempDir()
	logPath := filepath.Join(tempDir, "audit.log")

	// Pre-create a non-empty directory at logPath+".0" so active-file rename fails when maxBackups=1.
	if err := os.MkdirAll(filepath.Join(logPath+".0", "nested"), 0o755); err != nil {
		t.Fatalf("Failed to create blocking directory: %v", err)
	}

	writer, err := NewRollingFileWriter(logPath, 50, 1)
	if err != nil {
		t.Fatalf("Failed to create RollingFileWriter: %v", err)
	}
	defer writer.Close()

	msg1 := []byte("123456789012345678901234567890") // 30 bytes
	if _, err := writer.Write(msg1); err != nil {
		t.Fatalf("Write 1 failed: %v", err)
	}

	msg2 := []byte("abcdefghijklmnopqrstuvwxyz1234") // 30 bytes -> triggers rotation
	if _, err := writer.Write(msg2); err != nil {
		t.Fatalf("Write 2 failed: %v", err)
	}

	activeContent, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("Failed to read active log: %v", err)
	}
	expected := string(msg1) + string(msg2)
	if string(activeContent) != expected {
		t.Fatalf("Active log truncated or corrupted on rename failure! Got %q, want %q", string(activeContent), expected)
	}
}
