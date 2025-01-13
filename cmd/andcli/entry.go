package main

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"strings"

	"github.com/xlzd/gotp"
)

type entry struct {
	Secret    string
	Issuer    string
	Label     string
	Digits    int
	Type      string
	Algorithm string
	Thumbnail string
	Period    int
	Tags      []string
	Choice    string // text to display in the model list view
}

func (e entry) generateTOTP() (string, int64) {
	t := gotp.NewTOTP(e.Secret, e.Digits, e.Period, e.generateHasher())
	return t.NowWithExpiration()
}

func (e entry) generateHasher() *gotp.Hasher {

	defaultHashName := "sha1"
	if e.Algorithm != "" {
		defaultHashName = strings.ToLower(e.Algorithm)
	}

	// default values.
	h := &gotp.Hasher{
		HashName: defaultHashName,
		Digest:   sha1.New,
	}

	switch h.HashName {
	case "sha256":
		h.Digest = sha256.New
	case "sha512":
		h.Digest = sha512.New
	}

	return h
}

type entries []entry

func (e entries) filter(val string) entries {
	filtered := make(entries, 0)
	for _, item := range e {
		if strings.Contains(strings.ToLower(item.Choice), strings.ToLower(val)) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
