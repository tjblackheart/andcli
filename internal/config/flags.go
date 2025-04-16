package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"
	"github.com/tjblackheart/andcli/v2/internal/buildinfo"
	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

var (
	availableVaults = strings.Join(vaults.Types(), ", ")

	set     = flag.NewFlagSet("default", flag.ContinueOnError)
	vfile   = set.StringP("file", "f", "", "Path to the encrypted vault (deprecated: Pass the filename directly)")
	vtype   = set.StringP("type", "t", "", fmt.Sprintf("Vault type (%s)", availableVaults))
	cmd     = set.StringP("clipboard-cmd", "c", "", "A custom clipboard command, including args (xclip, wl-copy, pbcopy etc.)")
	version = set.BoolP("version", "v", false, "Prints version info and exits")
	help    = set.BoolP("help", "h", false, "Show this help")
)

// Parses given flags into the existing config.
func (cfg *Config) parseFlags() error {

	set.Usage = func() { usage(true) }
	set.SortFlags = false

	// FIXME: https://github.com/spf13/pflag/issues/352
	if err := set.Parse(os.Args[1:]); err != nil {
		if err == flag.ErrHelp {
			os.Exit(0)
		}

		log.Printf("andcli: %s", err)
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
	}

	if *vtype != "" {
		cfg.Type = *vtype
	}

	if *cmd != "" {
		cfg.ClipboardCmd = *cmd
	}

	if set.Arg(0) != "" {
		abs, err := filepath.Abs(set.Arg(0))
		if err != nil {
			return err
		}
		cfg.File = abs
	}

	return nil
}

// prints custom formatted usage information
func usage(includeBuildInfo bool) {
	msg := `
Usage: %s [options] <path/to/file>

Options:
`
	if includeBuildInfo {
		fmt.Print(buildinfo.Long(), "\n")
	}

	fmt.Fprintf(set.Output(), msg, os.Args[0])
	set.PrintDefaults()
}
