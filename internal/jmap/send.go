package jmap

import (
	"fmt"
	"time"

	"git.sr.ht/~rockorager/go-jmap"
	"git.sr.ht/~rockorager/go-jmap/mail"
	"git.sr.ht/~rockorager/go-jmap/mail/email"
	"git.sr.ht/~rockorager/go-jmap/mail/emailsubmission"
	"git.sr.ht/~rockorager/go-jmap/mail/identity"
)

// Identity represents a sending identity
type Identity struct {
	ID    string
	Name  string
	Email string
}

// GetIdentities fetches available sending identities
func (c *Client) GetIdentities() ([]Identity, error) {
	req := &jmap.Request{}
	req.Invoke(&identity.Get{
		Account: c.accountID,
	})

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get identities: %w", err)
	}

	var identities []Identity
	for _, inv := range resp.Responses {
		if getResp, ok := inv.Args.(*identity.GetResponse); ok {
			for _, id := range getResp.List {
				identities = append(identities, Identity{
					ID:    string(id.ID),
					Name:  id.Name,
					Email: id.Email,
				})
			}
		}
	}

	return identities, nil
}

// GetDefaultIdentity returns the first identity (usually the default)
func (c *Client) GetDefaultIdentity() (*Identity, error) {
	identities, err := c.GetIdentities()
	if err != nil {
		return nil, err
	}
	if len(identities) == 0 {
		return nil, fmt.Errorf("no sending identities found")
	}
	return &identities[0], nil
}

// SendEmail creates and sends an email using the default identity
func (c *Client) SendEmail(to, cc []string, subject, body string, inReplyTo, references []string) error {
	return c.SendEmailWithIdentity(to, cc, subject, body, inReplyTo, references, "")
}

// SendEmailWithIdentity creates and sends an email using a specific identity
func (c *Client) SendEmailWithIdentity(to, cc []string, subject, body string, inReplyTo, references []string, identityID string) error {
	// Get identity
	var ident *Identity
	var err error

	if identityID != "" {
		// Find specific identity
		identities, err := c.GetIdentities()
		if err != nil {
			return err
		}
		for i := range identities {
			if identities[i].ID == identityID {
				ident = &identities[i]
				break
			}
		}
		if ident == nil {
			return fmt.Errorf("identity not found: %s", identityID)
		}
	} else {
		// Use default identity
		ident, err = c.GetDefaultIdentity()
		if err != nil {
			return err
		}
	}

	// Get drafts mailbox for temporary storage
	mailboxes, err := c.GetMailboxes()
	if err != nil {
		return fmt.Errorf("failed to get mailboxes: %w", err)
	}

	var draftsID jmap.ID
	for _, mb := range mailboxes {
		if mb.Role == "drafts" {
			draftsID = jmap.ID(mb.ID)
			break
		}
	}
	if draftsID == "" {
		return fmt.Errorf("drafts mailbox not found")
	}

	// Build recipient addresses
	toAddrs := make([]*mail.Address, len(to))
	for i, addr := range to {
		toAddrs[i] = &mail.Address{Email: addr}
	}

	ccAddrs := make([]*mail.Address, len(cc))
	for i, addr := range cc {
		ccAddrs[i] = &mail.Address{Email: addr}
	}

	// Build envelope recipients (all To + CC)
	var rcptTo []*emailsubmission.Address
	for _, addr := range to {
		rcptTo = append(rcptTo, &emailsubmission.Address{Email: addr})
	}
	for _, addr := range cc {
		rcptTo = append(rcptTo, &emailsubmission.Address{Email: addr})
	}

	// Create the email
	now := time.Now()
	newEmail := &email.Email{
		MailboxIDs: map[jmap.ID]bool{draftsID: true},
		From:       []*mail.Address{{Name: ident.Name, Email: ident.Email}},
		To:         toAddrs,
		CC:         ccAddrs,
		Subject:    subject,
		SentAt:     &now,
		Keywords:   map[string]bool{"$seen": true},
		BodyValues: map[string]*email.BodyValue{
			"body": {Value: body},
		},
		TextBody: []*email.BodyPart{
			{PartID: "body", Type: "text/plain"},
		},
	}

	// Add reply headers if replying
	if len(inReplyTo) > 0 {
		newEmail.InReplyTo = inReplyTo
	}
	if len(references) > 0 {
		newEmail.References = references
	}

	req := &jmap.Request{}

	// Create the email
	emailCreateID := jmap.ID("draft")
	_ = req.Invoke(&email.Set{
		Account: c.accountID,
		Create: map[jmap.ID]*email.Email{
			emailCreateID: newEmail,
		},
	})

	// Submit the email for sending
	submissionID := jmap.ID("send")
	req.Invoke(&emailsubmission.Set{
		Account: c.accountID,
		Create: map[jmap.ID]*emailsubmission.EmailSubmission{
			submissionID: {
				IdentityID: jmap.ID(ident.ID),
				EmailID:    jmap.ID("#" + string(emailCreateID)),
				Envelope: &emailsubmission.Envelope{
					MailFrom: &emailsubmission.Address{Email: ident.Email},
					RcptTo:   rcptTo,
				},
			},
		},
		OnSuccessUpdateEmail: map[jmap.ID]jmap.Patch{
			"#" + submissionID: {
				"mailboxIds":      nil, // Remove from drafts
				"keywords/$draft": nil, // Remove draft keyword
				"keywords/$seen":  true,
			},
		},
	})

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send email: %w", err)
	}

	// Check for errors in responses
	for _, inv := range resp.Responses {
		if setResp, ok := inv.Args.(*email.SetResponse); ok {
			if len(setResp.NotCreated) > 0 {
				for _, setErr := range setResp.NotCreated {
					desc := "unknown error"
					if setErr.Description != nil {
						desc = *setErr.Description
					}
					return fmt.Errorf("failed to create email: %s", desc)
				}
			}
		}
		if setResp, ok := inv.Args.(*emailsubmission.SetResponse); ok {
			if len(setResp.NotCreated) > 0 {
				for _, setErr := range setResp.NotCreated {
					desc := "unknown error"
					if setErr.Description != nil {
						desc = *setErr.Description
					}
					return fmt.Errorf("failed to submit email: %s", desc)
				}
			}
		}
	}

	return nil
}
