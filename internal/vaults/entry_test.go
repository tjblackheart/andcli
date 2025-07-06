package vaults

import (
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"reflect"
	"testing"

	"github.com/xlzd/gotp"
)

func TestEntryGenerateHasher(t *testing.T) {
	tests := []struct {
		name  string
		entry *Entry
		want  *gotp.Hasher
	}{
		{
			"default",
			&Entry{},
			&gotp.Hasher{HashName: "sha1", Digest: sha1.New},
		},
		{
			"sha256",
			&Entry{Algorithm: "SHA256"},
			&gotp.Hasher{HashName: "sha256", Digest: sha256.New},
		},
		{
			"sha512",
			&Entry{Algorithm: "SHA512"},
			&gotp.Hasher{HashName: "sha512", Digest: sha512.New},
		},
		{
			"sha224",
			&Entry{Algorithm: "SHA-224"},
			&gotp.Hasher{HashName: "sha224", Digest: sha256.New224},
		},
		{
			"sha384",
			&Entry{Algorithm: "SHA-384"},
			&gotp.Hasher{HashName: "sha384", Digest: sha512.New384},
		},
		{
			"ignores cases",
			&Entry{Algorithm: "sha1"},
			&gotp.Hasher{HashName: "sha1", Digest: sha1.New},
		},
		{
			"handles dashes",
			&Entry{Algorithm: "SHA-1"},
			&gotp.Hasher{HashName: "sha1", Digest: sha1.New},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := tt.entry.hasher()
			if !reflect.DeepEqual(tt.want.HashName, h.HashName) {
				t.Errorf("entry.hasher() = %v, want %v", h, tt.want)
			}

			fn1 := reflect.Indirect(reflect.ValueOf(tt.want.Digest))
			fn2 := reflect.Indirect(reflect.ValueOf(h.Digest))
			if !reflect.DeepEqual(fn1, fn2) {
				t.Errorf("entry.hasher() = %v, want %v", h, tt.want)
			}
		})
	}
}

func TestEntry_Title(t *testing.T) {
	tests := []struct {
		name string
		e    Entry
		want string
	}{
		{"title from issuer", Entry{Label: "label", Issuer: "issuer"}, "issuer"},
		{"title from label", Entry{Label: "label", Issuer: ""}, "label"},
		{"title from label short", Entry{Label: "label - label2", Issuer: ""}, "label"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Title(); got != tt.want {
				t.Errorf("Entry.Title() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntry_Description(t *testing.T) {
	tests := []struct {
		name string
		e    Entry
		want string
	}{
		{"desc base", Entry{Label: "label"}, "label"},
		{"desc split 1", Entry{Label: "part1 - part2"}, "part2"},
		{"desc split 2", Entry{Label: "part1 - part2:part3"}, "part3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.Description(); got != tt.want {
				t.Errorf("Entry.Description() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntry_FilterValue(t *testing.T) {
	tests := []struct {
		name string
		e    Entry
		want string
	}{
		{"value: issuer", Entry{Label: "label", Issuer: "issuer"}, "issuer"},
		{"value: label", Entry{Label: "label", Issuer: ""}, "label"},
		{"value: label short", Entry{Label: "label - label2", Issuer: ""}, "label"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.e.FilterValue(); got != tt.want {
				t.Errorf("Entry.FilterValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntry_GenerateTOTP(t *testing.T) {
	tests := []struct {
		name string
		e    Entry
	}{
		{
			"entry1",
			Entry{
				Secret:    "4S62BZNFXXSZLCRO",
				Digits:    6,
				Period:    30,
				Algorithm: "sha1",
			},
		},
		{
			"entry2",
			Entry{
				Secret:    "4S62BZNFXXSZLCRO",
				Digits:    10,
				Period:    40,
				Algorithm: "sha1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s, i := tt.e.GenerateTOTP()
			if s == "" {
				t.Fatal("Entry.GenerateTOTP(): got empty string")
			}

			if len(s) != tt.e.Digits {
				t.Fatalf("Entry.GenerateTOTP(): len is %v, want %v", len(s), tt.e.Digits)
			}

			if i == 0 {
				t.Errorf("Entry.GenerateTOTP(): got empty time value")
			}
		})
	}
}

func TestEntry_SanitizeAndValidate(t *testing.T) {
	tests := []struct {
		name       string // description of this test case
		have, want *Entry
		fails      bool
	}{
		{"fails: missing secret", &Entry{Secret: ""}, nil, true},
		{"fails: wrong type", &Entry{Secret: "123", Type: "HOTP"}, nil, true},
		{
			"defaults: period",
			&Entry{
				Secret:    "123",
				Type:      "TOTP",
				Period:    0,
				Algorithm: "SHA1",
				Digits:    6,
			},
			&Entry{
				Secret:    "123",
				Type:      "TOTP",
				Period:    30,
				Algorithm: "SHA1",
				Digits:    6,
			},
			false,
		},
		{
			"defaults: algorithm",
			&Entry{
				Secret:    "123",
				Type:      "TOTP",
				Period:    30,
				Algorithm: "",
				Digits:    6,
			},
			&Entry{
				Secret:    "123",
				Type:      "TOTP",
				Period:    30,
				Algorithm: "SHA1",
				Digits:    6,
			},
			false,
		},
		{
			"defaults: digits",
			&Entry{
				Secret:    "123",
				Type:      "TOTP",
				Period:    30,
				Algorithm: "SHA1",
				Digits:    0,
			},
			&Entry{
				Secret:    "123",
				Type:      "TOTP",
				Period:    30,
				Algorithm: "SHA1",
				Digits:    6,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.have.SanitizeAndValidate()
			if tt.fails {
				if err == nil {
					t.Fatalf("SanitizeAndValidate(): %s: want err, got nil", tt.name)
				}
				return
			}

			if !reflect.DeepEqual(tt.have, tt.want) {
				t.Fatalf("SanitizeAndValidate(): %s: want %#v, got %#v", tt.name, tt.want, tt.have)
			}
		})
	}
}
