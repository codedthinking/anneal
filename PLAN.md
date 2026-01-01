# anneal — Implementation Plan

*email triage under noise*
by the9x.ac

## Overview

**anneal** is a personal, terminal-first email client designed to reduce inbox load without demanding completion. It reframes email as a noisy optimization problem and rewards local, irreversible progress rather than inbox zero.

- **Mental model:** simulated annealing for email
- **Promise:** steady reduction under uncertainty, no guilt
- **Backend:** Fastmail via JMAP (opinionated by design)

## Technology Stack

| Component | Choice | Rationale |
|-----------|--------|-----------|
| Language | **Go** | Fast, single binary, excellent TUI ecosystem |
| TUI Framework | **Bubble Tea** | Elm architecture, beautiful aesthetics, active community |
| Styling | **Lip Gloss** | Companion to Bubble Tea for styling |
| Components | **Bubbles** | Pre-built components (lists, text inputs, spinners) |
| JMAP Client | **git.sr.ht/~rockorager/go-jmap** | Complete RFC 8620/8621 implementation |
| Contacts/Calendar | **CalDAV/CardDAV** | Fastmail doesn't support JMAP for these yet |
| Config | **YAML** | Human-readable, easy multi-account config |
| Keyring | **zalando/go-keyring** | Secure credential storage |
| Storage | **SQLite** | Local cache for emails, mailboxes, sync state |

## Architecture

```
anneal/
├── main.go                 # Entry point
├── go.mod
├── go.sum
├── config.yaml.example
├── Makefile
├── BRAND.md                # Brand guidelines
│
├── internal/
│   ├── config/             # Configuration loading
│   │   └── config.go
│   │
│   ├── jmap/               # JMAP client wrapper
│   │   ├── client.go       # Session management
│   │   ├── mail.go         # Email operations
│   │   ├── mailbox.go      # Folder operations
│   │   └── identity.go     # Send identity
│   │
│   ├── storage/            # Local persistence
│   │   ├── db.go           # SQLite connection
│   │   ├── emails.go       # Email cache operations
│   │   ├── mailboxes.go    # Mailbox cache operations
│   │   └── sync.go         # Sync state tracking
│   │
│   ├── caldav/             # Calendar client
│   │   └── client.go
│   │
│   ├── carddav/            # Contacts client
│   │   └── client.go
│   │
│   ├── models/             # Domain models
│   │   ├── email.go
│   │   ├── mailbox.go
│   │   ├── contact.go
│   │   ├── event.go
│   │   └── account.go
│   │
│   └── ui/                 # Bubble Tea UI
│       ├── app.go          # Root model, view routing
│       ├── styles.go       # Lip Gloss styles (the9x.ac palette)
│       ├── keys.go         # Key bindings
│       │
│       ├── views/
│       │   ├── mailbox.go      # Folder list sidebar
│       │   ├── threadlist.go   # Thread/message list view
│       │   ├── emaillist.go    # Email list view
│       │   ├── emailreader.go  # Read email view
│       │   ├── compose.go      # Compose/reply view
│       │   ├── contacts.go     # Contacts list
│       │   ├── calendar.go     # Calendar view
│       │   ├── accounts.go     # Account switcher
│       │   └── search.go       # Search interface
│       │
│       └── components/
│           ├── statusbar.go    # Bottom status bar
│           ├── help.go         # Help overlay
│           └── modal.go        # Generic modal
│
└── docs/
    └── keybindings.md
```

## Brand-Aligned UX Principles

1. **Session-scoped success** — success is defined per session, never globally
2. **Directional feedback** — show deltas (−11), not totals
3. **No red states** — no failure indicators, ever
4. **Irreversibility** — actions feel final and satisfying
5. **Silence by default** — minimal chrome, minimal copy

## Visual Identity

anneal inherits **the9x.ac** colors:

| Role | Color | Usage |
|------|-------|-------|
| Background | `#1d1d40` | App background |
| Primary | `#d4d2e3` | Main text, selected items |
| Secondary | `#9795b5` | Subjects, quotes, muted text |
| Dim | `#5a5880` | Headers, dates, hints |
| Accent | `#e61e25` | Flags only (used sparingly) |

