package andotp

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
		{"decrypts", "testdata/andotp_test.json.aes", "andcli-test", false},
		{"fails: wrong password", "testdata/andotp_test.json.aes", "invalid", true},
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
				{Issuer: "iss-1", Digits: 6, Secret: "secret", Type: "TOTP"},
				{Issuer: "iss-2", Digits: 4, Secret: "secret", Type: "HOTP"},
				{Issuer: "iss-3", Digits: 0, Secret: "secret", Type: "TOTP", Period: 20},
				{Issuer: "iss-4", Digits: 4, Secret: "secret", Type: "TOTP", Algorithm: "SHA256"},
				{Issuer: "iss-5"},
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
			entries := (&vault{entries: tt.input}).Entries()
			if !reflect.DeepEqual(entries, tt.want) {
				t.Fatalf("Entries(): want %#v\nhave %#v", tt.want, entries)
			}
		})
	}
}
