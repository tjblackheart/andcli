package main

import (
	"encoding/json"

	"github.com/grijul/go-andotp/andotp"
)

type andotpEntry struct {
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

func (e andotpEntry) toEntry() *entry {
	return &entry{
		Secret:    e.Secret,
		Issuer:    e.Issuer,
		Label:     e.Label,
		Digits:    e.Digits,
		Type:      e.Type,
		Algorithm: e.Algorithm,
		Thumbnail: e.Thumbnail,
		Period:    e.Period,
		Tags:      e.Tags,
	}
}

//

func decryptANDOTP(data, password []byte) (entries, error) {
	b, err := andotp.Decrypt(data, string(password))
	if err != nil {
		return nil, err
	}

	var ae []andotpEntry
	if err := json.Unmarshal(b, &ae); err != nil {
		return nil, err
	}

	var list entries
	for _, e := range ae {
		list = append(list, *e.toEntry())
	}

	return list, nil
}
