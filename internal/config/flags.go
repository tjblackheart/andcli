package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	flag "github.com/spf13/pflag"
	"github.com/tjblackheart/andcli/v2/internal/buildinfo"
	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

var (
	set     = flag.NewFlagSet("default", flag.ExitOnError)
	vfile   = set.StringP("file", "f", "", "Path to the encrypted vault (deprecated: Pass the filename directly)")
	vtype   = set.StringP("type", "t", "", fmt.Sprintf("Vault type (%s)", vaults.StrTypes()))
	cmd     = set.StringP("clipboard-cmd", "c", "", "A custom clipboard command, including args (xclip, wl-copy, pbcopy etc.)")
	pwstdin = set.Bool("passwd-stdin", false, "Read the vault password from stdin. If set, skips the password input.")
	query   = set.StringP("query", "q", "", "Query the vault directly and skip TUI functionality")
	version = set.BoolP("version", "v", false, "Prints version info and exits")
	timeout = set.Int("timeout", 5, "Timeout for decrypting the vault file")
	help    = set.BoolP("help", "h", false, "Show this help")
)

// Parses given flags into the existing config.
func (cfg *Config) parseFlags() error {
	set.Usage = func() { usage(true) }

	if err := set.Parse(os.Args[1:]); err != nil {
		log.Printf("%s: %s", buildinfo.AppName, err)
		usage(false)
		os.Exit(1)
	}

	if *version {
		fmt.Println(buildinfo.Long())
		os.Exit(0)
	}

	if *help {
		usage(true)
		os.Exit(0)
	}

	if *vfile != "" {
		abs, err := filepath.Abs(*vfile)
		if err != nil {
			return err
		}
		cfg.File = abs
		cfg.dirty = true
	}

	if *vtype != "" {
		cfg.Type = vaults.Type(*vtype)
		cfg.dirty = true
	}

	if *cmd != "" {
		cfg.ClipboardCmd = *cmd
		cfg.dirty = true
	}

	if *pwstdin {
		cfg.passwordFromStdin = true
	}

	if *query != "" {
		cfg.query = *query
	}

	if set.Arg(0) != "" {
		abs, err := filepath.Abs(set.Arg(0))
		if err != nil {
			return err
		}
		cfg.File = abs
		cfg.dirty = true
	}

	cfg.timeout = *timeout
	if cfg.timeout <= 0 {
		cfg.timeout = 5
	}

	return nil
}

// prints custom formatted usage information
func usage(includeDescription bool) {
	if includeDescription {
		fmt.Printf("%s - %s\n", buildinfo.AppName, buildinfo.Description)
	}

	msg := `
Usage: %s [options] <path/to/file>

Options:
`

	fmt.Fprintf(set.Output(), msg, os.Args[0])
	set.PrintDefaults()
}
