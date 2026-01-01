package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/koren/tuimail/internal/config"
	"github.com/koren/tuimail/internal/jmap"
	"github.com/koren/tuimail/internal/models"
	"github.com/koren/tuimail/internal/ui/views"
)

// ViewState represents the navigation depth
// Flow: Folders → Messages (threads + emails) → [Thread contents] → Email
// Single emails skip the thread contents step
type ViewState int

const (
	ViewFolders  ViewState = iota // Sidebar focused, selecting mailbox
	ViewMessages                  // Message list: threads (multi-email) and standalone emails
	ViewThread                    // Inside a multi-email thread, selecting which email
	ViewEmail                     // Reading single email
)

// Thread represents a group of emails in a conversation
type Thread struct {
	ID        string
	Subject   string
	Emails    []models.Email
	Preview   string
	Date      string
	From      string
	UnreadCnt int
	Expanded  bool
}

// App is the main application model
type App struct {
	cfg       *config.Config
	client    *jmap.Client
	keys      KeyMap
	help      help.Model
	spinner   spinner.Model
	width     int
	height    int
	viewState ViewState
	loading   bool
	err       error

	// Data
	mailboxes       []models.Mailbox
	selectedMailbox int
	emails          []models.Email
	threads         []Thread
	selectedThread  int
	selectedInThread int
	currentEmail    *models.Email

	// Views
	mailboxView *views.MailboxView
	threadList  *views.ThreadListView
	emailReader *views.EmailReaderView
}

// NewApp creates a new application instance
func NewApp(cfg *config.Config, client *jmap.Client) *App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	return &App{
		cfg:       cfg,
		client:    client,
		keys:      DefaultKeyMap(),
		help:      help.New(),
		spinner:   s,
		viewState: ViewFolders,
		loading:   true,
	}
}

// Init initializes the application
func (a *App) Init() tea.Cmd {
	return tea.Batch(
		a.spinner.Tick,
		a.loadMailboxes,
	)
}

// Msg types for async operations
type mailboxesLoadedMsg struct {
	mailboxes []models.Mailbox
	err       error
}

type emailsLoadedMsg struct {
	emails []models.Email
	err    error
}

type emailLoadedMsg struct {
	email *models.Email
	err   error
}

type emailActionMsg struct {
	err error
}

func (a *App) loadMailboxes() tea.Msg {
	mailboxes, err := a.client.GetMailboxes()
	return mailboxesLoadedMsg{mailboxes: mailboxes, err: err}
}

func (a *App) loadEmails(mailboxID string) tea.Cmd {
	return func() tea.Msg {
		emails, err := a.client.GetEmails(mailboxID, a.cfg.PageSize)
		return emailsLoadedMsg{emails: emails, err: err}
	}
}

func (a *App) loadEmail(emailID string) tea.Cmd {
	return func() tea.Msg {
		email, err := a.client.GetEmail(emailID)
		return emailLoadedMsg{email: email, err: err}
	}
}

// convertToViewThreads converts app threads to view threads
func (a *App) convertToViewThreads() []views.Thread {
	viewThreads := make([]views.Thread, len(a.threads))
	for i, t := range a.threads {
		viewThreads[i] = views.Thread{
			ID:        t.ID,
			Subject:   t.Subject,
			Preview:   t.Preview,
			Date:      t.Date,
			From:      t.From,
			EmailCnt:  len(t.Emails),
			UnreadCnt: t.UnreadCnt,
			Expanded:  t.Expanded,
		}
	}
	return viewThreads
}

