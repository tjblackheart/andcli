package twofas

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/tjblackheart/andcli/v2/internal/vaults"
	"golang.org/x/crypto/pbkdf2"
)

const (
	numFields     int = 3
	authTagLength int = 16
)

type (
	vault struct {
		UpdatedAt         int
		SchemaVersion     int
		AppVersionCode    int
		AppVersionName    string
		AppOrigin         string
		ServicesEncrypted string
		//
		db []entry `json:"-"`
	}

	entry struct {
		Name      string
		Secret    string
		UpdatedAt int
		Otp       otp
		Order     struct{ Position int }
		Icon      struct {
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

	otp struct {
		Label     string
		Account   string
		Issuer    string
		Digits    int
		Period    int
		Algorithm string
		TokenType string `json:"tokenType"`
		Source    string
	}
)

func Open(filename string, pass []byte) (vaults.Vault, error) {

	var t = vaults.TYPE_TWOFAS
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

	plain, err := v.decryptDB(key)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", t, err)
	}

	if err := json.Unmarshal(plain, &v.db); err != nil {
		return nil, fmt.Errorf("%s: %w", t, err)
	}

	return v, nil
}

func (v vault) Entries() []vaults.Entry {

	entries := make([]vaults.Entry, 0)

	for _, e := range v.db {
		e.Otp.TokenType = strings.ToUpper(e.Otp.TokenType)
		if e.Otp.TokenType != "TOTP" {
			log.Printf("Ignoring entry %q: %s", e.Otp.Issuer, e.Otp.TokenType)
			continue
		}

		if e.Secret == "" {
			log.Printf("Ignoring entry %q: missing secret", e.Otp.Issuer)
			continue
		}

		if e.Otp.Period == 0 {
			log.Printf("Missing period for entry %q: using default (30)", e.Otp.Issuer)
			e.Otp.Period = 30
		}

		if e.Otp.Algorithm == "" {
			log.Printf("Missing algorithm for entry %q: using default (SHA1)", e.Otp.Issuer)
			e.Otp.Algorithm = "SHA1"
		}

		if e.Otp.Digits == 0 {
			log.Printf("Missing digits for entry %q: using default (6)", e.Otp.Issuer)
			e.Otp.Digits = 6
		}

		label := e.Otp.Label

		if label == "" {
			label = e.Otp.Account
		}

		entries = append(entries, vaults.Entry{
			Secret:    e.Secret,
			Issuer:    e.Otp.Issuer,
			Label:     label,
			Digits:    e.Otp.Digits,
			Type:      e.Otp.TokenType,
			Algorithm: e.Otp.Algorithm,
			Period:    e.Otp.Period,
		})
	}

	return entries
}

func (v vault) masterKeyFromPass(password []byte) ([]byte, error) {

	servicesEncrypted := strings.SplitN(v.ServicesEncrypted, ":", numFields+1)
	if len(servicesEncrypted) != numFields {
		return nil, fmt.Errorf("invalid vault file: number of fields is not %d", numFields)
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
		msg := "invalid vault file: length of cipher text with auth tag must be more than %d"
		return nil, fmt.Errorf(msg, authTagLength)
	}

	return pbkdf2.Key(password, salt, 10000, 32, sha256.New), nil
}

func (v vault) decryptDB(key []byte) ([]byte, error) {
	servicesEncrypted := strings.SplitN(v.ServicesEncrypted, ":", numFields+1)
	if len(servicesEncrypted) != numFields {
		return nil, fmt.Errorf("invalid vault file: number of fields is not %d", numFields)
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
