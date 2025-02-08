package config

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/tjblackheart/andcli/internal/buildinfo"
	"github.com/tjblackheart/andcli/internal/vaults"
)

var (
	types = strings.Join(vaults.Types(), ", ")

	vfile   = flag.String("f", "", "Path to the encrypted vault (deprecated: Pass the filename directly)")
	vtype   = flag.String("t", "", fmt.Sprintf("Vault type (%s)", types))
	cmd     = flag.String("c", "", "Clipboard command (xclip, wl-copy, pbcopy etc.)")
	version = flag.Bool("v", false, "Prints version info and exits")
)

// Parses given flags into the existing config.
func (cfg *Config) parseFlags() error {

	flag.Usage = usage
	flag.Parse()

	if *version {
		usage()
		os.Exit(0)
	}

	f := trim(*vfile)
	if f != "" {
		abs, err := filepath.Abs(f)
		if err != nil {
			return err
		}
		cfg.File = abs
	}

	t := trim(*vtype)
	if t != "" {
		cfg.Type = t
	}

	c := trim(*cmd)
	if c != "" {
		cfg.ClipboardCmd = c
	}

	if flag.Arg(0) != "" {
		cfg.File = trim(flag.Arg(0))
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

func trim(s string) string {
	return strings.ToLower(strings.TrimSpace(s))
}
