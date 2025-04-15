package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	flag "github.com/spf13/pflag"
	"github.com/tjblackheart/andcli/v2/internal/buildinfo"
	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

var (
	availableVaults = strings.Join(vaults.Types(), ", ")

	vfile   = flag.StringP("file", "f", "", "Path to the encrypted vault (deprecated: Pass the filename directly)")
	vtype   = flag.StringP("type", "t", "", fmt.Sprintf("Vault type (%s)", availableVaults))
	cmd     = flag.StringP("clipboard-cmd", "c", "", "A custom clipboard command, including args (xclip, wl-copy, pbcopy etc.)")
	version = flag.BoolP("version", "v", false, "Prints version info and exits")
	help    = flag.BoolP("help", "h", false, "Show this help")
)

// Parses given flags into the existing config.
func (cfg *Config) parseFlags() error {

	flag.CommandLine.SortFlags = false
	flag.Usage = usage
	flag.Parse()

	if *version {
		fmt.Println(buildinfo.Long())
		os.Exit(0)
	}

	if *help {
		usage()
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
