package tui

import "github.com/charmbracelet/bubbles/key"

type keyMap struct {
	Up, Down, Left, Right key.Binding
	Top, Bottom           key.Binding
	HalfDown, HalfUp      key.Binding
	PageDown, PageUp      key.Binding
	Help, Quit, Back      key.Binding
	Open, Select, Restore  key.Binding
}

func defaultKeyMap() keyMap {
	return keyMap{
		Up: key.NewBinding(key.WithKeys("k", "up"), key.WithHelp("k/↑", "up")),
		Down: key.NewBinding(key.WithKeys("j", "down"), key.WithHelp("j/↓", "down")),
		Left: key.NewBinding(key.WithKeys("h", "left"), key.WithHelp("h/←", "prev pane")),
		Right: key.NewBinding(key.WithKeys("l", "right"), key.WithHelp("l/→", "next pane")),
		Top: key.NewBinding(key.WithKeys("g"), key.WithHelp("gg", "top")),
		Bottom: key.NewBinding(key.WithKeys("G"), key.WithHelp("G", "bottom")),
		HalfDown: key.NewBinding(key.WithKeys("ctrl+d"), key.WithHelp("ctrl-d", "half page")),
		HalfUp: key.NewBinding(key.WithKeys("ctrl+u"), key.WithHelp("ctrl-u", "half page")),
		PageDown: key.NewBinding(key.WithKeys("ctrl+f"), key.WithHelp("ctrl-f", "page down")),
		PageUp: key.NewBinding(key.WithKeys("ctrl+b"), key.WithHelp("ctrl-b", "page up")),
		Help: key.NewBinding(key.WithKeys("?"), key.WithHelp("?", "help")),
		Quit: key.NewBinding(key.WithKeys("q", "ctrl+c"), key.WithHelp("q", "quit")),
		Back: key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		Open: key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
		Select: key.NewBinding(key.WithKeys(" "), key.WithHelp("space", "select")),
		Restore: key.NewBinding(key.WithKeys("r"), key.WithHelp("r", "restore")),
	}
}

func (k keyMap) ShortHelp() []key.Binding {
	return []key.Binding{k.Up, k.Down, k.Open, k.Select, k.Restore, k.Help, k.Quit}
}

func (k keyMap) FullHelp() [][]key.Binding {
	return [][]key.Binding{
		{k.Up, k.Down, k.Left, k.Right},
		{k.Top, k.Bottom, k.HalfDown, k.HalfUp, k.PageDown, k.PageUp},
		{k.Open, k.Select, k.Restore, k.Help, k.Back, k.Quit},
	}
}
