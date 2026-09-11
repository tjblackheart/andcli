package ente

import (
	"encoding/base64"
	"encoding/json"
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
	p_cost     = 1
	kdfKeySize = 32
	vaultType  = vaults.ENTE
)

var _ vaults.Vault = &ente{}

type (
	ente struct {
		Version         int
		KdfParams       params
		EncryptedData   string
		EncryptionNonce string
		//
		plain []byte
	}

	params struct {
		MemLimit uint32
		OpsLimit uint32
		Salt     string
	}
)

func init() { vaults.Register(vaultType, Open) }

func Open(filename string, pass []byte) (vaults.Vault, error) {
	var v ente

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

	v.plain, err = v.decrypt(pass)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", vaultType, err)
	}

	return &v, nil
}

func (v *ente) IsPlain(b []byte) bool {
	return len(b) == 0 || b[0] != '{' || !strings.Contains(string(b), "kdfParams")
}

func (v *ente) Entries() []vaults.Entry {
	entries := make([]vaults.Entry, 0)

	for line := range strings.SplitSeq(string(v.plain), "\n") {
		if line == "" {
			continue
		}

		u, err := url.Parse(line)
		if err != nil {
			log.Printf("parse entry: %s", err)
			continue
		}

		label := "-"
		if parts := strings.SplitN(u.Path, ":", 2); len(parts) == 2 {
			label = parts[1]
		}

		values := u.Query()
		digits, _ := strconv.Atoi(values.Get("digits"))
		period, _ := strconv.Atoi(values.Get("period"))

		entry := vaults.Entry{
			Secret:    values.Get("secret"),
			Issuer:    values.Get("issuer"),
			Label:     label,
			Digits:    digits,
			Type:      strings.ToUpper(u.Host),
			Algorithm: values.Get("algorithm"),
			Period:    period,
		}

		if err := entry.SanitizeAndValidate(); err == nil {
			entries = append(entries, entry)
		}
	}

	return entries
}

func (v *ente) decrypt(pass []byte) ([]byte, error) {
	salt, err := base64.StdEncoding.DecodeString(v.KdfParams.Salt)
	if err != nil {
		return nil, err
	}

	t_cost, m_cost := v.KdfParams.OpsLimit, v.KdfParams.MemLimit/1024
	if m_cost < 1024 || t_cost < 1 {
		return nil, fmt.Errorf("invalid kdf params")
	}
	key := argon2.IDKey(pass, salt, t_cost, m_cost, p_cost, kdfKeySize)

	data, err := base64.StdEncoding.DecodeString(v.EncryptedData)
	if err != nil {
		return nil, err
	}

	nonce, err := base64.StdEncoding.DecodeString(v.EncryptionNonce)
	if err != nil {
		return nil, err
	}

	dec, err := newStreamDecryptor(key, nonce)
	if err != nil {
		return nil, err
	}

	return dec.pull(data)
}
