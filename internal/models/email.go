package models

import "time"

// EmailAddress represents an email address with optional name
type EmailAddress struct {
	Name  string
	Email string
}

// String returns a formatted email address
func (e EmailAddress) String() string {
	if e.Name != "" {
		return e.Name + " <" + e.Email + ">"
	}
	return e.Email
}

// ShortName returns just the name or email for display
func (e EmailAddress) ShortName() string {
	if e.Name != "" {
		return e.Name
	}
	return e.Email
}

// Email represents an email message
type Email struct {
	ID           string
	ThreadID     string
	MailboxIDs   []string
	From         []EmailAddress
	To           []EmailAddress
	CC           []EmailAddress
	BCC          []EmailAddress
	ReplyTo      []EmailAddress
	Subject      string
	Preview      string
	TextBody     string
	HTMLBody     string
	ReceivedAt   time.Time
	Size         int
	IsUnread     bool
	IsFlagged    bool
	IsDraft      bool
	HasAttachment bool
	Attachments  []Attachment
}

// Attachment represents an email attachment
type Attachment struct {
	BlobID   string
	Name     string
	Type     string
	Size     int
	IsInline bool
}

// FromDisplay returns the primary sender for display
func (e *Email) FromDisplay() string {
	if len(e.From) > 0 {
		return e.From[0].ShortName()
	}
	return "(unknown)"
}

// DateDisplay returns a formatted date for list view
func (e *Email) DateDisplay() string {
	now := time.Now()
	if e.ReceivedAt.Year() == now.Year() &&
		e.ReceivedAt.YearDay() == now.YearDay() {
		return e.ReceivedAt.Format("3:04 PM")
	}
	if e.ReceivedAt.Year() == now.Year() {
		return e.ReceivedAt.Format("Jan 2")
	}
	return e.ReceivedAt.Format("Jan 2, 2006")
}
