package ui

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/the9x/anneal/internal/config"
	"github.com/the9x/anneal/internal/jmap"
	"github.com/the9x/anneal/internal/models"
	"github.com/the9x/anneal/internal/storage"
	"github.com/the9x/anneal/internal/ui/views"
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
	ViewCompose                   // Composing/replying to email
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
	store     *storage.Store
	syncer    *storage.Syncer
	keys      KeyMap
	help      help.Model
	spinner   spinner.Model
	width     int
	height    int
	viewState ViewState
	loading   bool
	syncing   bool // Background sync in progress
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
	composeView *views.ComposeView

	// State for compose
	prevViewState ViewState // Where to return after compose
}

// NewApp creates a new application instance
func NewApp(cfg *config.Config, client *jmap.Client, store *storage.Store) *App {
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = SpinnerStyle

	var syncer *storage.Syncer
	if store != nil {
		syncer = storage.NewSyncer(store, client)
	}

	return &App{
		cfg:       cfg,
		client:    client,
		store:     store,
		syncer:    syncer,
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
		a.loadMailboxesCacheFirst,
	)
}

// Msg types for async operations
type mailboxesLoadedMsg struct {
	mailboxes  []models.Mailbox
	fromCache  bool
	err        error
}

type emailsLoadedMsg struct {
	emails    []models.Email
	fromCache bool
	err       error
}

type emailLoadedMsg struct {
	email     *models.Email
	fromCache bool
	err       error
}

type syncCompleteMsg struct {
	mailboxResult *storage.SyncResult
	emailResult   *storage.SyncResult
	err           error
}

type emailActionMsg struct {
	err error
}

type emailSentMsg struct {
	err error
}

type attachmentOpenedMsg struct {
	err error
}

// loadMailboxesCacheFirst tries cache first, then falls back to network
func (a *App) loadMailboxesCacheFirst() tea.Msg {
	// Try cache first if syncer is available
	if a.syncer != nil {
		mailboxes, err := a.syncer.GetCachedMailboxes()
		if err == nil && len(mailboxes) > 0 {
			return mailboxesLoadedMsg{mailboxes: mailboxes, fromCache: true, err: nil}
		}
	}

	// Fall back to network
	mailboxes, err := a.client.GetMailboxes()
	return mailboxesLoadedMsg{mailboxes: mailboxes, fromCache: false, err: err}
}

func (a *App) loadMailboxes() tea.Msg {
	mailboxes, err := a.client.GetMailboxes()
	return mailboxesLoadedMsg{mailboxes: mailboxes, fromCache: false, err: err}
}

func (a *App) loadEmails(mailboxID string) tea.Cmd {
	return func() tea.Msg {
		// Try cache first
		if a.syncer != nil {
			emails, err := a.syncer.GetCachedEmails(mailboxID, a.cfg.PageSize)
			if err == nil && len(emails) > 0 {
				return emailsLoadedMsg{emails: emails, fromCache: true, err: nil}
			}
		}

		// Fall back to network
		emails, err := a.client.GetEmails(mailboxID, a.cfg.PageSize)

		// Cache the results
		if err == nil && a.syncer != nil && len(emails) > 0 {
			a.store.SaveEmails(a.client.AccountID(), emails)
		}

		return emailsLoadedMsg{emails: emails, fromCache: false, err: err}
	}
}

// loadEmailsFresh always fetches from network, skipping cache
func (a *App) loadEmailsFresh(mailboxID string) tea.Cmd {
	return func() tea.Msg {
		emails, err := a.client.GetEmails(mailboxID, a.cfg.PageSize)

		// Update the cache with fresh data
		if err == nil && a.store != nil && len(emails) > 0 {
			a.store.SaveEmails(a.client.AccountID(), emails)
		}

		return emailsLoadedMsg{emails: emails, fromCache: false, err: err}
	}
}

