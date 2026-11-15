package ente

import (
	"fmt"
	"strings"
	"testing"
)

// This test will take quite some time because of the ente crypto.
func TestOpen(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		password string
		fails    bool
		slow     bool
	}{
		{"decrypts", "testdata/ente-auth-codes.json", "andcli-test", false, true},
		{"rejects wrong password", "testdata/ente-auth-codes.json", "invalid", true, true},
		{"rejects plain data", "testdata/ente-auth-codes-plain.txt", "", true, false},
		{"rejects invalid version", "testdata/ente-noversion.json", "", true, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if testing.Short() && tt.slow {
				t.Skip("ente KDF fixture is 1 GiB, skipping. run this test without -short")
			}

			v, err := Open(tt.filename, []byte(tt.password))
			if tt.fails {
				if err == nil {
					t.Fatal("Open() expected error, got none")
				}
				return
			}

			entries := v.Entries()
			if len(entries) != 3 {
				t.Fatalf("Open() expected len to be 3, have %v", len(entries))
			}

			for i := range 3 {
				want := fmt.Sprintf("demo%d", i+1)
				if entries[i].Label != want {
					t.Fatalf("Open() have %v, %s", entries[i].Label, want)
				}
			}
		})
	}
}

func TestValidateKDFParams(t *testing.T) {
	tests := []struct {
		name     string
		memLimit uint32
		opsLimit uint32
		wantErr  bool
	}{
		{"valid", 1 << 30, 4, false},
		{"minimum", 1 << 20, 1, false},
		{"mem too small", (1 << 20) - 1, 4, true},
		{"mem too large", (1 << 30) + 1, 4, true},
		{"ops zero", 1 << 30, 0, true},
		{"ops too large", 1 << 30, 11, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateKDFParams(tt.memLimit, tt.opsLimit)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateKDFParams(%d, %d) error = %v, wantErr %v", tt.memLimit, tt.opsLimit, err, tt.wantErr)
			}
		})
	}
}

func TestOpenRejectsHugeKDF(t *testing.T) {
	const raw = `{"version":1,"kdfParams":{"memLimit":1073741825,"opsLimit":4,"salt":""},"encryptedData":"","encryptionNonce":""}`

	if _, err := open([]byte(raw), []byte("x")); err == nil {
		t.Fatal("open() error = nil, want an error")
	}
}

func TestValidateKDFParamsMessages(t *testing.T) {
	err := validateKDFParams((1<<30)+1, 4)
	if err == nil || !strings.Contains(err.Error(), "memLimit too large") {
		t.Errorf("error = %v, want memLimit message", err)
	}

	err = validateKDFParams(1<<20, 11)
	if err == nil || !strings.Contains(err.Error(), "opsLimit 11") {
		t.Errorf("error = %v, want opsLimit message", err)
	}
}

func FuzzOpen(f *testing.F) {
	f.Add([]byte(`{}`))
	f.Add([]byte(`{"version":1}`))
	f.Add([]byte("not a vault"))

	f.Fuzz(func(t *testing.T, data []byte) {
		v, err := open(data, []byte("andcli-test"))
		if err != nil {
			return
		}
		_ = v.Entries()
	})
}
