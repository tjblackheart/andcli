// why, yes, I *do* use this on android. all hail termux. :]
//
// the reason for this split is the "clipboard" package,
// which depends on x/mobile, which needs some weird CGO stuff to build

package clipboard

import (
	"fmt"
	"os/exec"
	"strings"
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
	return cb
}

func (cb Clipboard) Set(b []byte) error {
	pipe := fmt.Sprintf("echo -n %s | %s", string(b), cb.String())
	return exec.Command("sh", "-c", pipe).Run()
}

func (cb Clipboard) IsInitialized() bool {
	return cb.cmd != ""
}

// inits the clipboard with the given user values.
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

func (cb Clipboard) String() string {
	return fmt.Sprintf("%s %s", cb.cmd, strings.Join(cb.args, " "))
}
