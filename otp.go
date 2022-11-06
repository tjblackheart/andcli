package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/xlzd/gotp"
	"golang.org/x/term"
)

const (
	ANDOTP = "andotp"
	AEGIS  = "aegis"
)

type entry struct {
	Secret    string
	Issuer    string
	Label     string
	Digits    int
	Type      string
	Algorithm string
	Thumbnail string
	Period    int
	Tags      []string
}

func decrypt(vaultFile, vaultType string) ([]entry, error) {
	abs, err := filepath.Abs(vaultFile)
	if err != nil {
		return nil, err
	}

	fi, err := os.Stat(abs)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return nil, fmt.Errorf("not a file: %s", vaultFile)
	}

	fmt.Print("Enter password: ")
	pass, err := term.ReadPassword(int(os.Stdin.Fd()))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println()

	b, err := os.ReadFile(abs)
	if err != nil {
		return nil, err
	}

	switch vaultType {
	case ANDOTP:
		return decryptANDOTP(b, pass)
	case AEGIS:
		return decryptAEGIS(b, pass)
	default:
		return nil, fmt.Errorf("vault type %q: not implemented", vaultType)
	}
}

func generateTOTP(e *entry) (string, int64) {
	return gotp.NewDefaultTOTP(e.Secret).NowWithExpiration()
}
