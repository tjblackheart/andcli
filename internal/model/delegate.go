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

	tokenColor := green
	if until <= 15 && until > 10 {
		tokenColor = yellow
	}

	if until <= 10 {
		tokenColor = red
	}

	text := d.style.listItem.Render(entry.Title())

	if idx == m.Index() {
		state.currentToken = token

		formattedToken := "*** ***"
		if state.showToken {
			formattedToken = fmt.Sprintf("%s %s", token[:3], token[3:])
		}

		// TODO: integrate optional entry.Description
		text = fmt.Sprintf(
			"%s%s %s",
			d.style.activeItem.Render(entry.Title()),
			style.token.Background(tokenColor).Padding(0, 1, 0, 1).Render(formattedToken),
			style.until.Foreground(tokenColor).Render(fmt.Sprintf("%vs", until)),
		)
	}

	fmt.Fprint(w, text)
}
