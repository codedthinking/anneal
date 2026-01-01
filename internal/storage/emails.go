package storage

import (
	"database/sql"
	"encoding/json"
	"time"

	"github.com/the9x/anneal/internal/models"
)

// GetEmails retrieves emails for a mailbox
func (s *Store) GetEmails(mailboxID string, limit int) ([]models.Email, error) {
	rows, err := s.db.Query(`
		SELECT e.id, e.thread_id, e.subject, e.preview, e.from_json, e.to_json, e.cc_json,
		       e.reply_to_json, e.received_at, e.size, e.is_unread, e.is_flagged, e.is_draft, e.has_attachment
		FROM emails e
		JOIN email_mailboxes em ON e.id = em.email_id
		WHERE em.mailbox_id = ?
		ORDER BY e.received_at DESC
		LIMIT ?
	`, mailboxID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanEmails(rows)
}

// GetEmailsByThread retrieves all emails in a thread
func (s *Store) GetEmailsByThread(threadID string) ([]models.Email, error) {
	rows, err := s.db.Query(`
		SELECT id, thread_id, subject, preview, from_json, to_json, cc_json,
		       reply_to_json, received_at, size, is_unread, is_flagged, is_draft, has_attachment
		FROM emails
		WHERE thread_id = ?
		ORDER BY received_at ASC
	`, threadID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	return s.scanEmails(rows)
}

func (s *Store) scanEmails(rows *sql.Rows) ([]models.Email, error) {
	var emails []models.Email

	for rows.Next() {
		var e models.Email
		var fromJSON, toJSON, ccJSON, replyToJSON sql.NullString
		var receivedAt int64
		var isUnread, isFlagged, isDraft, hasAttachment int

		err := rows.Scan(
			&e.ID, &e.ThreadID, &e.Subject, &e.Preview,
			&fromJSON, &toJSON, &ccJSON, &replyToJSON,
			&receivedAt, &e.Size, &isUnread, &isFlagged, &isDraft, &hasAttachment,
		)
		if err != nil {
			return nil, err
		}

		e.ReceivedAt = time.Unix(receivedAt, 0)
		e.IsUnread = isUnread == 1
		e.IsFlagged = isFlagged == 1
		e.IsDraft = isDraft == 1
		e.HasAttachment = hasAttachment == 1

		// Parse JSON address fields
		if fromJSON.Valid {
			json.Unmarshal([]byte(fromJSON.String), &e.From)
		}
		if toJSON.Valid {
			json.Unmarshal([]byte(toJSON.String), &e.To)
		}
		if ccJSON.Valid {
			json.Unmarshal([]byte(ccJSON.String), &e.CC)
		}
		if replyToJSON.Valid {
			json.Unmarshal([]byte(replyToJSON.String), &e.ReplyTo)
		}

		emails = append(emails, e)
	}

	return emails, rows.Err()
}

// SaveEmails saves emails and their mailbox associations
func (s *Store) SaveEmails(accountID string, emails []models.Email) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	emailStmt, err := tx.Prepare(`
		INSERT OR REPLACE INTO emails
		(id, account_id, thread_id, subject, preview, from_json, to_json, cc_json, reply_to_json,
		 received_at, size, is_unread, is_flagged, is_draft, has_attachment, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}
	defer emailStmt.Close()

	mailboxStmt, err := tx.Prepare(`
		INSERT OR IGNORE INTO email_mailboxes (email_id, mailbox_id) VALUES (?, ?)
	`)
	if err != nil {
		return err
	}
	defer mailboxStmt.Close()

	now := time.Now().Unix()

	for _, e := range emails {
		fromJSON, _ := json.Marshal(e.From)
		toJSON, _ := json.Marshal(e.To)
		ccJSON, _ := json.Marshal(e.CC)
		replyToJSON, _ := json.Marshal(e.ReplyTo)

		isUnread := 0
		if e.IsUnread {
			isUnread = 1
		}
		isFlagged := 0
		if e.IsFlagged {
			isFlagged = 1
		}
		isDraft := 0
		if e.IsDraft {
			isDraft = 1
		}
		hasAttachment := 0
		if e.HasAttachment {
			hasAttachment = 1
		}

		_, err := emailStmt.Exec(
			e.ID, accountID, e.ThreadID, e.Subject, e.Preview,
			string(fromJSON), string(toJSON), string(ccJSON), string(replyToJSON),
			e.ReceivedAt.Unix(), e.Size, isUnread, isFlagged, isDraft, hasAttachment, now,
		)
		if err != nil {
			return err
		}

		// Save mailbox associations
		for _, mbID := range e.MailboxIDs {
			if _, err := mailboxStmt.Exec(e.ID, mbID); err != nil {
				return err
			}
		}
	}

	return tx.Commit()
}

// GetEmailBody retrieves the full body for an email
func (s *Store) GetEmailBody(emailID string) (*models.Email, error) {
	row := s.db.QueryRow(`
		SELECT e.id, e.thread_id, e.subject, e.preview, e.from_json, e.to_json, e.cc_json,
		       e.reply_to_json, e.received_at, e.size, e.is_unread, e.is_flagged, e.is_draft, e.has_attachment,
		       b.text_body, b.html_body, b.attachments_json
		FROM emails e
		LEFT JOIN email_bodies b ON e.id = b.email_id
		WHERE e.id = ?
	`, emailID)

	var e models.Email
	var fromJSON, toJSON, ccJSON, replyToJSON sql.NullString
	var textBody, htmlBody, attachmentsJSON sql.NullString
	var receivedAt int64
	var isUnread, isFlagged, isDraft, hasAttachment int

	err := row.Scan(
		&e.ID, &e.ThreadID, &e.Subject, &e.Preview,
		&fromJSON, &toJSON, &ccJSON, &replyToJSON,
		&receivedAt, &e.Size, &isUnread, &isFlagged, &isDraft, &hasAttachment,
		&textBody, &htmlBody, &attachmentsJSON,
	)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	e.ReceivedAt = time.Unix(receivedAt, 0)
	e.IsUnread = isUnread == 1
	e.IsFlagged = isFlagged == 1
	e.IsDraft = isDraft == 1
	e.HasAttachment = hasAttachment == 1

	if fromJSON.Valid {
		json.Unmarshal([]byte(fromJSON.String), &e.From)
	}
	if toJSON.Valid {
		json.Unmarshal([]byte(toJSON.String), &e.To)
	}
	if ccJSON.Valid {
		json.Unmarshal([]byte(ccJSON.String), &e.CC)
	}
	if replyToJSON.Valid {
		json.Unmarshal([]byte(replyToJSON.String), &e.ReplyTo)
	}

	if textBody.Valid {
		e.TextBody = textBody.String
	}
	if htmlBody.Valid {
		e.HTMLBody = htmlBody.String
	}
	if attachmentsJSON.Valid {
		json.Unmarshal([]byte(attachmentsJSON.String), &e.Attachments)
	}

	return &e, nil
}

// SaveEmailBody saves the full body for an email
func (s *Store) SaveEmailBody(email *models.Email) error {
	attachmentsJSON, _ := json.Marshal(email.Attachments)

	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO email_bodies (email_id, text_body, html_body, attachments_json, fetched_at)
		VALUES (?, ?, ?, ?, ?)
	`, email.ID, email.TextBody, email.HTMLBody, string(attachmentsJSON), time.Now().Unix())
	return err
}

// HasEmailBody checks if we have the body cached
func (s *Store) HasEmailBody(emailID string) (bool, error) {
	var count int
	err := s.db.QueryRow("SELECT COUNT(*) FROM email_bodies WHERE email_id = ?", emailID).Scan(&count)
	return count > 0, err
}

// DeleteEmail removes an email from the cache
func (s *Store) DeleteEmail(emailID string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM email_bodies WHERE email_id = ?", emailID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM email_mailboxes WHERE email_id = ?", emailID); err != nil {
		return err
	}
	if _, err := tx.Exec("DELETE FROM emails WHERE id = ?", emailID); err != nil {
		return err
	}

	return tx.Commit()
}

