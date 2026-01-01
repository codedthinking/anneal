package jmap

import (
	"fmt"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
	"git.sr.ht/~rockorager/go-jmap/mail/mailbox"
	"github.com/koren/tuimail/internal/models"
)

// Client wraps the JMAP client for Fastmail
type Client struct {
	client    *jmap.Client
	accountID jmap.ID
	email     string
}

// New creates a new JMAP client for Fastmail
func New(emailAddr, token string) (*Client, error) {
	client := &jmap.Client{
		SessionEndpoint: "https://api.fastmail.com/jmap/session",
	}
	client.WithAccessToken(token)

	// Authenticate and get session
	if err := client.Authenticate(); err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	// Get account ID for mail
	accountID := client.Session.PrimaryAccounts[mail.URI]
	if accountID == "" {
		return nil, fmt.Errorf("no mail account found")
	}

	return &Client{
		client:    client,
		accountID: accountID,
		email:     emailAddr,
	}, nil
}

// GetMailboxes fetches all mailboxes for the account
func (c *Client) GetMailboxes() ([]models.Mailbox, error) {
	req := &jmap.Request{}
	req.Invoke(&mailbox.Get{
		Account: c.accountID,
	})

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get mailboxes: %w", err)
	}

	var mailboxes []models.Mailbox
	for _, inv := range resp.Responses {
		if getResp, ok := inv.Args.(*mailbox.GetResponse); ok {
			for _, mb := range getResp.List {
				mailboxes = append(mailboxes, models.Mailbox{
					ID:          string(mb.ID),
					Name:        mb.Name,
					Role:        string(mb.Role),
					ParentID:    string(mb.ParentID),
					TotalEmails: int(mb.TotalEmails),
					UnreadCount: int(mb.UnreadEmails),
					SortOrder:   int(mb.SortOrder),
				})
			}
		}
	}

	return mailboxes, nil
}

// GetEmails fetches emails from a mailbox
func (c *Client) GetEmails(mailboxID string, limit int) ([]models.Email, error) {
	req := &jmap.Request{}

	// Query for email IDs in the mailbox
	queryCall := req.Invoke(&email.Query{
		Account: c.accountID,
		Filter: &email.FilterCondition{
			InMailbox: jmap.ID(mailboxID),
		},
		Sort: []*email.SortComparator{
			{Property: "receivedAt", IsAscending: false},
		},
		Limit: uint64(limit),
	})

	// Get email details using the query results
	req.Invoke(&email.Get{
		Account: c.accountID,
		ReferenceIDs: &jmap.ResultReference{
			ResultOf: queryCall,
			Name:     "Email/query",
			Path:     "/ids",
		},
		Properties: []string{
			"id", "threadId", "mailboxIds", "from", "to", "cc", "bcc",
			"replyTo", "subject", "preview", "receivedAt", "size",
			"keywords", "hasAttachment",
		},
	})

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get emails: %w", err)
	}

	var emails []models.Email
	for _, inv := range resp.Responses {
		if getResp, ok := inv.Args.(*email.GetResponse); ok {
			for _, e := range getResp.List {
				emails = append(emails, convertEmail(e))
			}
		}
	}

	return emails, nil
}

// GetEmail fetches a single email with full body
func (c *Client) GetEmail(emailID string) (*models.Email, error) {
	req := &jmap.Request{}
	req.Invoke(&email.Get{
		Account: c.accountID,
		IDs:     []jmap.ID{jmap.ID(emailID)},
		Properties: []string{
			"id", "threadId", "mailboxIds", "from", "to", "cc", "bcc",
			"replyTo", "subject", "preview", "receivedAt", "size",
			"keywords", "hasAttachment", "textBody", "htmlBody",
			"attachments", "bodyValues",
		},
		FetchAllBodyValues: true,
	})

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get email: %w", err)
	}

	for _, inv := range resp.Responses {
		if getResp, ok := inv.Args.(*email.GetResponse); ok {
			if len(getResp.List) > 0 {
				e := convertEmail(getResp.List[0])
				return &e, nil
			}
		}
	}

	return nil, fmt.Errorf("email not found")
}