## Core Features

### Phase 1: Email Core (MVP) ✓
1. **Authentication** - API token storage in system keyring
2. **Session Management** - JMAP session initialization
3. **Mailbox List** - Display folders with unread counts
4. **Email List** - Paginated email list with threading
5. **Email Viewer** - Read emails with plain text/HTML rendering
6. **Basic Actions** - Mark read/unread, archive (full thread), delete

### Phase 2: Compose & Reply

#### 2.1 Compose (Partially Done)
- [x] Inline compose view with To/CC/Subject/Body fields
- [x] Reply with quoted original, proper recipients
- [x] Reply All with CC population
- [x] Forward with message header block
- [x] Send via JMAP EmailSubmission
- [ ] $EDITOR integration (optional, `e` key in compose to open $EDITOR)
  - Write body to temp file, open editor, read back on close
  - `exec.Command(os.Getenv("EDITOR"), tempFile).Run()`

#### 2.2 Drafts
- [ ] Auto-save draft every 30s while composing
- [ ] Save draft on Esc (instead of discard)
- [ ] Load drafts from Drafts folder
- [ ] Resume editing a draft

#### 2.3 Attachments (Done)
- [x] Display attachment list (name, size) in email reader
- [x] Download blob via JMAP download URL
- [x] Save to temp file: `/tmp/anneal/attachments/{blobId}-{name}`
- [x] Open with system default: `exec.Command("open", filePath).Start()`
- [x] Navigation: `→/Enter` to enter attachment mode from email, `→/Enter` to open
- [ ] Cleanup temp files on exit (optional)

### Phase 3: Multi-Account
1. **Account Config** - Multiple Fastmail accounts in config
2. **Account Switcher** - Quick switch with keyboard shortcut
3. **Unified Inbox** - Optional combined view

### Phase 4: Contacts & Calendar
1. **CardDAV Client** - Fetch and display contacts
2. **Contact Autocomplete** - In compose view
3. **CalDAV Client** - Fetch calendar events
4. **Today View** - Show today's events in sidebar

### Phase 5: Polish
1. **Search** - JMAP search with filters
2. **Session Metrics** - Show deltas during session (−11 handled)
3. **Themes** - Configurable color schemes
4. **Notifications** - New mail notifications

### Phase 6: Local Storage & Performance

#### 6.1 SQLite Schema

```sql
-- Sync state tracking
CREATE TABLE sync_state (
    account_id TEXT PRIMARY KEY,
    mailbox_state TEXT,      -- JMAP state token for Mailbox
    email_state TEXT,        -- JMAP state token for Email
    last_sync INTEGER        -- Unix timestamp
);

-- Mailboxes cache
CREATE TABLE mailboxes (
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

-- Emails cache (metadata only, not full body)
CREATE TABLE emails (
    id TEXT PRIMARY KEY,
    account_id TEXT NOT NULL,
    thread_id TEXT,
    subject TEXT,
    preview TEXT,
    from_json TEXT,          -- JSON array of addresses
    to_json TEXT,
    cc_json TEXT,
    received_at INTEGER,
    size INTEGER,
    is_unread INTEGER DEFAULT 1,
    is_flagged INTEGER DEFAULT 0,
    has_attachment INTEGER DEFAULT 0,
    updated_at INTEGER NOT NULL
);

-- Email-to-mailbox mapping (many-to-many)
CREATE TABLE email_mailboxes (
    email_id TEXT NOT NULL,
    mailbox_id TEXT NOT NULL,
    PRIMARY KEY (email_id, mailbox_id)
);

-- Full email body cache (lazy-loaded, can be purged)
CREATE TABLE email_bodies (
    email_id TEXT PRIMARY KEY,
    text_body TEXT,
    html_body TEXT,
    attachments_json TEXT,   -- JSON array of attachment metadata
    fetched_at INTEGER
);

CREATE INDEX idx_emails_thread ON emails(thread_id);
CREATE INDEX idx_emails_received ON emails(received_at DESC);
CREATE INDEX idx_email_mailboxes_mailbox ON email_mailboxes(mailbox_id);
```

