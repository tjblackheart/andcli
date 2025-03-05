package vaults

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"strings"

	"github.com/xlzd/gotp"
)

// an Entry represents a generic vault entry.
type Entry struct {
	Secret    string
	Issuer    string
	Label     string
	Type      string
	Algorithm string
	Tags      []string
	Digits    int
	Period    int
}

// Returns a generated OTP and expiration time for the current entry.
func (e Entry) GenerateTOTP() (string, int64) {
	totp := gotp.NewTOTP(e.Secret, e.Digits, e.Period, e.hasher())
	return totp.NowWithExpiration()
}

// Returns a new gotp Hasher based on the entry algorithm.
func (e Entry) hasher() *gotp.Hasher {
	h := &gotp.Hasher{
		HashName: "sha1",
		Digest:   sha1.New,
	}

	switch strings.ToLower(e.Algorithm) {
	case "sha256", "sha-256":
		h.HashName = "sha256"
		h.Digest = sha256.New
	case "sha512", "sha-512":
		h.HashName = "sha512"
		h.Digest = sha512.New
	}

	return h
}

// Implementation of bubbletea listitem.Title()
func (e Entry) Title() string {
	title := strings.TrimSpace(e.Issuer)
	if title == "" {
		title = strings.Split(e.Label, " - ")[0]
	}
	return title
}

// Implementation of bubbletea listitem.Description()
func (e Entry) Description() string {
	desc := e.Label

	parts := strings.Split(desc, " - ")
	if len(parts) > 1 {
		desc = parts[1]
	}

	parts = strings.Split(desc, ":")
	desc = parts[0]
	if len(parts) > 1 {
		desc = parts[1]
	}

	return desc
}

// Implementation of bubbletea listitem.FilterValue()
func (e Entry) FilterValue() string {
	return e.Title()
}