// groupEmailsIntoThreads groups emails by thread ID
func (a *App) groupEmailsIntoThreads(emails []models.Email) []Thread {
	threadMap := make(map[string]*Thread)
	var threadOrder []string

	for _, email := range emails {
		tid := email.ThreadID
		if tid == "" {
			tid = email.ID // Fallback to email ID if no thread
		}

		if t, exists := threadMap[tid]; exists {
			t.Emails = append(t.Emails, email)
			if email.IsUnread {
				t.UnreadCnt++
			}
			// Update thread date to most recent
			if email.ReceivedAt.After(t.Emails[0].ReceivedAt) {
				t.Date = email.DateDisplay()
			}
		} else {
			threadOrder = append(threadOrder, tid)
			unread := 0
			if email.IsUnread {
				unread = 1
			}
			threadMap[tid] = &Thread{
				ID:        tid,
				Subject:   email.Subject,
				Emails:    []models.Email{email},
				Preview:   email.Preview,
				Date:      email.DateDisplay(),
				From:      email.FromDisplay(),
				UnreadCnt: unread,
				Expanded:  false,
			}
		}
	}

	// Build ordered slice
	threads := make([]Thread, 0, len(threadOrder))
	for _, tid := range threadOrder {
		threads = append(threads, *threadMap[tid])
	}

	return threads
}

// Update handles messages
func (a *App) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		a.width = msg.Width
		a.height = msg.Height
		a.help.Width = msg.Width
		return a, nil

	case tea.KeyMsg:
		// Global keys
		if key.Matches(msg, a.keys.Quit) {
			return a, tea.Quit
		}
		if key.Matches(msg, a.keys.Help) {
			a.help.ShowAll = !a.help.ShowAll
			return a, nil
		}

		// Handle navigation
		return a.handleKeyPress(msg)

	case spinner.TickMsg:
		var cmd tea.Cmd
		a.spinner, cmd = a.spinner.Update(msg)
		return a, cmd

	case mailboxesLoadedMsg:
		a.loading = false
		if msg.err != nil {
			a.err = msg.err
			return a, nil
		}
		a.mailboxes = msg.mailboxes
		a.mailboxView = views.NewMailboxView(a.mailboxes)

		// Find inbox and load emails
		for i, mb := range a.mailboxes {
			if mb.Role == "inbox" {
				a.selectedMailbox = i
				a.mailboxView.Select(i)
				a.loading = true
				return a, a.loadEmails(mb.ID)
			}
		}
		if len(a.mailboxes) > 0 {
			a.loading = true
			return a, a.loadEmails(a.mailboxes[0].ID)
		}
		return a, nil

	case emailsLoadedMsg:
		a.loading = false
		if msg.err != nil {
			a.err = msg.err
			return a, nil
		}
		a.emails = msg.emails
		oldThreadCount := len(a.threads)
		a.threads = a.groupEmailsIntoThreads(msg.emails)

		// Preserve selection on refresh, reset on initial load
		if oldThreadCount == 0 {
			a.selectedThread = 0
			a.selectedInThread = 0
		} else {
			// Make sure selection is still valid
			if a.selectedThread >= len(a.threads) {
				a.selectedThread = len(a.threads) - 1
				if a.selectedThread < 0 {
					a.selectedThread = 0
				}
			}
			a.selectedInThread = 0
		}

		if a.threadList == nil {
			a.threadList = views.NewThreadListView(a.width-26, a.height-6)
		}
		a.threadList.Select(a.selectedThread)
		a.viewState = ViewMessages
		return a, nil

	case emailLoadedMsg:
		a.loading = false
		if msg.err != nil {
			a.err = msg.err
			return a, nil
		}
		a.currentEmail = msg.email
		a.emailReader = views.NewEmailReaderView(msg.email, a.width-26, a.height-6)
		a.viewState = ViewEmail

		// Mark as read
		if msg.email.IsUnread {
			go a.client.MarkAsRead(msg.email.ID)
		}
		return a, nil

	case emailActionMsg:
		if msg.err != nil {
			a.err = msg.err
		}
		// Refresh emails after action
		if len(a.mailboxes) > 0 && a.selectedMailbox < len(a.mailboxes) {
			return a, a.loadEmails(a.mailboxes[a.selectedMailbox].ID)
		}
		return a, nil
	}

	return a, tea.Batch(cmds...)
}

