package main

import (
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/muesli/termenv"
)

const (
	list   = "list"
	detail = "detail"
)

type (
	model struct {
		filename string
		entries  []entry
		choices  []string
		cursor   int
		selected int
		view     string
		visible  bool
	}

	tickMsg struct{}
)

var (
	success = color.New(color.FgGreen, color.Bold)
	warning = color.New(color.FgYellow, color.Bold)
	danger  = color.New(color.FgRed, color.Bold)
	white   = color.New(color.FgWhite, color.Bold)
	muted   = color.New(color.FgHiWhite, color.Faint)

	copyCmd           = ""
	current           = "" // holds an unformatted copy of the current token
	copied            = false
	copiedVisibleSecs = 2
)

func newModel(file string, entries []entry) *model {
	m := &model{
		filename: file,
		entries:  entries,
		choices:  []string{},
		selected: -1,
		view:     list,
	}

	cmds := []string{"xclip", "pbcopy"} // linux, macos
	for _, c := range cmds {
		if err := exec.Command(c).Run(); err == nil {
			copyCmd = c
			break
		}
	}

	for _, e := range entries {
		choice := strings.TrimSpace(e.Issuer)
		if choice == "" {
			parts := strings.Split(e.Label, " - ")
			choice = parts[0]
		}

		label := e.Label
		parts := strings.Split(e.Label, " - ")
		if len(parts) > 1 {
			label = parts[1]
		}

		m.choices = append(m.choices, fmt.Sprintf("%s (%s)", choice, label))
	}

	return m
}

func (m model) Init() tea.Cmd { return tick() }

func (m model) View() string {
	s := m.header(fmt.Sprintf("%s %s: %s", app, tag, filepath.Base(m.filename)))
	if m.view == list {
		return s + m.list()
	}
	return s + m.detail()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			termenv.ClearScreen()
			return m, tea.Quit
		}
	}

	if m.selected != -1 {
		return m.updateDetail(msg)
	}

	return m.updateList(msg)
}

func (m *model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	last := len(m.choices) - 1

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "esc":
			termenv.ClearScreen()
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			} else {
				m.cursor = last
			}
		case "down", "j":
			if m.cursor < last {
				m.cursor++
			} else {
				m.cursor = 0
			}
		case "enter":
			m.selected = m.cursor
			m.view = detail
		case "pgdown":
			m.cursor = last
		case "pgup":
			m.cursor = 0
		}
	case tickMsg:
		return m, tick()
	}

	return m, nil
}

func (m *model) updateDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if msg.Type == tea.KeyEsc {
			m.selected = -1
			m.view = list
			m.visible = false
			current = ""
		}

		if msg.Type == tea.KeyEnter {
			m.visible = !m.visible
		}

		if msg.String() == "c" {
			if current != "" && copyCmd != "" {
				cmd := fmt.Sprintf("echo %s | %s", current, copyCmd)
				if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
					log.Println("copy:", err)
					return m, tea.Quit
				}
				copied = true
			}
		}
	case tickMsg:
		if copied {
			if copiedVisibleSecs > 0 {
				copiedVisibleSecs--
			} else {
				copied = false
				copiedVisibleSecs = 2
			}
		}

		return m, tick()
	}

	return m, nil
}

func (m model) list() string {
	s := fmt.Sprintf("Found %d entries. Select:\n\n", len(m.choices))

	for i, choice := range m.choices {
		cursor := " "
		if m.cursor == i {
			cursor = success.Sprint("> ")
			choice = white.Sprint(choice)
		}
		s += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	return s + m.footer()
}

func (m *model) detail() string {
	s := fmt.Sprintf("\n%s", m.choices[m.selected])

	token, exp := generateTOTP(&m.entries[m.selected])
	until := exp - time.Now().Unix()
	current = token

	if !m.visible {
		token = "******"
	}

	// format the token
	token = fmt.Sprintf("%s %s", token[:3], token[3:])
	fmtToken := success.Sprintf("%s", token)
	fmtUntil := white.Sprintf("%ds", until)

	if until <= 10 && until > 5 {
		fmtToken = warning.Sprintf("%s", token)
		fmtUntil = warning.Sprintf("%ds", until)
	}

	if until <= 5 {
		fmtToken = danger.Sprintf("%s", token)
		fmtUntil = danger.Sprintf("%ds", until)
	}

	if copied {
		fmtToken += success.Sprint(" ✓ ")
	}

	view := fmt.Sprintf("%s: %s\nValid: %s\n", s, fmtToken, fmtUntil)

	return view + m.footer()
}

func (m model) footer() string {
	footer := "[q, esc] quit | [enter] view"
	if m.view == detail {
		footer = "[q] quit | [enter] toggle visibility | [esc] go back"
		if copyCmd != "" {
			footer += " | [c] copy"
		}
	}
	return muted.Sprintf("\n%s\n", footer)
}

func (m model) header(s string) string {
	var line string
	for range s {
		line += "="
	}

	return fmt.Sprintf("%s\n%s\n", s, line)
}

func tick() tea.Cmd {
	return tea.Tick(time.Second, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}
