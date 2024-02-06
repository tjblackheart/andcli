package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"

	"golang.org/x/crypto/pbkdf2"
)

const numFields int = 3
const authTagLength int = 16

type (
	twofasVault struct {
		UpdatedAt         int
		SchemaVersion     int
		AppVersionCode    int
		AppVersionName    string
		AppOrigin         string
		ServicesEncrypted string
	}

	twofasDB []twofasEntry

	twofasEntry struct {
		Name      string
		Secret    string
		UpdatedAt int
		Otp       struct {
			Label     string
			Account   string
			Issuer    string
			Digits    int
			Period    int
			Algorithm string
			TokenType string `json:"tokenType"`
			Source    string
		}
		Order struct {
			Position int
		}
		Icon struct {
			Selected string
			Label    struct {
				Text            string
				BackgroundColor string `json:"backgroundColor"`
			}
			IconCollection struct {
				Id string
			} `json:"iconCollection"`
		}
	}
)

func (e twofasEntry) toEntry() *entry {
	return &entry{
		Secret:    e.Secret,
		Issuer:    e.Otp.Issuer,
		Label:     e.Otp.Label,
		Digits:    e.Otp.Digits,
		Type:      e.Otp.TokenType,
		Algorithm: e.Otp.Algorithm,
		Period:    e.Otp.Period,
	}
}

//

func decryptTWOFAS(data, password []byte) (entries, error) {

	var vault twofasVault
	if err := json.Unmarshal(data, &vault); err != nil {
		return nil, err
	}

	key, err := deriveTwoFasMasterKey(&vault, password)
	if err != nil {
		return nil, err
	}

	plain, err := decryptTwoFasDB(&vault, key)
	if err != nil {
		return nil, err
	}

	var db twofasDB
	if err := json.Unmarshal(plain, &db); err != nil {
		return nil, err
	}

	var list entries
	for _, e := range db {
		list = append(list, *e.toEntry())
	}

	return list, nil
}

func deriveTwoFasMasterKey(v *twofasVault, password []byte) ([]byte, error) {
	servicesEncrypted := strings.SplitN(v.ServicesEncrypted, ":", numFields+1)
	if len(servicesEncrypted) != numFields {
		return nil, fmt.Errorf("Invalid vault file. Number of fields is not %d", numFields)
	}
	var dbAndAuthTag, salt []byte
	var err error

	dbAndAuthTag, err = base64.StdEncoding.DecodeString(servicesEncrypted[0])
	if err != nil {
		return nil, err
	}
	salt, err = base64.StdEncoding.DecodeString(servicesEncrypted[1])
	if err != nil {
		return nil, err
	}

	if len(dbAndAuthTag) <= authTagLength {
		return nil, fmt.Errorf("Invalid vault file. Length of cipher text with auth tag must be more than %d", authTagLength)
	}

	return pbkdf2.Key(password, salt, 10000, 32, sha256.New), nil
}

func decryptTwoFasDB(v *twofasVault, key []byte) ([]byte, error) {

	servicesEncrypted := strings.SplitN(v.ServicesEncrypted, ":", numFields+1)
	if len(servicesEncrypted) != numFields {
		return nil, fmt.Errorf("Invalid vault file. Number of fields is not %d", numFields)
	}

	var dbAndAuthTag, b, tag, nonce []byte
	var err error

	dbAndAuthTag, err = base64.StdEncoding.DecodeString(servicesEncrypted[0])
	if err != nil {
		return nil, err
	}

	b = dbAndAuthTag[:len(dbAndAuthTag)-authTagLength]
	tag = dbAndAuthTag[len(dbAndAuthTag)-authTagLength:]

	nonce, err = base64.StdEncoding.DecodeString(servicesEncrypted[2])
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
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
