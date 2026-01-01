# anneal

*email triage under noise*

A calm, terminal-based email client for Fastmail. Navigate your inbox with vim-style keys, archive threads with a single keystroke, and never see a red notification badge again.

```
┌─────────────────────────────────────────────────────────────────────┐
│ ◈ anneal                                    work@fastmail.com  email │
├──────────────────┬──────────────────────────────────────────────────┤
│ ◈ mailboxes      │ from              subject                   date │
│                  ├──────────────────────────────────────────────────┤
│   ▶ inbox    12  │ ● alice chen      quarterly planning    10:30 am │
│     drafts    1  │   bob smith       re: api changes       09:15 am │
│     sent         │   github          [ci] build passed     yesterday │
│     archive      │ ▶3 design team    logo feedback           nov 28 │
│     trash        │   newsletter      weekly digest           nov 27 │
│                  │                                                  │
│ ◈ labels         │                                                  │
│     work         │                                                  │
│     personal     │                                                  │
├──────────────────┴──────────────────────────────────────────────────┤
│ ↑/↓: select  →/enter: open  ←/esc: back  a: archive  ?: help        │
└─────────────────────────────────────────────────────────────────────┘
```

## What it does

- Read, compose, reply, and forward emails
- Archive or delete entire conversation threads
- View and open attachments
- Cache emails locally for fast startup
- Store your API token securely in the system keyring

## What it doesn't do (yet)

- Multiple accounts (single Fastmail account only)
- Search
- Contacts or calendar
- Rich text compose (plain text only)
- Inline images
- Drafts auto-save

## Install

```bash
git clone https://github.com/the9x/anneal
cd anneal
make build
```

Requires **Go 1.21+** and a **Fastmail account**.

## First run

```bash
./bin/anneal
```

You'll be prompted for:
1. Your Fastmail email address
2. An API token (get one from Fastmail Settings → Privacy & Security → API tokens)

Your token is stored in the system keyring, not in a plain text file.

## How it works

The interface has a simple left-to-right flow:

```
folders → messages → thread → email → attachments
```

Press `→` or `Enter` to go deeper. Press `←` or `Esc` to go back.

### The sidebar

The left pane shows your mailboxes. Numbers indicate unread count.

```
◈ mailboxes
  ▶ inbox    12     ← selected, 12 unread
    drafts    1
    sent
    archive
```

### The message list

The main pane shows threads. A `●` means unread. A number like `▶3` means the thread has 3 emails.

```
● alice chen      quarterly planning    10:30 am   ← unread
  bob smith       re: api changes       09:15 am
▶3 design team    logo feedback           nov 28   ← 3-email thread
```

### Reading email

When you open an email, the content is displayed with basic markdown rendering. Scroll with `↑`/`↓`. If there are attachments, press `→` to select and open them.

### Composing

Press `c` to compose, `r` to reply, `R` to reply all, `f` to forward.

```
┌─────────────────────────────────────────────────────────────────────┐
│ to:      alice@example.com                                          │
│ cc:                                                                 │
│ subject: re: quarterly planning                                     │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│ Sounds good, let's meet Thursday.                                   │
│                                                                     │
│ > On Nov 28, alice wrote:                                           │
│ > Can we schedule a call to discuss Q1?                             │
│                                                                     │
├─────────────────────────────────────────────────────────────────────┤
│ tab: next field  ctrl+s: send  esc: cancel                          │
└─────────────────────────────────────────────────────────────────────┘
```

Use `Tab` to move between fields. `Ctrl+S` to send. `Esc` to cancel.

## Keybindings

### Navigation

| Key | Action |
|-----|--------|
| `→` or `Enter` | Open / go deeper |
| `←` or `Esc` | Back |
| `↑` or `k` | Up / scroll up |
| `↓` or `j` | Down / scroll down |
| `g` | Jump to top |
| `G` | Jump to bottom |

### Actions

| Key | Action |
|-----|--------|
| `c` | Compose new email |
| `r` | Reply |
| `R` | Reply all |
| `f` | Forward |
| `a` | Archive (whole thread) |
| `d` | Delete |
| `u` | Toggle read, or undelete in Trash |
| `?` | Show all keybindings |
| `Q` | Quit |

## Files

| Path | Purpose |
|------|---------|
| `~/.config/anneal/config.yaml` | Account settings |
| `~/.local/share/anneal/cache.db` | Local email cache |
| System keyring | API token (secure) |

## Troubleshooting

**"No API token found"** — Run the app again and enter your token, or check that your system keyring is working.

**Slow startup** — First run fetches all mailboxes and recent emails. Subsequent runs load from cache instantly.

**Attachments won't open** — anneal uses the `open` command (macOS). On Linux, you may need to adjust this.

## License

MIT
