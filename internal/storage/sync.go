package storage

import (
	"time"

	"github.com/the9x/anneal/internal/jmap"
	"github.com/the9x/anneal/internal/models"
)

// Syncer handles synchronization between JMAP and local storage
type Syncer struct {
	store  *Store
	client *jmap.Client
}

// NewSyncer creates a new syncer
func NewSyncer(store *Store, client *jmap.Client) *Syncer {
	return &Syncer{
		store:  store,
		client: client,
	}
}

// SyncResult contains the result of a sync operation
type SyncResult struct {
	MailboxesCreated   int
	MailboxesUpdated   int
	MailboxesDestroyed int
	EmailsCreated      int
	EmailsUpdated      int
	EmailsDestroyed    int
}

// SyncMailboxes synchronizes mailboxes with the server
func (s *Syncer) SyncMailboxes() (*SyncResult, error) {
	accountID := s.client.AccountID()
	result := &SyncResult{}

	// Get current sync state
	state, err := s.store.GetSyncState(accountID)
	if err != nil {
		return nil, err
	}

	// If no state, do full sync
	if state == nil || state.MailboxState == "" {
		return s.fullMailboxSync(accountID)
	}

	// Try incremental sync
	changes, err := s.client.GetMailboxChanges(state.MailboxState)
	if err != nil {
		// If state is too old, fall back to full sync
		return s.fullMailboxSync(accountID)
	}

	// Handle destroyed mailboxes
	for _, id := range changes.Destroyed {
		if err := s.store.DeleteMailbox(id); err != nil {
			return nil, err
		}
		result.MailboxesDestroyed++
	}

	// Handle created and updated mailboxes
	idsToFetch := append(changes.Created, changes.Updated...)
	if len(idsToFetch) > 0 {
		mailboxes, err := s.client.GetMailboxesByIDs(idsToFetch)
		if err != nil {
			return nil, err
		}

		for _, mb := range mailboxes {
			if err := s.store.UpdateMailbox(accountID, mb); err != nil {
				return nil, err
			}
		}

		result.MailboxesCreated = len(changes.Created)
		result.MailboxesUpdated = len(changes.Updated)
	}

	// Update sync state
	state.MailboxState = changes.NewState
	state.LastSync = time.Now()
	if err := s.store.SaveSyncState(state); err != nil {
		return nil, err
	}

	return result, nil
}

// fullMailboxSync does a complete mailbox sync
func (s *Syncer) fullMailboxSync(accountID string) (*SyncResult, error) {
	result := &SyncResult{}

	mailboxes, newState, err := s.client.MailboxesWithState()
	if err != nil {
		return nil, err
	}

	if err := s.store.SaveMailboxes(accountID, mailboxes); err != nil {
		return nil, err
	}

	result.MailboxesCreated = len(mailboxes)

	// Save sync state
	state := &SyncState{
		AccountID:    accountID,
		MailboxState: newState,
		LastSync:     time.Now(),
	}
	if err := s.store.SaveSyncState(state); err != nil {
		return nil, err
	}

	return result, nil
}

// SyncEmails synchronizes emails for a mailbox
func (s *Syncer) SyncEmails(mailboxID string, limit int) (*SyncResult, error) {
	accountID := s.client.AccountID()
	result := &SyncResult{}

	// Get current sync state
	state, err := s.store.GetSyncState(accountID)
	if err != nil {
		return nil, err
	}

	// If no state, do full sync
	if state == nil || state.EmailState == "" {
		return s.fullEmailSync(accountID, mailboxID, limit)
	}

	// Try incremental sync
	changes, err := s.client.GetEmailChanges(state.EmailState)
	if err != nil {
		// If state is too old, fall back to full sync
		return s.fullEmailSync(accountID, mailboxID, limit)
	}

	// Handle destroyed emails
	for _, id := range changes.Destroyed {
		if err := s.store.DeleteEmail(id); err != nil {
			return nil, err
		}
		result.EmailsDestroyed++
	}

	// Handle created and updated emails
	idsToFetch := append(changes.Created, changes.Updated...)
	if len(idsToFetch) > 0 {
		emails, err := s.client.GetEmailsByIDs(idsToFetch)
		if err != nil {
			return nil, err
		}

		if err := s.store.SaveEmails(accountID, emails); err != nil {
			return nil, err
		}

		result.EmailsCreated = len(changes.Created)
		result.EmailsUpdated = len(changes.Updated)
	}

	// Update sync state
	state.EmailState = changes.NewState
	state.LastSync = time.Now()
	if err := s.store.SaveSyncState(state); err != nil {
		return nil, err
	}

	return result, nil
}

// fullEmailSync does a complete email sync for a mailbox
func (s *Syncer) fullEmailSync(accountID, mailboxID string, limit int) (*SyncResult, error) {
	result := &SyncResult{}

	emails, newState, err := s.client.EmailsWithState(mailboxID, limit)
	if err != nil {
		return nil, err
	}

	if err := s.store.SaveEmails(accountID, emails); err != nil {
		return nil, err
	}

	result.EmailsCreated = len(emails)

	// Get or create sync state
	state, err := s.store.GetSyncState(accountID)
	if err != nil {
		return nil, err
	}
	if state == nil {
		state = &SyncState{AccountID: accountID}
	}

	state.EmailState = newState
	state.LastSync = time.Now()
	if err := s.store.SaveSyncState(state); err != nil {
		return nil, err
	}

	return result, nil
}

// GetCachedMailboxes returns cached mailboxes (instant)
func (s *Syncer) GetCachedMailboxes() ([]models.Mailbox, error) {
	return s.store.GetMailboxes(s.client.AccountID())
}

// GetCachedEmails returns cached emails for a mailbox (instant)
func (s *Syncer) GetCachedEmails(mailboxID string, limit int) ([]models.Email, error) {
	return s.store.GetEmails(mailboxID, limit)
}

// GetCachedEmailBody returns cached email body if available
func (s *Syncer) GetCachedEmailBody(emailID string) (*models.Email, error) {
	return s.store.GetEmailBody(emailID)
}

// FetchAndCacheEmailBody fetches email body from server and caches it
func (s *Syncer) FetchAndCacheEmailBody(emailID string) (*models.Email, error) {
	email, err := s.client.GetEmail(emailID)
	if err != nil {
		return nil, err
	}

	if err := s.store.SaveEmailBody(email); err != nil {
		return nil, err
	}

	return email, nil
}

// NeedsSync returns true if a sync is recommended
func (s *Syncer) NeedsSync(maxAge time.Duration) (bool, error) {
	state, err := s.store.GetSyncState(s.client.AccountID())
	if err != nil {
		return false, err
	}

	if state == nil {
		return true, nil
	}

	return time.Since(state.LastSync) > maxAge, nil
}

// HasCachedData returns true if there's any cached data
func (s *Syncer) HasCachedData() (bool, error) {
	mailboxes, err := s.store.GetMailboxes(s.client.AccountID())
	if err != nil {
		return false, err
	}
	return len(mailboxes) > 0, nil
}
