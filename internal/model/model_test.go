package model

import (
	"strings"
	"testing"

	tea "charm.land/bubbletea/v2"
	"github.com/tjblackheart/andcli/v2/internal/clipboard"
)

func TestModel_Update(t *testing.T) {
	t.Run("osc52", func(t *testing.T) {
		m := newTestModel(new(clipboard.Clipboard))

		updated, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Text: "c"})
		if cmd == nil {
			t.Fatal("Model.Update() command = nil, want OSC52 and status commands")
		}

		msg := cmd()
		batch, ok := msg.(tea.BatchMsg)
		if !ok {
			t.Fatalf("Model.Update() command returned %T, want tea.BatchMsg", msg)
		}
		if len(batch) != 2 {
			t.Fatalf("OSC52 fallback returned %d commands, want 2", len(batch))
		}

		view := updated.(Model).list.View()
		if !strings.Contains(view, "Token copied to terminal clipboard") {
			t.Errorf("list view does not contain OSC52 status: %q", view)
		}
	})

	t.Run("async", func(t *testing.T) {
		const command = "andcli-test-command-that-does-not-exist"
		m := newTestModel(clipboard.New(command))

		_, cmd := m.Update(tea.KeyPressMsg{Code: 'y', Text: "y"})
		if cmd == nil {
			t.Fatal("Model.Update() command = nil, want asynchronous copy")
		}

		msg := cmd()
		result, ok := msg.(cbMsg)
		if !ok {
			t.Fatalf("copy command returned %T, want cbMsg", msg)
		}
		if result.err == nil {
			t.Fatal("cbMsg.err = nil, want an execution error")
		}
		if result.command != command {
			t.Errorf("cbMsg.command = %q, want %q", result.command, command)
		}

		updated, status := m.Update(result)
		if status == nil {
			t.Fatal("cbMsg returned nil status command")
		}
		view := updated.(Model).list.View()
		if !strings.Contains(view, command) {
			t.Errorf("list view does not contain failed command: %q", view)
		}
	})

	t.Run("success", func(t *testing.T) {
		m := newTestModel(clipboard.New("test"))

		updated, status := m.Update(cbMsg{command: "test"})
		if status == nil {
			t.Fatal("cbMsg returned nil status command")
		}
		view := updated.(Model).list.View()
		if !strings.Contains(view, "Token copied to system clipboard") {
			t.Errorf("list view does not contain success status: %q", view)
		}
	})
}

func newTestModel(cb *clipboard.Clipboard) Model {
	state := &appState{currentOTP: &otp{token: "123456"}}
	style := newStyle()
	m := Model{
		state: state,
		style: style,
		cb:    cb,
	}
	m.list = initList(nil, &itemDelegate{style: style, state: state}, "")
	m.list.SetSize(200, 20)
	return m
}
