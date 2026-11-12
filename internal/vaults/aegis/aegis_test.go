package aegis

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/tjblackheart/andcli/v2/internal/vaults"
	"github.com/tjblackheart/andcli/v2/internal/vaults/vaulttest"
)

func TestOpen(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		password string
		fails    bool
		wantErr  error
	}{
		{"decrypts", "testdata/aegis-export-test.json", "andcli-test", false, nil},
		{"fails: wrong password", "testdata/aegis-export-test.json", "invalid", true, nil},
		{"fails: invalid file", "testdata/aegis-invalid-file.json", "invalid", true, nil},
		{"fails: plaintext vault", "testdata/aegis-export-test-plain.json", "", true, vaults.ErrIsPlain},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := Open(tt.filename, []byte(tt.password))
			if tt.fails {
				if err == nil {
					t.Fatal("Open() expected error, got none")
				}
				if tt.wantErr != nil && !errors.Is(err, tt.wantErr) {
					t.Fatalf("Open() error = %v, want %v", err, tt.wantErr)
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
				{Issuer: "iss-1", Info: info{Digits: 6, Secret: "secret"}, Type: "TOTP"},
				{Issuer: "iss-2", Info: info{Digits: 4, Secret: "secret"}, Type: "HOTP"},
				{Issuer: "iss-3", Info: info{Digits: 0, Secret: "secret", Period: 20}, Type: "TOTP"},
				{Issuer: "iss-4", Info: info{Digits: 4, Secret: "secret", Algo: "SHA256"}, Type: "TOTP"},
				{Issuer: "iss-5"},
			},
			[]vaults.Entry{
				{Issuer: "iss-1", Digits: 6, Secret: "SECRET", Type: "TOTP", Algorithm: "SHA1", Period: 30},
				{Issuer: "iss-3", Digits: 6, Secret: "SECRET", Type: "TOTP", Algorithm: "SHA1", Period: 20},
				{Issuer: "iss-4", Digits: 4, Secret: "SECRET", Type: "TOTP", Algorithm: "SHA256", Period: 30},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := (&aegis{db: db{Entries: tt.input}}).Entries()
			if !reflect.DeepEqual(entries, tt.want) {
				t.Fatalf("Entries(): want %#v\nhave %#v", tt.want, entries)
			}
		})
	}
}

func FuzzOpen(f *testing.F) {
	vaulttest.Seed(f, "testdata")

	f.Fuzz(func(t *testing.T, data []byte) {
		v, err := open(data, []byte("andcli-test"))
		if err != nil {
			return
		}
		_ = v.Entries()
	})
}

func TestDecryptDBRejectsInvalidNonce(t *testing.T) {
	var v aegis
	v.Header.Params.Nonce = "00"

	if _, err := v.decryptDB(make([]byte, 32)); err == nil {
		t.Fatal("decryptDB() error = nil, want an error")
	}
}

func TestMasterKeyFromPassRejectsInvalidNonce(t *testing.T) {
	const raw = `{"header":{"slots":[{"type":1,"n":2,"r":1,"p":1,"salt":"00","key":"00","key_params":{"nonce":"00","tag":"00"}}]}}`

	var v aegis
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatal(err)
	}

	if _, err := v.masterKeyFromPass([]byte("x")); err == nil {
		t.Fatal("masterKeyFromPass() error = nil, want an error")
	}
}
