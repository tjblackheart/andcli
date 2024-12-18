package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/muesli/termenv"
	"github.com/stretchr/testify/assert"
)

func TestDecrypt(t *testing.T) {
	tests := []struct {
		name         string
		vfile, vtype string
		pass         []byte
		fails        bool
	}{
		{"fails: invalid file", "somefile", "", nil, true},
		{"fails: directory", os.TempDir(), "", nil, true},
		{"fails: wrong password", "testdata/aegis-export-test.json", "", []byte("pass"), true},
		{"fails: wrong type", "testdata/aegis-export-test.json", "sometype", []byte("andcli-test"), true},
		{"succeeds: aegis", "testdata/aegis-export-test.json", TYPE_AEGIS, []byte("andcli-test"), false},
		{"succeeds: andotp", "testdata/andotp_test.json.aes", TYPE_ANDOTP, []byte("andcli-test"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e, err := decrypt(tt.vfile, tt.vtype, tt.pass)
			if tt.fails {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.IsType(t, entries{}, e)
		})
	}
}

func TestAEGIS(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		password string
		fails    bool
	}{
		{"decrypts", "testdata/aegis-export-test.json", "andcli-test", false},
		{"fails: wrong password", "testdata/aegis-export-test.json", "invalid", true},
		{"fails: invalid file", "testdata/aegis-invalid-file.json", "invalid", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := os.ReadFile(tt.filename)
			assert.NoError(t, err)

			entries, err := decryptAEGIS(b, []byte(tt.password))
			if tt.fails {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, entries, 1)
			assert.Equal(t, entries[0].Label, "andcli-test")
		})
	}
}

func TestANDOTP(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		password string
		fails    bool
	}{
		{"decrypts", "testdata/andotp_test.json.aes", "andcli-test", false},
		{"fails: wrong password", "testdata/andotp_test.json.aes", "invalid", true},
		//{"fails: invalid file", "testdata/aegis-invalid-file.json", "invalid", true}, // panic in external lib
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := os.ReadFile(tt.filename)
			assert.NoError(t, err)

			entries, err := decryptANDOTP(b, []byte(tt.password))
			if tt.fails {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, err)
			assert.Len(t, entries, 1)
			assert.Equal(t, entries[0].Label, "andcli-test")
		})
	}
}

func TestConvertANDOTP(t *testing.T) {
	have := andotpEntry{
		Secret:    "secret",
		Issuer:    "issuer",
		Label:     "name",
		Digits:    6,
		Type:      "type",
		Algorithm: "algo",
		Thumbnail: "",
		Period:    30,
		LastUsed:  0,
		UsedFreq:  0,
		Tags:      nil,
	}

	want := &entry{
		Secret:    "secret",
		Issuer:    "issuer",
		Label:     "name",
		Digits:    6,
		Type:      "type",
		Algorithm: "algo",
		Thumbnail: "",
		Period:    30,
		Tags:      nil,
	}

	assert.Equal(t, have.toEntry(), want)
}

func TestConvertAEGISEntry(t *testing.T) {
	have := aegisEntry{
		Type:     "type",
		UUID:     "1",
		Name:     "name",
		Issuer:   "issuer",
		Note:     "",
		Icon:     "",
		IconMime: "",
		Info: struct {
			Secret string
			Algo   string
			Digits int
			Period int
		}{"secret", "algo", 6, 30},
	}

	want := &entry{
		Secret:    "secret",
		Issuer:    "issuer",
		Label:     "name",
		Digits:    6,
		Type:      "type",
		Algorithm: "algo",
		Thumbnail: "",
		Period:    30,
		Tags:      nil,
	}

	assert.Equal(t, have.toEntry(), want)
}

func TestConfig(t *testing.T) {
	cwd, err := os.Getwd()
	assert.NoError(t, err)

	tests := []struct {
		name                           string
		vaultDir, vaultFile, vaultType string
		want                           *config
		fails                          bool
	}{
		{
			"creates new config",
			os.TempDir(), "test.aes", TYPE_AEGIS,
			&config{File: filepath.Join(os.TempDir(), "test.aes"), Type: TYPE_AEGIS},
			false,
		},
		{
			"saves abs path",
			".", "test2.aes", TYPE_AEGIS,
			&config{File: filepath.Join(cwd, "test2.aes"), Type: TYPE_AEGIS},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgDir = os.TempDir()
			cfgFile = filepath.Join(cfgDir, "config_test.yaml")

			cfg, err := newConfig(filepath.Join(tt.vaultDir, tt.vaultFile), tt.vaultType, "")
			if tt.fails {
				assert.Error(t, err)
				return
			}

			assert.NoError(t, cfg.save())
			assert.Equal(t, tt.want, cfg)
			assert.FileExists(t, cfgFile)
		})
	}
}

func TestChoices(t *testing.T) {
	tests := []struct {
		name    string
		entries entries
		want    *model
	}{
		{
			"creates choices",
			entries{
				{Label: "label1", Issuer: "issuer1"},
				{Label: "label2", Issuer: "issuer2"},
				{Label: "label3"},
				{Issuer: "issuer4"},
			},
			&model{
				items: entries{
					{Label: "label1", Issuer: "issuer1", Choice: "issuer1 (label1)"},
					{Label: "label2", Issuer: "issuer2", Choice: "issuer2 (label2)"},
					{Label: "label3", Choice: "label3 (label3)"},
					{Issuer: "issuer4", Choice: "issuer4"},
				},
			},
		},
		{
			"does not fail on empty list",
			make(entries, 0),
			&model{items: make(entries, 0)},
		},
	}

	o := termenv.DefaultOutput()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := newModel(o, "", "", tt.entries...)
			assert.Equal(t, tt.want.items, m.items)
		})
	}
}
