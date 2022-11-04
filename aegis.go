package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"

	"golang.org/x/crypto/scrypt"
)

type (
	aegisVault struct {
		Version int
		Header  struct {
			Slots []struct {
				Type      int
				UUID, Key string
				KeyParams struct{ Nonce, Tag string } `json:"key_params"`
				N, R, P   int
				Salt      string
				Repaired  bool
			}
			Params struct{ Nonce, Tag string }
		}
		DB string
	}

	aegisDB struct {
		Version int
		Entries []aegisEntry
	}

	aegisEntry struct {
		Type, UUID, Name   string
		Issuer, Note, Icon string
		IconMime           string `json:"icon_mime"`
		Info               struct {
			Secret, Algo   string
			Digits, Period int
		}
	}
)

func (e aegisEntry) toEntry() *entry {
	return &entry{
		Secret:    e.Info.Secret,
		Issuer:    e.Issuer,
		Label:     e.Name,
		Digits:    e.Info.Digits,
		Type:      e.Type,
		Algorithm: e.Info.Algo,
		Thumbnail: e.Icon,
		Period:    e.Info.Period,
	}
}

func decryptAEGIS(data, password []byte) ([]entry, error) {
	var vault aegisVault
	if err := json.Unmarshal(data, &vault); err != nil {
		return nil, fmt.Errorf("aegis: %s", err)
	}

	// decrypt master key
	masterKey, err := deriveAegisMasterKey(&vault, password)
	if err != nil {
		return nil, fmt.Errorf("aegis: master key: %s", err)
	}

	// decrypt DB
	plain, err := decryptAegisDB(&vault, masterKey)
	if err != nil {
		return nil, fmt.Errorf("aegis: decrypt DB: %s", err)
	}

	// convert entries
	var db aegisDB
	if err := json.Unmarshal(plain, &db); err != nil {
		return nil, fmt.Errorf("aegis: unmarshal: %s", err)
	}

	var entries []entry
	for _, e := range db.Entries {
		entries = append(entries, *e.toEntry())
	}

	return entries, nil
}

func deriveAegisMasterKey(v *aegisVault, password []byte) ([]byte, error) {
	var n, r, p int
	var salt, keyNonce, keyTag, key, derivedKey []byte
	var err error

	for _, s := range v.Header.Slots {
		if s.Type == 1 {
			n, r, p = s.N, s.R, s.P

			salt, err = hex.DecodeString(s.Salt)
			if err != nil {
				return nil, fmt.Errorf("aegis: decode salt: %s", err)
			}

			keyNonce, err = hex.DecodeString(s.KeyParams.Nonce)
			if err != nil {
				return nil, fmt.Errorf("aegis: decode keyNonce: %s", err)
			}

			keyTag, err = hex.DecodeString(s.KeyParams.Tag)
			if err != nil {
				return nil, fmt.Errorf("aegis: decode keyTag: %s", err)
			}

			key, err = hex.DecodeString(s.Key)
			if err != nil {
				return nil, fmt.Errorf("aegis: decode key: %s", err)
			}

			derivedKey, err = scrypt.Key(password, salt, n, r, p, 32)
			if err != nil {
				continue
			}
		}
	}

	if derivedKey == nil {
		return nil, fmt.Errorf("aegis: could not derive master key from password")
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, fmt.Errorf("aegis: new cipher: %s", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aegis: new gcm: %s", err)
	}

	var enc []byte
	enc = append(enc, key...)
	enc = append(enc, keyTag...)

	masterKey, err := gcm.Open(nil, keyNonce, enc, nil)
	if err != nil {
		return nil, fmt.Errorf("aegis: decrypt master key: %s", err)
	}

	return masterKey, nil
}

func decryptAegisDB(v *aegisVault, masterKey []byte) ([]byte, error) {
	// decrypt DB with master key
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, fmt.Errorf("aegis: new cipher: %s", err)
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("aegis: new gcm: %s", err)
	}

	nonce, err := hex.DecodeString(v.Header.Params.Nonce)
	if err != nil {
		return nil, fmt.Errorf("aegis: decode nonce: %s", err)
	}

	b, err := base64.StdEncoding.DecodeString(v.DB)
	if err != nil {
		return nil, fmt.Errorf("aegis: decrypt db: %s", err)
	}

	tag, err := hex.DecodeString(v.Header.Params.Tag)
	if err != nil {
		return nil, fmt.Errorf("aegis: tag: %s", err)
	}

	var enc []byte
	enc = append(enc, b...)
	enc = append(enc, tag...)

	plain, err := gcm.Open(nil, nonce, enc, nil)
	if err != nil {
		return nil, fmt.Errorf("aegis: decrypt: %s", err)
	}

	return plain, nil
}
