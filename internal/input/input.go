package input

import (
	"bufio"
	"errors"
	"fmt"
	"os"

	"golang.org/x/term"
)

// Prompts a question with hidden input. For testing, a default answer
// can be provided and will be returned as is (only the first list item)
func Hidden(question string, defaults ...[]byte) ([]byte, error) {
	defer fmt.Println()

	fd := int(os.Stdin.Fd())

	// atm this is used for testing only, hence the IsTerm check.
	if len(defaults) > 0 && !term.IsTerminal(fd) {
		return defaults[0], nil
	}

	fmt.Printf("%s", question)

	return term.ReadPassword(fd)
}

// Returns piped input
func Stdin() ([]byte, error) {

	fi, err := os.Stdin.Stat()
	if err != nil {
		return nil, err
	}

	if (fi.Mode() & os.ModeCharDevice) != 0 {
		return nil, errors.New("stdin: no input provided")
	}

	s := bufio.NewScanner(bufio.NewReader(os.Stdin))
	if s.Scan(); s.Err() != nil {
		return nil, s.Err()
	}

	return s.Bytes(), nil
}