func (a *App) handleKeyPress(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// Navigation: ← goes back, → goes forward, Enter opens, Esc goes back
	switch a.viewState {
	case ViewFolders:
		return a.handleFoldersKeys(msg)
	case ViewMessages:
		return a.handleMessagesKeys(msg)
	case ViewThread:
		return a.handleThreadKeys(msg)
	case ViewEmail:
		return a.handleEmailKeys(msg)
	}
	return a, nil
}

func (a *App) handleFoldersKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Up):
		if a.selectedMailbox > 0 {
			a.selectedMailbox--
			if a.mailboxView != nil {
				a.mailboxView.Select(a.selectedMailbox)
			}
		}
	case key.Matches(msg, a.keys.Down):
		if a.selectedMailbox < len(a.mailboxes)-1 {
			a.selectedMailbox++
			if a.mailboxView != nil {
				a.mailboxView.Select(a.selectedMailbox)
			}
		}
	case key.Matches(msg, a.keys.Right), key.Matches(msg, a.keys.Enter):
		// Open mailbox → go to thread list
		if len(a.mailboxes) > 0 {
			a.loading = true
			return a, a.loadEmails(a.mailboxes[a.selectedMailbox].ID)
		}
	case key.Matches(msg, a.keys.Back):
		// Already at leftmost level, quit
		return a, tea.Quit
	}
	return a, nil
}

func (a *App) handleMessagesKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Up):
		if a.selectedThread > 0 {
			a.selectedThread--
			if a.threadList != nil {
				a.threadList.Select(a.selectedThread)
			}
		}
	case key.Matches(msg, a.keys.Down):
		if a.selectedThread < len(a.threads)-1 {
			a.selectedThread++
			if a.threadList != nil {
				a.threadList.Select(a.selectedThread)
			}
		}
	case key.Matches(msg, a.keys.Top):
		a.selectedThread = 0
		if a.threadList != nil {
			a.threadList.Select(a.selectedThread)
		}
	case key.Matches(msg, a.keys.Bottom):
		a.selectedThread = len(a.threads) - 1
		if a.threadList != nil {
			a.threadList.Select(a.selectedThread)
		}
	case key.Matches(msg, a.keys.Right), key.Matches(msg, a.keys.Enter):
		// Open thread
		if len(a.threads) > 0 && a.selectedThread < len(a.threads) {
			thread := &a.threads[a.selectedThread]
			if len(thread.Emails) == 1 {
				// Single email thread - go directly to email
				a.loading = true
				return a, a.loadEmail(thread.Emails[0].ID)
			} else {
				// Multi-email thread - expand and go to thread view
				thread.Expanded = true
				a.selectedInThread = 0
				a.viewState = ViewThread
			}
		}
	case key.Matches(msg, a.keys.Expand):
		// Toggle expand/collapse
		if len(a.threads) > 0 && a.selectedThread < len(a.threads) {
			a.threads[a.selectedThread].Expanded = !a.threads[a.selectedThread].Expanded
			if a.threadList != nil {
				a.threadList.UpdateThreads(a.convertToViewThreads())
			}
		}
	case key.Matches(msg, a.keys.Left), key.Matches(msg, a.keys.Back):
		// Go back to folders
		a.viewState = ViewFolders
	case key.Matches(msg, a.keys.Delete):
		if len(a.threads) > 0 && a.selectedThread < len(a.threads) {
			thread := a.threads[a.selectedThread]
			// Delete first email in thread (or all?)
			if len(thread.Emails) > 0 {
				return a, a.deleteEmail(thread.Emails[0].ID)
			}
		}
	case key.Matches(msg, a.keys.MarkUnread):
		if len(a.threads) > 0 && a.selectedThread < len(a.threads) {
			thread := a.threads[a.selectedThread]
			if len(thread.Emails) > 0 {
				return a, a.toggleUnread(thread.Emails[0])
			}
		}
	case key.Matches(msg, a.keys.Archive):
		if len(a.threads) > 0 && a.selectedThread < len(a.threads) {
			thread := a.threads[a.selectedThread]
			if len(thread.Emails) > 0 {
				// If we're at the last thread, move selection up
				if a.selectedThread >= len(a.threads)-1 && a.selectedThread > 0 {
					a.selectedThread--
				}
				return a, a.archiveEmail(thread.Emails[0].ID)
			}
		}
	}
	return a, nil
}

