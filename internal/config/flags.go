package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/tjblackheart/andcli/internal/buildinfo"
	"github.com/tjblackheart/andcli/internal/vaults"
)

// Parses given flags into the existing config.
func (cfg *Config) parseFlags() error {
	var (
		vfile, vtype, cmd string
		version           bool
		types             string
	)

	types = strings.Join(vaults.Types(), ", ")

	flag.Usage = usage
	flag.StringVar(&vfile, "f", "", "Path to the encrypted vault (deprecated: Pass the filename directly)")
	flag.StringVar(&vtype, "t", "", fmt.Sprintf("Vault type (%s)", types))
	flag.StringVar(&cmd, "c", "", "Clipboard command (xclip, wl-copy, pbcopy etc.)")
	flag.BoolVar(&version, "v", false, "Prints version info and exits")
	flag.Parse()

	if version {
		usage()
		os.Exit(0)
	}

	if vfile != "" {
		cfg.File = vfile
	}

	if vtype != "" {
		cfg.Type = vtype
	}

	if cmd != "" {
		cfg.ClipboardCmd = cmd
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
