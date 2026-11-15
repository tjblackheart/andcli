package aegis

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/tjblackheart/andcli/v2/internal/vaults"
	"golang.org/x/crypto/scrypt"
)

const (
	vaultType = vaults.AEGIS

	maxScryptN   = 1 << 20
	maxScryptR   = 32
	maxScryptP   = 16
	maxScryptMem = 256 << 20
)

var _ vaults.Vault = &aegis{}

type (
	aegis struct {
		Version int
		Header  struct {
			Slots []struct {
				Type, N, R, P   int
				UUID, Key, Salt string
				Repaired        bool
				KeyParams       struct {
					Nonce, Tag string
				} `json:"key_params"`
			}
			Params struct {
				Nonce, Tag string
			}
		}
		DB string
		//
		db db
	}

	db struct {
		Version int
		Entries []entry
	}

	entry struct {
		Type, UUID, Name   string
		Issuer, Note, Icon string
		IconMime           string `json:"icon_mime"`
		Info               info
	}

	info struct {
		Secret, Algo   string
		Digits, Period int
	}
)

func init() { vaults.Register(vaultType, Open) }

func Open(filename string, pass []byte) (vaults.Vault, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", vaultType, err)
	}
	return open(b, pass)
}

func open(b, pass []byte) (vaults.Vault, error) {
	var v aegis

	if v.IsPlain(b) {
		return nil, vaults.ErrIsPlain
	}

	if err := json.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("%s: %w", vaultType, err)
	}

	key, err := v.masterKeyFromPass(pass)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", vaultType, err)
	}

	b, err = v.decryptDB(key)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", vaultType, err)
	}

	if err := json.Unmarshal(b, &v.db); err != nil {
		return nil, fmt.Errorf("%s: %w", vaultType, err)
	}

	return &v, nil
}

func (v aegis) Entries() []vaults.Entry {
	entries := make([]vaults.Entry, 0)

	for _, e := range v.db.Entries {
		entry := vaults.Entry{
			Secret:    e.Info.Secret,
			Issuer:    e.Issuer,
			Label:     e.Name,
			Digits:    e.Info.Digits,
			Type:      strings.ToUpper(e.Type),
			Algorithm: e.Info.Algo,
			Period:    e.Info.Period,
		}

		if err := entry.SanitizeAndValidate(); err == nil {
			entries = append(entries, entry)
		}
	}

	return entries
}

func (v aegis) IsPlain(b []byte) bool {
	var db struct {
		DB json.RawMessage `json:"db"`
	}

	json.Unmarshal(b, &db) // intentional skip err check
	if len(db.DB) > 0 && db.DB[0] != '"' {
		return true
	}
	return false
}

func (v aegis) masterKeyFromPass(password []byte) ([]byte, error) {
	var salt, keyNonce, keyTag, key, derivedKey []byte
	var err error

	for _, s := range v.Header.Slots {
		if s.Type != 1 {
			continue
		}

		if salt, err = hex.DecodeString(s.Salt); err != nil {
			return nil, err
		}

		if keyNonce, err = hex.DecodeString(s.KeyParams.Nonce); err != nil {
			return nil, err
		}

		if keyTag, err = hex.DecodeString(s.KeyParams.Tag); err != nil {
			return nil, err
		}

		if key, err = hex.DecodeString(s.Key); err != nil {
			return nil, err
		}

		if err := validateScryptParams(s.N, s.R, s.P); err != nil {
			return nil, fmt.Errorf("%s: %w", vaultType, err)
		}

		derivedKey, err = scrypt.Key(password, salt, s.N, s.R, s.P, 32)
		if err != nil {
			return nil, err
		}
	}

	if derivedKey == nil {
		return nil, errors.New("failed to derive master key: missing slot type 1")
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	if len(keyNonce) != gcm.NonceSize() {
		return nil, fmt.Errorf("invalid key nonce length %d, want %d", len(keyNonce), gcm.NonceSize())
	}

	var b []byte
	b = append(b, key...)
	b = append(b, keyTag...)

	masterKey, err := gcm.Open(nil, keyNonce, b, nil)
	if err != nil {
		return nil, err
	}

	return masterKey, nil
}

func (v aegis) decryptDB(key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce, err := hex.DecodeString(v.Header.Params.Nonce)
	if err != nil {
		return nil, err
	}

	db, err := base64.StdEncoding.DecodeString(v.DB)
	if err != nil {
		return nil, err
	}

	tag, err := hex.DecodeString(v.Header.Params.Tag)
	if err != nil {
		return nil, err
	}

	if len(nonce) != gcm.NonceSize() {
		return nil, fmt.Errorf("invalid nonce length %d, want %d", len(nonce), gcm.NonceSize())
	}

	var b []byte
	b = append(b, db...)
	b = append(b, tag...)

	plain, err := gcm.Open(nil, nonce, b, nil)
	if err != nil {
		return nil, err
	}

	return plain, nil
}

func validateScryptParams(n, r, p int) error {
	if n < 2 || n&(n-1) != 0 {
		return fmt.Errorf("invalid scrypt N %d: must be a power of two >= 2", n)
	}
	if n > maxScryptN {
		return fmt.Errorf("scrypt N too large: %d, max %d", n, maxScryptN)
	}
	if r < 1 || r > maxScryptR {
		return fmt.Errorf("invalid scrypt R %d: must be 1..%d", r, maxScryptR)
	}
	if p < 1 || p > maxScryptP {
		return fmt.Errorf("invalid scrypt P %d: must be 1..%d", p, maxScryptP)
	}
	if mem := int64(128) * int64(n) * int64(r); mem > maxScryptMem {
		return fmt.Errorf("scrypt parameters too large: N=%d R=%d need %d bytes, max %d", n, r, mem, maxScryptMem)
	}
	return nil
}
