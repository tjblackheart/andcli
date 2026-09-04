package ente_test

import (
	"fmt"
	"testing"

	"github.com/tjblackheart/andcli/v2/internal/vaults/ente"
)

// This test will take quite some time because of the ente crypto.
func TestOpen(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		password string
		fails    bool
	}{
		{"decrypts", "testdata/ente-auth-codes.json", "andcli-test", false},
		{"rejects wrong password", "testdata/ente-auth-codes.json", "invalid", true},
		{"rejects plain data", "testdata/ente-auth-codes-plain.txt", "", true},
		{"rejects invalid version", "testdata/ente-noversion.json", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := ente.Open(tt.filename, []byte(tt.password))
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
