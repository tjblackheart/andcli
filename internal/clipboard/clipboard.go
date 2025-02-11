//go:build !android

package clipboard

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
	"strings"

	"golang.design/x/clipboard"
)

type Clipboard struct {
	cmd  string
	args []string
}

var sysUtils = []string{"xclip", "xsel", "wl-copy", "pbcopy"}

// New inits a new Clipboard instance with a given comand string.
// If nothing is provided, it falls back to either system tools
// or, if that also fails, uses a even more generic solution as last resort.
func New(s string) *Clipboard {
	cb := &Clipboard{cmd: "", args: make([]string, 0)}
	if s != "" {
		return cb.initUser(s)
	}
	return cb.initSystem()
}

// Set writes b to the selected clipboard.
func (cb Clipboard) Set(b []byte) error {
	if cb.cmd == "clipboard" {
		clipboard.Write(clipboard.FmtText, b)
		if !bytes.Equal(clipboard.Read(clipboard.FmtText), b) {
			return errors.New("failed to set clipboard content")
		}
		return nil
	}

	pipe := fmt.Sprintf("echo -n %s | %s", string(b), cb.String())

	return exec.Command("sh", "-c", pipe).Run()
}

// Checks if a command is povided and the clipboard is usable
func (cb Clipboard) IsInitialized() bool {
	return cb.cmd != ""
}

// Inits the clipboard with the given user values.
// path validation is done in config already.
func (cb *Clipboard) initUser(s string) *Clipboard {
	parts := strings.SplitN(s, " ", 2)
	if parts[0] != "" {
		cb.cmd = parts[0]
		if len(parts) > 1 {
			cb.args = strings.Split(parts[1], " ")
		}
	}

	return cb
}

// Inits the clipboard with the first occurence found of the defined system tools.
// If none of these are available, use a generic solution.
func (cb *Clipboard) initSystem() *Clipboard {

	for _, v := range sysUtils {
		if path, err := exec.LookPath(v); err == nil {
			cb.cmd = path
			if v == "xclip" {
				cb.args = append(cb.args, "-selection", "clipboard")
			}

			if v == "xsel" {
				cb.args = append(cb.args, "-b")
			}

			return cb
		}
	}

	if err := clipboard.Init(); err == nil {
		cb.cmd = "clipboard"
	}

	return cb
}

// Return a formatted string built from the current command, including args.
func (cb Clipboard) String() string {
	args := strings.TrimSpace(strings.Join(cb.args, " "))
	return strings.TrimSpace(strings.Join([]string{cb.cmd, args}, " "))
}
