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
	"github.com/tjblackheart/andcli/v2/internal/vaults/stratum"
	"github.com/tjblackheart/andcli/v2/internal/vaults/twofas"
)

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

	m := model.New(vault.Entries(), cfg)
	p := tea.NewProgram(m, tea.WithAltScreen())

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

	var b []byte
	var err error

	if !c.ReadFromStdin() {
		b, err = input.AskHidden("Password: ")
		if err != nil {
			return nil, err
		}
	} else {
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

		b = s.Bytes()
	}

	switch c.Type {
	case vaults.TYPE_ANDOTP:
		return andotp.Open(c.File, b)
	case vaults.TYPE_AEGIS:
		return aegis.Open(c.File, b)
	case vaults.TYPE_TWOFAS:
		return twofas.Open(c.File, b)
	case vaults.TYPE_STRATUM:
		return stratum.Open(c.File, b)
	}

	return nil, fmt.Errorf("vault type %q: not implemented", c.Type)
}
