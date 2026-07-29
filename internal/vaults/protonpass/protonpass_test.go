package protonpass

import (
	"errors"
	"fmt"
	"testing"

	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

func TestMain(m *testing.M) {
	m.Run()
}

func TestOpen(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		password string
		fails    bool
		wantErr  error
	}{
		{"decrypts text", "testdata/protonpass-test.pgp", "andcli-test", false, nil},
		{"decrypts zip", "testdata/protonpass-test.pgp.zip", "andcli-test", false, nil},
		{"decrypts hidden zip", "testdata/protonpass-test.pgp.data", "andcli-test", false, nil},
		{"fails: wrong password", "testdata/protonpass-test.pgp", "", true, nil},
		{"fails: plaintext vault", "testdata/protonpass-test-plain.zip", "", true, vaults.ErrIsPlain},
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
