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

func New(s string) *Clipboard {
	cb := &Clipboard{cmd: "", args: make([]string, 0)}
	if s != "" {
		return cb.initUser(s)
	}
	return cb.initSystem()
}

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

// Inits the clipboard with the first occurence found of either
// "xclip", "xsel", "wl-copy" or "pbcopy".
// if none of these are available, use a generic solution.
func (cb *Clipboard) initSystem() *Clipboard {

	system := []string{"xclip", "xsel", "wl-copy", "pbcopy"}

	for _, v := range system {
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

func (cb Clipboard) String() string {
	return fmt.Sprintf("%s %s", cb.cmd, strings.Join(cb.args, " "))
}