func (a *App) loadEmail(emailID string) tea.Cmd {
	return func() tea.Msg {
		// Try cache first (for full body)
		if a.syncer != nil {
			email, err := a.syncer.GetCachedEmailBody(emailID)
			if err == nil && email != nil && (email.TextBody != "" || email.HTMLBody != "") {
				return emailLoadedMsg{email: email, fromCache: true, err: nil}
			}
		}

		// Fall back to network
		email, err := a.client.GetEmail(emailID)

		// Cache the body
		if err == nil && email != nil && a.store != nil {
			a.store.SaveEmailBody(email)
		}

		return emailLoadedMsg{email: email, fromCache: false, err: err}
	}
}

// syncInBackground triggers a background sync
func (a *App) syncInBackground(mailboxID string) tea.Cmd {
	return func() tea.Msg {
		if a.syncer == nil {
			return syncCompleteMsg{err: nil}
		}

		mailboxResult, err := a.syncer.SyncMailboxes()
		if err != nil {
			return syncCompleteMsg{err: err}
		}

		var emailResult *storage.SyncResult
		if mailboxID != "" {
			emailResult, err = a.syncer.SyncEmails(mailboxID, 100)
		}

		return syncCompleteMsg{
			mailboxResult: mailboxResult,
			emailResult:   emailResult,
			err:           err,
		}
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

		// Cache mailboxes if from network
		if !msg.fromCache && a.store != nil {
			a.store.SaveMailboxes(a.client.AccountID(), a.mailboxes)
		}

		// Find inbox and load emails
		var inboxID string
		for i, mb := range a.mailboxes {
			if mb.Role == "inbox" {
				a.selectedMailbox = i
				a.mailboxView.Select(i)
				inboxID = mb.ID
				break
			}
		}
		if inboxID == "" && len(a.mailboxes) > 0 {
			inboxID = a.mailboxes[0].ID
		}

		if inboxID != "" {
			a.loading = true
			cmds := []tea.Cmd{a.loadEmails(inboxID)}

			// Trigger background sync if loaded from cache
			if msg.fromCache {
				a.syncing = true
				cmds = append(cmds, a.syncInBackground(inboxID))
			}

			return a, tea.Batch(cmds...)
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
		// Force refresh from network after action (skip cache)
		if len(a.mailboxes) > 0 && a.selectedMailbox < len(a.mailboxes) {
			return a, a.loadEmailsFresh(a.mailboxes[a.selectedMailbox].ID)
		}
		return a, nil

	case emailSentMsg:
		if msg.err != nil {
			a.err = msg.err
		}
		// Refresh to show sent email in sent folder if viewing it
		if len(a.mailboxes) > 0 && a.selectedMailbox < len(a.mailboxes) {
			return a, a.loadEmails(a.mailboxes[a.selectedMailbox].ID)
		}
		return a, nil

	case attachmentOpenedMsg:
		if msg.err != nil {
			a.err = msg.err
		}
		// Exit attachment mode after opening
		if a.emailReader != nil && a.emailReader.InAttachmentMode() {
			a.emailReader.ToggleAttachmentMode()
		}
		return a, nil

	case syncCompleteMsg:
		a.syncing = false
		if msg.err != nil {
			// Sync errors are non-fatal, just log them
			return a, nil
		}

		// If there were changes, refresh the data
		hasChanges := false
		if msg.mailboxResult != nil {
			hasChanges = msg.mailboxResult.MailboxesCreated > 0 ||
				msg.mailboxResult.MailboxesUpdated > 0 ||
				msg.mailboxResult.MailboxesDestroyed > 0
		}
		if msg.emailResult != nil {
			hasChanges = hasChanges ||
				msg.emailResult.EmailsCreated > 0 ||
				msg.emailResult.EmailsUpdated > 0 ||
				msg.emailResult.EmailsDestroyed > 0
		}

		if hasChanges {
			// Reload from cache (which now has synced data)
			var cmds []tea.Cmd

			// Reload mailboxes if they changed
			if msg.mailboxResult != nil &&
				(msg.mailboxResult.MailboxesCreated > 0 ||
					msg.mailboxResult.MailboxesUpdated > 0 ||
					msg.mailboxResult.MailboxesDestroyed > 0) {
				cmds = append(cmds, func() tea.Msg {
					mailboxes, err := a.syncer.GetCachedMailboxes()
					return mailboxesLoadedMsg{mailboxes: mailboxes, fromCache: true, err: err}
				})
			}

			// Reload emails if they changed
			if msg.emailResult != nil &&
				(msg.emailResult.EmailsCreated > 0 ||
					msg.emailResult.EmailsUpdated > 0 ||
					msg.emailResult.EmailsDestroyed > 0) {
				if len(a.mailboxes) > 0 && a.selectedMailbox < len(a.mailboxes) {
					mailboxID := a.mailboxes[a.selectedMailbox].ID
					cmds = append(cmds, func() tea.Msg {
						emails, err := a.syncer.GetCachedEmails(mailboxID, a.cfg.PageSize)
						return emailsLoadedMsg{emails: emails, fromCache: true, err: err}
					})
				}
			}

			if len(cmds) > 0 {
				return a, tea.Batch(cmds...)
			}
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
	case ViewCompose:
		return a.handleComposeKeys(msg)
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
				// In trash, "u" undeletes (moves to inbox)
				if a.isInTrash() {
					if a.selectedThread >= len(a.threads)-1 && a.selectedThread > 0 {
						a.selectedThread--
					}
					return a, a.undeleteThread(thread.Emails)
				}
				// Otherwise toggle read/unread
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
				// Archive all emails in the thread
				emailIDs := make([]string, len(thread.Emails))
				for i, e := range thread.Emails {
					emailIDs[i] = e.ID
				}
				return a, a.archiveThread(emailIDs)
			}
		}
	case key.Matches(msg, a.keys.Compose):
		return a.startCompose(nil, views.ModeCompose)
	case key.Matches(msg, a.keys.Reply):
		if len(a.threads) > 0 && a.selectedThread < len(a.threads) {
			thread := a.threads[a.selectedThread]
			if len(thread.Emails) > 0 {
				return a.startCompose(&thread.Emails[0], views.ModeReply)
			}
		}
	case key.Matches(msg, a.keys.ReplyAll):
		if len(a.threads) > 0 && a.selectedThread < len(a.threads) {
			thread := a.threads[a.selectedThread]
			if len(thread.Emails) > 0 {
				return a.startCompose(&thread.Emails[0], views.ModeReplyAll)
			}
		}
	case key.Matches(msg, a.keys.Forward):
		if len(a.threads) > 0 && a.selectedThread < len(a.threads) {
			thread := a.threads[a.selectedThread]
			if len(thread.Emails) > 0 {
				return a.startCompose(&thread.Emails[0], views.ModeForward)
			}
		}
	case key.Matches(msg, a.keys.Refresh):
		// Force refresh from network
		if len(a.mailboxes) > 0 && a.selectedMailbox < len(a.mailboxes) {
			a.loading = true
			return a, a.loadEmailsFresh(a.mailboxes[a.selectedMailbox].ID)
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
		// Archive entire thread, go back to messages
		if len(thread.Emails) > 0 {
			thread.Expanded = false
			a.viewState = ViewMessages
			// Adjust selection if at end
			if a.selectedThread >= len(a.threads)-1 && a.selectedThread > 0 {
				a.selectedThread--
			}
			// Archive all emails in the thread
			emailIDs := make([]string, len(thread.Emails))
			for i, e := range thread.Emails {
				emailIDs[i] = e.ID
			}
			return a, a.archiveThread(emailIDs)
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
	// Handle attachment mode separately
	if a.emailReader != nil && a.emailReader.InAttachmentMode() {
		return a.handleAttachmentKeys(msg)
	}

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
		if a.selectedThread < len(a.threads) {
			thread := a.threads[a.selectedThread]
			a.currentEmail = nil
			a.viewState = ViewMessages
			// Adjust selection if at end
			if a.selectedThread >= len(a.threads)-1 && a.selectedThread > 0 {
				a.selectedThread--
			}
			// Archive all emails in the thread
			emailIDs := make([]string, len(thread.Emails))
			for i, e := range thread.Emails {
				emailIDs[i] = e.ID
			}
			return a, a.archiveThread(emailIDs)
		}
	case key.Matches(msg, a.keys.Compose):
		return a.startCompose(nil, views.ModeCompose)
	case key.Matches(msg, a.keys.Reply):
		if a.currentEmail != nil {
			return a.startCompose(a.currentEmail, views.ModeReply)
		}
	case key.Matches(msg, a.keys.ReplyAll):
		if a.currentEmail != nil {
			return a.startCompose(a.currentEmail, views.ModeReplyAll)
		}
	case key.Matches(msg, a.keys.Forward):
		if a.currentEmail != nil {
			return a.startCompose(a.currentEmail, views.ModeForward)
		}
	case key.Matches(msg, a.keys.Right), key.Matches(msg, a.keys.Enter):
		// Navigate forward to attachments if email has any
		if a.emailReader != nil && a.emailReader.HasAttachments() {
			a.emailReader.ToggleAttachmentMode()
		}
	}
	return a, nil
}

func (a *App) handleAttachmentKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, a.keys.Up):
		a.emailReader.PrevAttachment()
	case key.Matches(msg, a.keys.Down):
		a.emailReader.NextAttachment()
	case key.Matches(msg, a.keys.Left), key.Matches(msg, a.keys.Back):
		// Go back to email
		a.emailReader.ToggleAttachmentMode()
	case key.Matches(msg, a.keys.Right), key.Matches(msg, a.keys.Enter):
		// Open selected attachment
		att := a.emailReader.SelectedAttachment()
		if att != nil {
			return a, a.openAttachment(att)
		}
	}
	return a, nil
}

