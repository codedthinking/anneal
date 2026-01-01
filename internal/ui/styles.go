package ui

import "github.com/charmbracelet/lipgloss"

// Blade Runner / Cyberpunk color palette
var (
	// Neon accents
	ColorNeonPink   = lipgloss.Color("#FF2E97")
	ColorNeonCyan   = lipgloss.Color("#00FFFF")
	ColorNeonPurple = lipgloss.Color("#BD00FF")
	ColorNeonBlue   = lipgloss.Color("#00D4FF")
	ColorNeonOrange = lipgloss.Color("#FF6B35")

	// Background layers
	ColorBgDark    = lipgloss.Color("#0D0D0D")
	ColorBgMid     = lipgloss.Color("#1A1A2E")
	ColorBgLight   = lipgloss.Color("#16213E")
	ColorBgHover   = lipgloss.Color("#1F1F3D")
	ColorBgSelect  = lipgloss.Color("#2D2D5A")

	// Text
	ColorTextBright = lipgloss.Color("#EAEAEA")
	ColorTextNormal = lipgloss.Color("#B8B8B8")
	ColorTextMuted  = lipgloss.Color("#5C5C7A")
	ColorTextDim    = lipgloss.Color("#3D3D5C")

	// Status
	ColorSuccess = lipgloss.Color("#00FF9F")
	ColorWarning = lipgloss.Color("#FFE66D")
	ColorError   = lipgloss.Color("#FF4757")
	ColorInfo    = lipgloss.Color("#70A1FF")
)

// Gradient-like borders using special characters
var (
	BorderNeon = lipgloss.Border{
		Top:         "━",
		Bottom:      "━",
		Left:        "┃",
		Right:       "┃",
		TopLeft:     "┏",
		TopRight:    "┓",
		BottomLeft:  "┗",
		BottomRight: "┛",
	}

	BorderSoft = lipgloss.Border{
		Top:         "─",
		Bottom:      "─",
		Left:        "│",
		Right:       "│",
		TopLeft:     "╭",
		TopRight:    "╮",
		BottomLeft:  "╰",
		BottomRight: "╯",
	}
)

// App frame
var (
	AppStyle = lipgloss.NewStyle().
			Background(ColorBgDark)
)

// Header bar - neon gradient feel
var (
	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorTextBright).
			Background(ColorBgMid).
			BorderStyle(lipgloss.NormalBorder()).
			BorderBottom(true).
			BorderForeground(ColorNeonPink).
			Padding(0, 2).
			Bold(true)

	HeaderTitleStyle = lipgloss.NewStyle().
				Foreground(ColorNeonCyan).
				Bold(true)

	HeaderAccountStyle = lipgloss.NewStyle().
				Foreground(ColorNeonPink)

	LogoStyle = lipgloss.NewStyle().
			Foreground(ColorNeonCyan).
			Bold(true)
)

// Sidebar
var (
	SidebarStyle = lipgloss.NewStyle().
			Width(24).
			Background(ColorBgMid).
			BorderStyle(BorderSoft).
			BorderRight(true).
			BorderForeground(ColorTextDim).
			Padding(1, 0)

	SidebarActiveStyle = lipgloss.NewStyle().
				Width(24).
				Background(ColorBgMid).
				BorderStyle(BorderNeon).
				BorderRight(true).
				BorderForeground(ColorNeonPurple).
				Padding(1, 0)

	SidebarTitleStyle = lipgloss.NewStyle().
				Foreground(ColorNeonPurple).
				Bold(true).
				Padding(0, 2).
				MarginBottom(1)

	MailboxStyle = lipgloss.NewStyle().
			Foreground(ColorTextNormal).
			Padding(0, 2)

	MailboxSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorNeonCyan).
				Background(ColorBgSelect).
				Bold(true).
				Padding(0, 2)

	MailboxUnreadStyle = lipgloss.NewStyle().
				Foreground(ColorNeonPink).
				Bold(true)
)

