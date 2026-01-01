# TuiMail - TUI Email Client for Fastmail

## Overview

A terminal-based email client built with Go and Bubble Tea, connecting to Fastmail via JMAP protocol. Features email management, multi-account support, contacts, and calendar access.

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

## Architecture

```
tuimail/
├── main.go                 # Entry point
├── go.mod
├── go.sum
├── config.yaml.example
├── Makefile
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
│       ├── styles.go       # Lip Gloss styles
│       ├── keys.go         # Key bindings
│       │
│       ├── views/
│       │   ├── mailbox.go      # Folder list sidebar
│       │   ├── emaillist.go    # Email list view
│       │   ├── emailview.go    # Read email view
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

## Core Features

### Phase 1: Email Core (MVP)
1. **Authentication** - API token storage in system keyring
2. **Session Management** - JMAP session initialization
3. **Mailbox List** - Display folders with unread counts
4. **Email List** - Paginated email list with threading
5. **Email Viewer** - Read emails with plain text/HTML rendering
6. **Basic Actions** - Mark read/unread, archive, delete, move

### Phase 2: Compose & Reply
1. **Compose** - New email with $EDITOR integration
2. **Reply/Forward** - Quote original, manage recipients
3. **Drafts** - Auto-save drafts
4. **Attachments** - View and save attachments

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
2. **Keyboard Shortcuts** - Vim-style navigation
3. **Themes** - Configurable color schemes
4. **Notifications** - New mail notifications

## UI Layout

```
┌─────────────────────────────────────────────────────────────────┐
│ TuiMail                              work@example.com  [?] Help │
├───────────────┬─────────────────────────────────────────────────┤
│ MAILBOXES     │ INBOX (847)                          Search: /  │
│               ├─────────────────────────────────────────────────┤
│ ▶ Inbox    12 │ ● John Doe         Project Update      10:30 AM │
│   Drafts    2 │   Jane Smith       Re: Meeting notes   09:15 AM │
│   Sent        │   GitHub       ●   [notifications] ... Yesterday│
│   Archive     │   Alice Wong       Quarterly report    Yesterday│
│   Trash       │ ▶ Bob Miller       Design feedback     Nov 28   │
│               │   Newsletter       Weekly digest       Nov 27   │
│ LABELS        │                                                 │
│   Work        │                                                 │
│   Personal    │                                                 │
│   Finance     │                                                 │
│               │                                                 │
├───────────────┴─────────────────────────────────────────────────┤
│ j/k: navigate  Enter: open  c: compose  r: reply  d: delete     │
└─────────────────────────────────────────────────────────────────┘
```

## Key Bindings (Vim-inspired)

| Key | Action |
|-----|--------|
| `←/h` | Navigate back (folder/list/thread) |
| `→/l` | Navigate forward / open |
| `↑/k` | Navigate up in list |
| `↓/j` | Navigate down in list |
| `g/G` | Go to top/bottom |
| `Enter` | Open / expand thread |
| `Esc/q` | Back |
| `Space/Tab` | Expand/collapse thread |
| `c` | Compose new email |
| `r` | Reply |
| `R` | Reply all |
| `f` | Forward |
| `d` | Delete |
| `a` | Archive |
| `m` | Move to folder |
| `s` | Star/flag |
| `u` | Mark unread |
| `/` | Search |
| `1-5` | Switch account |
| `?` | Help |
| `Q/Ctrl+C` | Quit |

## Configuration

```yaml
# ~/.config/tuimail/config.yaml
accounts:
  - name: Work
    email: work@example.com
    # Token stored in system keyring
    default: true

  - name: Personal
    email: personal@fastmail.com

theme: dark  # dark, light, nord, dracula
editor: $EDITOR  # or vim, nvim, etc.
preview_pane: true
threading: true
page_size: 50
```

## Implementation Steps

### Step 1: Project Setup
- Initialize Go module
- Set up directory structure
- Add dependencies (bubble tea, lip gloss, go-jmap)
- Create Makefile

### Step 2: JMAP Client
- Implement session initialization
- Mailbox fetching
- Email list fetching with pagination
- Email content fetching

### Step 3: Basic TUI
- Root app model with view state
- Mailbox sidebar component
- Email list component
- Navigation between views

### Step 4: Email Viewer
- Plain text rendering
- Basic HTML to text conversion
- Header display
- Attachment listing

### Step 5: Actions
- Mark read/unread
- Delete/archive
- Move between mailboxes

### Step 6: Compose
- External editor integration
- Reply with quoting
- Send via JMAP EmailSubmission

### Step 7: Multi-Account
- Config file parsing
- Keyring integration
- Account switcher UI

### Step 8: Contacts/Calendar
- CardDAV client
- CalDAV client
- UI integration

## Dependencies

```go
require (
    github.com/charmbracelet/bubbletea
    github.com/charmbracelet/bubbles
    github.com/charmbracelet/lipgloss
    git.sr.ht/~rockorager/go-jmap
    github.com/zalando/go-keyring
    gopkg.in/yaml.v3
    github.com/emersion/go-webdav  // CardDAV/CalDAV
)
```

## Success Criteria

- [x] Can authenticate with Fastmail API token
- [x] Can view mailboxes and emails
- [x] Can read email content (with markdown rendering)
- [ ] Can compose and send emails
- [ ] Can switch between multiple accounts
- [ ] Can view contacts
- [ ] Can view calendar events
- [x] Smooth, responsive TUI experience
- [x] Vim-style keyboard navigation

## Implementation Progress

### Completed (Phase 1)
- [x] Project setup with Go modules and Makefile
- [x] JMAP client wrapper with session management
- [x] Mailbox list with unread counts
- [x] Email list with threading support
- [x] Thread list view with collapse/expand
- [x] Email viewer with 80-char width limit
- [x] HTML → Markdown conversion
- [x] Glamour markdown rendering
- [x] Mark read/unread, archive, delete actions
- [x] Cyberpunk/Blade Runner aesthetic
- [x] Arrow key navigation (←/→ between panes)
- [x] Breadcrumb navigation indicator

### In Progress
- [ ] Compose with $EDITOR integration

### Not Started
- [ ] Reply/Forward functionality
- [ ] Multi-account switching
- [ ] CardDAV contacts
- [ ] CalDAV calendar
- [ ] Full-text search
