package main

import (
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
	return gotp.NewDefaultTOTP(e.Secret).NowWithExpiration()
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