// startCompose initializes the compose view
func (a *App) startCompose(email *models.Email, mode views.ComposeMode) (tea.Model, tea.Cmd) {
	a.composeView = views.NewComposeView(a.width-26, a.height-8)

	switch mode {
	case views.ModeReply:
		a.composeView.SetReply(email, false)
		a.composeView.RemoveSelfFromCC(a.client.Email())
	case views.ModeReplyAll:
		a.composeView.SetReply(email, true)
		a.composeView.RemoveSelfFromCC(a.client.Email())
	case views.ModeForward:
		a.composeView.SetForward(email)
	}

	a.prevViewState = a.viewState
	a.viewState = ViewCompose

	return a, nil
}

// handleComposeKeys handles input in compose view
func (a *App) handleComposeKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "esc":
		// Cancel compose
		a.viewState = a.prevViewState
		a.composeView = nil
		return a, nil
	case "ctrl+s":
		// Send email
		if a.composeView == nil {
			return a, nil
		}

		// Validate
		if !a.composeView.HasRecipients() {
			return a, nil
		}
		if a.composeView.IsEmpty() {
			return a, nil
		}

		to, cc, subject, body := a.composeView.GetValues()
		original := a.composeView.Original

		// Return to previous view
		a.viewState = a.prevViewState
		a.composeView = nil

		return a, a.sendEmail(to, cc, subject, body, original)
	}

	// Pass to compose view
	var cmd tea.Cmd
	a.composeView, cmd = a.composeView.Update(msg)
	return a, cmd
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

