package keepass

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/tjblackheart/andcli/v2/internal/vaults"
	keepass "github.com/tobischo/gokeepasslib/v3"
)

type vault struct {
	entries []keepass.Entry
}

func Open(filename string, pass []byte) (vaults.Vault, error) {

	var t = vaults.TYPE_KEEPASS
	var v = vault{entries: make([]keepass.Entry, 0)}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", t, err)
	}

	db := keepass.NewDatabase()
	db.Credentials = keepass.NewPasswordCredentials(string(pass))

	if err := keepass.NewDecoder(file).Decode(db); err != nil {
		return nil, fmt.Errorf("%s: %s", t, err)
	}

	if err := db.UnlockProtectedEntries(); err != nil {
		return nil, fmt.Errorf("%s: %s", t, err)
	}

	if len(db.Content.Root.Groups) == 0 {
		return nil, fmt.Errorf("%s: no content", t)
	}

	v.entries = append(v.entries, parseGroups(db.Content.Root.Groups)...)

	return v, nil
}

func (v vault) Entries() []vaults.Entry {

	entries := make([]vaults.Entry, 0)
	for _, e := range v.entries {
		issuer := e.GetTitle()

		v := e.GetContent("otp")
		if v == "" {
			continue
		}

		otp, err := url.Parse(v)
		if err != nil {
			log.Printf("Ignoring entry %q: %s", issuer, err)
			continue
		}

		otpType := strings.ToUpper(otp.Host)
		if otpType != "TOTP" {
			log.Printf("Ignoring entry %q: %s", issuer, otpType)
			continue
		}

		period, _ := strconv.Atoi(otp.Query().Get("period"))
		if period == 0 {
			log.Printf("Missing period for entry %q: using default (30)", issuer)
			period = 30
		}

		digits, _ := strconv.Atoi(otp.Query().Get("digits"))
		if digits == 0 {
			log.Printf("Missing digits for entry %q: using default (6)", issuer)
			digits = 6
		}

		entries = append(entries, vaults.Entry{
			Secret:    otp.Query().Get("secret"),
			Issuer:    issuer,
			Label:     e.GetContent("UserName"),
			Digits:    digits,
			Type:      otpType,
			Algorithm: otp.Query().Get("algorithm"),
			Period:    period,
		})
	}

	return entries
}

func parseGroups(groups []keepass.Group) []keepass.Entry {
	entries := make([]keepass.Entry, 0)
	for _, group := range groups {
		entries = append(entries, group.Entries...)
		entries = append(entries, parseGroups(group.Groups)...)
	}
	return entries
}
