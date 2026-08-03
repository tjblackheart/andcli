package vaults

import (
	"errors"
)

// Vault is the basic skeleton of a vault implementation.
type Vault interface {
	Entries() []Entry
	IsPlain([]byte) bool
}

var ErrIsPlain error = errors.New("plaintext vaults are unsupported")
