package main

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"
	"golang.design/x/clipboard"
)

type (
	model struct {
		filename           string
		items              entries
		filtered           entries
		cursor             int
		selected           int
		view               string
		visible            bool
		query              string
		output             *termenv.Output
		copied             bool
		copyFailed         bool
		copiedVisibleMSecs int
	}

	tickMsg struct{}
)

func newModel(o *termenv.Output, filename, cfgClipboardCmd string, entries ...entry) *model {
	m := &model{
		filename:           filename,
		items:              entries,
		selected:           -1,
		view:               VIEW_LIST,
		output:             o,
		copied:             false,
		copyFailed:         false,
		copiedVisibleMSecs: 2000,
	}

	copyCmd = setupClipboard(cfgClipboardCmd)

	for i, e := range m.items {
		issuer := strings.TrimSpace(e.Issuer)
		if issuer == "" {
			parts := strings.Split(e.Label, " - ")
			issuer = parts[0]
		}

		label := e.Label
		parts := strings.Split(e.Label, " - ")
		if len(parts) > 1 {
			label = parts[1]
		}

		m.items[i].Choice = issuer
		if label != "" {
			m.items[i].Choice = fmt.Sprintf("%s (%s)", issuer, label)
		}
	}

	return m
}

func (m model) Init() tea.Cmd { return tick() }

func (m model) View() string {
	s := m.header(fmt.Sprintf("%s %s: %s", APP_NAME, tag, filepath.Base(m.filename)))
	if m.view == VIEW_LIST {
		return s + m.list()
	}
	return s + m.detail()
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			m.output.ClearScreen()
			return m, tea.Quit
		}
	}

	if m.selected != -1 {
		return m.updateDetail(msg)
	}

	return m.updateList(msg)
}

func (m *model) updateList(msg tea.Msg) (tea.Model, tea.Cmd) {
	m.filtered = m.items.filter(m.query)

	last := len(m.filtered) - 1
	if last < 0 {
		last = 0
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			if m.query == "" {
				m.output.ClearScreen()
				return m, tea.Quit
			}
			m.query = ""
			m.filtered = m.items
		case tea.KeyUp:
			m.cursor--
			if m.cursor < 0 {
				m.cursor = last
			}
		case tea.KeyDown:
			m.cursor++
			if m.cursor > last {
				m.cursor = 0
			}
		case tea.KeyEnter:
			if len(m.filtered) > 0 {
				m.selected = m.cursor
				m.view = VIEW_DETAIL
			}
		case tea.KeyPgDown:
			m.cursor = last
		case tea.KeyPgUp:
			m.cursor = 0
		case tea.KeyBackspace:
			if len(m.query) > 0 {
				m.query = m.query[:len(m.query)-1]
			}
		case tea.KeyRunes:
			m.cursor = 0
			m.query += msg.String()
		}
	case tickMsg:
		return m, tick()
	}

	return m, nil
}

func (m *model) updateDetail(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEsc:
			m.query = ""
			m.selected = -1
			m.view = VIEW_LIST
			m.visible = false
			current = ""
		case tea.KeyEnter:
			m.visible = !m.visible
		case tea.KeyRunes:
			if msg.String() == "q" {
				m.output.ClearScreen()
				return m, tea.Quit
			}
			if msg.String() == "c" {
				if current != "" && copyCmd != "" {
					if copyCmd == "clipboard" {
						currentBytes := []byte(current)
						clipboard.Write(clipboard.FmtText, currentBytes)
						if !bytes.Equal(clipboard.Read(clipboard.FmtText), currentBytes) {
							log.Println("copy: failed")
							return m, tea.Quit
						}
					} else {
						cmd := fmt.Sprintf("echo -n %s | %s", current, copyCmd)
						if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
							log.Println("copy:", err)
							return m, tea.Quit
						}
					}
					m.copyFailed = false
					m.copied = true
				}
			}
		}
	case tickMsg:
		if m.copied {
			if m.copiedVisibleMSecs > 0 {
				m.copiedVisibleMSecs--
			} else {
				m.copied = false
				m.copiedVisibleMSecs = 2000
			}
		}

		return m, tick()
	}

	return m, nil
}

func (m model) list() string {
	list := ""

	for i, e := range m.filtered {
		cursor := " "
		choice := e.Choice
		if m.cursor == i {
			cursor = success.Sprint("> ")
			choice = white.Sprint(e.Choice)
		}
		list += fmt.Sprintf("%s %s\n", cursor, choice)
	}

	return list + m.footer()
}

func (m *model) detail() string {
	name := fmt.Sprintf("\n%s", m.filtered[m.selected].Choice)
	entry := m.filtered[m.selected]

	token, exp := entry.generateTOTP()
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

	if m.copied {
		fmtToken += success.Sprint(" ✓ ")
	}

	if m.copyFailed {
		fmtToken += danger.Sprint(" ✗ \nCopy command (" + copyCmd + ") failed, check your configuration!")
	}

	view := fmt.Sprintf("%s: %s\nValid: %s\n", name, fmtToken, fmtUntil)

	return view + m.footer()
}

func (m model) footer() string {
	footer := "[esc] quit"
	if len(m.query) > 0 {
		footer = "[esc] clear search"
	}

	if len(m.filtered) > 0 {
		footer += " | [enter] view"
	}

	if m.view == VIEW_DETAIL {
		footer = "[esc] back | [q] quit | [enter] toggle visibility"
		if copyCmd != "" {
			footer += " | [c] copy"
		}
	}

	return muted.Sprintf("\n%s\n", footer)
}

func (m model) header(s string) string {
	if s == "" {
		return "\n"
	}

	var line string
	for range s {
		line += "="
	}

	word := "entries"
	length := len(m.filtered)
	if length == 1 {
		word = "entry"
	}

	counter := fmt.Sprintf("%d %s.", length, word)

	header := fmt.Sprintf("%s\n%s\n%s\n", s, line, counter)
	if m.view != VIEW_DETAIL {
		header += "\nType to search: " + white.Sprint(m.query)
		header += "\n\n"
	}

	return header
}

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond, func(time.Time) tea.Msg {
		return tickMsg{}
	})
}