func (a *App) archiveThread(emailIDs []string) tea.Cmd {
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
		// Archive all emails in the thread
		for _, emailID := range emailIDs {
			if err := a.client.MoveEmail(emailID, "", archiveID); err != nil {
				return emailActionMsg{err: err}
			}
		}
		return emailActionMsg{err: nil}
	}
}

// isInTrash returns true if currently viewing the trash folder
func (a *App) isInTrash() bool {
	if a.selectedMailbox < len(a.mailboxes) {
		return a.mailboxes[a.selectedMailbox].Role == "trash"
	}
	return false
}

// undeleteThread moves emails from trash back to inbox
func (a *App) undeleteThread(emails []models.Email) tea.Cmd {
	return func() tea.Msg {
		var inboxID string
		for _, mb := range a.mailboxes {
			if mb.Role == "inbox" {
				inboxID = mb.ID
				break
			}
		}
		if inboxID == "" {
			return emailActionMsg{err: fmt.Errorf("inbox not found")}
		}
		// Move all emails in the thread to inbox
		for _, email := range emails {
			if err := a.client.MoveEmail(email.ID, "", inboxID); err != nil {
				return emailActionMsg{err: err}
			}
		}
		return emailActionMsg{err: nil}
	}
}

func (a *App) openAttachment(att *models.Attachment) tea.Cmd {
	return func() tea.Msg {
		// Create cache directory
		cacheDir := filepath.Join(os.TempDir(), "anneal", "attachments")
		if err := os.MkdirAll(cacheDir, 0755); err != nil {
			return attachmentOpenedMsg{err: fmt.Errorf("failed to create cache dir: %w", err)}
		}

		// Download blob
		data, err := a.client.DownloadBlob(att.BlobID, att.Name)
		if err != nil {
			return attachmentOpenedMsg{err: err}
		}

		// Save to temp file
		filePath := filepath.Join(cacheDir, fmt.Sprintf("%s-%s", att.BlobID, att.Name))
		if err := os.WriteFile(filePath, data, 0644); err != nil {
			return attachmentOpenedMsg{err: fmt.Errorf("failed to save file: %w", err)}
		}

		// Open with system default (non-blocking)
		cmd := exec.Command("open", filePath)
		if err := cmd.Start(); err != nil {
			return attachmentOpenedMsg{err: fmt.Errorf("failed to open file: %w", err)}
		}

		return attachmentOpenedMsg{err: nil}
	}
}

