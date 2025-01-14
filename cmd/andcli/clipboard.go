package main

import (
	"fmt"
	"log"
	"os/exec"
	"strings"

	"golang.design/x/clipboard"
)

func setupClipboard(cfgCmd string) string {
	cmd := ""

	// use any given user values first and validate system paths
	parts := strings.SplitN(cfgCmd, " ", 2)
	if parts[0] != "" {
		cmd, err := exec.LookPath(parts[0])
		if err != nil {
			log.Fatalf("%s Configured clipboard command not found in $PATH: %s", danger.Sprint("[ERR]"), parts[0])
		}

		if len(parts) > 1 {
			cmd = fmt.Sprintf("%s %s", cmd, parts[1])
		}

		return cmd
	}

	// if no options where given, use the first available system util found.
	// xorg (x2), wayland, macos
	sysUtils := []string{"xclip", "xsel", "wl-copy", "pbcopy"}
	for _, v := range sysUtils {
		if cmd, err := exec.LookPath(v); err == nil {
			args := ""

			if v == "xclip" {
				args = "-selection clipboard" // force xclip copy to system clipboard.
			}

			if v == "xsel" {
				args = "--input --clipboard" // force xsel copy to system clipboard.
			}

			return strings.Join([]string{cmd, args}, " ")
		}
	}

	// if nothing matched, try to use a more generic solution.
	if err := clipboard.Init(); err == nil {
		cmd = "clipboard"
	}

	return cmd
}
