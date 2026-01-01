package storage

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite"
)

// Store handles all local persistence
type Store struct {
	db *sql.DB
}

// SyncState tracks JMAP state tokens for incremental sync
type SyncState struct {
	AccountID    string
	MailboxState string
	EmailState   string
	LastSync     time.Time
}

// New creates a new Store, initializing the database if needed
func New() (*Store, error) {
	dbPath, err := getDBPath()
	if err != nil {
		return nil, err
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrent access
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to enable WAL: %w", err)
	}

	store := &Store{db: db}

	if err := store.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return store, nil
}

// Close closes the database connection
func (s *Store) Close() error {
	return s.db.Close()
}

// getDBPath returns the path to the SQLite database file
func getDBPath() (string, error) {
	// Use XDG data directory or fallback
	dataDir := os.Getenv("XDG_DATA_HOME")
	if dataDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dataDir = filepath.Join(home, ".local", "share")
	}

	appDir := filepath.Join(dataDir, "anneal")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return "", err
	}

	return filepath.Join(appDir, "cache.db"), nil
}

// migrate runs database migrations
func (s *Store) migrate() error {
	// Create migrations table if not exists
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS schema_version (
			version INTEGER PRIMARY KEY
		)
	`)
	if err != nil {
		return err
	}

	// Get current version
	var version int
	row := s.db.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_version")
	if err := row.Scan(&version); err != nil {
		return err
	}

	// Run migrations
	migrations := []string{
		migration001,
	}

	for i, migration := range migrations {
		migrationVersion := i + 1
		if migrationVersion <= version {
			continue
		}

		if _, err := s.db.Exec(migration); err != nil {
			return fmt.Errorf("migration %d failed: %w", migrationVersion, err)
		}

		if _, err := s.db.Exec("INSERT INTO schema_version (version) VALUES (?)", migrationVersion); err != nil {
			return err
		}
	}

	return nil
}

const migration001 = `
-- Sync state tracking
CREATE TABLE IF NOT EXISTS sync_state (
    account_id TEXT PRIMARY KEY,
    mailbox_state TEXT,
    email_state TEXT,
    last_sync INTEGER
);

-- Mailboxes cache
CREATE TABLE IF NOT EXISTS mailboxes (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    name TEXT NOT NULL,
    role TEXT,
    parent_id TEXT,
    total_emails INTEGER DEFAULT 0,
    unread_count INTEGER DEFAULT 0,
    sort_order INTEGER DEFAULT 0,
    updated_at INTEGER NOT NULL
);

-- Emails cache (metadata only)
CREATE TABLE IF NOT EXISTS emails (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    thread_id TEXT,
    subject TEXT,
    preview TEXT,
    from_json TEXT,
    to_json TEXT,
    cc_json TEXT,
    reply_to_json TEXT,
    received_at INTEGER,
    size INTEGER,
    is_unread INTEGER DEFAULT 1,
    is_flagged INTEGER DEFAULT 0,
    is_draft INTEGER DEFAULT 0,
    has_attachment INTEGER DEFAULT 0,
    updated_at INTEGER NOT NULL
);

-- Email-to-mailbox mapping
CREATE TABLE IF NOT EXISTS email_mailboxes (
    email_id TEXT NOT NULL,
    mailbox_id TEXT NOT NULL,
    PRIMARY KEY (email_id, mailbox_id)
);

-- Full email body cache (lazy-loaded)
CREATE TABLE IF NOT EXISTS email_bodies (
    email_id TEXT PRIMARY KEY,
    text_body TEXT,
    html_body TEXT,
    attachments_json TEXT,
    fetched_at INTEGER
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_mailboxes_account ON mailboxes(account_id);
CREATE INDEX IF NOT EXISTS idx_emails_account ON emails(account_id);
CREATE INDEX IF NOT EXISTS idx_emails_thread ON emails(thread_id);
CREATE INDEX IF NOT EXISTS idx_emails_received ON emails(received_at DESC);
CREATE INDEX IF NOT EXISTS idx_email_mailboxes_mailbox ON email_mailboxes(mailbox_id);
`

// GetSyncState retrieves the sync state for an account
func (s *Store) GetSyncState(accountID string) (*SyncState, error) {
	row := s.db.QueryRow(`
		SELECT account_id, mailbox_state, email_state, last_sync
		FROM sync_state WHERE account_id = ?
	`, accountID)

	state := &SyncState{}
	var lastSync int64
	var mailboxState, emailState sql.NullString

	err := row.Scan(&state.AccountID, &mailboxState, &emailState, &lastSync)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	state.MailboxState = mailboxState.String
	state.EmailState = emailState.String
	state.LastSync = time.Unix(lastSync, 0)

	return state, nil
}

// SaveSyncState saves the sync state for an account
func (s *Store) SaveSyncState(state *SyncState) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO sync_state (account_id, mailbox_state, email_state, last_sync)
		VALUES (?, ?, ?, ?)
	`, state.AccountID, state.MailboxState, state.EmailState, state.LastSync.Unix())
	return err
}

// ClearCache removes all cached data (for debugging/reset)
func (s *Store) ClearCache() error {
	tables := []string{"email_bodies", "email_mailboxes", "emails", "mailboxes", "sync_state"}
	for _, table := range tables {
		if _, err := s.db.Exec("DELETE FROM " + table); err != nil {
			return err
		}
	}
	return nil
}
