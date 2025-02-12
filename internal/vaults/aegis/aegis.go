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

	"github.com/tjblackheart/andcli/internal/vaults"
	"golang.org/x/crypto/scrypt"
)

type (
	vault struct {
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
		db db `json:"-"`
	}

	db struct {
		Version int
		Entries []entry
	}

	entry struct {
		Type, UUID, Name   string
		Issuer, Note, Icon string
		IconMime           string `json:"icon_mime"`
		Info               struct {
			Secret, Algo   string
			Digits, Period int
		}
	}
)

func Open(filename string, pass []byte) (vaults.Vault, error) {

	var t = vaults.TYPE_AEGIS
	var v vault

	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", t, err)
	}

	if err := json.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("%s: %w", t, err)
	}

	key, err := v.masterKeyFromPass(pass)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", t, err)
	}

	b, err = v.decryptDB(key)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", t, err)
	}

	if err := json.Unmarshal(b, &v.db); err != nil {
		return nil, fmt.Errorf("%s: %w", t, err)
	}

	return v, nil
}

func (v vault) Entries() []vaults.Entry {

	entries := make([]vaults.Entry, 0)

	for _, e := range v.db.Entries {
		entries = append(entries, vaults.Entry{
			Secret:    e.Info.Secret,
			Issuer:    e.Issuer,
			Label:     e.Name,
			Digits:    e.Info.Digits,
			Type:      e.Type,
			Algorithm: e.Info.Algo,
			Period:    e.Info.Period,
		})
	}

	return entries
}

func (v vault) masterKeyFromPass(password []byte) ([]byte, error) {
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

	var cipher []byte
	cipher = append(cipher, key...)
	cipher = append(cipher, keyTag...)

	masterKey, err := gcm.Open(nil, keyNonce, cipher, nil)
	if err != nil {
		return nil, err
	}

	return masterKey, nil
}

func (v vault) decryptDB(key []byte) ([]byte, error) {

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

	var cipher []byte
	cipher = append(cipher, db...)
	cipher = append(cipher, tag...)

	plain, err := gcm.Open(nil, nonce, cipher, nil)
	if err != nil {
		return nil, err
	}

	return plain, nil
}
