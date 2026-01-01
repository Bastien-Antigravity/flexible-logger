package models

// -----------------------------------------------------------------------------
// Notification represents a message to be sent to a notification server
type Notification struct {
	Message    string
	Attachment string
	Tags       []string
}
