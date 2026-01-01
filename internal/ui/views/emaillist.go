package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/the9x/anneal/internal/models"
)

// anneal brand colors
var (
	listColorPrimary   = lipgloss.Color("#d4d2e3")
	listColorSecondary = lipgloss.Color("#9795b5")
	listColorDim       = lipgloss.Color("#5a5880")
	listColorBg        = lipgloss.Color("#1d1d40")
	listColorBgSelect  = lipgloss.Color("#2d2d5a")
	listColorAccent    = lipgloss.Color("#e61e25") // used sparingly

	emailListHeaderStyle = lipgloss.NewStyle().
				Foreground(listColorDim).
				Background(listColorBg).
				Padding(0, 1)

	emailRowStyle = lipgloss.NewStyle().
			Foreground(listColorPrimary).
			Padding(0, 1)

	emailRowSelectedStyle = lipgloss.NewStyle().
				Foreground(listColorPrimary).
				Background(listColorBgSelect).
				Bold(true).
				Padding(0, 1)

	emailUnreadDotStyle = lipgloss.NewStyle().
				Foreground(listColorPrimary)

	emailFromStyle = lipgloss.NewStyle().
			Foreground(listColorPrimary)

	emailFromUnreadStyle = lipgloss.NewStyle().
				Foreground(listColorPrimary).
				Bold(true)

	emailSubjectStyle = lipgloss.NewStyle().
				Foreground(listColorSecondary)

	emailSubjectUnreadStyle = lipgloss.NewStyle().
				Foreground(listColorPrimary).
				Bold(true)

	emailDateStyle = lipgloss.NewStyle().
			Foreground(listColorDim)

	emailFlagStyle = lipgloss.NewStyle().
			Foreground(listColorAccent)

	emailAttachStyle = lipgloss.NewStyle().
				Foreground(listColorDim)

	emptyListStyle = lipgloss.NewStyle().
			Foreground(listColorDim).
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
		fromWidth, "from",
		subjectWidth, "subject",
		dateWidth, "date")
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
			Foreground(listColorDim).
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
