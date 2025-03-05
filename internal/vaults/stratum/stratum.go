package stratum

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/tjblackheart/andcli/v2/internal/vaults"
	"golang.org/x/crypto/argon2"
)

// force interface impl
var _ vaults.Vault = &vault{}

const (
	stKeyLength    = 32
	stHeader       = "AUTHENTICATORPRO"
	stLegacyHeader = "AuthenticatorPro"
	stSaltLength   = 16
	stIVLength     = 12
	stParallel     = 4
	stIterations   = 3
	stMemSize      = 65536
)

type (
	vault struct {
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

	v := &vault{Authenticators: make([]entry, 0)}
	t := vaults.TYPE_STRATUM

	switch string(b[:len(stHeader)]) {
	case stHeader:
		b, err := v.decrypt(b, pass)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", t, err)
		}

		if err := json.Unmarshal(b, &v); err != nil {
			return nil, fmt.Errorf("%s: %w", t, err)
		}

	case stLegacyHeader:
		b, err := v.decryptLegacy(b, pass)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", t, err)
		}

		if err := json.Unmarshal(b, &v); err != nil {
			return nil, fmt.Errorf("%s: %w", t, err)
		}
	}

	return v, nil
}

func (v vault) Entries() []vaults.Entry {

	// https://github.com/stratumauth/app/blob/master/doc/BACKUP_FORMAT.md
	// Algorithm (applies to HOTP and TOTP): 0 = SHA-1, 1 = SHA-256, 2 = SHA-512
	// Type: 1 = HOTP, 2 = TOTP, 3 = Mobile-Otp, 4 = Steam, 5 = Yandex

	list := make([]vaults.Entry, 0)
	for _, e := range v.Authenticators {
		// TODO: ignore everything but TOTP
		if e.Type != 2 {
			continue
		}

		alg := "sha1"
		switch e.Algorithm {
		case 1:
			alg = "sha256"
		case 2:
			alg = "sha512"
		}

		list = append(list, vaults.Entry{
			Secret:    e.Secret,
			Issuer:    e.Issuer,
			Digits:    int(e.Digits),
			Type:      "TOTP",
			Algorithm: alg,
			Period:    e.Period,
			Label:     e.Username,
		})
	}

	return list
}

func (v vault) decrypt(b, pass []byte) ([]byte, error) {

	salt := b[len(stHeader) : len(stHeader)+stSaltLength]
	nonce := b[len(stHeader)+stSaltLength : len(stHeader)+stSaltLength+stIVLength]
	payload := b[len(stHeader)+stSaltLength+stIVLength:]
	key := argon2.IDKey(pass, salt, stIterations, stMemSize, stParallel, stKeyLength)

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

func (v vault) decryptLegacy(_, _ []byte) ([]byte, error) {
	return nil, errors.New("legacy decryption support is not implemented")
}
