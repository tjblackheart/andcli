package andotp

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	gao "github.com/grijul/go-andotp/andotp"
	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

const (
	vaultType     = vaults.ANDOTP
	iterationLen  = 4
	saltLen       = 12
	ivLen         = 12
	gcmTagLen     = 16
	minVaultLen   = iterationLen + saltLen + ivLen + gcmTagLen
	maxIterations = 1_000_000
)

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

func init() { vaults.Register(vaultType, Open) }

func Open(filename string, pass []byte) (vaults.Vault, error) {
	b, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", vaultType, err)
	}
	return open(b, pass)
}

func open(data, pass []byte) (vaults.Vault, error) {
	v := andotp{entries: make([]entry, 0)}

	if v.IsPlain(data) {
		return nil, vaults.ErrIsPlain
	}

	if len(data) < minVaultLen {
		return nil, fmt.Errorf(
			"%s: %w: truncated vault",
			vaultType, vaults.ErrInvalidData,
		)
	}

	i := binary.BigEndian.Uint32(data[:iterationLen])
	if i == 0 || i > maxIterations {
		return nil, fmt.Errorf(
			"%s: %w: invalid iteration count %d",
			vaultType, vaults.ErrInvalidData, i,
		)
	}

	plain, err := gao.Decrypt(data, string(pass))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", vaultType, err)
	}

	if err := json.Unmarshal(plain, &v.entries); err != nil {
		return nil, fmt.Errorf("%s: %w", vaultType, err)
	}

	return &v, nil
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

func (v andotp) IsPlain(b []byte) bool {
	return len(b) > 0 && b[0] == '['
}
