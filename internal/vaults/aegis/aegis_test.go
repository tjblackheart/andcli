package aegis

import (
	"encoding/json"
	"errors"
	"reflect"
	"strings"
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
				{Issuer: "iss-1", Info: info{Digits: 6, Secret: "jbswy3dp"}, Type: "TOTP"},
				{Issuer: "iss-2", Info: info{Digits: 4, Secret: "jbswy3dp"}, Type: "HOTP"},
				{Issuer: "iss-3", Info: info{Digits: 0, Secret: "jbswy3dp", Period: 20}, Type: "TOTP"},
				{Issuer: "iss-4", Info: info{Digits: 4, Secret: "jbswy3dp", Algo: "SHA256"}, Type: "TOTP"},
				{Issuer: "iss-5"},
			},
			[]vaults.Entry{
				{Issuer: "iss-1", Digits: 6, Secret: "JBSWY3DP", Type: "TOTP", Algorithm: "SHA1", Period: 30},
				{Issuer: "iss-3", Digits: 6, Secret: "JBSWY3DP", Type: "TOTP", Algorithm: "SHA1", Period: 20},
				{Issuer: "iss-4", Digits: 4, Secret: "JBSWY3DP", Type: "TOTP", Algorithm: "SHA256", Period: 30},
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

func TestValidateScryptParams(t *testing.T) {
	tests := []struct {
		name    string
		n, r, p int
		wantErr bool
	}{
		{"valid", 32768, 8, 1, false},
		{"minimum valid", 2, 1, 1, false},
		{"max n", 1 << 20, 1, 1, false},      // 128 MiB
		{"max memory", 1 << 18, 8, 1, false}, // 256 MiB
		{"n too small", 1, 8, 1, true},
		{"n not power of two", 3, 8, 1, true},
		{"n too large", 1 << 21, 1, 1, true},
		{"n negative", -1, 8, 1, true},
		{"r zero", 32768, 0, 1, true},
		{"r too large", 32768, 33, 1, true},
		{"p zero", 32768, 8, 0, true},
		{"p too large", 32768, 8, 17, true},
		{"memory too large", 1 << 19, 8, 1, true}, // 512 MiB
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateScryptParams(tt.n, tt.r, tt.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateScryptParams(%d, %d, %d) error = %v, wantErr %v", tt.n, tt.r, tt.p, err, tt.wantErr)
			}
		})
	}
}

func TestMasterKeyFromPassRejectsHugeKDF(t *testing.T) {
	const raw = `{"header":{"slots":[{"type":1,"n":1048576,"r":32,"p":1,"salt":"00","key":"00","key_params":{"nonce":"00","tag":"00"}}]}}`

	var v aegis
	if err := json.Unmarshal([]byte(raw), &v); err != nil {
		t.Fatal(err)
	}

	if _, err := v.masterKeyFromPass([]byte("x")); err == nil {
		t.Fatal("masterKeyFromPass() error = nil, want an error")
	}
}

func TestValidateScryptParamsMessage(t *testing.T) {
	err := validateScryptParams(1<<19, 8, 1)
	if err == nil {
		t.Fatal("validateScryptParams() error = nil, want an error")
	}

	for _, want := range []string{"N=524288", "R=8", "max 268435456"} {
		if !strings.Contains(err.Error(), want) {
			t.Errorf("error = %q, want it to contain %q", err, want)
		}
	}
}
