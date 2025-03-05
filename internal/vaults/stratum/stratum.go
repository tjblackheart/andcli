package stratum

// https://github.com/stratumauth/app/blob/master/doc/BACKUP_FORMAT.md
// https://github.com/stratumauth/app/blob/master/extra/decrypt_backup.py

import (
	"errors"

	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

// force interface impl
var _ vaults.Vault = &vault{}

const (
	KEY_LENGTH = 32

	// Default
	HEADER      = "AUTHENTICATORPRO"
	SALT_LENGTH = 16
	IV_LENGTH   = 12
	PARALLELISM = 4
	ITERATIONS  = 3
	MEMORY_SIZE = 65536

	// Legacy
	LEGACY_HEADER      = "AuthenticatorPro"
	LEGACY_HASH_MODE   = "sha1"
	LEGACY_ITERATIONS  = 64000
	LEGACY_SALT_LENGTH = 20
	LEGACY_IV_LENGTH   = 16
)

type vault struct {
}

func Open(filename string, pass []byte) (vaults.Vault, error) {

	return nil, errors.New("not implemented")
}

func (v vault) Entries() []vaults.Entry {
	list := make([]vaults.Entry, 0)

	return list
}

func (v vault) decrypt() {

}

func (v vault) decryptLegacy() {

}
