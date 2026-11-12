package model

import (
	"fmt"
	"io"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

type itemDelegate struct {
	style *appStyle
	state *appState
}

func (d itemDelegate) Height() int                         { return 1 }
func (d itemDelegate) Spacing() int                        { return 0 }
func (d itemDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, idx int, li list.Item) {
	entry, _ := li.(vaults.Entry)
	text := d.style.listItem.Render(entry.Title())

	if idx != m.Index() {
		fmt.Fprint(w, text)
		return
	}

	otp := d.state.currentOTP
	until := max(otp.exp-time.Now().Unix(), 0)
	val := format(otp.token, d.state.showToken)

	bgColor, fgColor := green, white
	if until <= 10 && until > 5 {
		bgColor, fgColor = yellow, black
	}

	if until <= 5 {
		bgColor = red
	}

	item := d.style.activeItem.BorderForeground(bgColor).Render(entry.Title())
	if d.state.showUsernames {
		user := d.style.username.Render(fmt.Sprintf("(%s) ", entry.Description()))
		item = fmt.Sprintf("%s%s", item, user)
	}

	text = fmt.Sprintf(
		"%s%s %s",
		item,
		d.style.token.Background(bgColor).Foreground(fgColor).Render(val),
		d.style.until.Foreground(bgColor).Render(fmt.Sprintf("%vs", until)),
	)

	fmt.Fprint(w, text)
}

func format(token string, show bool) string {
	if !show || len(token) == 0 {
		return "*** ***"
	}

	if len(token) < 4 {
		return token
	}

	mid := len(token) / 2
	return token[:mid] + " " + token[mid:]
}
