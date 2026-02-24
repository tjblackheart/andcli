package keepass

import (
	"fmt"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/tjblackheart/andcli/v2/internal/vaults"
	"github.com/tobischo/gokeepasslib/v3"
)

const vaultType = vaults.KEEPASS

var _ vaults.Vault = &keepass{}

type keepass struct{ entries []gokeepasslib.Entry }

func Open(filename string, pass []byte) (vaults.Vault, error) {
	v := keepass{entries: make([]gokeepasslib.Entry, 0)}

	file, err := os.Open(filename)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", vaultType, err)
	}

	db := gokeepasslib.NewDatabase()
	db.Credentials = gokeepasslib.NewPasswordCredentials(string(pass))

	if err := gokeepasslib.NewDecoder(file).Decode(db); err != nil {
		return nil, fmt.Errorf("%s: %s", vaultType, err)
	}

	if err := db.UnlockProtectedEntries(); err != nil {
		return nil, fmt.Errorf("%s: %s", vaultType, err)
	}

	if len(db.Content.Root.Groups) == 0 {
		return nil, fmt.Errorf("%s: no content", vaultType)
	}

	v.entries = append(v.entries, parseGroups(db.Content.Root.Groups)...)

	return v, nil
}

func (v keepass) Entries() []vaults.Entry {
	entries := make([]vaults.Entry, 0)
	for _, e := range v.entries {
		issuer := e.GetTitle()

		v := e.GetContent("otp")
		if v == "" {
			continue
		}

		otp, err := url.Parse(v)
		if err != nil {
			log.Printf("%q: %s", issuer, err)
			continue
		}

		period, _ := strconv.Atoi(otp.Query().Get("period"))
		digits, _ := strconv.Atoi(otp.Query().Get("digits"))

		entry := vaults.Entry{
			Secret:    otp.Query().Get("secret"),
			Issuer:    issuer,
			Label:     e.GetContent("UserName"),
			Digits:    digits,
			Type:      strings.ToUpper(otp.Host),
			Algorithm: otp.Query().Get("algorithm"),
			Period:    period,
		}

		if err := entry.SanitizeAndValidate(); err == nil {
			entries = append(entries, entry)
		}
	}

	return entries
}

func parseGroups(groups []gokeepasslib.Group) []gokeepasslib.Entry {
	entries := make([]gokeepasslib.Entry, 0)
	for _, group := range groups {
		entries = append(entries, group.Entries...)
		entries = append(entries, parseGroups(group.Groups)...)
	}
	return entries
}
