package model

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/tjblackheart/andcli/internal/buildinfo"
	"github.com/tjblackheart/andcli/internal/config"
	"github.com/tjblackheart/andcli/internal/vaults"
)

type (
	// tea.Model implementation
	Model struct {
		list  list.Model
		style lipgloss.Style
	}

	tickMsg  struct{}
	frameMsg struct{}

	appState struct {
		showToken bool
		//showDescription bool
	}
)

var (
	style *appStyle
	state *appState
)

func New(entries []vaults.Entry, cfg *config.Config) Model {
	items := make([]list.Item, 0)
	for _, e := range entries {
		items = append(items, e)
	}

	state = &appState{}
	style = newDefaultStyle()

	delegate := itemDelegate{style}
	keys := []key.Binding{
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "show/hide token")),
		key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy")),
	}

	l := list.New(items, delegate, 0, 0)
	l.FilterInput.Prompt = "Search for: "
	l.FilterInput.PromptStyle = style.filterPrompt
	l.FilterInput.Cursor.Style = style.filterCursor
	l.Styles.Title = style.title
	l.InfiniteScrolling = true
	l.StatusMessageLifetime = 3 * time.Second
	l.Title = fmt.Sprintf("%s: %s", buildinfo.AppName, filepath.Base(cfg.File))
	l.AdditionalShortHelpKeys = func() []key.Binding { return keys }
	l.AdditionalFullHelpKeys = func() []key.Binding { return keys }

	return Model{list: l}
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
		switch msg.String() {
		case "enter":
			if m.list.FilterState() != list.Filtering {
				state.showToken = !state.showToken
			}
		case "c":
			m.list.NewStatusMessage("copy token: TODO")
			/* c := exec.Command("vim", "file.txt")
			cmd := ExecProcess(c, func(err error) Msg {
				return VimFinishedMsg{err: err}
			}) */
		}

	case tickMsg:
		return m, tick()

	case frameMsg:
		return m, frame()

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

func frame() tea.Cmd {
	return tea.Tick(time.Second/60, func(time.Time) tea.Msg {
		return frameMsg{}
	})
}
