package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjblackheart/andcli/v2/internal/buildinfo"
	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

var (
	availableVaults = strings.Join(vaults.Types(), ", ")

	vfile   = flag.String("f", "", "Path to the encrypted vault (deprecated: Pass the filename directly)")
	vtype   = flag.String("t", "", fmt.Sprintf("Vault type (%s)", availableVaults))
	cmd     = flag.String("c", "", "Clipboard command (xclip, wl-copy, pbcopy etc.)")
	version = flag.Bool("v", false, "Prints version info and exits")
)

// Parses given flags into the existing config.
func (cfg *Config) parseFlags() error {

	flag.Usage = usage
	flag.Parse()

	if *version {
		fmt.Println(buildinfo.Long())
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

	if flag.Arg(0) != "" {
		cfg.File = flag.Arg(0)
	}

	return nil
}

// prints custom formatted usage information
func usage() {
	msg := `Usage: %s [options] <path/to/file>

Options:
`

	fmt.Print(buildinfo.Long(), "\n")
	fmt.Fprintf(flag.CommandLine.Output(), msg, os.Args[0])
	flag.PrintDefaults()
}
