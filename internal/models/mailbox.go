package models

// Mailbox represents a mail folder
type Mailbox struct {
	ID          string
	Name        string
	Role        string // inbox, drafts, sent, trash, archive, junk
	ParentID    string
	TotalEmails int
	UnreadCount int
	SortOrder   int
}

// IsSystem returns true if this is a system mailbox
func (m *Mailbox) IsSystem() bool {
	return m.Role != ""
}

// DisplayName returns the name to show in the UI
func (m *Mailbox) DisplayName() string {
	switch m.Role {
	case "inbox":
		return "Inbox"
	case "drafts":
		return "Drafts"
	case "sent":
		return "Sent"
	case "trash":
		return "Trash"
	case "archive":
		return "Archive"
	case "junk":
		return "Junk"
	default:
		return m.Name
	}
}
