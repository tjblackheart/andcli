package protonauth

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/tjblackheart/andcli/v2/internal/vaults"
	"golang.org/x/crypto/argon2"
)

const (
	// Argon2 / AES constants taken from Proton Authenticator.
	// See github.com/protonpass/proton-pass-common
	// and github.com/protonpass/android-authenticator
	m_cost     = 19 * 1024
	t_cost     = 2
	p_cost     = 1
	keySize    = 32
	saltSize   = 16
	ivSize     = 12
	tagSize    = 16
	additional = "proton.authenticator.export.v1"
	vaultType  = vaults.PROTON_AUTH
)

var _ vaults.Vault = &protonAuthenticator{}

type (
	protonAuthenticator struct {
		Version int
		Salt    string
		Content string
		//
		db db
	}

	db struct{ Entries []entry }

	entry struct {
		ID      string
		Content content
		Note    string
	}

	content struct {
		URI  string
		Type string `json:"entry_type"`
		Name string
	}
)

func init() { vaults.Register(vaultType, Open) }

func Open(filename string, pass []byte) (vaults.Vault, error) {
	var v protonAuthenticator

	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", vaultType, err)
	}

	if v.IsPlain(b) {
		return nil, vaults.ErrIsPlain
	}

	if err := json.Unmarshal(b, &v); err != nil {
		return nil, fmt.Errorf("%s: %w", vaultType, err)
	}

	if v.Version != 1 {
		return nil, fmt.Errorf("%s: unsupported version", vaultType)
	}

	db, err := v.decrypt(pass)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", vaultType, err)
	}

	if err := json.Unmarshal(db, &v.db); err != nil {
		return nil, fmt.Errorf("%s: %w", vaultType, err)
	}

	return &v, nil
}

func (v *protonAuthenticator) IsPlain(b []byte) bool {
	return !strings.Contains(string(b), "salt")
}

func (v *protonAuthenticator) Entries() []vaults.Entry {
	entries := make([]vaults.Entry, 0)

	for _, e := range v.db.Entries {
		tokenType := strings.ToUpper(e.Content.Type)
		if tokenType != "TOTP" {
			continue
		}

		u, err := url.Parse(e.Content.URI)
		if err != nil {
			log.Printf("%q: %s", e.Content.Name, err)
			continue
		}

		values := u.Query()
		digits, _ := strconv.Atoi(values.Get("digits"))
		period, _ := strconv.Atoi(values.Get("period"))

		entry := vaults.Entry{
			Secret:    values.Get("secret"),
			Issuer:    values.Get("issuer"),
			Label:     e.Content.Name,
			Digits:    digits,
			Type:      tokenType,
			Algorithm: values.Get("algorithm"),
			Period:    period,
		}

		if err := entry.SanitizeAndValidate(); err == nil {
			entries = append(entries, entry)
		}
	}

	return entries
}

func (v *protonAuthenticator) decrypt(pass []byte) ([]byte, error) {
	salt, err := base64.StdEncoding.DecodeString(v.Salt)
	if err != nil {
		return nil, err
	}

	key := argon2.IDKey(pass, salt, t_cost, m_cost, p_cost, keySize)

	content, err := base64.StdEncoding.DecodeString(v.Content)
	if err != nil {
		return nil, err
	}

	if len(content) < ivSize+tagSize {
		return nil, errors.New("unexpected content length")
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	iv, ct := content[:ivSize], content[ivSize:]

	return gcm.Open(nil, iv, ct, []byte(additional))
}
