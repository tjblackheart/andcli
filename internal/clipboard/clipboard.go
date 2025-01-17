//go:build !android

package clipboard

import (
	"bytes"
	"errors"
	"fmt"
	"log"
	"os/exec"
	"strings"

	"golang.design/x/clipboard"
)

type CB struct {
	cmd  string
	args []string
}

func New(cfgCmd string) *CB {
	cb := &CB{cmd: "", args: make([]string, 0)}

	// use any given user values first and validate system paths
	parts := strings.SplitN(cfgCmd, " ", 2)
	if parts[0] != "" {
		path, err := exec.LookPath(parts[0])
		if err != nil {
			log.Fatalf("Error: Configured clipboard command not found in $PATH: %s", parts[0])
		}
		cb.cmd = path

		if len(parts) > 1 {
			cb.args = strings.Split(parts[1], " ")
		}

		return cb
	}

	// if no options where given, use the first available system util found.
	// xorg (x2), wayland, macos
	sysUtils := []string{"xclip", "xsel", "wl-copy", "pbcopy"}
	for _, v := range sysUtils {
		if path, err := exec.LookPath(v); err == nil {
			args := make([]string, 0)
			if v == "xclip" {
				args = append(args, "-selection", "clipboard") // force xclip copy to system clipboard.
			}

			if v == "xsel" {
				args = append(args, "--input", "--clipboard") // force xsel copy to system clipboard.
			}

			cb.cmd = path
			cb.args = args

			return cb
		}
	}

	// if nothing matched, try to use a more generic solution.
	if err := clipboard.Init(); err == nil {
		cb.cmd = "clipboard"
	}

	return cb
}

func (cb CB) Set(b []byte) error {

	// just ignore this.
	if !cb.IsInitialized() {
		return nil
	}

	switch cb.cmd {
	case "clipboard":
		clipboard.Write(clipboard.FmtText, b)
		if !bytes.Equal(clipboard.Read(clipboard.FmtText), b) {
			return errors.New("")
		}
	default:
		cmd := fmt.Sprintf("%s %s", cb.cmd, strings.Join(cb.args, " "))
		cmd = fmt.Sprintf("echo -n %s | %s", string(b), cmd)
		if err := exec.Command("sh", "-c", cmd).Run(); err != nil {
			return err
		}
	}

	return nil
}

func (cb CB) IsInitialized() bool {
	return cb.cmd != ""
}
