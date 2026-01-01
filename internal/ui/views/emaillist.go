package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/koren/tuimail/internal/models"
)

// Cyberpunk styles for email list
var (
	// Colors
	listColorNeonCyan   = lipgloss.Color("#00FFFF")
	listColorNeonPink   = lipgloss.Color("#FF2E97")
	listColorNeonOrange = lipgloss.Color("#FF6B35")
	listColorTextBright = lipgloss.Color("#EAEAEA")
	listColorTextNormal = lipgloss.Color("#B8B8B8")
	listColorTextMuted  = lipgloss.Color("#5C5C7A")
	listColorTextDim    = lipgloss.Color("#3D3D5C")
	listColorBgLight    = lipgloss.Color("#16213E")
	listColorBgSelect   = lipgloss.Color("#2D2D5A")

	emailListHeaderStyle = lipgloss.NewStyle().
				Foreground(listColorTextMuted).
				Background(listColorBgLight).
				Bold(true).
				Padding(0, 1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(listColorTextDim)

	emailRowStyle = lipgloss.NewStyle().
			Foreground(listColorTextNormal).
			Padding(0, 1)

	emailRowSelectedStyle = lipgloss.NewStyle().
				Foreground(listColorTextBright).
				Background(listColorBgSelect).
				Padding(0, 1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderLeft(true).
				BorderForeground(listColorNeonCyan)

	emailUnreadDotStyle = lipgloss.NewStyle().
				Foreground(listColorNeonPink).
				Bold(true)

	emailFromStyle = lipgloss.NewStyle().
			Foreground(listColorTextNormal)

	emailFromUnreadStyle = lipgloss.NewStyle().
				Foreground(listColorNeonCyan).
				Bold(true)

	emailSubjectStyle = lipgloss.NewStyle().
				Foreground(listColorTextNormal)

	emailSubjectUnreadStyle = lipgloss.NewStyle().
				Foreground(listColorTextBright).
				Bold(true)

	emailDateStyle = lipgloss.NewStyle().
			Foreground(listColorTextMuted)

	emailFlagStyle = lipgloss.NewStyle().
			Foreground(listColorNeonOrange)

	emailAttachStyle = lipgloss.NewStyle().
				Foreground(listColorTextMuted)

	emptyListStyle = lipgloss.NewStyle().
			Foreground(listColorTextMuted).
			Padding(2).
			Align(lipgloss.Center)
)

// EmailListView displays a list of emails
type EmailListView struct {
	emails   []models.Email
	selected int
	offset   int
	width    int
	height   int
}

// NewEmailListView creates a new email list view
func NewEmailListView(emails []models.Email, width, height int) *EmailListView {
	return &EmailListView{
		emails:   emails,
		selected: 0,
		offset:   0,
		width:    width,
		height:   height,
	}
}

// Select sets the selected email
func (v *EmailListView) Select(index int) {
	if index >= 0 && index < len(v.emails) {
		v.selected = index

		// Adjust scroll offset to keep selection visible
		visibleRows := v.height - 2
		if v.selected < v.offset {
			v.offset = v.selected
		} else if v.selected >= v.offset+visibleRows {
			v.offset = v.selected - visibleRows + 1
		}
	}
}

// Selected returns the currently selected email
func (v *EmailListView) Selected() *models.Email {
	if v.selected < len(v.emails) {
		return &v.emails[v.selected]
	}
	return nil
}

// SetSize updates the view dimensions
func (v *EmailListView) SetSize(width, height int) {
	v.width = width
	v.height = height
}

// View renders the email list
func (v *EmailListView) View() string {
	if len(v.emails) == 0 {
		emptyMsg := emptyListStyle.Render("◇ No messages in this folder")
		return lipgloss.Place(v.width, v.height, lipgloss.Center, lipgloss.Center, emptyMsg)
	}

	var b strings.Builder

	// Calculate column widths
	fromWidth := 22
	dateWidth := 10
	flagsWidth := 4
	subjectWidth := v.width - fromWidth - dateWidth - flagsWidth - 8
	if subjectWidth < 10 {
		subjectWidth = 10
	}

	// Render header
	header := fmt.Sprintf("  %-*s  %-*s  %*s",
		fromWidth, "FROM",
		subjectWidth, "SUBJECT",
		dateWidth, "DATE")
	b.WriteString(emailListHeaderStyle.Width(v.width).Render(header))
	b.WriteString("\n")

	// Calculate visible range
	visibleRows := v.height - 3
	if visibleRows < 1 {
		visibleRows = 1
	}

	endIdx := v.offset + visibleRows
	if endIdx > len(v.emails) {
		endIdx = len(v.emails)
	}

	// Render visible emails
	for i := v.offset; i < endIdx; i++ {
		email := v.emails[i]
		isSelected := i == v.selected

		b.WriteString(v.renderEmailRow(email, isSelected, fromWidth, subjectWidth, dateWidth))
		if i < endIdx-1 {
			b.WriteString("\n")
		}
	}

	// Scroll indicator
	if len(v.emails) > visibleRows {
		b.WriteString("\n")
		scrollInfo := fmt.Sprintf(" ▾ %d/%d ", v.selected+1, len(v.emails))
		scrollStyle := lipgloss.NewStyle().
			Foreground(listColorTextMuted).
			Align(lipgloss.Right).
			Width(v.width)
		b.WriteString(scrollStyle.Render(scrollInfo))
	}

	return b.String()
}

func (v *EmailListView) renderEmailRow(email models.Email, selected bool, fromWidth, subjectWidth, dateWidth int) string {
	// Unread indicator
	unreadDot := " "
	if email.IsUnread {
		unreadDot = emailUnreadDotStyle.Render("●")
	}

	// Flags (starred, attachment)
	flags := ""
	if email.IsFlagged {
		flags += emailFlagStyle.Render("★")
	} else {
		flags += " "
	}
	if email.HasAttachment {
		flags += emailAttachStyle.Render("◈")
	} else {
		flags += " "
	}

	// From
	from := email.FromDisplay()
	if len(from) > fromWidth {
		from = from[:fromWidth-1] + "…"
	}
	fromStyle := emailFromStyle
	if email.IsUnread {
		fromStyle = emailFromUnreadStyle
	}
	fromStr := fromStyle.Width(fromWidth).Render(from)

	// Subject
	subject := email.Subject
	if subject == "" {
		subject = "(no subject)"
	}
	if len(subject) > subjectWidth {
		subject = subject[:subjectWidth-1] + "…"
	}
	subjectStyle := emailSubjectStyle
	if email.IsUnread {
		subjectStyle = emailSubjectUnreadStyle
	}
	subjectStr := subjectStyle.Width(subjectWidth).Render(subject)

	// Date
	dateStr := emailDateStyle.Width(dateWidth).Render(email.DateDisplay())

	// Combine row
	row := fmt.Sprintf("%s%s %s  %s  %s", unreadDot, flags, fromStr, subjectStr, dateStr)

	// Apply selection style
	if selected {
		return emailRowSelectedStyle.Width(v.width).Render(row)
	}
	return emailRowStyle.Width(v.width).Render(row)
}
