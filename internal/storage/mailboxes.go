package storage

import (
	"time"

	"github.com/the9x/anneal/internal/models"
)

// GetMailboxes retrieves all mailboxes for an account
func (s *Store) GetMailboxes(accountID string) ([]models.Mailbox, error) {
	rows, err := s.db.Query(`
		SELECT id, name, role, parent_id, total_emails, unread_count, sort_order
		FROM mailboxes
		WHERE account_id = ?
		ORDER BY sort_order, name
	`, accountID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var mailboxes []models.Mailbox
	for rows.Next() {
		var mb models.Mailbox
		var role, parentID *string

		err := rows.Scan(&mb.ID, &mb.Name, &role, &parentID, &mb.TotalEmails, &mb.UnreadCount, &mb.SortOrder)
		if err != nil {
			return nil, err
		}

		if role != nil {
			mb.Role = *role
		}
		if parentID != nil {
			mb.ParentID = *parentID
		}

		mailboxes = append(mailboxes, mb)
	}

	return mailboxes, rows.Err()
}

// SaveMailboxes saves mailboxes for an account (replaces existing)
func (s *Store) SaveMailboxes(accountID string, mailboxes []models.Mailbox) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing mailboxes for this account
	if _, err := tx.Exec("DELETE FROM mailboxes WHERE account_id = ?", accountID); err != nil {
		return err
	}

	// Insert new mailboxes
	stmt, err := tx.Prepare(`
		INSERT INTO mailboxes (id, account_id, name, role, parent_id, total_emails, unread_count, sort_order, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now().Unix()
	for _, mb := range mailboxes {
		var role, parentID *string
		if mb.Role != "" {
			role = &mb.Role
		}
		if mb.ParentID != "" {
			parentID = &mb.ParentID
		}

		_, err := stmt.Exec(mb.ID, accountID, mb.Name, role, parentID, mb.TotalEmails, mb.UnreadCount, mb.SortOrder, now)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateMailbox updates a single mailbox
func (s *Store) UpdateMailbox(accountID string, mb models.Mailbox) error {
	var role, parentID *string
	if mb.Role != "" {
		role = &mb.Role
	}
	if mb.ParentID != "" {
		parentID = &mb.ParentID
	}

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO mailboxes (id, account_id, name, role, parent_id, total_emails, unread_count, sort_order, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, mb.ID, accountID, mb.Name, role, parentID, mb.TotalEmails, mb.UnreadCount, mb.SortOrder, time.Now().Unix())
	return err
}

// DeleteMailbox removes a mailbox
func (s *Store) DeleteMailbox(mailboxID string) error {
	_, err := s.db.Exec("DELETE FROM mailboxes WHERE id = ?", mailboxID)
	return err
}