func (a *App) handleThreadKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if a.selectedThread >= len(a.threads) {
		return a, nil
	}
	thread := &a.threads[a.selectedThread]

	switch {
	case key.Matches(msg, a.keys.Up):
		if a.selectedInThread > 0 {
			a.selectedInThread--
		}
	case key.Matches(msg, a.keys.Down):
		if a.selectedInThread < len(thread.Emails)-1 {
			a.selectedInThread++
		}
	case key.Matches(msg, a.keys.Right), key.Matches(msg, a.keys.Enter):
		// Open email
		if a.selectedInThread < len(thread.Emails) {
			a.loading = true
			return a, a.loadEmail(thread.Emails[a.selectedInThread].ID)
		}
	case key.Matches(msg, a.keys.Left), key.Matches(msg, a.keys.Back):
		// Go back to thread list
		thread.Expanded = false
		a.viewState = ViewMessages
	case key.Matches(msg, a.keys.Collapse):
		// Collapse and go back
		thread.Expanded = false
		a.viewState = ViewMessages
	case key.Matches(msg, a.keys.Archive):
		// Archive selected email in thread, go back to messages
		if a.selectedInThread < len(thread.Emails) {
			thread.Expanded = false
			a.viewState = ViewMessages
			// Adjust selection if at end
			if a.selectedThread >= len(a.threads)-1 && a.selectedThread > 0 {
				a.selectedThread--
			}
			return a, a.archiveEmail(thread.Emails[a.selectedInThread].ID)
		}
	case key.Matches(msg, a.keys.Delete):
		// Delete selected email in thread, go back to messages
		if a.selectedInThread < len(thread.Emails) {
			thread.Expanded = false
			a.viewState = ViewMessages
			if a.selectedThread >= len(a.threads)-1 && a.selectedThread > 0 {
				a.selectedThread--
			}
			return a, a.deleteEmail(thread.Emails[a.selectedInThread].ID)
		}
	}
	return a, nil
}

func (a *App) handleEmailKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Up):
		if a.emailReader != nil {
			a.emailReader.ScrollUp()
		}
	case key.Matches(msg, a.keys.Down):
		if a.emailReader != nil {
			a.emailReader.ScrollDown()
		}
	case key.Matches(msg, a.keys.Left), key.Matches(msg, a.keys.Back):
		// Go back
		a.currentEmail = nil
		// Check if thread has multiple emails
		if a.selectedThread < len(a.threads) && len(a.threads[a.selectedThread].Emails) > 1 {
			a.viewState = ViewThread
		} else {
			a.viewState = ViewMessages
		}
	case key.Matches(msg, a.keys.Delete):
		if a.currentEmail != nil {
			emailID := a.currentEmail.ID
			a.currentEmail = nil
			a.viewState = ViewMessages
			// Adjust selection if at end
			if a.selectedThread >= len(a.threads)-1 && a.selectedThread > 0 {
				a.selectedThread--
			}
			return a, a.deleteEmail(emailID)
		}
	case key.Matches(msg, a.keys.Archive):
		if a.currentEmail != nil {
			emailID := a.currentEmail.ID
			a.currentEmail = nil
			a.viewState = ViewMessages
			// Adjust selection if at end
			if a.selectedThread >= len(a.threads)-1 && a.selectedThread > 0 {
				a.selectedThread--
			}
			return a, a.archiveEmail(emailID)
		}
	}
	return a, nil
}

func (a *App) deleteEmail(emailID string) tea.Cmd {
	return func() tea.Msg {
		var trashID string
		for _, mb := range a.mailboxes {
			if mb.Role == "trash" {
				trashID = mb.ID
				break
			}
		}
		if trashID == "" {
			return emailActionMsg{err: fmt.Errorf("trash mailbox not found")}
		}
		err := a.client.DeleteEmail(emailID, trashID)
		return emailActionMsg{err: err}
	}
}

