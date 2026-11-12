package protonauth

import (
	"fmt"
	"testing"

	"github.com/tjblackheart/andcli/v2/internal/vaults/vaulttest"
)

func TestOpen(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		password string
		fails    bool
	}{
		{"decrypts", "testdata/protonauth-test.json", "andcli-test", false},
		{"rejects wrong password", "testdata/protonauth-test.json", "invalid", true},
		{"rejects plain data", "testdata/protonauth-test-plain.json", "", true},
		{"rejects invalid version", "testdata/protonauth-test-noversion.json", "", true},
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
