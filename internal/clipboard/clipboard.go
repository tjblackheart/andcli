package clipboard

import (
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

var utils = map[string][]sysUtil{
	"windows": {{cmd: "clip.exe"}},
	"darwin":  {{cmd: "pbcopy"}},
	"linux": {
		{cmd: "xclip", args: []string{"-selection", "clipboard"}},
		{cmd: "xsel", args: []string{"-b"}},
		{cmd: "wl-copy"},
	},
	"android": {{cmd: "termux-clipboard-set"}},
}

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

	pipe, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := pipe.Write(b); err != nil {
		return err
	}

	if err := pipe.Close(); err != nil {
		return err
	}

	return cmd.Wait()
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
func (cb *Clipboard) initSystem() *Clipboard {

	utils, ok := utils[runtime.GOOS]
	if !ok {
		return cb
	}

	for _, v := range utils {
		if path, err := exec.LookPath(v.cmd); err == nil {
			cb.cmd = path
			cb.args = v.args
			break
		}
	}

	return cb
}

// Return a formatted string built from the current command, including args.
func (cb Clipboard) String() string {
	args := strings.TrimSpace(strings.Join(cb.args, " "))
	return strings.TrimSpace(strings.Join([]string{cb.cmd, args}, " "))
}
