package model

import "github.com/charmbracelet/lipgloss"

const (
	green    = lipgloss.Color("#39A02E")
	yellow   = lipgloss.Color("#DB9F1F")
	red      = lipgloss.Color("#f10000")
	darkGrey = lipgloss.Color("#424242")
)

type appStyle struct {
	title, listItem, activeItem lipgloss.Style
	filterCursor, filterPrompt  lipgloss.Style
	token, until                lipgloss.Style
}

func newDefaultStyle() *appStyle {
	return &appStyle{
		title: lipgloss.NewStyle().
			Background(green).
			Padding(0, 1).
			MarginTop(1),

		listItem: lipgloss.NewStyle().
			PaddingLeft(3).
			MarginLeft(1).
			Faint(true),

		activeItem: lipgloss.NewStyle().
			Padding(0, 1, 0, 2).
			MarginLeft(1).
			Bold(true).
			Background(darkGrey).
			Border(lipgloss.ThickBorder(), false, false, false, true).
			BorderForeground(green).
			Faint(false),

		filterPrompt: lipgloss.NewStyle().
			Foreground(green).
			MarginTop(1),

		filterCursor: lipgloss.NewStyle().
			Background(green),

		token: lipgloss.NewStyle().Bold(true).Padding(0, 1, 0, 1),

		until: lipgloss.NewStyle().Bold(false),
	}
}
