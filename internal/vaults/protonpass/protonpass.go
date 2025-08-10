package protonpass

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/ProtonMail/gopenpgp/v3/crypto"
	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

const t = vaults.TYPE_PROTON

type (
	envelope struct{ Vaults map[string]vault }

	// protonvault only implements the essentials for reading OTP data.
	vault struct {
		Name, Description string
		Items             []struct {
			Data struct {
				Metadata struct{ Name string }
				Type     string
				Content  struct {
					Username string `json:"itemUsername"`
					TOTPUri  string `json:"totpUri"`
				}
			}
		}
	}
)

func Open(filename string, pass []byte) (vaults.Vault, error) {

	b, err := read(filename)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", t, err)
	}

	hnd, err := crypto.PGP().Decryption().Password(pass).New()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", t, err)
	}

	result, err := hnd.Decrypt(b, crypto.Armor)
	if err != nil {
		return nil, fmt.Errorf("%s: %s", t, err)
	}

	var e envelope
	if err := json.Unmarshal(result.Bytes(), &e); err != nil {
		return nil, fmt.Errorf("%s: %s", t, err)
	}

	return e, nil
}

func (e envelope) Entries() []vaults.Entry {
	var entries = make([]vaults.Entry, 0)

	for _, v := range e.Vaults {
		for _, i := range v.Items {
			d := i.Data
			if strings.ToLower(d.Type) != "login" || d.Content.TOTPUri == "" {
				continue
			}

			issuer := d.Metadata.Name
			uri, err := url.Parse(d.Content.TOTPUri)
			if err != nil {
				log.Printf("%q: %s", issuer, err)
				continue
			}

			period, _ := strconv.Atoi(uri.Query().Get("period"))
			digits, _ := strconv.Atoi(uri.Query().Get("digits"))

			entry := vaults.Entry{
				Secret:    uri.Query().Get("secret"),
				Issuer:    issuer,
				Label:     d.Content.Username,
				Digits:    digits,
				Type:      strings.ToUpper(uri.Host),
				Algorithm: uri.Query().Get("algorithm"),
				Period:    period,
			}

			if err := entry.SanitizeAndValidate(); err == nil {
				entries = append(entries, entry)
			}
		}
	}

	return entries
}

// opens, reads and returns file content, handles zip if necessary.
func read(filename string) ([]byte, error) {

	sig := []byte{0x50, 0x4b, 0x03, 0x04}

	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	head := make([]byte, 4)
	if _, err := f.ReadAt(head, 0); err != nil {
		return nil, err
	}

	// not a zip file
	if !bytes.Equal(head, sig) {
		return os.ReadFile(filename)
	}

	r, err := zip.OpenReader(filename)
	if err != nil {
		return nil, err
	}

	if len(r.File) == 0 {
		return nil, errors.New("archive has no content")
	}

	// read only the first entry.
	rc, err := r.File[0].Open()
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	return io.ReadAll(rc)
}
