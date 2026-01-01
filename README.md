# TuiMail

A beautiful TUI email client for Fastmail, built with [Bubble Tea](https://github.com/charmbracelet/bubbletea).

![TuiMail Screenshot](docs/screenshot.png)

## Features

- **JMAP Protocol** - Native Fastmail API integration
- **Vim-style Navigation** - j/k, g/G, and familiar keybindings
- **Multi-Account Support** - Switch between accounts with number keys
- **Secure Token Storage** - API tokens stored in system keyring
- **Beautiful TUI** - Charm.sh aesthetic with Lip Gloss styling

## Installation

### From Source

```bash
git clone https://github.com/koren/tuimail
cd tuimail
make build
make install  # Optional: install to /usr/local/bin
```

### Requirements

- Go 1.21+
- Fastmail account with API access

## Setup

1. **Get your Fastmail API token:**
   - Go to Fastmail Settings → Privacy & Security → Integrations
   - Under "API tokens", click "Manage"
   - Create a new token with Mail access

2. **Run TuiMail:**
   ```bash
   ./bin/tuimail
   ```

   On first run, you'll be prompted to enter your email and API token.

## Configuration

Configuration is stored in `~/.config/tuimail/config.yaml`:

```yaml
accounts:
  - name: Work
    email: work@fastmail.com
    default: true

  - name: Personal
    email: personal@fastmail.com

theme: dark
editor: vim
preview_pane: true
threading: true
page_size: 50
```

API tokens are stored securely in your system's keyring (macOS Keychain, GNOME Keyring, etc.).

## Keybindings

### Navigation
| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `g` | Go to top |
| `G` | Go to bottom |
| `Enter` | Open email/mailbox |
| `Esc` / `q` | Back |
| `Tab` | Switch pane |

### Email Actions
| Key | Action |
|-----|--------|
| `c` | Compose new email |
| `r` | Reply |
| `R` | Reply all |
| `f` | Forward |
| `d` | Delete |
| `a` | Archive |
| `s` | Star/flag |
| `u` | Mark unread |

### Other
| Key | Action |
|-----|--------|
| `/` | Search |
| `Ctrl+R` | Refresh |
| `1-5` | Switch account |
| `?` | Help |
| `Q` | Quit |

## Architecture

```
tuimail/
├── main.go                 # Entry point
├── internal/
│   ├── config/             # Configuration & keyring
│   ├── jmap/               # JMAP client wrapper
│   ├── models/             # Domain models
│   └── ui/                 # Bubble Tea UI
│       ├── app.go          # Main app model
│       ├── styles.go       # Lip Gloss styles
│       ├── keys.go         # Keybindings
│       └── views/          # View components
```

## Dependencies

- [Bubble Tea](https://github.com/charmbracelet/bubbletea) - TUI framework
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) - Styling
- [Bubbles](https://github.com/charmbracelet/bubbles) - TUI components
- [go-jmap](https://git.sr.ht/~rockorager/go-jmap) - JMAP client
- [go-keyring](https://github.com/zalando/go-keyring) - Secure credential storage

## License

MIT
