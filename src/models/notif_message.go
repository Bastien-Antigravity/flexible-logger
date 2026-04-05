package models

type NotifMessage struct {
	Message    string   `json:"message"`
	Attachment string   `json:"attachment"`
	Tags       []string `json:"tags"`
}
