package model

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tjblackheart/andcli/internal/vaults"
)

type itemDelegate struct{ style *defaultStyle }

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, idx int, i list.Item) {

	entry, _ := i.(vaults.Entry)
	text := d.style.listItem.Render(entry.Title())

	if idx != m.Index() {
		fmt.Fprint(w, text)
		return
	}

	token, exp := entry.GenerateTOTP()
	until := exp - time.Now().Unix()
	state.currentOTP = token

	bgColor, fgColor := green, white
	if until <= 10 && until > 5 {
		bgColor, fgColor = yellow, black
	}

	if until <= 5 {
		bgColor = red
	}

	formatted := "*** ***"
	if state.showToken {
		formatted = fmt.Sprintf("%s %s", state.currentOTP[:3], state.currentOTP[3:])
	}

	item := d.style.activeItem.Render(entry.Title())
	if state.showUsernames {
		user := style.username.Render(fmt.Sprintf("(%s) ", entry.Description()))
		item = fmt.Sprintf("%s%s", item, user)
	}

	text = fmt.Sprintf(
		"%s%s %s",
		item,
		style.token.Background(bgColor).Foreground(fgColor).Render(formatted),
		style.until.Foreground(bgColor).Render(fmt.Sprintf("%vs", until)),
	)

	fmt.Fprint(w, text)
}
