package vaults

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"log"
	"strings"

	"github.com/xlzd/gotp"
)

// Entry represents a generic vault entry.
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

var (
	ErrMissingSecret = errors.New("missing secret value")
	ErrInvalidType   = errors.New("entry is not a TOTP")
)

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
	case "sha224", "sha-224":
		h.HashName = "sha224"
		h.Digest = sha256.New224
	case "sha256", "sha-256":
		h.HashName = "sha256"
		h.Digest = sha256.New
	case "sha384", "sha-384":
		h.HashName = "sha384"
		h.Digest = sha512.New384
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

// SanitizeAndValidate will add missing defaults if necessary
// (to prevent division by zero, for example). If there are crucial fields
// missing (i.e. secret), it will return an error.
func (e *Entry) SanitizeAndValidate() error {
	if e.Secret == "" {
		log.Printf("%q: ignoring: missing secret", e.Issuer)
		return ErrMissingSecret
	}

	if strings.ToUpper(e.Type) != "TOTP" {
		log.Printf("%q: ignoring: %s", e.Issuer, e.Type)
		return ErrInvalidType
	}

	if e.Period == 0 {
		log.Printf("%q: missing period, using default (30)", e.Issuer)
		e.Period = 30
	}

	if e.Algorithm == "" {
		log.Printf("%q: missing algorithm: using default (SHA1)", e.Issuer)
		e.Algorithm = "SHA1"
	}

	if e.Digits == 0 {
		log.Printf("%q: missing digits: using default (6)", e.Issuer)
		e.Digits = 6
	}

	return nil
}
