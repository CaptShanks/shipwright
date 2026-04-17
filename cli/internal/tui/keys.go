package tui

import "github.com/charmbracelet/bubbles/key"

type KeyMap struct {
	Tab      key.Binding
	ShiftTab key.Binding
	Num1     key.Binding
	Num2     key.Binding
	Num3     key.Binding
	Enter    key.Binding
	Esc      key.Binding
	Quit     key.Binding
	Help     key.Binding
	Filter   key.Binding
}

var Keys = KeyMap{
	Tab: key.NewBinding(
		key.WithKeys("tab"),
		key.WithHelp("tab", "next tab"),
	),
	ShiftTab: key.NewBinding(
		key.WithKeys("shift+tab"),
		key.WithHelp("shift+tab", "prev tab"),
	),
	Num1: key.NewBinding(
		key.WithKeys("1"),
		key.WithHelp("1", "dashboard"),
	),
	Num2: key.NewBinding(
		key.WithKeys("2"),
		key.WithHelp("2", "marketplace"),
	),
	Num3: key.NewBinding(
		key.WithKeys("3"),
		key.WithHelp("3", "installed"),
	),
	Enter: key.NewBinding(
		key.WithKeys("enter"),
		key.WithHelp("enter", "select"),
	),
	Esc: key.NewBinding(
		key.WithKeys("esc"),
		key.WithHelp("esc", "back"),
	),
	Quit: key.NewBinding(
		key.WithKeys("q", "ctrl+c"),
		key.WithHelp("q", "quit"),
	),
	Help: key.NewBinding(
		key.WithKeys("?"),
		key.WithHelp("?", "help"),
	),
	Filter: key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("/", "filter"),
	),
}
