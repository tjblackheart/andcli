package twofas

import (
	"reflect"
	"testing"

	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

func TestOpen(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		password string
		fails    bool
	}{
		{"decrypts", "testdata/twofas-export-test.2fas", "andcli-test", false},
		{"fails: wrong password", "testdata/twofas-export-test.2fas", "invalid", true},
		{"fails: invalid file", "testdata/twofas-invalid-file.2fas", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := Open(tt.filename, []byte(tt.password))
			if tt.fails {
				if err == nil {
					t.Fatal("Open() expected error, got none")
				}
				return
			}

			entries := v.Entries()
			if len(entries) != 1 {
				t.Fatalf("Open() expected len to be 1, have %v", len(entries))
			}

			if entries[0].Label != "andcli-test" {
				t.Fatalf("Open() have %v, want andcli-test", entries[0].Label)
			}
		})
	}
}

func TestEntries(t *testing.T) {
	tests := []struct {
		name  string
		input []entry
		want  []vaults.Entry
	}{
		{
			"mitigates missing fields",
			[]entry{
				{Secret: "secret", Otp: otp{Issuer: "iss-1", Digits: 6, TokenType: "TOTP"}},
				{Secret: "secret", Otp: otp{Issuer: "iss-2", Digits: 4, TokenType: "HOTP"}},
				{Secret: "secret", Otp: otp{Issuer: "iss-3", Digits: 0, TokenType: "TOTP", Period: 20}},
				{Secret: "secret", Otp: otp{Issuer: "iss-4", Digits: 4, TokenType: "TOTP", Algorithm: "SHA256"}},
				{Otp: otp{Issuer: "iss-5"}},
			},
			[]vaults.Entry{
				{Issuer: "iss-1", Digits: 6, Secret: "secret", Type: "TOTP", Algorithm: "SHA1", Period: 30},
				{Issuer: "iss-3", Digits: 6, Secret: "secret", Type: "TOTP", Algorithm: "SHA1", Period: 20},
				{Issuer: "iss-4", Digits: 4, Secret: "secret", Type: "TOTP", Algorithm: "SHA256", Period: 30},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := (&vault{db: tt.input}).Entries()
			if !reflect.DeepEqual(entries, tt.want) {
				t.Fatalf("Entries(): want %#v\nhave %#v", tt.want, entries)
			}
		})
	}
}
