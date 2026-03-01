package model

import (
	"charm.land/lipgloss/v2"
	"github.com/tjblackheart/andcli/v2/internal/config"
)

type appStyle struct {
	lipgloss.Style
	title, listItem, activeItem lipgloss.Style
	username, filterCursor      lipgloss.Style
	filterPrompt, token, until  lipgloss.Style
}

var (
	base   = lipgloss.Color(config.DefaultTheme.Base)
	green  = lipgloss.Color(config.DefaultTheme.Green)
	yellow = lipgloss.Color(config.DefaultTheme.Yellow)
	red    = lipgloss.Color(config.DefaultTheme.Red)
	grey   = lipgloss.Color(config.DefaultTheme.Grey)
	black  = lipgloss.Color(config.DefaultTheme.Black)
	white  = lipgloss.Color(config.DefaultTheme.White)
)

func newStyle() *appStyle {
	ls := lipgloss.NewStyle()

	as := &appStyle{
		title:        ls.Background(base).Padding(0, 1),
		listItem:     ls.PaddingLeft(2).Faint(true),
		username:     ls.Background(grey),
		filterPrompt: ls.Foreground(base),
		filterCursor: ls.Background(base),
		token:        ls.Bold(true).Padding(0, 1, 0, 1),
		until:        ls.Bold(true),
		activeItem: ls.
			Padding(0, 1).
			Bold(true).
			Background(grey).
			Border(lipgloss.ThickBorder(), false, false, false, true).
			Faint(false),

		Style: ls,
	}

	return as
}

func newThemedStyle(theme *config.Theme) *appStyle {
	base = lipgloss.Color(theme.Base)
	green = lipgloss.Color(theme.Green)
	yellow = lipgloss.Color(theme.Yellow)
	red = lipgloss.Color(theme.Red)
	grey = lipgloss.Color(theme.Grey)
	black = lipgloss.Color(theme.Black)
	white = lipgloss.Color(theme.White)

	return newStyle()
}
