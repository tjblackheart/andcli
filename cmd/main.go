package main

import (
	"fmt"
	"log"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tjblackheart/andcli/internal/config"
	"github.com/tjblackheart/andcli/internal/input"
	"github.com/tjblackheart/andcli/internal/model"
	"github.com/tjblackheart/andcli/internal/vaults"
	"github.com/tjblackheart/andcli/internal/vaults/aegis"
	"github.com/tjblackheart/andcli/internal/vaults/andotp"
	"github.com/tjblackheart/andcli/internal/vaults/twofas"
)

func main() {
	log.SetFlags(0)

	cfg, err := config.Create()
	if err != nil {
		log.Fatal(err)
	}

	vault, err := openFile(cfg)
	if err != nil {
		log.Fatal(err)
	}

	m := model.New(vault.Entries(), cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	if err := cfg.Persist(); err != nil {
		log.Fatal(err)
	}
}

func openFile(c *config.Config) (vaults.Vault, error) {
	log.Printf("Opening %s ...", c.File)

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
