package stratum

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/tjblackheart/andcli/v2/internal/vaults"
	"golang.org/x/crypto/argon2"
)

const vaultType = vaults.STRATUM

var _ vaults.Vault = &stratum{}

const (
	KEY_LENGTH = 32

	// Default
	HEADER      = "AUTHENTICATORPRO"
	SALT_LENGTH = 16
	IV_LENGTH   = 12
	THREADS     = 4
	ITERATIONS  = 3
	MEM_SIZE    = 65536

	// Legacy
	LEGACY_HEADER      = "AuthenticatorPro"
	LEGACY_HASH_MODE   = "sha1"
	LEGACY_ITERATIONS  = 64000
	LEGACY_SALT_LENGTH = 20
	LEGACY_IV_LENGTH   = 16
)

type (
	stratum struct {
		AuthenticatorCategories []any // ignored as of now
		Authenticators          []entry
	}

	entry struct {
		Algorithm uint8
		CopyCount int
		Counter   int
		Digits    uint8
		Icon      string
		Issuer    string
		Period    int
		Pin       string
		Ranking   int
		Secret    string
		Type      uint8
		Username  string
	}
)

func Open(filename string, pass []byte) (vaults.Vault, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	v := &stratum{Authenticators: make([]entry, 0)}

	switch string(b[:len(HEADER)]) {
	case HEADER:
		b, err := v.decrypt(b, pass)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", vaultType, err)
		}

		if err := json.Unmarshal(b, &v); err != nil {
			return nil, fmt.Errorf("%s: %w", vaultType, err)
		}

	case LEGACY_HEADER:
		b, err := v.decryptLegacy(b, pass)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", vaultType, err)
		}

		if err := json.Unmarshal(b, &v); err != nil {
			return nil, fmt.Errorf("%s: %w", vaultType, err)
		}
	}

	return v, nil
}

func (v stratum) Entries() []vaults.Entry {
	// https://github.com/stratumauth/app/blob/master/doc/BACKUP_FORMAT.md
	// Algorithm (applies to HOTP and TOTP): 0 = SHA-1, 1 = SHA-256, 2 = SHA-512
	// Type: 1 = HOTP, 2 = TOTP, 3 = Mobile-Otp, 4 = Steam, 5 = Yandex

	list := make([]vaults.Entry, 0)
	for _, e := range v.Authenticators {

		alg := "SHA1"
		switch e.Algorithm {
		case 1:
			alg = "SHA256"
		case 2:
			alg = "SHA512"
		}

		entry := vaults.Entry{
			Secret:    e.Secret,
			Issuer:    e.Issuer,
			Digits:    int(e.Digits),
			Type:      e.typeToString(),
			Algorithm: alg,
			Period:    e.Period,
			Label:     e.Username,
		}

		if err := entry.SanitizeAndValidate(); err == nil {
			list = append(list, entry)
		}
	}

	return list
}

func (v stratum) decrypt(b, pass []byte) ([]byte, error) {
	salt := b[len(HEADER) : len(HEADER)+SALT_LENGTH]
	nonce := b[len(HEADER)+SALT_LENGTH : len(HEADER)+SALT_LENGTH+IV_LENGTH]
	payload := b[len(HEADER)+SALT_LENGTH+IV_LENGTH:]
	key := argon2.IDKey(pass, salt, ITERATIONS, MEM_SIZE, THREADS, KEY_LENGTH)

	cb, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(cb)
	if err != nil {
		return nil, err
	}

	return gcm.Open(nil, nonce, payload, nil)
}

func (v stratum) decryptLegacy(_, _ []byte) ([]byte, error) {
	return nil, errors.New("decryption of stratum legacy vaults is not implemented")
}

func (e entry) typeToString() string {
	s := "unknown"

	switch e.Type {
	case 1:
		s = "hotp"
	case 2:
		s = "totp"
	case 3:
		s = "mobile"
	case 4:
		s = "steam"
	case 5:
		s = "yandex"
	}

	return strings.ToUpper(s)
}
