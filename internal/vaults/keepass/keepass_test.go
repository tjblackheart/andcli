package keepass

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/tjblackheart/andcli/v2/internal/vaults"
	"github.com/tobischo/gokeepasslib/v3"
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
		{"decrypts", "testdata/keepass-test.kdbx", "andcli-test", false},
		{"fails: wrong password", "testdata/keepass-test.kdbx", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v, err := Open(tt.filename, []byte(tt.password))
			if tt.fails {
				if err == nil {
					t.Fatal("Open() expected error, got nil")
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
		input []gokeepasslib.Entry
		want  []vaults.Entry
	}{
		{
			"mitigates missing fields",
			[]gokeepasslib.Entry{
				{Values: []gokeepasslib.ValueData{
					{Key: "Title", Value: gokeepasslib.V{Content: "iss-1"}},
					{Key: "UserName", Value: gokeepasslib.V{Content: "demo1"}},
					{Key: "otp", Value: gokeepasslib.V{Content: "otpauth://totp/otp.provider.dev%3Ademo1?secret=secret&period=30&digits=6&issuer=otp.provider.dev&algorithm=SHA1"}},
				}},
				{Values: []gokeepasslib.ValueData{
					{Key: "Title", Value: gokeepasslib.V{Content: "iss-2"}},
					{Key: "UserName", Value: gokeepasslib.V{Content: "demo2"}},
					{Key: "otp", Value: gokeepasslib.V{Content: "otpauth://hotp/otp.provider.dev%3Ademo2?secret=secret&digits=6&issuer=otp.provider.dev&algorithm=SHA1"}},
				}},
				{Values: []gokeepasslib.ValueData{
					{Key: "Title", Value: gokeepasslib.V{Content: "iss-3"}},
					{Key: "UserName", Value: gokeepasslib.V{Content: "demo3"}},
					{Key: "otp", Value: gokeepasslib.V{Content: "otpauth://totp/otp.provider.dev%3Ademo3?secret=secret&period=20&issuer=otp.provider.dev&algorithm=SHA1"}},
				}},
				{Values: []gokeepasslib.ValueData{
					{Key: "Title", Value: gokeepasslib.V{Content: "iss-4"}},
					{Key: "UserName", Value: gokeepasslib.V{Content: "demo4"}},
					{Key: "otp", Value: gokeepasslib.V{Content: "otpauth://totp/otp.provider.dev%3Ademo1?secret=secret&digits=4&issuer=otp.provider.dev&algorithm=SHA256"}},
				}},
				{Values: []gokeepasslib.ValueData{
					{Key: "Title", Value: gokeepasslib.V{Content: "iss-5"}},
				}},
			},
			[]vaults.Entry{
				{Issuer: "iss-1", Label: "demo1", Digits: 6, Secret: "secret", Type: "TOTP", Algorithm: "SHA1", Period: 30},
				{Issuer: "iss-3", Label: "demo3", Digits: 6, Secret: "secret", Type: "TOTP", Algorithm: "SHA1", Period: 20},
				{Issuer: "iss-4", Label: "demo4", Digits: 4, Secret: "secret", Type: "TOTP", Algorithm: "SHA256", Period: 30},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			entries := (&vault{tt.input}).Entries()
			if !reflect.DeepEqual(entries, tt.want) {
				t.Fatalf("Entries(): want %#v\nhave %#v", tt.want, entries)
			}
		})
	}
}

func Test_parseGroups(t *testing.T) {

	tests := []struct {
		name   string
		groups []gokeepasslib.Group
		want   []gokeepasslib.Entry
	}{
		{
			"recursive subgroups",
			[]gokeepasslib.Group{
				{Name: "g1", Entries: []gokeepasslib.Entry{{IconID: 1}}},
				{Name: "g2", Groups: []gokeepasslib.Group{
					{Name: "g2-1", Entries: []gokeepasslib.Entry{{IconID: 2}}},
					{Name: "g2-2", Entries: []gokeepasslib.Entry{{IconID: 3}, {IconID: 4}}},
					{Name: "g2-3"},
				}},
				{Name: "g3", Groups: []gokeepasslib.Group{
					{Name: "g3-1", Groups: []gokeepasslib.Group{
						{Name: "g3-1-1"},
						{Name: "g3-1-2", Entries: []gokeepasslib.Entry{{IconID: 5}}},
					}},
				}},
			},
			[]gokeepasslib.Entry{
				{IconID: 1},
				{IconID: 2},
				{IconID: 3},
				{IconID: 4},
				{IconID: 5},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseGroups(tt.groups)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseGroups() = %v, want %v", got, tt.want)
			}
		})
	}
}