// SetEmailKeywords updates email keywords (read/unread, flagged, etc.)
func (c *Client) SetEmailKeywords(emailID string, keywords map[string]bool) error {
	req := &jmap.Request{}

	// Build patch for each keyword
	patch := jmap.Patch{}
	for k, v := range keywords {
		patch["keywords/"+k] = v
	}

	req.Invoke(&email.Set{
		Account: c.accountID,
		Update: map[jmap.ID]jmap.Patch{
			jmap.ID(emailID): patch,
		},
	})

	_, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to update email: %w", err)
	}

	return nil
}

// MarkAsRead marks an email as read
func (c *Client) MarkAsRead(emailID string) error {
	return c.SetEmailKeywords(emailID, map[string]bool{
		"$seen": true,
	})
}

// MarkAsUnread marks an email as unread
func (c *Client) MarkAsUnread(emailID string) error {
	return c.SetEmailKeywords(emailID, map[string]bool{
		"$seen": false,
	})
}

// MoveEmail moves an email to a different mailbox
func (c *Client) MoveEmail(emailID string, fromMailboxID, toMailboxID string) error {
	req := &jmap.Request{}

	patch := jmap.Patch{
		"mailboxIds": map[jmap.ID]bool{
			jmap.ID(toMailboxID): true,
		},
	}

	req.Invoke(&email.Set{
		Account: c.accountID,
		Update: map[jmap.ID]jmap.Patch{
			jmap.ID(emailID): patch,
		},
	})

	_, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to move email: %w", err)
	}

	return nil
}

// DeleteEmail moves an email to trash
func (c *Client) DeleteEmail(emailID, trashMailboxID string) error {
	return c.MoveEmail(emailID, "", trashMailboxID)
}

// convertEmail converts a JMAP email to our model
func convertEmail(e *email.Email) models.Email {
	result := models.Email{
		ID:            string(e.ID),
		ThreadID:      string(e.ThreadID),
		Subject:       e.Subject,
		Preview:       e.Preview,
		Size:          int(e.Size),
		HasAttachment: e.HasAttachment,
	}

	// Handle pointer to time
	if e.ReceivedAt != nil {
		result.ReceivedAt = *e.ReceivedAt
	}

	// Convert mailbox IDs
	for id := range e.MailboxIDs {
		result.MailboxIDs = append(result.MailboxIDs, string(id))
	}

	// Convert addresses
	for _, addr := range e.From {
		result.From = append(result.From, models.EmailAddress{
			Name:  addr.Name,
			Email: addr.Email,
		})
	}
	for _, addr := range e.To {
		result.To = append(result.To, models.EmailAddress{
			Name:  addr.Name,
			Email: addr.Email,
		})
	}
	for _, addr := range e.CC {
		result.CC = append(result.CC, models.EmailAddress{
			Name:  addr.Name,
			Email: addr.Email,
		})
	}
	for _, addr := range e.ReplyTo {
		result.ReplyTo = append(result.ReplyTo, models.EmailAddress{
			Name:  addr.Name,
			Email: addr.Email,
		})
	}

	// Check keywords
	if seen, ok := e.Keywords["$seen"]; ok && seen {
		result.IsUnread = false
	} else {
		result.IsUnread = true
	}
	if flagged, ok := e.Keywords["$flagged"]; ok && flagged {
		result.IsFlagged = true
	}
	if draft, ok := e.Keywords["$draft"]; ok && draft {
		result.IsDraft = true
	}

	// Get body content from body values
	for _, part := range e.TextBody {
		if val, ok := e.BodyValues[part.PartID]; ok {
			result.TextBody += val.Value
		}
	}
	for _, part := range e.HTMLBody {
		if val, ok := e.BodyValues[part.PartID]; ok {
			result.HTMLBody += val.Value
		}
	}

	// Convert attachments
	for _, att := range e.Attachments {
		result.Attachments = append(result.Attachments, models.Attachment{
			BlobID:   string(att.BlobID),
			Name:     att.Name,
			Type:     att.Type,
			Size:     int(att.Size),
			IsInline: att.Disposition == "inline",
		})
	}

	return result
}

// AccountID returns the JMAP account ID
func (c *Client) AccountID() string {
	return string(c.accountID)
}

// Email returns the email address for this client
func (c *Client) Email() string {
	return c.email
}