func (a *App) toggleUnread(email models.Email) tea.Cmd {
	return func() tea.Msg {
		var err error
		if email.IsUnread {
			err = a.client.MarkAsRead(email.ID)
		} else {
			err = a.client.MarkAsUnread(email.ID)
		}
		return emailActionMsg{err: err}
	}
}

func (a *App) archiveEmail(emailID string) tea.Cmd {
	return func() tea.Msg {
		var archiveID string
		for _, mb := range a.mailboxes {
			if mb.Role == "archive" {
				archiveID = mb.ID
				break
			}
		}
		if archiveID == "" {
			return emailActionMsg{err: fmt.Errorf("archive mailbox not found")}
		}
		err := a.client.MoveEmail(emailID, "", archiveID)
		return emailActionMsg{err: err}
	}
}

// View renders the application
func (a *App) View() string {
	if a.width == 0 {
		return LoadingStyle.Render("  ▓▒░ INITIALIZING ░▒▓")
	}

	header := a.renderHeader()
	content := a.renderContent()
	statusBar := a.renderStatusBar()
	helpView := a.renderHelp()

	headerHeight := lipgloss.Height(header)
	statusHeight := lipgloss.Height(statusBar)
	helpHeight := lipgloss.Height(helpView)
	contentHeight := a.height - headerHeight - statusHeight - helpHeight

	content = lipgloss.NewStyle().Height(contentHeight).Render(content)

	return lipgloss.JoinVertical(
		lipgloss.Left,
		header,
		content,
		statusBar,
		helpView,
	)
}

func (a *App) renderHeader() string {
	logo := LogoStyle.Render("▀█▀ █ █ █ █▀▄▀█ ▄▀█ █ █   ")
	logo2 := HeaderTitleStyle.Render(" █  █▄█ █ █ ▀ █ █▀█ █ █▄▄ ")

	var titleBlock string
	if a.width > 60 {
		titleBlock = lipgloss.JoinVertical(lipgloss.Left, logo, logo2)
	} else {
		titleBlock = LogoStyle.Render("◈ TUIMAIL")
	}

	accountLabel := StatusDescStyle.Render("▸ ")
	account := HeaderAccountStyle.Render(a.client.Email())
	accountBlock := accountLabel + account

	// Mode indicator based on view state
	var modeIndicator string
	switch a.viewState {
	case ViewFolders:
		modeIndicator = StatusModeStyle.Render(" FOLDERS ")
	case ViewMessages:
		modeIndicator = StatusModeStyle.Render(" MESSAGES ")
	case ViewThread:
		modeIndicator = StatusModeStyle.Render(" THREAD ")
	case ViewEmail:
		modeIndicator = StatusModeStyle.Render(" EMAIL ")
	}

	titleWidth := lipgloss.Width(titleBlock)
	accountWidth := lipgloss.Width(accountBlock)
	modeWidth := lipgloss.Width(modeIndicator)
	gap := a.width - titleWidth - accountWidth - modeWidth - 8
	if gap < 0 {
		gap = 0
	}
	spacer := lipgloss.NewStyle().Width(gap).Render("")

	headerContent := lipgloss.JoinHorizontal(
		lipgloss.Center,
		titleBlock,
		spacer,
		modeIndicator,
		"  ",
		accountBlock,
	)

	return HeaderStyle.Width(a.width).Render(headerContent)
}

