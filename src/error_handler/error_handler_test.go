package error_handler

// =============================================================================
// ESSENTIAL PROCESS: Unit tests for internal error reporting and fallback stderr logging mechanisms.
//
// DATA FLOW:
//   1. Redirects and intercepts standard error stream.
//   2. Invokes ReportInternalError with simulated infrastructure faults.
//   3. Validates formatted fallback error message delivery.
//
// KEY PARAMETERS:
//   - t: Testing context.
//   - source: Component identifier where simulated error originated.
// =============================================================================

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
