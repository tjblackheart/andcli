package andotp

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	gao "github.com/grijul/go-andotp/andotp"
	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

const vaultType = vaults.ANDOTP

var _ vaults.Vault = &andotp{}

type (
	andotp struct{ entries []entry }

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
		return nil, fmt.Errorf("%s: %w", vaultType, err)
	}

	b, err = gao.Decrypt(b, string(pass))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", vaultType, err)
	}

	entries := make([]entry, 0)
	if err := json.Unmarshal(b, &entries); err != nil {
		return nil, fmt.Errorf("%s: %w", vaultType, err)
	}

	return &andotp{entries}, nil
}

func (v andotp) Entries() []vaults.Entry {
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
