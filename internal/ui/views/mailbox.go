package views

import (
	"fmt"
	"sort"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/the9x/anneal/internal/models"
)

// anneal brand colors
var (
	mbColorPrimary   = lipgloss.Color("#d4d2e3")
	mbColorSecondary = lipgloss.Color("#9795b5")
	mbColorDim       = lipgloss.Color("#5a5880")
	mbColorBg        = lipgloss.Color("#1d1d40")
	mbColorBgSelect  = lipgloss.Color("#2d2d5a")

	mailboxTitleStyle = lipgloss.NewStyle().
				Foreground(mbColorSecondary).
				Padding(0, 1).
				MarginBottom(1)

	mailboxItemStyle = lipgloss.NewStyle().
				Foreground(mbColorPrimary).
				PaddingLeft(1)

	mailboxSelectedStyle = lipgloss.NewStyle().
				Foreground(mbColorPrimary).
				Background(mbColorBgSelect).
				Bold(true).
				PaddingLeft(1)

	mailboxUnreadStyle = lipgloss.NewStyle().
				Foreground(mbColorPrimary)

	mailboxIconStyle = lipgloss.NewStyle().
				Foreground(mbColorDim)

	mailboxIconActiveStyle = lipgloss.NewStyle().
				Foreground(mbColorPrimary)
)

// MailboxView displays the mailbox list
type MailboxView struct {
	mailboxes []models.Mailbox
	selected  int
	width     int
	height    int
}

// NewMailboxView creates a new mailbox view
func NewMailboxView(mailboxes []models.Mailbox) *MailboxView {
	// Sort mailboxes: system first by role, then custom alphabetically
	sorted := make([]models.Mailbox, len(mailboxes))
	copy(sorted, mailboxes)

	roleOrder := map[string]int{
		"inbox":   0,
		"drafts":  1,
		"sent":    2,
		"archive": 3,
		"trash":   4,
		"junk":    5,
	}

	sort.Slice(sorted, func(i, j int) bool {
		ri, oki := roleOrder[sorted[i].Role]
		rj, okj := roleOrder[sorted[j].Role]

		if oki && okj {
			return ri < rj
		}
		if oki {
			return true
		}
		if okj {
			return false
		}
		return sorted[i].Name < sorted[j].Name
	})

	return &MailboxView{
		mailboxes: sorted,
		selected:  0,
	}
}

// Select sets the selected mailbox
func (v *MailboxView) Select(index int) {
	if index >= 0 && index < len(v.mailboxes) {
		v.selected = index
	}
}

// Selected returns the currently selected mailbox
func (v *MailboxView) Selected() *models.Mailbox {
	if v.selected < len(v.mailboxes) {
		return &v.mailboxes[v.selected]
	}
	return nil
}

// SetSize updates the view dimensions
func (v *MailboxView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

// View renders the mailbox list
func (v *MailboxView) View() string {
	var b strings.Builder

	// Title
	title := mailboxTitleStyle.Render("◈ mailboxes")
	b.WriteString(title)
	b.WriteString("\n")

	// Separate system and custom mailboxes
	var system, custom []models.Mailbox
	for _, mb := range v.mailboxes {
		if mb.IsSystem() {
			system = append(system, mb)
		} else {
			custom = append(custom, mb)
		}
	}

	// Render system mailboxes
	idx := 0
	for _, mb := range system {
		b.WriteString(v.renderMailbox(mb, idx == v.selected))
		b.WriteString("\n")
		idx++
	}

	// Render custom mailboxes if any
	if len(custom) > 0 {
		b.WriteString("\n")
		labelTitle := mailboxTitleStyle.Render("◈ labels")
		b.WriteString(labelTitle)
		b.WriteString("\n")

		for _, mb := range custom {
			b.WriteString(v.renderMailbox(mb, idx == v.selected))
			b.WriteString("\n")
			idx++
		}
	}

	return b.String()
}

func (v *MailboxView) renderMailbox(mb models.Mailbox, selected bool) string {
	name := mb.DisplayName()
	icon := v.getIcon(mb.Role, selected)

	// Truncate name if too long
	maxNameLen := 12
	if len(name) > maxNameLen {
		name = name[:maxNameLen-1] + "…"
	}

	// Format unread count
	var countStr string
	if mb.UnreadCount > 0 {
		countStr = mailboxUnreadStyle.Render(fmt.Sprintf(" %d", mb.UnreadCount))
	}

	// Build the line
	style := mailboxItemStyle
	if selected {
		style = mailboxSelectedStyle
	}

	line := fmt.Sprintf("%s %s", icon, name)
	styledLine := style.Render(line)

	// Add count at the end
	return styledLine + countStr
}

func (v *MailboxView) getIcon(role string, selected bool) string {
	style := mailboxIconStyle
	if selected {
		style = mailboxIconActiveStyle
	}

	var icon string
	switch role {
	case "inbox":
		icon = "▶"
	case "drafts":
		icon = "◇"
	case "sent":
		icon = "△"
	case "archive":
		icon = "▣"
	case "trash":
		icon = "▽"
	case "junk":
		icon = "⊘"
	default:
		icon = "◆"
	}
	return style.Render(icon)
}
