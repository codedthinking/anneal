package ui

import "github.com/charmbracelet/bubbles/key"

// KeyMap defines the keybindings for the application
type KeyMap struct {
	Up          key.Binding
	Down        key.Binding
	Left        key.Binding
	Right       key.Binding
	Top         key.Binding
	Bottom      key.Binding
	Enter       key.Binding
	Back        key.Binding
	Quit        key.Binding
	Compose     key.Binding
	Reply       key.Binding
	ReplyAll    key.Binding
	Forward     key.Binding
	Delete      key.Binding
	Archive     key.Binding
	Move        key.Binding
	Star        key.Binding
	MarkUnread  key.Binding
	Search      key.Binding
	Refresh     key.Binding
	Expand      key.Binding
	Collapse    key.Binding
	Help        key.Binding
	Account1    key.Binding
	Account2    key.Binding
	Account3    key.Binding
	Account4    key.Binding
	Account5    key.Binding
}

// DefaultKeyMap returns the default keybindings
func DefaultKeyMap() KeyMap {
	return KeyMap{
		Up: key.NewBinding(
			key.WithKeys("up", "k"),
			key.WithHelp("k/↑", "up"),
		),
		Down: key.NewBinding(
			key.WithKeys("down", "j"),
			key.WithHelp("j/↓", "down"),
		),
		Left: key.NewBinding(
			key.WithKeys("left", "h"),
			key.WithHelp("←/h", "back"),
		),
		Right: key.NewBinding(
			key.WithKeys("right", "l"),
			key.WithHelp("→/l", "open"),
		),
		Top: key.NewBinding(
			key.WithKeys("g", "home"),
			key.WithHelp("g", "top"),
		),
		Bottom: key.NewBinding(
			key.WithKeys("G", "end"),
			key.WithHelp("G", "bottom"),
		),
		Enter: key.NewBinding(
			key.WithKeys("enter"),
			key.WithHelp("enter", "open/expand"),
		),
		Back: key.NewBinding(
			key.WithKeys("esc", "q"),
			key.WithHelp("esc/q", "back"),
		),
		Quit: key.NewBinding(
			key.WithKeys("ctrl+c", "Q"),
			key.WithHelp("Q", "quit"),
		),
		Compose: key.NewBinding(
			key.WithKeys("c"),
			key.WithHelp("c", "compose"),
		),
		Reply: key.NewBinding(
			key.WithKeys("r"),
			key.WithHelp("r", "reply"),
		),
		ReplyAll: key.NewBinding(
			key.WithKeys("R"),
			key.WithHelp("R", "reply all"),
		),
		Forward: key.NewBinding(
			key.WithKeys("f"),
			key.WithHelp("f", "forward"),
		),
		Delete: key.NewBinding(
			key.WithKeys("d", "delete"),
			key.WithHelp("d", "delete"),
		),
		Archive: key.NewBinding(
			key.WithKeys("a"),
			key.WithHelp("a", "archive"),
		),
		Move: key.NewBinding(
			key.WithKeys("m"),
			key.WithHelp("m", "move"),
		),
		Star: key.NewBinding(
			key.WithKeys("s"),
			key.WithHelp("s", "star"),
		),
		MarkUnread: key.NewBinding(
			key.WithKeys("u"),
			key.WithHelp("u", "mark unread"),
		),
		Search: key.NewBinding(
			key.WithKeys("/"),
			key.WithHelp("/", "search"),
		),
		Refresh: key.NewBinding(
			key.WithKeys("ctrl+r"),
			key.WithHelp("ctrl+r", "refresh"),
		),
		Expand: key.NewBinding(
			key.WithKeys("space", "tab"),
			key.WithHelp("space", "expand thread"),
		),
		Collapse: key.NewBinding(
			key.WithKeys("shift+tab"),
			key.WithHelp("shift+tab", "collapse"),
		),
		Help: key.NewBinding(
			key.WithKeys("?"),
			key.WithHelp("?", "help"),
		),
		Account1: key.NewBinding(
			key.WithKeys("1"),
			key.WithHelp("1", "account 1"),
		),
		Account2: key.NewBinding(
			key.WithKeys("2"),
			key.WithHelp("2", "account 2"),
		),
		Account3: key.NewBinding(
			key.WithKeys("3"),
			key.WithHelp("3", "account 3"),
		),
		Account4: key.NewBinding(
			key.WithKeys("4"),
			key.WithHelp("4", "account 4"),
		),
		Account5: key.NewBinding(
			key.WithKeys("5"),
			key.WithHelp("5", "account 5"),
		),
	}
}

// ShortHelp returns keybindings for the short help view
func (k KeyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Left, k.Right, k.Enter, k.Help}
}

// FullHelp returns keybindings for the expanded help view
func (k KeyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Enter, k.Back, k.Expand},
		{k.Compose, k.Reply, k.ReplyAll, k.Forward},
		{k.Delete, k.Archive, k.Star, k.MarkUnread},
		{k.Search, k.Refresh, k.Help, k.Quit},
	}
}
