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

	// use any given user values first and validate any system paths
	parts := strings.SplitN(cfgCmd, " ", 2)
	if parts[0] != "" {
		path, err := exec.LookPath(parts[0])
		if err != nil {
			log.Fatalf("%s Configured clipboard command not found in $PATH: %s", danger.Sprint("[ERR]"), parts[0])
			return ""
		}

		cmd = parts[0]
		if len(parts) > 1 {
			cmd = fmt.Sprintf("%s %s", path, parts[1])
		}

		return cmd
	}

	// if no options where given, use the first available binary found.
	sysUtils := []string{"xclip", "wl-copy", "pbcopy"} // xorg, wayland, macos
	for _, v := range sysUtils {
		if path, err := exec.LookPath(v); err == nil {
			cmd = path
			if v == "xclip" {
				// force xclip copy to system clipboard.
				cmd = fmt.Sprintf("%s -selection clipboard", path)
			}
			return cmd
		}
	}

	// if nothing matched, use a generic go based solution.
	if err := clipboard.Init(); err == nil {
		cmd = "clipboard"
	}

	return cmd
}
