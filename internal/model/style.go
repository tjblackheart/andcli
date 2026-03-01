package model

import (
	"github.com/charmbracelet/lipgloss"
	"github.com/tjblackheart/andcli/v2/internal/config"
)

type appStyle struct {
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
	return &appStyle{
		title:        lipgloss.NewStyle().Background(base).Padding(0, 1),
		listItem:     lipgloss.NewStyle().PaddingLeft(2).Faint(true),
		username:     lipgloss.NewStyle().Background(grey),
		filterPrompt: lipgloss.NewStyle().Foreground(base),
		filterCursor: lipgloss.NewStyle().Background(base),
		token:        lipgloss.NewStyle().Bold(true).Padding(0, 1, 0, 1),
		until:        lipgloss.NewStyle().Bold(true),
		activeItem: lipgloss.NewStyle().
			Padding(0, 1).
			Bold(true).
			Background(grey).
			Border(lipgloss.ThickBorder(), false, false, false, true).
			Faint(false),
	}
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
