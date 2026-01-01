package ui

import "github.com/charmbracelet/lipgloss"

// anneal color palette — the9x.ac brand
var (
	// Core colors
	ColorBg        = lipgloss.Color("#1d1d40") // background
	ColorPrimary   = lipgloss.Color("#d4d2e3") // primary text
	ColorSecondary = lipgloss.Color("#9795b5") // secondary text
	ColorAccent    = lipgloss.Color("#e61e25") // accent (used sparingly)

	// Derived shades
	ColorBgLight  = lipgloss.Color("#252550") // slightly lighter bg
	ColorBgSelect = lipgloss.Color("#2d2d5a") // selection bg
	ColorDim      = lipgloss.Color("#5a5880") // dim text
)

// Minimal borders
var (
	BorderMinimal = lipgloss.Border{
		Top:    "─",
		Bottom: "─",
	}
)

// App frame
var (
	AppStyle = lipgloss.NewStyle().
			Background(ColorBg)
)

// Header - minimal, just the name
var (
	HeaderStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Background(ColorBg).
			Padding(0, 2)

	HeaderTitleStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Bold(true)

	HeaderAccountStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary)

	LogoStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary).
			Bold(true)
)

// No sidebar in anneal - single pane focus
var (
	SidebarStyle = lipgloss.NewStyle().
			Width(24).
			Background(ColorBg).
			Padding(1, 0)

	SidebarActiveStyle = SidebarStyle

	SidebarTitleStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary).
				Padding(0, 2).
				MarginBottom(1)

	MailboxStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Padding(0, 2)

	MailboxSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Background(ColorBgSelect).
				Padding(0, 2)

	MailboxUnreadStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary)
)

// Message list
var (
	EmailListStyle = lipgloss.NewStyle().
			Background(ColorBg).
			Padding(0, 1)

	EmailListHeaderStyle = lipgloss.NewStyle().
				Foreground(ColorDim).
				Background(ColorBg).
				Padding(0, 1)

	EmailItemStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Padding(0, 1)

	EmailItemSelectedStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				Background(ColorBgSelect).
				Padding(0, 1)

	EmailUnreadDotStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary)

	EmailFromStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	EmailFromUnreadStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary)

	EmailSubjectStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary)

	EmailSubjectUnreadStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary)

	EmailPreviewStyle = lipgloss.NewStyle().
				Foreground(ColorDim)

	EmailDateStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	EmailFlagStyle = lipgloss.NewStyle().
			Foreground(ColorAccent)

	EmailAttachmentStyle = lipgloss.NewStyle().
				Foreground(ColorDim)
)

// Email reader
var (
	EmailReaderStyle = lipgloss.NewStyle().
				Background(ColorBg).
				Padding(1, 2)

	EmailReaderHeaderStyle = lipgloss.NewStyle().
				Background(ColorBg).
				Padding(1, 0).
				MarginBottom(1)

	EmailReaderLabelStyle = lipgloss.NewStyle().
				Foreground(ColorDim).
				Width(8)

	EmailReaderValueStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary)

	EmailReaderSubjectStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				MarginTop(1).
				MarginBottom(1)

	EmailReaderBodyStyle = lipgloss.NewStyle().
				Foreground(ColorSecondary)

	EmailReaderAttachmentStyle = lipgloss.NewStyle().
					Foreground(ColorDim).
					MarginTop(1)

	EmailReaderScrollStyle = lipgloss.NewStyle().
				Foreground(ColorDim).
				Align(lipgloss.Right)
)

// Status bar - minimal
var (
	StatusBarStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Background(ColorBg).
			Padding(0, 2)

	StatusKeyStyle = lipgloss.NewStyle().
			Foreground(ColorPrimary)

	StatusDescStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	StatusModeStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)
)

// Help - minimal
var (
	HelpStyle = lipgloss.NewStyle().
			Foreground(ColorDim).
			Background(ColorBg).
			Padding(0, 2)

	HelpKeyStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	HelpDescStyle = lipgloss.NewStyle().
			Foreground(ColorDim)

	HelpSepStyle = lipgloss.NewStyle().
			Foreground(ColorDim)
)

// Loading - calm, no urgency
var (
	SpinnerStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	LoadingStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)
)

// No red error states per brand guide
var (
	ErrorStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary).
			Background(ColorBg).
			Padding(1, 2)

	SuccessStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)

	WarningStyle = lipgloss.NewStyle().
			Foreground(ColorSecondary)
)

// Dialog - minimal
var (
	DialogStyle = lipgloss.NewStyle().
			Background(ColorBgLight).
			Padding(2, 4)

	DialogTitleStyle = lipgloss.NewStyle().
				Foreground(ColorPrimary).
				MarginBottom(1)
)
