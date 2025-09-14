package model

import "github.com/charmbracelet/lipgloss"

// TODO: configurable colors via config file
var (
	base   = lipgloss.Color("#39A02E")
	green  = lipgloss.Color("#39A02E")
	yellow = lipgloss.Color("#DB9F1F")
	red    = lipgloss.Color("#f10000")
	grey   = lipgloss.Color("#424242")
	black  = lipgloss.Color("#000000")
	white  = lipgloss.Color("#ffffff")
)

type defaultStyle struct {
	title, listItem, activeItem lipgloss.Style
	username, filterCursor      lipgloss.Style
	filterPrompt, token, until  lipgloss.Style
}

var ns = lipgloss.NewStyle

func newDefaultStyle() *defaultStyle {
	return &defaultStyle{
		title:        ns().Background(base).Padding(0, 1),
		listItem:     ns().PaddingLeft(2).Faint(true),
		username:     ns().Background(grey),
		filterPrompt: ns().Foreground(base),
		filterCursor: ns().Background(base),
		token:        ns().Bold(true).Padding(0, 1, 0, 1),
		until:        ns().Bold(true),
		activeItem: ns().
			Padding(0, 1).
			Bold(true).
			Background(grey).
			Border(lipgloss.ThickBorder(), false, false, false, true).
			Faint(false),
	}
}
