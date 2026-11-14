package vaults

import (
	"errors"
	"reflect"
	"testing"
)

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
				Algorithm: "sha-1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.normalizeAlgo(); err != nil {
				t.Fatal(err)
			}
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

func TestEntry_GenerateTOTPInvalid(t *testing.T) {
	tests := []struct {
		name string
		e    Entry
	}{
		{"unknown algorithm", Entry{Secret: "JBSWY3DP", Digits: 6, Period: 30, Algorithm: "MD5"}},
		{"invalid secret", Entry{Secret: "123", Digits: 6, Period: 30, Algorithm: "SHA1"}},
		{"invalid digits", Entry{Secret: "JBSWY3DP", Digits: 11, Period: 30, Algorithm: "SHA1"}},
		{"invalid period", Entry{Secret: "JBSWY3DP", Digits: 6, Period: 0, Algorithm: "SHA1"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, exp := tt.e.GenerateTOTP()
			if token != "" || exp != 0 {
				t.Errorf("GenerateTOTP() = (%q, %d), want empty", token, exp)
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
		{"fails: invalid secret", &Entry{Secret: "123"}, nil, true},
		{"fails: wrong type", &Entry{Secret: "jbswy3dp", Type: "HOTP"}, nil, true},
		{
			"defaults: period",
			&Entry{
				Secret:    "jbswy3dp",
				Type:      "TOTP",
				Period:    0,
				Algorithm: "SHA1",
				Digits:    6,
			},
			&Entry{
				Secret:    "JBSWY3DP",
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
				Secret:    "jbswy3dp",
				Type:      "TOTP",
				Period:    30,
				Algorithm: "",
				Digits:    6,
			},
			&Entry{
				Secret:    "JBSWY3DP",
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
				Secret:    "jbswy3dp",
				Type:      "TOTP",
				Period:    30,
				Algorithm: "SHA1",
				Digits:    0,
			},
			&Entry{
				Secret:    "JBSWY3DP",
				Type:      "TOTP",
				Period:    30,
				Algorithm: "SHA1",
				Digits:    6,
			},
			false,
		},
		{
			"normalizes: secret to uppercase",
			&Entry{
				Secret:    "jbswy3dp",
				Type:      "TOTP",
				Period:    30,
				Algorithm: "SHA1",
				Digits:    6,
			},
			&Entry{
				Secret:    "JBSWY3DP",
				Type:      "TOTP",
				Period:    30,
				Algorithm: "SHA1",
				Digits:    6,
			},
			false,
		},
		{
			"accepts: one digit",
			&Entry{
				Secret:    "jbswy3dp",
				Type:      "TOTP",
				Period:    30,
				Algorithm: "SHA1",
				Digits:    1,
			},
			&Entry{
				Secret:    "JBSWY3DP",
				Type:      "TOTP",
				Period:    30,
				Algorithm: "SHA1",
				Digits:    1,
			},
			false,
		},
		{
			"accepts: two digits",
			&Entry{
				Secret:    "jbswy3dp",
				Type:      "TOTP",
				Period:    30,
				Algorithm: "SHA1",
				Digits:    2,
			},
			&Entry{
				Secret:    "JBSWY3DP",
				Type:      "TOTP",
				Period:    30,
				Algorithm: "SHA1",
				Digits:    2,
			},
			false,
		},
		{
			"normalizes: algorithm",
			&Entry{
				Secret:    "jbswy3dp",
				Type:      "TOTP",
				Period:    30,
				Algorithm: "sha-256",
				Digits:    6,
			},
			&Entry{
				Secret:    "JBSWY3DP",
				Type:      "TOTP",
				Period:    30,
				Algorithm: "SHA256",
				Digits:    6,
			},
			false,
		},
		{
			"normalizes: secret spaces",
			&Entry{
				Secret:    "jbsw y3dp",
				Type:      "TOTP",
				Period:    30,
				Algorithm: "SHA1",
				Digits:    6,
			},
			&Entry{
				Secret:    "JBSWY3DP",
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

func TestEntry_SanitizeAndValidateErrors(t *testing.T) {
	tests := []struct {
		name    string
		e       Entry
		wantErr error
	}{
		{
			"invalid digits",
			Entry{Secret: "JBSWY3DP", Type: "TOTP", Period: 30, Algorithm: "SHA1", Digits: 11},
			ErrInvalidDigits,
		},
		{
			"invalid period",
			Entry{Secret: "JBSWY3DP", Type: "TOTP", Period: -1, Algorithm: "SHA1", Digits: 6},
			ErrInvalidPeriod,
		},
		{
			"unsupported algorithm",
			Entry{Secret: "JBSWY3DP", Type: "TOTP", Period: 30, Algorithm: "MD5", Digits: 6},
			ErrUnsupportedAlgorithm,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.e.SanitizeAndValidate(); !errors.Is(err, tt.wantErr) {
				t.Errorf("SanitizeAndValidate() error = %v, want %v", err, tt.wantErr)
			}
		})
	}
}

func TestFind(t *testing.T) {
	entries := []Entry{
		{Issuer: "Google", Label: "user@gmail.com"},
		{Issuer: "GitHub", Label: "user@github.com"},
		{Issuer: "gitlab.com", Label: "user@gitlab.com"},
		{Issuer: "Some White Space", Label: "username"},
		{Issuer: "**Special ~ Chars**", Label: "username"},
	}

	tests := []struct {
		query   string
		want    *Entry
		wantErr bool
	}{
		{"Google", &entries[0], false},
		{"goog", &entries[0], false},
		{"GITHUB", &entries[1], false},
		{"git", nil, true},
		{"nomatch", nil, true},
		{"", nil, true},
		{"white space", &entries[3], false},
		{"~", &entries[4], false},
		{"[something]", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.query, func(t *testing.T) {
			got, err := Find(tt.query, entries)
			if (err != nil) != tt.wantErr {
				t.Errorf("Find() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Find() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEntry_normalizeAlgo(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    string
		wantErr error
	}{
		{"empty", "", "SHA1", nil},
		{"sha1", "SHA1", "SHA1", nil},
		{"dash", "SHA-1", "SHA1", nil},
		{"lower", "sha1", "SHA1", nil},
		{"sha224", "SHA224", "SHA224", nil},
		{"sha256", "sha256", "SHA256", nil},
		{"sha384", "SHA384", "SHA384", nil},
		{"sha512", "SHA512", "SHA512", nil},
		{"unsupported", "MD5", "MD5", ErrUnsupportedAlgorithm},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := Entry{Algorithm: tt.in}
			err := e.normalizeAlgo()
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("normalizeAlgo() error = %v, want %v", err, tt.wantErr)
			}
			if e.Algorithm != tt.want {
				t.Errorf("Algorithm = %q, want %q", e.Algorithm, tt.want)
			}
		})
	}
}
