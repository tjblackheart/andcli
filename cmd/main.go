package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/muesli/termenv"

	"github.com/tjblackheart/andcli/internal/config"
	"github.com/tjblackheart/andcli/internal/input"
	"github.com/tjblackheart/andcli/internal/model"
	"github.com/tjblackheart/andcli/internal/vaults"
	"github.com/tjblackheart/andcli/internal/vaults/aegis"
	"github.com/tjblackheart/andcli/internal/vaults/andotp"
	"github.com/tjblackheart/andcli/internal/vaults/twofas"
)

var output = termenv.DefaultOutput()

func main() {
	log.SetFlags(0)

	cfg, err := config.Create()
	if err != nil {
		log.Fatal(err)
	}

	vault, err := open(cfg)
	if err != nil {
		log.Fatal(err)
	}

	p := tea.NewProgram(
		model.New(vault.Entries(), cfg),
		tea.WithFilter(clearScreenOnExit),
	)

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	if err := cfg.Persist(); err != nil {
		log.Fatal(err)
	}
}

func open(c *config.Config) (vaults.Vault, error) {
	name := c.File
	if os.Getenv("ANDCLI_HIDE_ABSPATH") != "" {
		name = filepath.Base(c.File)
	}

	log.Printf("Opening %s ...", name)

	b, err := input.AskHidden("Password: ")
	if err != nil {
		return nil, err
	}

	switch c.Type {
	case vaults.TYPE_ANDOTP:
		return andotp.Open(c.File, b)
	case vaults.TYPE_AEGIS:
		return aegis.Open(c.File, b)
	case vaults.TYPE_TWOFAS:
		return twofas.Open(c.File, b)
	}

	return nil, fmt.Errorf("vault type %q: not implemented", c.Type)
}

// clears the screen before quitting the program
func clearScreenOnExit(m tea.Model, msg tea.Msg) tea.Msg {
	if _, ok := msg.(tea.QuitMsg); !ok {
		return msg
	}

	output.ClearScreen()

	return msg
}
