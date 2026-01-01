package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Thread represents a group of emails in a conversation
type Thread struct {
	ID        string
	Subject   string
	Preview   string
	Date      string
	From      string
	EmailCnt  int
	UnreadCnt int
	Expanded  bool
}

// anneal brand colors
var (
	thColorPrimary   = lipgloss.Color("#d4d2e3")
	thColorSecondary = lipgloss.Color("#9795b5")
	thColorDim       = lipgloss.Color("#5a5880")
	thColorBg        = lipgloss.Color("#1d1d40")
	thColorBgSelect  = lipgloss.Color("#2d2d5a")

	threadHeaderStyle = lipgloss.NewStyle().
				Foreground(thColorDim).
				Background(thColorBg).
				Padding(0, 1)

	threadRowStyle = lipgloss.NewStyle().
			Foreground(thColorPrimary).
			Padding(0, 1)

	threadRowSelectedStyle = lipgloss.NewStyle().
				Foreground(thColorPrimary).
				Background(thColorBgSelect).
				Bold(true).
				Padding(0, 1)

	threadUnreadDotStyle = lipgloss.NewStyle().
				Foreground(thColorPrimary)

	threadFromStyle = lipgloss.NewStyle().
			Foreground(thColorPrimary)

	threadFromUnreadStyle = lipgloss.NewStyle().
				Foreground(thColorPrimary).
				Bold(true)

	threadSubjectStyle = lipgloss.NewStyle().
				Foreground(thColorSecondary)

	threadSubjectUnreadStyle = lipgloss.NewStyle().
					Foreground(thColorPrimary).
					Bold(true)

	threadCountStyle = lipgloss.NewStyle().
				Foreground(thColorPrimary)

	threadDateStyle = lipgloss.NewStyle().
			Foreground(thColorDim)

	threadExpandedStyle = lipgloss.NewStyle().
				Foreground(thColorPrimary)

	threadEmptyStyle = lipgloss.NewStyle().
				Foreground(thColorDim).
				Padding(2).
				Align(lipgloss.Center)
)

const maxListWidth = 80

// ThreadListView displays a list of threads
type ThreadListView struct {
	threads      []Thread
	selected     int
	offset       int
	width        int
	contentWidth int
	height       int
}

// NewThreadListView creates a new thread list view
func NewThreadListView(width, height int) *ThreadListView {
	contentWidth := width
	if contentWidth > maxListWidth {
		contentWidth = maxListWidth
	}
	return &ThreadListView{
		threads:      []Thread{},
		selected:     0,
		offset:       0,
		width:        width,
		contentWidth: contentWidth,
		height:       height,
	}
}

// UpdateThreads updates the thread list
func (v *ThreadListView) UpdateThreads(threads []Thread) {
	v.threads = threads
}

// Select sets the selected thread
func (v *ThreadListView) Select(index int) {
	if index >= 0 && index < len(v.threads) {
		v.selected = index

		// Adjust scroll offset
		visibleRows := v.height - 2
		if v.selected < v.offset {
			v.offset = v.selected
		} else if v.selected >= v.offset+visibleRows {
			v.offset = v.selected - visibleRows + 1
		}
	}
}

// SetSize updates the view dimensions
func (v *ThreadListView) SetSize(width, height int) {
	v.width = width
	v.height = height
	v.contentWidth = width
	if v.contentWidth > maxListWidth {
		v.contentWidth = maxListWidth
	}
}

