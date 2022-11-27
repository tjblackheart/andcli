package main

import (
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