// Email list
var (
	EmailListStyle = lipgloss.NewStyle().
			Background(ColorBgDark).
			Padding(0, 1)

	EmailListHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorTextMuted).
				Background(ColorBgLight).
				Bold(true).
				Padding(0, 1).
				BorderStyle(lipgloss.NormalBorder()).
				BorderBottom(true).
				BorderForeground(ColorTextDim)

	EmailItemStyle = lipgloss.NewStyle().
			Foreground(ColorTextNormal).
			Padding(0, 1)

	EmailItemSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorTextBright).
				Background(ColorBgSelect).
				BorderStyle(lipgloss.NormalBorder()).
				BorderLeft(true).
				BorderForeground(ColorNeonCyan).
				Padding(0, 1)

	EmailUnreadDotStyle = lipgloss.NewStyle().
				Foreground(ColorNeonPink).
				Bold(true)

	EmailFromStyle = lipgloss.NewStyle().
			Foreground(ColorTextNormal)

	EmailFromUnreadStyle = lipgloss.NewStyle().
				Foreground(ColorNeonCyan).
				Bold(true)

	EmailSubjectStyle = lipgloss.NewStyle().
				Foreground(ColorTextNormal)

	EmailSubjectUnreadStyle = lipgloss.NewStyle().
				Foreground(ColorTextBright).
				Bold(true)

	EmailPreviewStyle = lipgloss.NewStyle().
				Foreground(ColorTextMuted)

	EmailDateStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted)

	EmailFlagStyle = lipgloss.NewStyle().
			Foreground(ColorNeonOrange)

	EmailAttachmentStyle = lipgloss.NewStyle().
				Foreground(ColorTextMuted)
)

// Email reader
var (
	EmailReaderStyle = lipgloss.NewStyle().
				Background(ColorBgDark).
				Padding(1, 2)

	EmailReaderHeaderStyle = lipgloss.NewStyle().
				Background(ColorBgLight).
				BorderStyle(BorderSoft).
				BorderForeground(ColorTextDim).
				Padding(1, 2).
				MarginBottom(1)

	EmailReaderLabelStyle = lipgloss.NewStyle().
				Foreground(ColorNeonPurple).
				Bold(true).
				Width(10)

	EmailReaderValueStyle = lipgloss.NewStyle().
				Foreground(ColorTextBright)

	EmailReaderSubjectStyle = lipgloss.NewStyle().
				Foreground(ColorNeonCyan).
				Bold(true).
				MarginTop(1).
				MarginBottom(1)

	EmailReaderBodyStyle = lipgloss.NewStyle().
				Foreground(ColorTextNormal)

	EmailReaderAttachmentStyle = lipgloss.NewStyle().
					Foreground(ColorTextMuted).
					MarginTop(1)

	EmailReaderScrollStyle = lipgloss.NewStyle().
				Foreground(ColorNeonPink).
				Align(lipgloss.Right)
)

// Status bar - bottom bar with neon accent
var (
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted).
			Background(ColorBgMid).
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(ColorNeonPurple).
			Padding(0, 2)

	StatusKeyStyle = lipgloss.NewStyle().
			Foreground(ColorNeonCyan).
			Bold(true)

	StatusDescStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted)

	StatusModeStyle = lipgloss.NewStyle().
			Foreground(ColorBgDark).
			Background(ColorNeonPink).
			Bold(true).
			Padding(0, 1)
)

// Help overlay
var (
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorTextMuted).
			Background(ColorBgMid).
			BorderStyle(BorderSoft).
			BorderForeground(ColorNeonPurple).
			Padding(1, 2)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorNeonCyan).
			Bold(true)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorTextNormal)

	HelpSepStyle = lipgloss.NewStyle().
			Foreground(ColorTextDim)
)

// Spinner/Loading
var (
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(ColorNeonCyan)

	LoadingStyle = lipgloss.NewStyle().
			Foreground(ColorNeonPink).
			Bold(true)
)

// Error/Success messages
var (
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorError).
			Background(ColorBgMid).
			BorderStyle(BorderSoft).
			BorderForeground(ColorError).
			Bold(true).
			Padding(1, 2)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSuccess).
			Bold(true)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorWarning).
			Bold(true)
)

// Dialog/Modal
var (
	DialogStyle = lipgloss.NewStyle().
			Background(ColorBgLight).
			BorderStyle(BorderNeon).
			BorderForeground(ColorNeonPink).
			Padding(2, 4)

	DialogTitleStyle = lipgloss.NewStyle().
				Foreground(ColorNeonCyan).
				Bold(true).
				MarginBottom(1)
)