func (a *App) sendEmail(to, cc []string, subject, body string, original *models.Email) tea.Cmd {
	return func() tea.Msg {
		var inReplyTo, references []string

		// Set reply headers if this is a reply
		if original != nil {
			inReplyTo = []string{original.ID}
			// Could add references chain here if needed
		}

		err := a.client.SendEmail(to, cc, subject, body, inReplyTo, references)
		return emailSentMsg{err: err}
	}
}

// View renders the application
func (a *App) View() string {
	if a.width == 0 {
		return LoadingStyle.Render("  ◇ initializing...")
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
	titleBlock := LogoStyle.Render("◈ anneal")

	accountLabel := StatusDescStyle.Render("▸ ")
	account := HeaderAccountStyle.Render(a.client.Email())
	accountBlock := accountLabel + account

	// Mode indicator based on view state
	var modeIndicator string
	switch a.viewState {
	case ViewFolders:
		modeIndicator = StatusModeStyle.Render(" folders ")
	case ViewMessages:
		modeIndicator = StatusModeStyle.Render(" messages ")
	case ViewThread:
		modeIndicator = StatusModeStyle.Render(" thread ")
	case ViewEmail:
		modeIndicator = StatusModeStyle.Render(" email ")
	case ViewCompose:
		modeIndicator = StatusModeStyle.Render(" compose ")
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
		}
		if a.isInTrash() {
			keys = append(keys, struct{ key, desc string }{"u", "undelete"})
		} else {
			keys = append(keys,
				struct{ key, desc string }{"c", "compose"},
				struct{ key, desc string }{"r", "reply"},
				struct{ key, desc string }{"a", "archive"},
			)
		}
		keys = append(keys, struct{ key, desc string }{"?", "help"})
	case ViewThread:
		keys = []struct{ key, desc string }{
			{"↑/↓", "select"},
			{"→/enter", "read"},
			{"←/esc", "messages"},
			{"a", "archive"},
			{"?", "help"},
		}
	case ViewEmail:
		if a.emailReader != nil && a.emailReader.InAttachmentMode() {
			keys = []struct{ key, desc string }{
				{"↑/↓", "select"},
				{"→/enter", "open"},
				{"←/esc", "email"},
			}
		} else {
			keys = []struct{ key, desc string }{
				{"↑/↓", "scroll"},
				{"←/esc", "back"},
				{"r", "reply"},
				{"R", "reply all"},
				{"f", "forward"},
				{"a", "archive"},
			}
			// Show attachments hint if email has attachments
			if a.emailReader != nil && a.emailReader.HasAttachments() {
				keys = append(keys, struct{ key, desc string }{"→", "attachments"})
			}
			keys = append(keys, struct{ key, desc string }{"?", "help"})
		}
	case ViewCompose:
		keys = []struct{ key, desc string }{
			{"tab", "next field"},
			{"ctrl+s", "send"},
			{"esc", "cancel"},
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
		Foreground(ColorDim).
		Background(ColorBg).
		Padding(0, 2).
		Width(a.width).
		Render(helpText)
}

func (a *App) renderContent() string {
	if a.err != nil {
		errBox := lipgloss.JoinVertical(lipgloss.Center,
			ErrorStyle.Render("◇ something went wrong"),
			"",
			lipgloss.NewStyle().Foreground(ColorSecondary).Render(fmt.Sprintf("%v", a.err)),
		)
		return lipgloss.Place(a.width, 10, lipgloss.Center, lipgloss.Center, errBox)
	}

	if a.loading {
		loadingBox := lipgloss.JoinVertical(lipgloss.Center,
			SpinnerStyle.Render(a.spinner.View()),
			"",
			LoadingStyle.Render("connecting..."),
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
	case ViewCompose:
		main = a.renderComposeView(mainWidth)
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
		Foreground(ColorPrimary).
		MarginBottom(1)

	b.WriteString(headerStyle.Render("◈ " + thread.Subject))
	b.WriteString("\n")

	countStyle := lipgloss.NewStyle().
		Foreground(ColorDim)
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
				Foreground(ColorPrimary).
				Render(indent[:len(indent)-2] + "▶ "))
		} else {
			b.WriteString(indent)
		}

		// Email info
		fromStyle := lipgloss.NewStyle().Foreground(ColorSecondary)
		if email.IsUnread {
			fromStyle = lipgloss.NewStyle().Foreground(ColorPrimary)
		}

		dateStyle := lipgloss.NewStyle().Foreground(ColorDim)

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
				Foreground(ColorDim).
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

func (a *App) renderComposeView(width int) string {
	if a.composeView == nil {
		return a.renderEmptyMain(width, "No compose view")
	}
	a.composeView.SetSize(width, a.height-8)
	return a.composeView.View()
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
				Foreground(ColorPrimary).
				Render(fmt.Sprintf(" ● %d unread", mb.UnreadCount))
			leftPart += unread
		}
	}

	// Breadcrumb navigation indicator
	breadcrumb := ""
	switch a.viewState {
	case ViewFolders:
		breadcrumb = StatusKeyStyle.Render("folders")
	case ViewMessages:
		breadcrumb = StatusDescStyle.Render("folders ") +
			StatusKeyStyle.Render("→ messages")
	case ViewThread:
		breadcrumb = StatusDescStyle.Render("folders → messages ") +
			StatusKeyStyle.Render("→ thread")
	case ViewEmail:
		// Check if in attachment mode
		if a.emailReader != nil && a.emailReader.InAttachmentMode() {
			breadcrumb = StatusDescStyle.Render("... → email ") +
				StatusKeyStyle.Render("→ attachments")
		} else if a.selectedThread < len(a.threads) && len(a.threads[a.selectedThread].Emails) > 1 {
			breadcrumb = StatusDescStyle.Render("... → thread ") +
				StatusKeyStyle.Render("→ email")
		} else {
			breadcrumb = StatusDescStyle.Render("... → messages ") +
				StatusKeyStyle.Render("→ email")
		}
	case ViewCompose:
		breadcrumb = StatusDescStyle.Render("... ") +
			StatusKeyStyle.Render("→ compose")
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