#### 6.2 Sync Strategy

**Startup flow:**
1. Load mailboxes from SQLite → display immediately
2. Load emails for current mailbox from SQLite → display immediately
3. Background: call JMAP with current state token
4. If state changed: fetch changes, update SQLite, refresh UI

**JMAP state tokens:**
- Each JMAP type (Mailbox, Email) has a `state` string
- Use `Mailbox/changes` and `Email/changes` to get deltas since last state
- Returns: created IDs, updated IDs, destroyed IDs

**Incremental sync (on state change):**
```
1. Call Mailbox/changes with stored mailbox_state
   → Get list of changed mailbox IDs
   → Fetch changed mailboxes with Mailbox/get
   → Update SQLite, update mailbox_state

2. Call Email/changes with stored email_state
   → Get list of changed email IDs
   → Fetch changed emails with Email/get
   → Update SQLite, update email_state
```

**Full sync (on first run or state token expired):**
```
1. Fetch all mailboxes → replace in SQLite
2. For each mailbox: fetch recent emails → replace in SQLite
3. Store new state tokens
```

#### 6.3 Cache Invalidation

**Automatic invalidation:**
- State token mismatch → incremental sync
- `cannotCalculateChanges` error → full sync for that type
- Network error → use stale cache, retry later

**Manual invalidation:**
- User action (archive, delete, mark read) → update local immediately
- Background sync confirms or reverts

**TTL-based refresh:**
- Check for changes every 60s while app is open
- On app start if last_sync > 5 minutes ago

#### 6.4 Implementation Files

```
internal/storage/
├── db.go           # Open/close SQLite, migrations
├── mailboxes.go    # CRUD for mailboxes table
├── emails.go       # CRUD for emails table
├── sync.go         # State tracking, sync orchestration
└── migrations/     # SQL migration files
    └── 001_initial.sql
```

#### 6.5 API Changes

**New interfaces:**
```go
type Store interface {
    // Mailboxes
    GetMailboxes(accountID string) ([]models.Mailbox, error)
    SaveMailboxes(accountID string, mailboxes []models.Mailbox) error

    // Emails
    GetEmails(mailboxID string, limit int) ([]models.Email, error)
    SaveEmails(emails []models.Email) error
    GetEmailBody(emailID string) (*models.Email, error)
    SaveEmailBody(email *models.Email) error

    // Sync state
    GetSyncState(accountID string) (*SyncState, error)
    SaveSyncState(state *SyncState) error

    // Invalidation
    DeleteEmail(emailID string) error
    UpdateEmailMailboxes(emailID string, mailboxIDs []string) error
}
```

**App startup changes:**
```go
// 1. Load from cache (instant)
mailboxes := store.GetMailboxes(accountID)
emails := store.GetEmails(inboxID, 50)
// Display UI immediately

// 2. Sync in background
go func() {
    changes := client.SyncChanges(store.GetSyncState())
    store.ApplyChanges(changes)
    // Send message to UI to refresh
}()
```

#### 6.6 Tasks

- [x] Add `modernc.org/sqlite` dependency (pure Go, no CGO)
- [x] Create `internal/storage/db.go` with migrations
- [x] Implement mailbox storage
- [x] Implement email metadata storage
- [x] Implement email body storage (lazy)
- [x] Add JMAP `/changes` endpoints to client
- [x] Implement sync orchestration
- [x] Update app.go to load from cache first
- [x] Add background sync goroutine
- [ ] Handle optimistic updates for user actions

### Phase 7: UI/UX Improvements
1. **Responsive Grid** - Dynamic column widths based on terminal size ✓
2. **Scroll Feedback** - Visual scrollbar, better position indicator
3. **Compose UX** - Better focus indication, placeholder contrast
4. **Email Rendering** - Signature collapse, cleaner link display