// UpdateEmailMailboxes updates the mailbox associations for an email
func (s *Store) UpdateEmailMailboxes(emailID string, mailboxIDs []string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Delete existing associations
	if _, err := tx.Exec("DELETE FROM email_mailboxes WHERE email_id = ?", emailID); err != nil {
		return err
	}

	// Insert new associations
	stmt, err := tx.Prepare("INSERT INTO email_mailboxes (email_id, mailbox_id) VALUES (?, ?)")
	if err != nil {
		return err
	}
	defer stmt.Close()

	for _, mbID := range mailboxIDs {
		if _, err := stmt.Exec(emailID, mbID); err != nil {
			return err
		}
	}

	return tx.Commit()
}

// UpdateEmailFlags updates read/flagged status
func (s *Store) UpdateEmailFlags(emailID string, isUnread, isFlagged bool) error {
	unread := 0
	if isUnread {
		unread = 1
	}
	flagged := 0
	if isFlagged {
		flagged = 1
	}

	_, err := s.db.Exec(`
		UPDATE emails SET is_unread = ?, is_flagged = ?, updated_at = ?
		WHERE id = ?
	`, unread, flagged, time.Now().Unix(), emailID)
	return err
}

// PurgeOldBodies removes bodies older than the given duration
func (s *Store) PurgeOldBodies(olderThan time.Duration) (int64, error) {
	cutoff := time.Now().Add(-olderThan).Unix()
	result, err := s.db.Exec("DELETE FROM email_bodies WHERE fetched_at < ?", cutoff)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}
