package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tjblackheart/andcli/v2/internal/buildinfo"
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
	log.SetPrefix(fmt.Sprintf("%s: ", buildinfo.AppName))

	cfg, err := config.Create()
	if err != nil {
		log.Fatalln(err)
	}

	vault, err := open(cfg)
	if err != nil {
		log.Fatalln(err)
	}

	entries := vault.Entries()
	if cfg.Query() != "" {
		entry, err := vaults.Find(cfg.Query(), entries)
		if err != nil {
			log.Fatalln(err)
		}

		token, exp := entry.GenerateTOTP()
		until := max(exp-time.Now().Unix(), 0)

		fmt.Printf("%s %s %ds\n", entry.Issuer, token, until)
		os.Exit(0)
	}

	m := model.New(entries, cfg)
	if _, err := tea.NewProgram(m).Run(); err != nil {
		log.Fatalln(err)
	}

	if err := cfg.Persist(); err != nil {
		log.Fatalln(err)
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

	done := make(chan struct{})

	var vault vaults.Vault
	go func() {
		switch cfg.Type {
		case vaults.ANDOTP:
			vault, err = andotp.Open(cfg.File, pw)
		case vaults.AEGIS:
			vault, err = aegis.Open(cfg.File, pw)
		case vaults.TWOFAS:
			vault, err = twofas.Open(cfg.File, pw)
		case vaults.STRATUM:
			vault, err = stratum.Open(cfg.File, pw)
		case vaults.KEEPASS:
			vault, err = keepass.Open(cfg.File, pw)
		case vaults.PROTON:
			vault, err = protonpass.Open(cfg.File, pw)
		default:
			vault, err = nil, fmt.Errorf("vault type %q: not implemented", cfg.Type)
		}
		done <- struct{}{}
	}()

	select {
	case <-done:
	case <-time.After(cfg.Timeout()):
		return nil, fmt.Errorf("decrypt: operation timed out. wrong type?")
	}

	return vault, err
}

func password(piped bool) ([]byte, error) {
	if !piped {
		return input.Hidden("Password: ")
	}

	log.Printf("Reading password from stdin ...")
	return input.Stdin()
}