func (a *App) renderHelp() string {
	if a.help.ShowAll {
		return HelpStyle.Width(a.width).Render(a.help.View(a.keys))
	}

	// Context-aware help based on view
	var keys []struct{ key, desc string }
	switch a.viewState {
	case ViewFolders:
		keys = []struct{ key, desc string }{
			{"↑/↓", "select"},
			{"→/enter", "open"},
			{"q", "quit"},
			{"?", "help"},
		}
	case ViewMessages:
		keys = []struct{ key, desc string }{
			{"↑/↓", "select"},
			{"→/enter", "open"},
			{"←/esc", "folders"},
			{"space", "expand"},
			{"a", "archive"},
			{"?", "help"},
		}
	case ViewThread:
		keys = []struct{ key, desc string }{
			{"↑/↓", "select"},
			{"→/enter", "read"},
			{"←/esc", "messages"},
			{"a", "archive"},
			{"?", "help"},
		}
	case ViewEmail:
		keys = []struct{ key, desc string }{
			{"↑/↓", "scroll"},
			{"←/esc", "back"},
			{"a", "archive"},
			{"d", "delete"},
			{"?", "help"},
		}
	}

	var parts []string
	for _, k := range keys {
		parts = append(parts,
			HelpKeyStyle.Render(k.key)+
				HelpSepStyle.Render(":")+
				HelpDescStyle.Render(k.desc))
	}

	helpText := ""
	for i, part := range parts {
		if i > 0 {
			helpText += HelpSepStyle.Render(" │ ")
		}
		helpText += part
	}

	return lipgloss.NewStyle().
		Foreground(ColorTextMuted).
		Background(ColorBgMid).
		Padding(0, 2).
		Width(a.width).
		Render(helpText)
}

func (a *App) renderContent() string {
	if a.err != nil {
		errBox := lipgloss.JoinVertical(lipgloss.Center,
			ErrorStyle.Render("▓▒░ ERROR ░▒▓"),
			"",
			lipgloss.NewStyle().Foreground(ColorError).Render(fmt.Sprintf("%v", a.err)),
		)
		return lipgloss.Place(a.width, 10, lipgloss.Center, lipgloss.Center, errBox)
	}

	if a.loading {
		loadingBox := lipgloss.JoinVertical(lipgloss.Center,
			SpinnerStyle.Render(a.spinner.View()),
			"",
			LoadingStyle.Render("▓▒░ CONNECTING ░▒▓"),
		)
		return lipgloss.Place(a.width, 10, lipgloss.Center, lipgloss.Center, loadingBox)
	}

	// Sidebar
	sidebarWidth := 24
	sidebar := a.renderSidebar(sidebarWidth)

	// Main content
	mainWidth := a.width - sidebarWidth - 1
	var main string
	switch a.viewState {
	case ViewFolders:
		main = a.renderEmptyMain(mainWidth, "← Select a folder")
	case ViewMessages:
		main = a.renderMessageList(mainWidth)
	case ViewThread:
		main = a.renderThreadContents(mainWidth)
	case ViewEmail:
		main = a.renderEmailReader(mainWidth)
	}

	return lipgloss.JoinHorizontal(lipgloss.Top, sidebar, main)
}

func (a *App) renderSidebar(width int) string {
	var style lipgloss.Style
	if a.viewState == ViewFolders {
		style = SidebarActiveStyle.Width(width)
	} else {
		style = SidebarStyle.Width(width)
	}

	if a.mailboxView == nil {
		return style.Render(StatusDescStyle.Render("No mailboxes"))
	}

	return style.Render(a.mailboxView.View())
}

func (a *App) renderEmptyMain(width int, msg string) string {
	return lipgloss.Place(
		width, a.height-8,
		lipgloss.Center, lipgloss.Center,
		StatusDescStyle.Render(msg),
	)
}

func (a *App) renderMessageList(width int) string {
	if a.threadList == nil {
		return a.renderEmptyMain(width, "No messages")
	}
	a.threadList.SetSize(width, a.height-6)
	a.threadList.UpdateThreads(a.convertToViewThreads())
	return a.threadList.View()
}

