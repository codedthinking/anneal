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

// Cyberpunk styles for thread list
var (
	threadColorNeonCyan   = lipgloss.Color("#00FFFF")
	threadColorNeonPink   = lipgloss.Color("#FF2E97")
	threadColorNeonPurple = lipgloss.Color("#BD00FF")
	threadColorTextBright = lipgloss.Color("#EAEAEA")
	threadColorTextNormal = lipgloss.Color("#B8B8B8")
	threadColorTextMuted  = lipgloss.Color("#5C5C7A")
	threadColorTextDim    = lipgloss.Color("#3D3D5C")
	threadColorBgLight    = lipgloss.Color("#16213E")
	threadColorBgSelect   = lipgloss.Color("#2D2D5A")

	threadHeaderStyle = lipgloss.NewStyle().
				Foreground(threadColorTextMuted).
				Background(threadColorBgLight).
				Bold(true).
				Padding(0, 1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(threadColorTextDim)

	threadRowStyle = lipgloss.NewStyle().
			Foreground(threadColorTextNormal).
			Padding(0, 1)

	threadRowSelectedStyle = lipgloss.NewStyle().
				Foreground(threadColorTextBright).
				Background(threadColorBgSelect).
				Padding(0, 1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderLeft(true).
				BorderForeground(threadColorNeonCyan)

	threadUnreadDotStyle = lipgloss.NewStyle().
				Foreground(threadColorNeonPink).
				Bold(true)

	threadFromStyle = lipgloss.NewStyle().
			Foreground(threadColorTextNormal)

	threadFromUnreadStyle = lipgloss.NewStyle().
				Foreground(threadColorNeonCyan).
				Bold(true)

	threadSubjectStyle = lipgloss.NewStyle().
				Foreground(threadColorTextNormal)

	threadSubjectUnreadStyle = lipgloss.NewStyle().
					Foreground(threadColorTextBright).
					Bold(true)

	threadCountStyle = lipgloss.NewStyle().
				Foreground(threadColorNeonPurple).
				Bold(true)

	threadDateStyle = lipgloss.NewStyle().
			Foreground(threadColorTextMuted)

	threadExpandedStyle = lipgloss.NewStyle().
				Foreground(threadColorNeonCyan)

	threadEmptyStyle = lipgloss.NewStyle().
				Foreground(threadColorTextMuted).
				Padding(2).
				Align(lipgloss.Center)
)

// ThreadListView displays a list of threads
type ThreadListView struct {
	threads  []Thread
	selected int
	offset   int
	width    int
	height   int
}

// NewThreadListView creates a new thread list view
func NewThreadListView(width, height int) *ThreadListView {
	return &ThreadListView{
		threads:  []Thread{},
		selected: 0,
		offset:   0,
		width:    width,
		height:   height,
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
}

// View renders the thread list
func (v *ThreadListView) View() string {
	if len(v.threads) == 0 {
		emptyMsg := threadEmptyStyle.Render("◇ No messages in this folder")
		return lipgloss.Place(v.width, v.height, lipgloss.Center, lipgloss.Center, emptyMsg)
	}

	var b strings.Builder

	// Calculate column widths
	fromWidth := 20
	dateWidth := 10
	countWidth := 4
	subjectWidth := v.width - fromWidth - dateWidth - countWidth - 10
	if subjectWidth < 10 {
		subjectWidth = 10
	}

	// Render header
	header := fmt.Sprintf("  %-*s  %-*s  %*s",
		fromWidth, "FROM",
		subjectWidth, "SUBJECT",
		dateWidth, "DATE")
	b.WriteString(threadHeaderStyle.Width(v.width).Render(header))
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
			Foreground(threadColorTextMuted).
			Align(lipgloss.Right).
			Width(v.width)
		b.WriteString(scrollStyle.Render(scrollInfo))
	}

	return b.String()
}

func (v *ThreadListView) renderThreadRow(thread Thread, selected bool, fromWidth, subjectWidth, dateWidth int) string {
	// Unread indicator
	unreadDot := " "
	if thread.UnreadCnt > 0 {
		unreadDot = threadUnreadDotStyle.Render("●")
	}

	// Thread/email indicator
	// Single email: no indicator (direct open)
	// Multi-email thread: show count with expand/collapse icon
	countStr := ""
	if thread.EmailCnt > 1 {
		if thread.Expanded {
			countStr = threadExpandedStyle.Render(fmt.Sprintf("▼%d", thread.EmailCnt))
		} else {
			countStr = threadCountStyle.Render(fmt.Sprintf("▶%d", thread.EmailCnt))
		}
	} else {
		// Single email - show mail icon to distinguish from threads
		countStr = threadDateStyle.Render("◇ ")
	}

	// From
	from := thread.From
	if len(from) > fromWidth {
		from = from[:fromWidth-1] + "…"
	}
	fromStyle := threadFromStyle
	if thread.UnreadCnt > 0 {
		fromStyle = threadFromUnreadStyle
	}
	fromStr := fromStyle.Width(fromWidth).Render(from)

	// Subject
	subject := thread.Subject
	if subject == "" {
		subject = "(no subject)"
	}
	if len(subject) > subjectWidth {
		subject = subject[:subjectWidth-1] + "…"
	}
	subjectStyle := threadSubjectStyle
	if thread.UnreadCnt > 0 {
		subjectStyle = threadSubjectUnreadStyle
	}
	subjectStr := subjectStyle.Width(subjectWidth).Render(subject)

	// Date
	dateStr := threadDateStyle.Width(dateWidth).Render(thread.Date)

	// Combine row
	row := fmt.Sprintf("%s%s %s  %s  %s", unreadDot, countStr, fromStr, subjectStr, dateStr)

	// Apply selection style
	if selected {
		return threadRowSelectedStyle.Width(v.width).Render(row)
	}
	return threadRowStyle.Width(v.width).Render(row)
}