## UI Layout

```
┌─────────────────────────────────────────────────────────────────┐
│ ◈ anneal                           work@example.com    messages │
├───────────────┬─────────────────────────────────────────────────┤
│ ◈ mailboxes   │ from              subject                  date │
│               ├─────────────────────────────────────────────────┤
│   ▶ inbox  12 │ ● john doe        project update       10:30 am │
│     drafts  2 │   jane smith      re: meeting notes    09:15 am │
│     sent      │   github      ●   [notifications] ... yesterday │
│     archive   │   alice wong      quarterly report     yesterday │
│     trash     │ ▶3 bob miller     design feedback        nov 28 │
│               │   newsletter      weekly digest          nov 27 │
│ ◈ labels      │                                                 │
│     work      │                                                 │
│     personal  │                                                 │
│               │                                                 │
├───────────────┴─────────────────────────────────────────────────┤
│ j/k: navigate  enter: open  a: archive  d: delete  ?: help      │
└─────────────────────────────────────────────────────────────────┘
```

## Key Bindings

| Key | Action |
|-----|--------|
| `←/h` | Navigate back (folder/list/thread) |
| `→/l/Enter` | Navigate forward / open |
| `↑/k` | Navigate up in list |
| `↓/j` | Navigate down in list |
| `g/G` | Go to top/bottom |
| `Esc/q` | Back |
| `c` | Compose new email |
| `r` | Reply |
| `R` | Reply all |
| `f` | Forward |
| `d` | Delete |
| `a` | Archive (entire thread) |
| `m` | Move to folder |
| `s` | Star/flag |
| `u` | Toggle read/unread |
| `/` | Search |
| `1-5` | Switch account |
| `?` | Help |
| `Q/Ctrl+C` | Quit |

## Configuration

```yaml
# ~/.config/anneal/config.yaml
accounts:
  - name: work
    email: work@example.com
    # Token stored in system keyring
    default: true

  - name: personal
    email: personal@fastmail.com

editor: $EDITOR  # or vim, nvim, etc.
preview_pane: true
threading: true
page_size: 50
```

## Implementation Progress

### Completed (Phase 1)
- [x] Project setup with Go modules and Makefile
- [x] JMAP client wrapper with session management
- [x] Mailbox list with unread counts
- [x] Email list with threading support
- [x] Thread list view with collapse/expand
- [x] Email viewer with 100-char width limit
- [x] HTML → Markdown conversion
- [x] Glamour markdown rendering
- [x] Mark read/unread, archive, delete actions
- [x] Archive operates on entire thread
- [x] the9x.ac brand colors and aesthetic
- [x] Arrow key navigation (←/→ between panes)
- [x] Breadcrumb navigation indicator
- [x] Lowercase, minimal UI copy
- [x] Responsive thread list grid (dynamic column widths)

### Completed (Phase 2 - Partial)
- [x] Compose view with inline fields
- [x] Reply/Reply All/Forward functionality
- [x] Send email via JMAP
- [x] Attachment list display
- [x] Attachment download & open (`t` to enter mode, `o` to open)

### Completed (Phase 6)
- [x] SQLite local cache with WAL mode
- [x] JMAP state-based incremental sync
- [x] Cache-first loading with background refresh

### In Progress
- [ ] Session metrics (show deltas: −11 handled)

### Not Started
- [ ] $EDITOR integration for compose
- [ ] Draft auto-save
- [ ] Multi-account switching
- [ ] CardDAV contacts
- [ ] CalDAV calendar
- [ ] Full-text search

## Success Criteria

- [x] Can authenticate with Fastmail API token
- [x] Can view mailboxes and emails
- [x] Can read email content (with markdown rendering)
- [x] Can compose and send emails
- [ ] Can switch between multiple accounts
- [ ] Can view contacts
- [ ] Can view calendar events
- [x] Smooth, responsive TUI experience
- [x] Vim-style keyboard navigation
- [x] Brand-aligned calm, minimal aesthetic
