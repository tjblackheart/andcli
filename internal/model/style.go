package model

import "github.com/charmbracelet/lipgloss"

var (
	base   = lipgloss.Color("#39A02E")
	green  = lipgloss.Color("#39A02E")
	yellow = lipgloss.Color("#DB9F1F")
	red    = lipgloss.Color("#f10000")
	grey   = lipgloss.Color("#424242")
	black  = lipgloss.Color("#000000")
	white  = lipgloss.Color("#ffffff")
)

type appStyle struct {
	title, listItem, activeItem, username lipgloss.Style
	filterCursor, filterPrompt            lipgloss.Style
	token, until                          lipgloss.Style
}

func newDefaultStyle() *appStyle {
	return &appStyle{
		title: lipgloss.NewStyle().
			Background(base).
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
			Background(grey).
			Border(lipgloss.ThickBorder(), false, false, false, true).
			BorderForeground(base).
			Faint(false),

		username: lipgloss.NewStyle().
			Background(grey),

		filterPrompt: lipgloss.NewStyle().
			Foreground(base).
			MarginTop(1),

		filterCursor: lipgloss.NewStyle().
			Background(base),

		token: lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1, 0, 1),

		until: lipgloss.NewStyle().Bold(true),
	}
}
