package vaults

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/sahilm/fuzzy"
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
	ErrMissingSecret        = errors.New("missing secret value")
	ErrInvalidSecret        = errors.New("invalid secret value")
	ErrInvalidType          = errors.New("entry is not a TOTP")
	ErrInvalidDigits        = errors.New("invalid digits")
	ErrInvalidPeriod        = errors.New("invalid period")
	ErrUnsupportedAlgorithm = errors.New("unsupported algorithm")

	hashers = map[string]*gotp.Hasher{
		"SHA1":   {HashName: "sha1", Digest: sha1.New},
		"SHA224": {HashName: "sha224", Digest: sha256.New224},
		"SHA256": {HashName: "sha256", Digest: sha256.New},
		"SHA384": {HashName: "sha384", Digest: sha512.New384},
		"SHA512": {HashName: "sha512", Digest: sha512.New},
	}
)

// Returns a generated OTP and expiration time for the current entry.
func (e Entry) GenerateTOTP() (string, int64) {
	hasher, ok := hashers[e.Algorithm]
	if !ok || !gotp.IsSecretValid(e.Secret) || e.Digits < 1 || e.Digits > 10 || e.Period < 1 {
		return "", 0
	}
	return gotp.NewTOTP(e.Secret, e.Digits, e.Period, hasher).NowWithExpiration()
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

	e.Secret = strings.ToUpper(e.Secret)
	e.Secret = strings.ReplaceAll(e.Secret, " ", "")

	if !gotp.IsSecretValid(e.Secret) {
		log.Printf("%q: ignoring: invalid secret", e.Issuer)
		return ErrInvalidSecret
	}

	if strings.ToUpper(e.Type) != "TOTP" {
		log.Printf("%q: ignoring: %s", e.Issuer, e.Type)
		return ErrInvalidType
	}

	if e.Period == 0 {
		log.Printf("%q: missing period, using default (30)", e.Issuer)
		e.Period = 30
	}

	if e.Period < 1 {
		log.Printf("%q: ignoring: invalid period %d", e.Issuer, e.Period)
		return ErrInvalidPeriod
	}

	if err := e.normalizeAlgo(); err != nil {
		log.Printf("%q: ignoring: %s", e.Issuer, err)
		return err
	}

	if e.Digits == 0 {
		log.Printf("%q: missing digits: using default (6)", e.Issuer)
		e.Digits = 6
	}

	if e.Digits < 1 || e.Digits > 10 {
		log.Printf("%q: ignoring: invalid digits %d", e.Issuer, e.Digits)
		return ErrInvalidDigits
	}

	return nil
}

func (e *Entry) normalizeAlgo() error {
	s := strings.ToUpper(e.Algorithm)
	s = strings.ReplaceAll(strings.TrimSpace(s), "-", "")

	switch s {
	case "", "SHA1":
		e.Algorithm = "SHA1"
	case "SHA224", "SHA256", "SHA384", "SHA512":
		e.Algorithm = s
	default:
		return fmt.Errorf("%w: %q", ErrUnsupportedAlgorithm, s)
	}

	return nil
}

// Find fuzzy filters a list of entries for s.
func Find(s string, entries []Entry) (*Entry, error) {
	stack := make([]string, 0, len(entries))
	for _, e := range entries {
		stack = append(stack, e.FilterValue())
	}

	matches := fuzzy.Find(s, stack)
	switch len(matches) {
	case 0:
		return nil, fmt.Errorf("no results for %q", s)
	case 1:
		return &entries[matches[0].Index], nil
	default:
		hits := []string{}
		for _, m := range matches {
			hits = append(hits, m.Str)
		}
		return nil, fmt.Errorf("multiple matches for %q: %s", s, strings.Join(hits, ", "))
	}
}
