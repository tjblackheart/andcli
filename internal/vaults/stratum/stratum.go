package stratum

// https://github.com/stratumauth/app/blob/master/extra/decrypt_backup.py

import (
	"errors"

	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

// force interface impl
var _ vaults.Vault = &vault{}

type vault struct {
}

func Open(filename string, pass []byte) (vaults.Vault, error) {
	return nil, errors.New("not implemented")
}

func (v vault) Entries() []vaults.Entry {
	list := make([]vaults.Entry, 0)

	return list
}
