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
}

func (e entry) generateTOTP() (string, int64) {
	return gotp.NewDefaultTOTP(e.Secret).NowWithExpiration()
}
