package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/the9x/anneal/internal/models"
)

// ComposeMode indicates the type of composition
type ComposeMode int

const (
	ModeCompose ComposeMode = iota
	ModeReply
	ModeReplyAll
	ModeForward
)

// anneal brand colors for compose
var (
	composeColorPrimary   = lipgloss.Color("#d4d2e3")
	composeColorSecondary = lipgloss.Color("#9795b5")
	composeColorDim       = lipgloss.Color("#5a5880")
	composeColorBg        = lipgloss.Color("#1d1d40")
	composeColorBgSelect  = lipgloss.Color("#2d2d5a")

	composeLabelStyle = lipgloss.NewStyle().
				Foreground(composeColorDim).
				Width(10).
				Align(lipgloss.Right)

	composeInputStyle = lipgloss.NewStyle().
				Foreground(composeColorPrimary)

	composeHeaderStyle = lipgloss.NewStyle().
				Foreground(composeColorSecondary).
				MarginBottom(1)

	composeFocusedStyle = lipgloss.NewStyle().
				Foreground(composeColorPrimary).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(composeColorSecondary)

	composeBlurredStyle = lipgloss.NewStyle().
				Foreground(composeColorSecondary).
				BorderStyle(lipgloss.NormalBorder()).
				BorderForeground(composeColorDim)

	composeHelpStyle = lipgloss.NewStyle().
				Foreground(composeColorDim).
				MarginTop(1)
)

// Identity represents a sending identity
type Identity struct {
	ID    string
	Name  string
	Email string
}

// ComposeField indicates which field is focused
type ComposeField int

const (
	FieldFrom ComposeField = iota
	FieldTo
	FieldCc
	FieldSubject
	FieldBody
)

// ComposeView is the inline email composition view
type ComposeView struct {
	Mode     ComposeMode
	Original *models.Email

	identities       []Identity
	selectedIdentity int

	to      textinput.Model
	cc      textinput.Model
	subject textinput.Model
	body    textarea.Model

	focused ComposeField
	width   int
	height  int
}

// NewComposeView creates a new compose view
func NewComposeView(width, height int, identities []Identity) *ComposeView {
	// To field
	to := textinput.New()
	to.Placeholder = "recipient@example.com"
	to.Focus()
	to.CharLimit = 500
	to.Width = width - 14
	to.PromptStyle = lipgloss.NewStyle().Foreground(composeColorDim)
	to.TextStyle = lipgloss.NewStyle().Foreground(composeColorPrimary)

	// Cc field
	cc := textinput.New()
	cc.Placeholder = ""
	cc.CharLimit = 500
	cc.Width = width - 14
	cc.PromptStyle = lipgloss.NewStyle().Foreground(composeColorDim)
	cc.TextStyle = lipgloss.NewStyle().Foreground(composeColorPrimary)

	// Subject field
	subject := textinput.New()
	subject.Placeholder = ""
	subject.CharLimit = 200
	subject.Width = width - 14
	subject.PromptStyle = lipgloss.NewStyle().Foreground(composeColorDim)
	subject.TextStyle = lipgloss.NewStyle().Foreground(composeColorPrimary)

	// Body textarea
	body := textarea.New()
	body.Placeholder = "compose your message..."
	body.CharLimit = 0 // No limit
	body.SetWidth(width - 4)
	body.SetHeight(height - 12)
	body.FocusedStyle.Base = lipgloss.NewStyle().Foreground(composeColorPrimary)
	body.BlurredStyle.Base = lipgloss.NewStyle().Foreground(composeColorSecondary)
	body.FocusedStyle.CursorLine = lipgloss.NewStyle().Background(composeColorBgSelect)
	body.ShowLineNumbers = false

	// Start on From field if multiple identities, otherwise To field
	startField := FieldTo
	if len(identities) > 1 {
		startField = FieldFrom
	}

	return &ComposeView{
		Mode:             ModeCompose,
		identities:       identities,
		selectedIdentity: 0,
		to:               to,
		cc:               cc,
		subject:          subject,
		body:             body,
		focused:          startField,
		width:            width,
		height:           height,
	}
}

// SetSize updates the view dimensions
func (v *ComposeView) SetSize(width, height int) {
	v.width = width
	v.height = height
	v.to.Width = width - 14
	v.cc.Width = width - 14
	v.subject.Width = width - 14
	v.body.SetWidth(width - 4)
	v.body.SetHeight(height - 12)
}

