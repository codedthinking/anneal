# TuiMail - Project Memory

## Project Overview

TUI email client for Fastmail via JMAP protocol, built with Go and Bubble Tea framework.

## Current Status: Phase 1 Complete

### Completed Features

1. **JMAP Integration**
   - Connection to Fastmail via JMAP API
   - Mailbox listing
   - Email fetching with threading support
   - Mark read/unread functionality
   - Authentication via API token stored in system keyring

2. **TUI Interface**
   - Cyberpunk/Blade Runner aesthetic (neon colors: #FF2E97, #00FFFF, #BD00FF)
   - Three-panel layout: sidebar (mailboxes) + main content
   - Navigation flow: Folders → Messages → [Thread] → Email
     - Single emails: Folders → Messages → Email (skips thread step)
     - Multi-email threads: Folders → Messages → Thread → Email
   - Visual distinction: ◇ for single emails, ▶N for threads with N emails
   - Arrow key navigation (←/→) with vim keys (h/j/k/l)
   - Thread collapsing/expansion with space/tab
   - Breadcrumb navigation indicator in status bar
   - Context-aware help bar

3. **Email Rendering**
   - 80-character max width for readability
   - Glamour markdown rendering
   - HTML → Markdown conversion (headers, bold, italic, links, lists, blockquotes)
   - HTML entity decoding
   - Quoted text styling (different color for `>` lines)
   - Attachment display

### Key Technical Decisions

- **go-jmap library**: Uses `git.sr.ht/~rockorager/go-jmap` v0.5.3
- **Filter type**: Must use `email.FilterCondition` (not `email.Filter` interface directly)
- **Patch type**: Updates use `jmap.Patch` not `*email.Email`
- **Thread grouping**: Emails grouped by ThreadID in app.go
- **Config storage**: `~/.config/tuimail/config.yaml` with token in system keyring

### File Structure

```
tuimail/
├── main.go                    # Entry point
├── Makefile                   # Build commands
├── go.mod / go.sum           # Dependencies
├── PLAN.md                    # Original implementation plan
├── internal/
│   ├── config/config.go      # Config loading with keyring
│   ├── jmap/client.go        # JMAP API wrapper
│   ├── models/               # Data models
│   │   ├── email.go
│   │   ├── mailbox.go
│   │   └── attachment.go
│   └── ui/
│       ├── app.go            # Main Bubble Tea app (Model/Update/View)
│       ├── keys.go           # Keybindings
│       ├── styles.go         # Lip Gloss styles, colors
│       └── views/
│           ├── sidebar.go        # Mailbox list
│           ├── emaillist.go      # Email list (flat)
│           ├── threadlist.go     # Thread list with counts
│           └── emailreader.go    # Email viewer with markdown
```

### Keybindings

| Key | Action |
|-----|--------|
| ←/h | Navigate back (folder/list/thread) |
| →/l | Navigate forward / open |
| ↑/k | Move up |
| ↓/j | Move down |
| Enter | Open / expand thread |
| Esc/q | Go back |
| Space/Tab | Expand/collapse thread |
| c | Compose (not implemented) |
| r | Reply (not implemented) |
| R | Reply All (not implemented) |
| f | Forward (not implemented) |
| d | Delete |
| a | Archive (quick archive from thread list or email view) |
| s | Star/flag |
| u | Mark unread |
| / | Search (not implemented) |
| 1-5 | Switch accounts |
| Q/Ctrl+C | Quit |

## Remaining Work

### Phase 2: Compose & Reply
- [ ] Open $EDITOR for composing emails
- [ ] Reply with quoted text
- [ ] Reply all
- [ ] Forward with attachments
- [ ] Draft saving

### Phase 3: Advanced Email
- [ ] Full-text search
- [ ] Attachment download
- [ ] Multiple account support
- [ ] Keyboard shortcuts for batch operations

### Phase 4: Contacts (CardDAV)
- [ ] Fetch contacts from Fastmail
- [ ] Contact search/autocomplete in compose
- [ ] Contact viewer

### Phase 5: Calendar (CalDAV)
- [ ] Calendar view
- [ ] Event creation/editing
- [ ] Reminders

## Configuration Example

```yaml
# ~/.config/tuimail/config.yaml
accounts:
  - name: personal
    email: user@fastmail.com
    jmap_url: https://api.fastmail.com/jmap/session
    token_keyring: tuimail-personal  # Token stored in system keyring
```

## Running

```bash
make build   # Build binary
make run     # Run application
```

## Dependencies

- github.com/charmbracelet/bubbletea
- github.com/charmbracelet/bubbles
- github.com/charmbracelet/lipgloss
- github.com/charmbracelet/glamour
- git.sr.ht/~rockorager/go-jmap
- github.com/zalando/go-keyring
- gopkg.in/yaml.v3
