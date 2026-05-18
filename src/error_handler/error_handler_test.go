package error_handler

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

func TestReportInternalError_WritesToStderr(t *testing.T) {
	// Capture stderr
	old := os.Stderr
	r, w, _ := os.Pipe()
	os.Stderr = w

	ReportInternalError("TestLogger", "test-source", fmt.Errorf("simulated error"), "original message")

	w.Close()
	var buf bytes.Buffer
	io.Copy(&buf, r)
	os.Stderr = old

	output := buf.String()
	if !strings.Contains(output, "INTERNAL ERROR [test-source]: simulated error") {
		t.Errorf("Expected output to contain error message, got: %s", output)
	}
	if !strings.Contains(output, "Original: original message") {
		t.Errorf("Expected output to contain original message, got: %s", output)
	}
	if !strings.Contains(output, "TestLogger") {
		t.Errorf("Expected output to contain logger name, got: %s", output)
	}
}
