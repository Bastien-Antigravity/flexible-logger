package models

// =============================================================================
// ESSENTIAL PROCESS: Data model representing high-priority alerts and notifications forwarded to notif-server.
//
// DATA FLOW:
//   1. Constructed by LogEngine upon encountering Warning or Error levels.
//   2. Decorated with tags, severity, and optional file attachments.
//   3. Dispatched via Notifier implementations.
//
// KEY PARAMETERS:
//   - NotifMessage: Alert payload structure.
// =============================================================================

type NotifMessage struct {
	Message    string   `json:"message"`
	Attachment string   `json:"attachment"`
	Tags       []string `json:"tags"`
	Level      string   `json:"level"`
}
