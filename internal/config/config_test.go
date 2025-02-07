package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/tjblackheart/andcli/internal/buildinfo"
	"gopkg.in/yaml.v3"
)

func Test_createDirectories(t *testing.T) {
	wantDir := filepath.Join(os.TempDir(), "andcli_test", "config", buildinfo.AppName)

	tests := []struct {
		name  string
		dir   string
		fails bool
	}{
		{"creates directories", wantDir, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := createDirectories(tt.dir)
			if (err != nil) != tt.fails {
				t.Errorf("createDirectories() error = %v, wantErr %v", err, tt.fails)
				return
			}

			if _, err := os.Stat(tt.dir); err != nil {
				t.Errorf("createDirectories() = %v, want %v", got, tt.dir)
			}

			if err := os.RemoveAll(tt.dir); err != nil {
				t.Error(err)
			}
		})
	}
}

func TestConfig_mergeExisting(t *testing.T) {
	path := filepath.Join(os.TempDir(), "andcli_test_config.yaml")

	tests := []struct {
		name  string
		have  *Config
		want  *Config
		fails bool
	}{
		{
			"merges existing",
			&Config{File: "", Type: "", ClipboardCmd: "", path: path},
			&Config{File: "/tmp/test.json", Type: "aegis", ClipboardCmd: "", path: path},
			false,
		},
		{
			"merges nonexisting",
			&Config{File: "test.json", Type: "test", ClipboardCmd: "", path: path},
			nil,
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want != nil {
				b, err := yaml.Marshal(tt.want)
				if err != nil {
					t.Error(err)
				}
				if err := os.WriteFile(tt.want.path, b, 0644); err != nil {
					t.Error(err)
				}
			} else {
				tt.want = tt.have
			}

			err := tt.have.mergeExisting()
			if (err != nil) != tt.fails {
				t.Errorf("mergeExisting() error = %v, wantErr %v", err, tt.fails)
			}

			if !reflect.DeepEqual(tt.have, tt.want) {
				t.Errorf("mergeExisting() have = %v, want %v", tt.have, tt.want)
			}

			os.Remove(path)
		})
	}
}

func TestConfig_validate(t *testing.T) {
	path := filepath.Join(os.TempDir(), "andcli_test.json")
	defer os.Remove(path)

	os.WriteFile(path, nil, 0644)

	tests := []struct {
		name     string
		have     *Config
		fails    bool
		contains string
	}{
		{
			"validates missing file name",
			&Config{File: "", Type: "", ClipboardCmd: ""},
			true,
			"no vault file",
		},
		{
			"validates missing file type",
			&Config{File: "test.json", Type: "", ClipboardCmd: ""},
			true,
			"no vault type",
		},
		{
			"validates clipboard binary",
			&Config{File: path, Type: "test", ClipboardCmd: "nosuchbinary"},
			true,
			"file not found",
		},
		{
			"vaildates isDir",
			&Config{File: os.TempDir(), Type: "test", ClipboardCmd: ""},
			true,
			"is a directory",
		},
		{
			"passes correct config",
			&Config{File: path, Type: "test", ClipboardCmd: ""},
			false,
			"",
		},
		{
			"passes with clipboard binary",
			&Config{File: path, Type: "test", ClipboardCmd: "ls"},
			false,
			"",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.have.validate()
			if (err != nil) != tt.fails {
				t.Errorf("Config.validate() error = %v, wantErr %v", err, tt.fails)
			}

			if !tt.fails {
				return
			}

			if !strings.Contains(err.Error(), tt.contains) {
				t.Errorf("Config.validate() have = %q, want %q", err.Error(), tt.contains)
			}
		})
	}
}

func Test_persist(t *testing.T) {
	fname := filepath.Join(os.TempDir(), "andcli_test_config.yaml")
	defer os.RemoveAll(fname)

	cfg := &Config{File: "test.json", Type: "aegis", ClipboardCmd: "/usr/bin/test", path: fname}
	if err := cfg.Persist(); err != nil {
		t.Errorf("Config.Persist() error = %v, expected none", err)
		return
	}

	if _, err := os.Stat(fname); err != nil {
		if os.IsNotExist(err) {
			t.Errorf("Config.Persist() error = expected file at %q", fname)
			return
		}
		t.Errorf("Config.Persist() error = %s", err)
	}

	cfg2 := new(Config)
	b, _ := os.ReadFile(fname)
	if err := yaml.Unmarshal(b, &cfg2); err != nil {
		t.Fatal(err)
	}
	cfg2.path = fname

	if !reflect.DeepEqual(cfg, cfg2) {
		t.Errorf("Config.Persist() cfg2 = %v, want %v", cfg2, cfg)
	}
}

func Test_create(t *testing.T) {
	cfgDir := os.TempDir()
	*vfile = filepath.Join("testdata", "empty.json")
	*vtype = "aegis"
	abs, _ := filepath.Abs(*vfile)

	cfg, err := create(cfgDir)
	if err != nil {
		t.Fatal(err)
	}

	want := &Config{
		File:         abs,
		Type:         *vtype,
		ClipboardCmd: "",
		path:         filepath.Join(cfgDir, buildinfo.AppName, "config.yaml"),
	}

	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("want: %v, have: %v", want, cfg)
	}
}
