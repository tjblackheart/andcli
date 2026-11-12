package andotp

import (
	"bytes"
	"encoding/binary"
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
		{"decrypts", "testdata/andotp_test.json.aes", "andcli-test", false, nil},
		{"fails: wrong password", "testdata/andotp_test.json.aes", "invalid", true, nil},
		{"fails: plaintext vault", "testdata/andotp_test_plain.json", "", true, vaults.ErrIsPlain},
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
				{Issuer: "iss-1", Digits: 6, Secret: "secret", Type: "TOTP"},
				{Issuer: "iss-2", Digits: 4, Secret: "secret", Type: "HOTP"},
				{Issuer: "iss-3", Digits: 0, Secret: "secret", Type: "TOTP", Period: 20},
				{Issuer: "iss-4", Digits: 4, Secret: "secret", Type: "TOTP", Algorithm: "SHA256"},
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
			entries := (&andotp{entries: tt.input}).Entries()
			if !reflect.DeepEqual(entries, tt.want) {
				t.Fatalf("Entries(): want %#v\nhave %#v", tt.want, entries)
			}
		})
	}
}

func TestErrTruncated(t *testing.T) {
	for _, n := range []int{0, 1, 4, 27, 28, 43} {
		if _, err := open(bytes.Repeat([]byte{'0'}, n), []byte("x")); !errors.Is(err, vaults.ErrInvalidData) {
			t.Fatalf("len %d: err = %v, want ErrInvalidData", n, err)
		}
	}
}

func TestErrIterations(t *testing.T) {
	data := make([]byte, minVaultLen)
	binary.BigEndian.PutUint32(data, 4_000_000_000)
	if _, err := open(data, []byte("x")); !errors.Is(err, vaults.ErrInvalidData) {
		t.Fatalf("err = %v, want ErrInvalidData", err)
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
