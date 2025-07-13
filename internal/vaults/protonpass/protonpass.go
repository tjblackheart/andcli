package protonpass

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

const t = vaults.TYPE_PROTON

type (
	envelope struct{ Vaults map[string]vault }

	// protonvault only implements the essentials for reading TOTP info.
	vault struct {
		Name, Description string
		Items             []struct {
			Data struct {
				Metadata struct{ Name string }
				Type     string
				Content  struct {
					Username string `json:"itemUsername"`
					TOTPUri  string `json:"totpUri"`
				}
			}
		}
	}
)

func Open(filename string, pass []byte) (vaults.Vault, error) {

	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", t, err)
	}

	hnd, err := crypto.PGP().Decryption().Password(pass).New()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", t, err)
	}

	result, err := hnd.Decrypt(b, crypto.Armor)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", t, err)
	}

	var e envelope
	if err := json.Unmarshal(result.Bytes(), &e); err != nil {
		return nil, fmt.Errorf("%s: %s", t, err)
	}

	return e, nil
}

func (e envelope) Entries() []vaults.Entry {
	var entries = make([]vaults.Entry, 0)

	for _, v := range e.Vaults {
		for _, i := range v.Items {
			d := i.Data
			if strings.ToLower(d.Type) != "login" || d.Content.TOTPUri == "" {
				continue
			}

			issuer := d.Metadata.Name
			uri, err := url.Parse(d.Content.TOTPUri)
			if err != nil {
				log.Printf("%q: %s", issuer, err)
				continue
			}

			period, _ := strconv.Atoi(uri.Query().Get("period"))
			digits, _ := strconv.Atoi(uri.Query().Get("digits"))

			entry := vaults.Entry{
				Secret:    uri.Query().Get("secret"),
				Issuer:    issuer,
				Label:     d.Content.Username,
				Digits:    digits,
				Type:      strings.ToUpper(uri.Host),
				Algorithm: uri.Query().Get("algorithm"),
				Period:    period,
			}

			if err := entry.SanitizeAndValidate(); err == nil {
				entries = append(entries, entry)
			}
		}
	}

	return entries
}
