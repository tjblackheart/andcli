package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/fatih/color"
	"github.com/muesli/termenv"
	"golang.org/x/term"
)

const (
	APP_NAME    = "andcli"
	VIEW_LIST   = "list"
	VIEW_DETAIL = "detail"
	TYPE_ANDOTP = "andotp"
	TYPE_AEGIS  = "aegis"
	TYPE_TWOFAS = "twofas"
)

var (
	// config
	cfgDir  = "."
	cfgFile = "config.yaml"

	// colors
	success = color.New(color.FgGreen, color.Bold)
	warning = color.New(color.FgYellow, color.Bold)
	danger  = color.New(color.FgRed, color.Bold)
	white   = color.New(color.FgWhite, color.Bold)
	muted   = color.New(color.FgHiWhite, color.Faint)

	// global ui stuff
	copyCmd            = ""
	current            = "" // holds an unformatted copy of the current token
	copied             = false
	copiedVisibleMSecs = 2000

	// build vars
	commit = ""
	date   = ""
	arch   = ""
	gover  = ""
	tag    = ""
)

func init() {
	var err error
	cfgDir, err = os.UserConfigDir()
	if err != nil {
		log.Fatal("[ERR] open config dir: ", err)
	}

	cfgDir = filepath.Join(cfgDir, APP_NAME)
	cfgFile = filepath.Join(cfgDir, cfgFile)

	log.SetFlags(0)
}

func main() {
	var vaultFile, vaultType string
	var showVersion bool

	flag.StringVar(&vaultFile, "f", "", "Path to the encrypted vault")
	flag.StringVar(&vaultType, "t", "", "Vault type (andotp, aegis, twofas)")
	flag.BoolVar(&showVersion, "v", false, "Show current version")
	flag.Parse()

	if showVersion {
		if tag != "" && arch != "" && commit != "" && date != "" && gover != "" {
			fmt.Printf(
				"%s %s %s (%s) built on %s with Go %s\n",
				APP_NAME, tag, arch, commit, date, gover,
			)
		} else {
			fmt.Printf("%s (direct install)\n", APP_NAME)
		}
		os.Exit(0)
	}

	prefix := danger.Sprint("[ERR]")

	cfg, err := newConfig(vaultFile, vaultType)
	if err != nil {
		log.Fatalf("%s: %s\n", prefix, err.Error())
	}

	if cfg.File == "" {
		log.Fatalf("%s: missing input file, specify one with -f\n", prefix)
	}

	if cfg.Type == "" {
		log.Fatalf("%s: missing vault type, specify one with -t\n", prefix)
	}

	entries, err := decrypt(cfg.File, cfg.Type)
	if err != nil {
		log.Fatalf("%s: %s\n", prefix, err.Error())
	}

	if err := cfg.save(); err != nil {
		log.Fatalf("%s: %s\n", prefix, err.Error())
	}

	output := termenv.DefaultOutput()
	output.ClearScreen()

	p := tea.NewProgram(newModel(output, cfg.File, entries...))
	if _, err := p.Run(); err != nil {
		log.Fatalf("%s: %s\n", prefix, err.Error())
	}
}

func decrypt(vaultFile, vaultType string, p ...[]byte) (entries, error) {
	fi, err := os.Stat(vaultFile)
	if err != nil {
		return nil, err
	}

	if fi.IsDir() {
		return nil, fmt.Errorf("not a file: %s", vaultFile)
	}

	var pass []byte
	if len(p) > 0 {
		pass = p[0]
	}

	fd := int(os.Stdin.Fd())
	if term.IsTerminal(fd) {
		fmt.Print("Enter password: ")
		pass, err = term.ReadPassword(fd)
		if err != nil {
			return nil, err
		}
		fmt.Println()
	}

	b, err := os.ReadFile(vaultFile)
	if err != nil {
		return nil, err
	}

	switch vaultType {
	case TYPE_ANDOTP:
		return decryptANDOTP(b, pass)
	case TYPE_AEGIS:
		return decryptAEGIS(b, pass)
	case TYPE_TWOFAS:
		return decryptTWOFAS(b, pass)
	}

	return nil, fmt.Errorf("vault type %q: not implemented", vaultType)
}
