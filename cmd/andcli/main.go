package main

import (
	"bufio"
	"errors"
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

	b, err := getPass(cfg)
	if err != nil {
		return nil, err
	}

	switch cfg.Type {
	case vaults.TYPE_ANDOTP:
		return andotp.Open(cfg.File, b)
	case vaults.TYPE_AEGIS:
		return aegis.Open(cfg.File, b)
	case vaults.TYPE_TWOFAS:
		return twofas.Open(cfg.File, b)
	case vaults.TYPE_STRATUM:
		return stratum.Open(cfg.File, b)
	case vaults.TYPE_KEEPASS:
		return keepass.Open(cfg.File, b)
	case vaults.TYPE_PROTON:
		return protonpass.Open(cfg.File, b)
	}

	return nil, fmt.Errorf("vault type %q: not implemented", cfg.Type)
}

func getPass(cfg *config.Config) ([]byte, error) {

	if !cfg.PasswdStdin() {
		return input.AskHidden("Password: ")
	}

	log.Printf("Reading pass from stdin ...")

	stat, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}

	if (stat.Mode() & os.ModeCharDevice) != 0 {
		return nil, errors.New("stdin: no input provided")
	}

	s := bufio.NewScanner(bufio.NewReader(os.Stdin))
	if s.Scan(); s.Err() != nil {
		return nil, s.Err()
	}

	return s.Bytes(), nil
}
