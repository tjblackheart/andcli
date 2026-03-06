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
		style *appStyle
		cb    *clipboard.Clipboard
	}

	appState struct {
		showToken     bool
		showUsernames bool
		currentOTP    *otp
	}

	otp struct {
		token string
		exp   int64
	}

	tickMsg struct{}
)

var (
	copyOK  = lipgloss.NewStyle().Foreground(green).Render(`✓`)
	copyErr = lipgloss.NewStyle().Foreground(red).Render(`✕`)
)

func New(entries []vaults.Entry, cfg *config.Config) Model {
	state := &appState{
		showToken:     cfg.Options.ShowTokens,
		showUsernames: cfg.Options.ShowUsernames,
		currentOTP:    &otp{},
	}

	items := make([]list.Item, 0)
	for _, e := range entries {
		items = append(items, e)
	}

	style := newThemedStyle(cfg.Theme)
	title := fmt.Sprintf("%s: %s", buildinfo.AppName, filepath.Base(cfg.File))
	dlg := &itemDelegate{style, state}

	m := Model{
		list:  initList(items, dlg, title),
		state: state,
		style: style,
		cb:    clipboard.New(cfg.ClipboardCmd),
	}

	m.updateToken()

	return m
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
			if err := m.cb.Set([]byte(m.state.currentOTP.token)); err != nil {
				msg = fmt.Sprintf("%s %s: %s", copyErr, m.cb.String(), err)
			}

			return m, m.list.NewStatusMessage(msg)
		}

	case tickMsg:
		m.updateToken()
		return m, tick()

	case tea.WindowSizeMsg:
		h, v := m.style.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)

	return m, cmd
}

func (m Model) View() tea.View {
	view := tea.NewView(m.style.Render(m.list.View()))
	view.AltScreen = true
	view.WindowTitle = m.list.Title
	return view
}

func (m *Model) updateToken() {
	item := m.list.SelectedItem()
	if item == nil {
		return
	}

	entry, ok := item.(vaults.Entry)
	if !ok {
		return
	}

	token, exp := entry.GenerateTOTP()
	m.state.currentOTP.token = token
	m.state.currentOTP.exp = exp
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func initList(items []list.Item, delegate *itemDelegate, title string) list.Model {
	lst := list.New(items, delegate, 0, 0)
	style := delegate.style

	keys := []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "toggle token")),
		key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "toggle usernames")),
		key.NewBinding(key.WithKeys("c", "y"), key.WithHelp("c/y", "yank to clipboard")),
	}

	lst.FilterInput.Prompt = "Search for: "
	lst.Styles.Filter.Focused.Prompt = style.filterPrompt
	lst.Styles.Filter.Focused.Text = style.filterCursor
	lst.Styles.Title = style.title
	lst.InfiniteScrolling = true
	lst.Title = title
	lst.AdditionalShortHelpKeys = func() []key.Binding { return keys }
	lst.AdditionalFullHelpKeys = func() []key.Binding { return keys }

	return lst
}
