package stratum

import (
	"fmt"
	"reflect"
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
	}{
		{"decrypts", "testdata/backup-andcli-test.stratum", "andcli-test", false},
		{"fails: wrong password", "testdata/backup-andcli-test.stratum", "", true},
		{"fails: legacy", "testdata/backup-legacy-andcli-test.stratum", "", true},
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

func TestEntries(t *testing.T) {
	tests := []struct {
		name  string
		input []entry
		want  []vaults.Entry
	}{
		{
			"mitigates missing fields",
			[]entry{
				{Issuer: "iss-1", Digits: 6, Secret: "secret", Type: 2},
				{Issuer: "iss-2", Digits: 4, Secret: "secret", Type: 1},
				{Issuer: "iss-3", Digits: 0, Secret: "secret", Type: 2, Period: 20},
				{Issuer: "iss-4", Digits: 4, Secret: "secret", Type: 2, Algorithm: 1},
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
			entries := (&vault{Authenticators: tt.input}).Entries()
			if !reflect.DeepEqual(entries, tt.want) {
				t.Fatalf("Entries(): want %#v\nhave %#v", tt.want, entries)
			}
		})
	}
}

func Test_entry_typeToString(t *testing.T) {
	tests := []struct {
		name string
		e    entry
		want string
	}{
		{"hotp", entry{Type: 1}, "HOTP"},
		{"totp", entry{Type: 2}, "TOTP"},
		{"mobile", entry{Type: 3}, "MOBILE"},
		{"steam", entry{Type: 4}, "STEAM"},
		{"yandex", entry{Type: 5}, "YANDEX"},
		{"unknown", entry{Type: 0}, "UNKNOWN"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.typeToString(); got != tt.want {
				t.Errorf("entry.typeToString() = %v, want %v", got, tt.want)
			}
		})
	}
}
