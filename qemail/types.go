package qemail

import "time"

// Message is the send payload for Mailbox.Send (From is taken from the Mailbox).
type Message struct {
	To          []string
	Cc          []string
	Bcc         []string
	Subject     string
	TextBody    string
	HtmlBody    string
	Attachments []Attachment
}

// MessageSummary holds key fields for displaying a message in a list view.
type MessageSummary struct {
	UID     string
	From    string
	To      []string
	Subject string
	Date    time.Time
	Seen    bool
	Flagged bool
	Size    uint32
}

// FullMessage represents a fully parsed email message with all MIME parts.
type FullMessage struct {
	From        string
	To          []string
	Cc          []string
	Subject     string
	Date        time.Time
	TextBody    string
	HtmlBody    string
	RawHeaders  string
	Attachments []Attachment
}

// Attachment holds the content of an email attachment.
type Attachment struct {
	Filename    string
	ContentType string
	Data        []byte
}
