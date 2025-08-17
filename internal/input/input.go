package input

import (
	"fmt"
	"os"

	"golang.org/x/term"
)

// Prompts a question with hidden input. For testing, a default answer
// can be provided and will be returned as is (only the first list item)
func AskHidden(question string, defaults ...[]byte) ([]byte, error) {
	fd := int(os.Stdin.Fd())

	// atm this is used for testing only, hence the IsTerm check.
	if len(defaults) > 0 && !term.IsTerminal(fd) {
		return defaults[0], nil
	}

	fmt.Printf("%s", question)
	password, err := term.ReadPassword(fd)
	fmt.Println()
	return password, err
}