// SetReply configures the view for replying
func (v *ComposeView) SetReply(email *models.Email, replyAll bool) {
	v.Original = email

	if replyAll {
		v.Mode = ModeReplyAll
	} else {
		v.Mode = ModeReply
	}

	// Set To address
	replyTo := ""
	if len(email.ReplyTo) > 0 {
		replyTo = email.ReplyTo[0].Email
	} else if len(email.From) > 0 {
		replyTo = email.From[0].Email
	}
	v.to.SetValue(replyTo)

	// Set CC for reply-all (excluding self - caller should pass myEmail)
	if replyAll {
		var ccAddrs []string
		for _, addr := range email.To {
			if addr.Email != replyTo {
				ccAddrs = append(ccAddrs, addr.Email)
			}
		}
		for _, addr := range email.CC {
			if addr.Email != replyTo {
				ccAddrs = append(ccAddrs, addr.Email)
			}
		}
		v.cc.SetValue(strings.Join(ccAddrs, ", "))
	}

	// Set subject with Re: prefix
	subject := email.Subject
	if !strings.HasPrefix(strings.ToLower(subject), "re:") {
		subject = "Re: " + subject
	}
	v.subject.SetValue(subject)

	// Quote original message
	v.body.SetValue(v.quoteText(email.TextBody, email.From))

	// Focus body for typing
	v.focusField(FieldBody)
}

// SetForward configures the view for forwarding
func (v *ComposeView) SetForward(email *models.Email) {
	v.Original = email
	v.Mode = ModeForward

	// Set subject with Fwd: prefix
	subject := email.Subject
	if !strings.HasPrefix(strings.ToLower(subject), "fwd:") {
		subject = "Fwd: " + subject
	}
	v.subject.SetValue(subject)

	// Build forwarded message
	fromStr := ""
	if len(email.From) > 0 {
		fromStr = email.From[0].String()
	}
	toStr := ""
	for i, addr := range email.To {
		if i > 0 {
			toStr += ", "
		}
		toStr += addr.String()
	}

	forwarded := fmt.Sprintf("\n---------- Forwarded message ----------\nFrom: %s\nDate: %s\nSubject: %s\nTo: %s\n\n%s",
		fromStr,
		email.ReceivedAt.Format("Mon, Jan 2, 2006 at 3:04 PM"),
		email.Subject,
		toStr,
		email.TextBody)

	v.body.SetValue(forwarded)

	// Focus To field since it's empty
	v.focusField(FieldTo)
}

func (v *ComposeView) quoteText(body string, from []models.EmailAddress) string {
	if body == "" {
		return ""
	}

	fromStr := "someone"
	if len(from) > 0 {
		if from[0].Name != "" {
			fromStr = from[0].Name
		} else {
			fromStr = from[0].Email
		}
	}

	var quoted strings.Builder
	quoted.WriteString(fmt.Sprintf("\n\nOn %s wrote:\n", fromStr))

	lines := strings.Split(body, "\n")
	for _, line := range lines {
		quoted.WriteString("> ")
		quoted.WriteString(line)
		quoted.WriteString("\n")
	}

	return quoted.String()
}

// RemoveSelfFromCC removes the user's own email from CC
func (v *ComposeView) RemoveSelfFromCC(myEmail string) {
	ccVal := v.cc.Value()
	if ccVal == "" {
		return
	}

	var filtered []string
	for _, addr := range strings.Split(ccVal, ",") {
		addr = strings.TrimSpace(addr)
		if addr != "" && addr != myEmail {
			filtered = append(filtered, addr)
		}
	}
	v.cc.SetValue(strings.Join(filtered, ", "))
}

func (v *ComposeView) focusField(field ComposeField) {
	// Skip From field if only one identity
	if field == FieldFrom && len(v.identities) <= 1 {
		field = FieldTo
	}

	v.focused = field
	v.to.Blur()
	v.cc.Blur()
	v.subject.Blur()
	v.body.Blur()

	switch field {
	case FieldFrom:
		// No input widget for From, just visual focus
	case FieldTo:
		v.to.Focus()
	case FieldCc:
		v.cc.Focus()
	case FieldSubject:
		v.subject.Focus()
	case FieldBody:
		v.body.Focus()
	}
}

