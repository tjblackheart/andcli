package main

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"

	"golang.org/x/crypto/scrypt"
)

type (
	aegisVault struct {
		Version int
		Header  struct {
			Slots []struct {
				Type, N, R, P   int
				UUID, Key, Salt string
				Repaired        bool
				KeyParams       struct{ Nonce, Tag string } `json:"key_params"`
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

//

func decryptAEGIS(data, password []byte) ([]entry, error) {

	var vault aegisVault
	if err := json.Unmarshal(data, &vault); err != nil {
		return nil, err
	}

	key, err := deriveAegisMasterKey(&vault, password)
	if err != nil {
		return nil, err
	}

	plain, err := decryptAegisDB(&vault, key)
	if err != nil {
		return nil, err
	}

	var db aegisDB
	if err := json.Unmarshal(plain, &db); err != nil {
		return nil, err
	}

	var entries []entry
	for _, e := range db.Entries {
		entries = append(entries, *e.toEntry())
	}

	return entries, nil
}

func deriveAegisMasterKey(v *aegisVault, password []byte) ([]byte, error) {

	var salt, keyNonce, keyTag, key, derivedKey []byte
	var err error

	for _, s := range v.Header.Slots {
		if s.Type == 1 {
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
				continue
			}
		}
	}

	if derivedKey == nil {
		return nil, errors.New("could not derive master key from password")
	}

	block, err := aes.NewCipher(derivedKey)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	var c []byte
	c = append(c, key...)
	c = append(c, keyTag...)

	masterKey, err := gcm.Open(nil, keyNonce, c, nil)
	if err != nil {
		return nil, err
	}

	return masterKey, nil
}

func decryptAegisDB(v *aegisVault, key []byte) ([]byte, error) {

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

	b, err := base64.StdEncoding.DecodeString(v.DB)
	if err != nil {
		return nil, err
	}

	tag, err := hex.DecodeString(v.Header.Params.Tag)
	if err != nil {
		return nil, err
	}

	var c []byte
	c = append(c, b...)
	c = append(c, tag...)

	plain, err := gcm.Open(nil, nonce, c, nil)
	if err != nil {
		return nil, err
	}

	return plain, nil
}
