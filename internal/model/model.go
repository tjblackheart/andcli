package model

import (
	"fmt"
	"path/filepath"
	"time"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/tjblackheart/andcli/v2/internal/buildinfo"
	"github.com/tjblackheart/andcli/v2/internal/clipboard"
	"github.com/tjblackheart/andcli/v2/internal/config"
	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

type (
	Model struct {
		list  list.Model
		state *appState
		cb    *clipboard.Clipboard
	}

	appState struct {
		showToken     bool
		showUsernames bool
		currentOTP    string
	}

	tickMsg struct{}
)

var (
	copyOK  = lipgloss.NewStyle().Foreground(green).Render(`✓`)
	copyErr = lipgloss.NewStyle().Foreground(red).Render(`✕`)
)

func New(entries []vaults.Entry, cfg *config.Config) Model {
	style := newThemedStyle(cfg.Theme)

	state := &appState{
		showToken:     cfg.Options.ShowTokens,
		showUsernames: cfg.Options.ShowUsernames,
	}

	items := make([]list.Item, 0)
	for _, e := range entries {
		items = append(items, e)
	}

	title := fmt.Sprintf("%s: %s", buildinfo.AppName, filepath.Base(cfg.File))
	keys := initKeys()
	dlg := &itemDelegate{style, state}

	return Model{
		list:  initList(items, dlg, keys, title),
		state: state,
		cb:    clipboard.New(cfg.ClipboardCmd),
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tick(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if m.list.FilterState() == list.Filtering {
			break
		}

		switch msg.String() {
		case "enter":
			m.state.showToken = !m.state.showToken
		case "u":
			m.state.showUsernames = !m.state.showUsernames
		case "c", "y":
			if !m.cb.IsInitialized() {
				msg := fmt.Sprintf("%s No clipboard command available", copyErr)
				return m, m.list.NewStatusMessage(msg)
			}

			msg := fmt.Sprintf("%s Token copied to clipboard", copyOK)
			if err := m.cb.Set([]byte(m.state.currentOTP)); err != nil {
				msg = fmt.Sprintf("%s %s: %s", copyErr, m.cb.String(), err)
			}
			return m, m.list.NewStatusMessage(msg)
		}

	case tickMsg:
		return m, tick()

	case tea.WindowSizeMsg:
		h, v := lipgloss.NewStyle().GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() tea.View {
	view := tea.NewView(lipgloss.NewStyle().Render(m.list.View()))
	view.AltScreen = true
	view.WindowTitle = m.list.Title
	return view
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func initKeys() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "toggle token")),
		key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "toggle usernames")),
		key.NewBinding(key.WithKeys("c", "y"), key.WithHelp("c/y", "yank to clipboard")),
	}
}

func initList(items []list.Item, delegate *itemDelegate, bindings []key.Binding, title string) list.Model {
	l := list.New(items, delegate, 0, 0)
	s := delegate.style

	l.FilterInput.Prompt = "Search for: "
	l.Styles.Filter.Focused.Prompt = s.filterPrompt
	l.Styles.Filter.Focused.Text = s.filterCursor
	l.Styles.Title = s.title
	l.InfiniteScrolling = true
	l.Title = title
	l.AdditionalShortHelpKeys = func() []key.Binding { return bindings }
	l.AdditionalFullHelpKeys = func() []key.Binding { return bindings }

	return l
}
