// Why, yes, I *do* use this on android. All hail Termux! :]
//
// The reason for this split is the "clipboard" package: depends on x/mobile,
// which needs enabled CGO and the whole Android NDK, and thats a little bit overkill.

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
	return cb.initUser(s)
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