func (a *App) renderThreadContents(width int) string {
	if a.selectedThread >= len(a.threads) {
		return a.renderEmptyMain(width, "No thread selected")
	}
	thread := a.threads[a.selectedThread]

	var b strings.Builder

	// Thread header
	headerStyle := lipgloss.NewStyle().
		Foreground(ColorNeonCyan).
		Bold(true).
		MarginBottom(1)

	b.WriteString(headerStyle.Render("◈ " + thread.Subject))
	b.WriteString("\n")

	countStyle := lipgloss.NewStyle().
		Foreground(ColorTextMuted)
	b.WriteString(countStyle.Render(fmt.Sprintf("  %d messages in thread", len(thread.Emails))))
	b.WriteString("\n\n")

	// Render emails in thread with indentation
	for i, email := range thread.Emails {
		isSelected := i == a.selectedInThread

		indent := "    "
		if i > 0 {
			indent = "        " // Deeper indent for replies
		}

		// Selection indicator
		if isSelected {
			b.WriteString(lipgloss.NewStyle().
				Foreground(ColorNeonCyan).
				Render(indent[:len(indent)-2] + "▶ "))
		} else {
			b.WriteString(indent)
		}

		// Email info
		fromStyle := lipgloss.NewStyle().Foreground(ColorTextNormal)
		if email.IsUnread {
			fromStyle = lipgloss.NewStyle().Foreground(ColorNeonCyan).Bold(true)
		}

		dateStyle := lipgloss.NewStyle().Foreground(ColorTextMuted)

		line := fromStyle.Render(email.FromDisplay()) +
			"  " +
			dateStyle.Render(email.DateDisplay())

		if isSelected {
			line = lipgloss.NewStyle().
				Background(ColorBgSelect).
				Render(line)
		}

		b.WriteString(line)
		b.WriteString("\n")

		// Preview for selected
		if isSelected {
			preview := email.Preview
			if len(preview) > 60 {
				preview = preview[:57] + "..."
			}
			previewStyle := lipgloss.NewStyle().
				Foreground(ColorTextMuted).
				PaddingLeft(len(indent))
			b.WriteString(previewStyle.Render(preview))
			b.WriteString("\n")
		}
	}

	return b.String()
}

func (a *App) renderEmailReader(width int) string {
	if a.emailReader == nil {
		return a.renderEmptyMain(width, "No email selected")
	}
	a.emailReader.SetSize(width, a.height-6)
	return a.emailReader.View()
}

func (a *App) renderStatusBar() string {
	var leftPart, rightPart string

	if len(a.mailboxes) > 0 && a.selectedMailbox < len(a.mailboxes) {
		mb := a.mailboxes[a.selectedMailbox]

		mailboxName := StatusKeyStyle.Render(mb.DisplayName())
		threadCount := StatusDescStyle.Render(fmt.Sprintf(" ◇ %d threads", len(a.threads)))
		leftPart = mailboxName + threadCount

		if mb.UnreadCount > 0 {
			unread := lipgloss.NewStyle().
				Foreground(ColorNeonPink).
				Bold(true).
				Render(fmt.Sprintf(" ● %d unread", mb.UnreadCount))
			leftPart += unread
		}
	}

	// Breadcrumb navigation indicator
	breadcrumb := ""
	switch a.viewState {
	case ViewFolders:
		breadcrumb = StatusKeyStyle.Render("FOLDERS")
	case ViewMessages:
		breadcrumb = StatusDescStyle.Render("folders ") +
			StatusKeyStyle.Render("→ MESSAGES")
	case ViewThread:
		breadcrumb = StatusDescStyle.Render("folders → messages ") +
			StatusKeyStyle.Render("→ THREAD")
	case ViewEmail:
		// Check if we came from a thread or directly from messages
		if a.selectedThread < len(a.threads) && len(a.threads[a.selectedThread].Emails) > 1 {
			breadcrumb = StatusDescStyle.Render("... → thread ") +
				StatusKeyStyle.Render("→ EMAIL")
		} else {
			breadcrumb = StatusDescStyle.Render("... → messages ") +
				StatusKeyStyle.Render("→ EMAIL")
		}
	}
	rightPart = breadcrumb

	gap := a.width - lipgloss.Width(leftPart) - lipgloss.Width(rightPart) - 6
	if gap < 0 {
		gap = 0
	}
	spacer := lipgloss.NewStyle().Width(gap).Render("")

	content := lipgloss.JoinHorizontal(lipgloss.Center, leftPart, spacer, rightPart)
	return StatusBarStyle.Width(a.width).Render(content)
}