// View renders the thread list
func (v *ThreadListView) View() string {
	if len(v.threads) == 0 {
		emptyMsg := threadEmptyStyle.Render("◇ No messages in this folder")
		return lipgloss.Place(v.width, v.height, lipgloss.Center, lipgloss.Center, emptyMsg)
	}

	var b strings.Builder

	// Calculate column widths based on contentWidth (capped at maxListWidth)
	fromWidth := 18
	dateWidth := 8
	countWidth := 3
	subjectWidth := v.contentWidth - fromWidth - dateWidth - countWidth - 8
	if subjectWidth < 10 {
		subjectWidth = 10
	}

	// Render header
	header := fmt.Sprintf("    %-*s %-*s %*s",
		fromWidth, "from",
		subjectWidth, "subject",
		dateWidth, "date")
	if len(header) > v.contentWidth {
		header = header[:v.contentWidth]
	}
	b.WriteString(threadHeaderStyle.MaxWidth(v.contentWidth).Render(header))
	b.WriteString("\n")

	// Calculate visible range
	visibleRows := v.height - 3
	if visibleRows < 1 {
		visibleRows = 1
	}

	endIdx := v.offset + visibleRows
	if endIdx > len(v.threads) {
		endIdx = len(v.threads)
	}

	// Render visible threads
	for i := v.offset; i < endIdx; i++ {
		thread := v.threads[i]
		isSelected := i == v.selected

		b.WriteString(v.renderThreadRow(thread, isSelected, fromWidth, subjectWidth, dateWidth))
		if i < endIdx-1 {
			b.WriteString("\n")
		}
	}

	// Scroll indicator
	if len(v.threads) > visibleRows {
		b.WriteString("\n")
		scrollInfo := fmt.Sprintf(" ▾ %d/%d ", v.selected+1, len(v.threads))
		scrollStyle := lipgloss.NewStyle().
			Foreground(thColorDim).
			Align(lipgloss.Right).
			MaxWidth(v.contentWidth)
		b.WriteString(scrollStyle.Render(scrollInfo))
	}

	// Center the content if width is larger than contentWidth
	content := b.String()
	if v.width > v.contentWidth {
		return lipgloss.Place(v.width, 0, lipgloss.Center, lipgloss.Top, content)
	}
	return content
}

func (v *ThreadListView) renderThreadRow(thread Thread, selected bool, fromWidth, subjectWidth, dateWidth int) string {
	// Build plain text first, then style

	// Unread indicator (1 char)
	unreadDot := " "
	if thread.UnreadCnt > 0 {
		unreadDot = "●"
	}

	// Thread/email indicator (3 chars max)
	countStr := "   "
	if thread.EmailCnt > 1 {
		if thread.Expanded {
			countStr = fmt.Sprintf("▼%-2d", thread.EmailCnt)
		} else {
			countStr = fmt.Sprintf("▶%-2d", thread.EmailCnt)
		}
	}

	// From - truncate and pad
	from := thread.From
	if len(from) > fromWidth {
		from = from[:fromWidth-1] + "…"
	}
	from = fmt.Sprintf("%-*s", fromWidth, from)

	// Subject - truncate and pad
	subject := thread.Subject
	if subject == "" {
		subject = "(no subject)"
	}
	if len(subject) > subjectWidth {
		subject = subject[:subjectWidth-1] + "…"
	}
	subject = fmt.Sprintf("%-*s", subjectWidth, subject)

	// Date - right align
	date := fmt.Sprintf("%*s", dateWidth, thread.Date)

	// Build the row as plain text
	row := fmt.Sprintf("%s%s%s %s %s", unreadDot, countStr, from, subject, date)

	// Truncate to contentWidth to prevent any overflow
	if len(row) > v.contentWidth {
		row = row[:v.contentWidth]
	}

	// Now apply styling to the complete row
	if selected {
		return threadRowSelectedStyle.MaxWidth(v.contentWidth).Render(row)
	}

	// For unselected, style individual parts
	var styled strings.Builder
	if thread.UnreadCnt > 0 {
		styled.WriteString(threadUnreadDotStyle.Render(unreadDot))
		styled.WriteString(threadCountStyle.Render(countStr))
		styled.WriteString(threadFromUnreadStyle.Render(from))
		styled.WriteString(" ")
		styled.WriteString(threadSubjectUnreadStyle.Render(subject))
	} else {
		styled.WriteString(threadDateStyle.Render(unreadDot))
		styled.WriteString(threadDateStyle.Render(countStr))
		styled.WriteString(threadFromStyle.Render(from))
		styled.WriteString(" ")
		styled.WriteString(threadSubjectStyle.Render(subject))
	}
	styled.WriteString(" ")
	styled.WriteString(threadDateStyle.Render(date))

	return threadRowStyle.MaxWidth(v.contentWidth).Render(styled.String())
}