// Update handles input for the compose view
func (v *ComposeView) Update(msg tea.Msg) (*ComposeView, tea.Cmd) {
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		// Handle From field identity selection
		if v.focused == FieldFrom && len(v.identities) > 1 {
			switch msg.String() {
			case "left", "h":
				v.selectedIdentity--
				if v.selectedIdentity < 0 {
					v.selectedIdentity = len(v.identities) - 1
				}
				return v, nil
			case "right", "l", "enter":
				v.selectedIdentity++
				if v.selectedIdentity >= len(v.identities) {
					v.selectedIdentity = 0
				}
				return v, nil
			}
		}

		switch msg.String() {
		case "tab", "down":
			// Move to next field
			if v.focused < FieldBody {
				v.focusField(v.focused + 1)
				return v, nil
			}
		case "shift+tab", "up":
			// Move to previous field (only from header fields)
			minField := FieldTo
			if len(v.identities) > 1 {
				minField = FieldFrom
			}
			if v.focused > minField && v.focused < FieldBody {
				v.focusField(v.focused - 1)
				return v, nil
			}
			if v.focused == FieldBody {
				// From body, up just moves cursor in textarea
			}
		}
	}

	// Update the focused component
	var cmd tea.Cmd
	switch v.focused {
	case FieldFrom:
		// From field doesn't have an input widget
	case FieldTo:
		v.to, cmd = v.to.Update(msg)
	case FieldCc:
		v.cc, cmd = v.cc.Update(msg)
	case FieldSubject:
		v.subject, cmd = v.subject.Update(msg)
	case FieldBody:
		v.body, cmd = v.body.Update(msg)
	}
	cmds = append(cmds, cmd)

	return v, tea.Batch(cmds...)
}

// View renders the compose view
func (v *ComposeView) View() string {
	var b strings.Builder

	// Header
	modeStr := "compose"
	switch v.Mode {
	case ModeReply:
		modeStr = "reply"
	case ModeReplyAll:
		modeStr = "reply all"
	case ModeForward:
		modeStr = "forward"
	}
	header := composeHeaderStyle.Render("◈ " + modeStr)
	b.WriteString(header)
	b.WriteString("\n\n")

	// From field (only if multiple identities)
	if len(v.identities) > 1 {
		fromLabel := composeLabelStyle.Render("from: ")
		b.WriteString(fromLabel)

		// Get current identity display
		identityStr := ""
		if v.selectedIdentity < len(v.identities) {
			id := v.identities[v.selectedIdentity]
			if id.Name != "" {
				identityStr = fmt.Sprintf("%s <%s>", id.Name, id.Email)
			} else {
				identityStr = id.Email
			}
		}

		// Style based on focus
		if v.focused == FieldFrom {
			// Show arrows for cycling
			fromStyle := lipgloss.NewStyle().
				Foreground(composeColorPrimary).
				Background(composeColorBgSelect)
			b.WriteString(fromStyle.Render(fmt.Sprintf("◀ %s ▶", identityStr)))
		} else {
			fromStyle := lipgloss.NewStyle().Foreground(composeColorSecondary)
			b.WriteString(fromStyle.Render(identityStr))
		}
		b.WriteString("\n")
	}

	// To field
	toLabel := composeLabelStyle.Render("to: ")
	b.WriteString(toLabel)
	b.WriteString(v.to.View())
	b.WriteString("\n")

	// Cc field
	ccLabel := composeLabelStyle.Render("cc: ")
	b.WriteString(ccLabel)
	b.WriteString(v.cc.View())
	b.WriteString("\n")

	// Subject field
	subjectLabel := composeLabelStyle.Render("subject: ")
	b.WriteString(subjectLabel)
	b.WriteString(v.subject.View())
	b.WriteString("\n\n")

	// Body
	b.WriteString(v.body.View())
	b.WriteString("\n")

	// Help - add arrows hint if on From field
	helpText := "tab: next field │ ctrl+s: send │ esc: cancel"
	if v.focused == FieldFrom {
		helpText = "←/→: change identity │ tab: next field │ ctrl+s: send │ esc: cancel"
	}
	help := composeHelpStyle.Render(helpText)
	b.WriteString(help)

	return b.String()
}

// GetValues returns the composed email values
func (v *ComposeView) GetValues() (to, cc []string, subject, body string) {
	// Parse To addresses
	toVal := v.to.Value()
	if toVal != "" {
		for _, addr := range strings.Split(toVal, ",") {
			addr = strings.TrimSpace(addr)
			if addr != "" {
				to = append(to, addr)
			}
		}
	}

	// Parse CC addresses
	ccVal := v.cc.Value()
	if ccVal != "" {
		for _, addr := range strings.Split(ccVal, ",") {
			addr = strings.TrimSpace(addr)
			if addr != "" {
				cc = append(cc, addr)
			}
		}
	}

	subject = v.subject.Value()
	body = v.body.Value()

	return
}

// IsEmpty returns true if the body is empty (cancel condition)
func (v *ComposeView) IsEmpty() bool {
	return strings.TrimSpace(v.body.Value()) == ""
}

// HasRecipients returns true if there's at least one recipient
func (v *ComposeView) HasRecipients() bool {
	return strings.TrimSpace(v.to.Value()) != ""
}

// GetIdentity returns the selected sending identity
func (v *ComposeView) GetIdentity() *Identity {
	if v.selectedIdentity < len(v.identities) {
		return &v.identities[v.selectedIdentity]
	}
	return nil
}
