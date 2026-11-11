package clipboard

import (
	"bytes"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

type (
	Clipboard struct {
		cmd  string
		args []string
	}

	sysUtil struct {
		cmd  string
		args []string
	}
)

// New inits a new Clipboard instance with a given comand string.
// If nothing is provided, it falls back to available system tools.
func New(s string) *Clipboard {
	cb := &Clipboard{cmd: "", args: make([]string, 0)}
	if s != "" {
		return cb.initUser(s)
	}
	return cb.initSystem()
}

// Set writes b to the selected clipboard.
func (cb Clipboard) Set(b []byte) error {
	cmd := exec.Command(cb.cmd, cb.args...)
	cmd.Stdin = bytes.NewReader(b)
	return cmd.Run()
}

// Checks if a command is povided and the clipboard is usable
func (cb Clipboard) IsInitialized() bool {
	return cb.cmd != ""
}

// Inits the clipboard with the given user values.
// path validation is done in config already.
func (cb *Clipboard) initUser(s string) *Clipboard {
	parts := strings.Fields(s)
	if len(parts) == 0 {
		return cb
	}

	cb.cmd = parts[0]
	cb.args = parts[1:]

	return cb
}

// Inits the clipboard with the first occurrence found of the defined system tools.
func (cb *Clipboard) initSystem() *Clipboard {
	lookPath := func(name string) (string, bool) {
		path, err := exec.LookPath(name)
		return path, err == nil
	}

	cb.cmd, cb.args = detect(runtime.GOOS, os.Getenv, lookPath)
	return cb
}

// detect resolves the first available system clipboard command.
func detect(goos string, getenv func(string) string, lookPath func(string) (string, bool)) (cmd string, args []string) {
	cmds := make([]sysUtil, 0, 4)

	switch goos {
	case "linux":
		if getenv("TERMUX_VERSION") != "" {
			cmds = append(cmds, sysUtil{cmd: "termux-clipboard-set"})
		}

		isWayland := getenv("WAYLAND_DISPLAY") != ""
		if isWayland {
			cmds = append(cmds, sysUtil{cmd: "wl-copy"})
		}

		isX11 := getenv("DISPLAY") != ""
		if isX11 {
			cmds = append(cmds,
				sysUtil{cmd: "xclip", args: []string{"-selection", "clipboard"}},
				sysUtil{cmd: "xsel", args: []string{"-b"}},
			)
		}

		if !isWayland && !isX11 {
			cmds = append(cmds,
				sysUtil{cmd: "wl-copy"},
				sysUtil{cmd: "xclip", args: []string{"-selection", "clipboard"}},
				sysUtil{cmd: "xsel", args: []string{"-b"}},
			)
		}

	case "darwin":
		cmds = append(cmds, sysUtil{cmd: "pbcopy"})
	case "windows":
		cmds = append(cmds, sysUtil{cmd: "clip.exe"})
	case "android":
		cmds = append(cmds, sysUtil{cmd: "termux-clipboard-set"})
	}

	for _, v := range cmds {
		if path, ok := lookPath(v.cmd); ok {
			return path, v.args
		}
	}

	return "", []string{}
}

// Return a formatted string built from the current command, including args.
func (cb Clipboard) String() string {
	args := strings.TrimSpace(strings.Join(cb.args, " "))
	return strings.TrimSpace(strings.Join([]string{cb.cmd, args}, " "))
}
