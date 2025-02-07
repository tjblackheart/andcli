package model

import (
	"fmt"
	"io"
	"time"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/tjblackheart/andcli/internal/vaults"
)

type itemDelegate struct{ style *appStyle }

func (d itemDelegate) Height() int                             { return 1 }
func (d itemDelegate) Spacing() int                            { return 0 }
func (d itemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }

func (d itemDelegate) Render(w io.Writer, m list.Model, idx int, i list.Item) {

	entry, _ := i.(vaults.Entry)
	token, exp := entry.GenerateTOTP()
	until := exp - time.Now().Unix()
	text := d.style.listItem.Render(entry.Title())

	bgColor, fgColor := green, white
	if until <= 10 && until > 5 {
		bgColor, fgColor = yellow, black
	}

	if until <= 5 {
		bgColor = red
	}

	if idx == m.Index() {
		state.currentToken = token

		formatted := "*** ***"
		if state.showToken {
			formatted = fmt.Sprintf("%s %s", token[:3], token[3:])
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
	}

	fmt.Fprint(w, text)
}
