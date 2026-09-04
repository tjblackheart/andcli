package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	tea "charm.land/bubbletea/v2"

	"github.com/tjblackheart/andcli/v2/internal/config"
	"github.com/tjblackheart/andcli/v2/internal/input"
	"github.com/tjblackheart/andcli/v2/internal/model"
	"github.com/tjblackheart/andcli/v2/internal/spinner"
	"github.com/tjblackheart/andcli/v2/internal/vaults"

	_ "github.com/tjblackheart/andcli/v2/internal/vaults/aegis"
	_ "github.com/tjblackheart/andcli/v2/internal/vaults/andotp"
	_ "github.com/tjblackheart/andcli/v2/internal/vaults/ente"
	_ "github.com/tjblackheart/andcli/v2/internal/vaults/keepass"
	_ "github.com/tjblackheart/andcli/v2/internal/vaults/protonauth"
	_ "github.com/tjblackheart/andcli/v2/internal/vaults/protonpass"
	_ "github.com/tjblackheart/andcli/v2/internal/vaults/stratum"
	_ "github.com/tjblackheart/andcli/v2/internal/vaults/twofas"
)

func main() {
	log.SetFlags(0)

	cfg, err := config.Create()
	if err != nil {
		log.Fatalln(err)
	}

	vault, err := open(cfg)
	if err != nil {
		log.Fatalf("Error reading file: %s\n", err)
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
	if _, ok := os.LookupEnv("ANDCLI_HIDE_ABSPATH"); ok {
		name = filepath.Base(cfg.File)
	}
	log.Printf("Opening %s ...", name)

	pw, err := password(cfg.PasswdStdin())
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.DecryptionTimeoutD())
	defer cancel()

	s := spinner.Brew.SetSuffix("Decrypting ...")
	s.Start()
	defer s.Stop()

	return vaults.Open(ctx, cfg.File, pw, cfg.Type)
}

func password(piped bool) ([]byte, error) {
	if !piped {
		return input.Hidden("Password: ")
	}

	log.Printf("Reading password from stdin ...")
	return input.Stdin()
}
