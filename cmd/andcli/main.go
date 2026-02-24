package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/tjblackheart/andcli/v2/internal/config"
	"github.com/tjblackheart/andcli/v2/internal/input"
	"github.com/tjblackheart/andcli/v2/internal/model"
	"github.com/tjblackheart/andcli/v2/internal/vaults"
	"github.com/tjblackheart/andcli/v2/internal/vaults/aegis"
	"github.com/tjblackheart/andcli/v2/internal/vaults/andotp"
	"github.com/tjblackheart/andcli/v2/internal/vaults/keepass"
	"github.com/tjblackheart/andcli/v2/internal/vaults/protonpass"
	"github.com/tjblackheart/andcli/v2/internal/vaults/stratum"
	"github.com/tjblackheart/andcli/v2/internal/vaults/twofas"
)

func main() {
	log.SetFlags(0)

	cfg, err := config.Create()
	if err != nil {
		log.Fatalf("andcli: %s", err)
	}

	vault, err := open(cfg)
	if err != nil {
		log.Fatalf("andcli: %s", err)
	}

	m := model.New(vault.Entries(), cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		log.Fatalf("andcli: %s", err)
	}

	if err := cfg.Persist(); err != nil {
		log.Fatalf("andcli: %s", err)
	}
}

func open(cfg *config.Config) (vaults.Vault, error) {
	name := cfg.File
	if os.Getenv("ANDCLI_HIDE_ABSPATH") != "" {
		name = filepath.Base(cfg.File)
	}

	log.Printf("Opening %s ...", name)

	pw, err := password(cfg.PasswdStdin())
	if err != nil {
		return nil, err
	}

	defer func() {
		for i := range pw {
			pw[i] = 0
		}
	}()

	switch cfg.Type {
	case vaults.TYPE_ANDOTP:
		return andotp.Open(cfg.File, pw)
	case vaults.TYPE_AEGIS:
		return aegis.Open(cfg.File, pw)
	case vaults.TYPE_TWOFAS:
		return twofas.Open(cfg.File, pw)
	case vaults.TYPE_STRATUM:
		return stratum.Open(cfg.File, pw)
	case vaults.TYPE_KEEPASS:
		return keepass.Open(cfg.File, pw)
	case vaults.TYPE_PROTON:
		return protonpass.Open(cfg.File, pw)
	}

	return nil, fmt.Errorf("vault type %q: not implemented", cfg.Type)
}

func password(piped bool) ([]byte, error) {
	if !piped {
		return input.Hidden("Password: ")
	}

	log.Printf("Reading password from stdin ...")
	return input.Stdin()
}
