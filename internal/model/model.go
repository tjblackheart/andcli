package model

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/tjblackheart/andcli/v2/internal/buildinfo"
	"github.com/tjblackheart/andcli/v2/internal/clipboard"
	"github.com/tjblackheart/andcli/v2/internal/config"
	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

type (
	// tea.Model implementation
	Model struct {
		list  list.Model
		style lipgloss.Style
	}

	appState struct {
		showToken           bool
		showUsernames       bool
		alwaysShowUsernames bool
		currentOTP          string
	}

	tickMsg struct{}
)

var (
	style *defaultStyle
	state *appState
	cb    *clipboard.Clipboard
)

func New(entries []vaults.Entry, cfg *config.Config) Model {
	items := make([]list.Item, 0)
	for _, e := range entries {
		items = append(items, e)
	}

	cb = clipboard.New(cfg.ClipboardCmd)
	state = &appState{
		showToken:           cfg.Options.ShowTokens,
		showUsernames:       cfg.Options.ShowUsernames,
		alwaysShowUsernames: cfg.Options.AlwaysShowUsernames,
	}
	style = newDefaultStyle()
	title := fmt.Sprintf("%s: %s", buildinfo.AppName, filepath.Base(cfg.File))

	keys := initKeys()
	d := &itemDelegate{style}
	list := initList(items, d, keys, title)

	return Model{list: list}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		tea.SetWindowTitle(m.list.Title),
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
			state.showToken = !state.showToken
		case "u":
			state.showUsernames = !state.showUsernames
		case "c":
			if !cb.IsInitialized() {
				break
			}

			msg := "Token copied to clipboard"
			if err := cb.Set([]byte(state.currentOTP)); err != nil {
				msg = fmt.Sprintf("%s: %s", cb.String(), err)
			}

			return m, m.list.NewStatusMessage(msg)
		}

	case tickMsg:
		return m, tick()

	case tea.WindowSizeMsg:
		h, v := m.style.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m Model) View() string {
	return m.style.Render(m.list.View())
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}

func initKeys() []key.Binding {
	keys := []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "show/hide token")),
		key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "show/hide usernames")),
	}

	if cb != nil && cb.IsInitialized() {
		keys = append(keys, key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy")))
	}

	return keys
}

func initList(i []list.Item, d *itemDelegate, k []key.Binding, title string) list.Model {
	l := list.New(i, d, 0, 0)

	l.FilterInput.Prompt = "Search for: "
	l.FilterInput.PromptStyle = style.filterPrompt
	l.FilterInput.Cursor.Style = style.filterCursor
	l.Styles.Title = style.title
	l.InfiniteScrolling = true
	l.StatusMessageLifetime = 3 * time.Second
	l.Title = title
	l.AdditionalShortHelpKeys = func() []key.Binding { return k }
	l.AdditionalFullHelpKeys = func() []key.Binding { return k }

	return l
}
