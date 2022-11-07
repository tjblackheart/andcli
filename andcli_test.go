package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

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
			os.TempDir(), "test.aes", AEGIS,
			&config{File: filepath.Join(os.TempDir(), "test.aes"), Type: AEGIS},
			false,
		},
		{
			"saves abs path",
			".", "test2.aes", AEGIS,
			&config{File: filepath.Join(cwd, "test2.aes"), Type: AEGIS},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfgDir = os.TempDir()
			cfgFile = filepath.Join(cfgDir, "config_test.yaml")

			cfg, err := newConfig(filepath.Join(tt.vaultDir, tt.vaultFile), tt.vaultType)
			if tt.fails {
				t.Log(cfg)
				assert.Error(t, err)
				return
			}

			assert.NoError(t, cfg.save())
			assert.Equal(t, tt.want, cfg)
			assert.FileExists(t, cfgFile)
		})
	}
}
