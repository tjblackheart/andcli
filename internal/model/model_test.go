package model

import (
	"reflect"
	"testing"

	"github.com/charmbracelet/bubbles/key"
	"github.com/tjblackheart/andcli/internal/clipboard"
)

func Test_initKeys(t *testing.T) {
	tests := []struct {
		name         string
		useClipboard bool
		want         []key.Binding
	}{
		{
			"inits keys without clipboard",
			false,
			[]key.Binding{
				key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "show/hide token")),
				key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "show/hide usernames")),
			},
		},
		{
			"inits keys with clipboard",
			true,
			[]key.Binding{
				key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "show/hide token")),
				key.NewBinding(key.WithKeys("u"), key.WithHelp("u", "show/hide usernames")),
				key.NewBinding(key.WithKeys("c"), key.WithHelp("c", "copy")),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cb = nil
			if tt.useClipboard {
				cb = clipboard.New("test")
			}

			if got := initKeys(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("initKeys() = %v, want %v", got, tt.want)
			}
		})
	}
}
