package andotp

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/grijul/go-andotp/andotp"
	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

const t = vaults.TYPE_ANDOTP

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
		entry := vaults.Entry{
			Secret:    e.Secret,
			Issuer:    e.Issuer,
			Label:     e.Label,
			Type:      strings.ToUpper(e.Type),
			Algorithm: e.Algorithm,
			Tags:      e.Tags,
			Digits:    e.Digits,
			Period:    e.Period,
		}

		if err := entry.SanitizeAndValidate(); err == nil {
			entries = append(entries, entry)
		}
	}

	return entries
}
