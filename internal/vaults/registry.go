package vaults

import (
	"context"
	"fmt"
	"slices"
	"strings"
)

const (
	ANDOTP  VaultType = "andotp"
	AEGIS   VaultType = "aegis"
	TWOFAS  VaultType = "twofas"
	STRATUM VaultType = "stratum"
	KEEPASS VaultType = "keepass"
	PROTON  VaultType = "proton"
)

type (
	OpenFn    func(path string, pass []byte) (Vault, error)
	VaultType string
)

func (vt VaultType) String() string { return string(vt) }

var registry = make(map[VaultType]OpenFn)

// Register registers a vault type with the given open function.
func Register(vt VaultType, fn OpenFn) { registry[vt] = fn }

// Open opens a vault of the given type at the given path using the given password.
func Open(ctx context.Context, path string, pass []byte, vt VaultType) (Vault, error) {
	open, ok := registry[vt]
	if !ok {
		return nil, fmt.Errorf("vault type %q: not implemented", vt)
	}

	type data struct {
		v   Vault
		err error
	}

	done := make(chan data, 1)
	go func() {
		defer func() {
			for i := range pass {
				pass[i] = 0
			}
		}()
		v, err := open(path, pass)
		done <- data{v, err}
	}()

	select {
	case r := <-done:
		return r.v, r.err
	case <-ctx.Done():
		return nil, fmt.Errorf("open: operation timed out. wrong type?")
	}
}

// Types returns a slice of all registered vault types.
func Types() []VaultType {
	var s []VaultType
	for t := range registry {
		s = append(s, t)
	}
	slices.Sort(s)
	return s
}

// StrTypes returns a comma-separated string of all registered vault types.
func StrTypes() string {
	var s []string
	for _, vt := range Types() {
		s = append(s, vt.String())
	}
	return strings.Join(s, ", ")
}
