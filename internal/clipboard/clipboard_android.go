// why, yes, I *do* use this on android. all hail termux. :]
//
// the reason for this split is the "clipboard" package,
// which depends on x/mobile, which needs some weird CGO stuff to build

package clipboard

import (
	"fmt"
	"log"
	"os/exec"
	"strings"
)

type CB struct {
	cmd  string
	args []string
}

func New(cfgCmd string) *CB {
	cb := &CB{cmd: "", args: make([]string, 0)}
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
	}

	return cb
}

func (cb CB) Set(b []byte) error {

	// just ignore this.
	if !cb.IsInitialized() {
		return nil
	}

	cmd := fmt.Sprintf("%s %s", cb.cmd, strings.Join(cb.args, " "))
	cmd = fmt.Sprintf("echo -n %s | %s", string(b), cmd)

	return exec.Command("sh", "-c", cmd).Run()

}

func (cb CB) IsInitialized() bool {
	return cb.cmd != ""
}
