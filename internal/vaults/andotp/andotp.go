package andotp

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/grijul/go-andotp/andotp"
	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

type (
	vault struct {
		entries []entry
	}

	entry struct {
		Secret    string
		Issuer    string
		Label     string
		Digits    int
		Type      string
		Algorithm string
		Thumbnail string
		Period    int
		LastUsed  int64 `json:"last_used"`
		UsedFreq  int   `json:"used_frequency"`
		Tags      []string
	}
)

func Open(filename string, pass []byte) (vaults.Vault, error) {

	var t = vaults.TYPE_ANDOTP

	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", t, err)
	}

	b, err = andotp.Decrypt(b, string(pass))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", t, err)
	}

	entries := make([]entry, 0)
	if err := json.Unmarshal(b, &entries); err != nil {
		return nil, fmt.Errorf("%s: %w", t, err)
	}

	return &vault{entries}, nil

}

func (v vault) Entries() []vaults.Entry {

	entries := make([]vaults.Entry, 0)

	for _, e := range v.entries {

		e.Type = strings.ToUpper(e.Type)
		if e.Type != "TOTP" {
			log.Printf("Ignoring entry %q: %s", e.Issuer, e.Type)
			continue
		}

		if e.Secret == "" {
			log.Printf("Ignoring entry %q: missing secret", e.Issuer)
			continue
		}

		if e.Period == 0 {
			log.Printf("Missing period for entry %q: using default (30)", e.Issuer)
			e.Period = 30
		}

		if e.Algorithm == "" {
			log.Printf("Missing algorithm for entry %q: using default (SHA1)", e.Issuer)
			e.Algorithm = "SHA1"
		}

		if e.Digits == 0 {
			log.Printf("Missing digits for entry %q: using default (6)", e.Issuer)
			e.Digits = 6
		}

		entries = append(entries, vaults.Entry{
			Secret:    e.Secret,
			Issuer:    e.Issuer,
			Label:     e.Label,
			Type:      e.Type,
			Algorithm: e.Algorithm,
			Tags:      e.Tags,
			Digits:    e.Digits,
			Period:    e.Period,
		})
	}

	return entries
}
