package config

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/goccy/go-yaml"
	"github.com/tjblackheart/andcli/v2/internal/buildinfo"
	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

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
			"ignores nonexisting",
			&Config{File: "test.json", Type: "test", ClipboardCmd: "", path: path},
			nil,
			false,
		},
		{
			"handles default options",
			&Config{path: path},
			&Config{
				File:         "test.json",
				Type:         "test",
				ClipboardCmd: "",
				Options: &Opts{
					ShowUsernames: false,
					ShowTokens:    false,
				},
				path: path,
			},
			false,
		},
		{
			"handles custom options",
			&Config{
				Options: &Opts{
					ShowUsernames: true,
					ShowTokens:    true,
				},
				path: path,
			},
			&Config{
				File:         "test.json",
				Type:         "test",
				ClipboardCmd: "",
				Options: &Opts{
					ShowUsernames: true,
					ShowTokens:    true,
				},
				path: path,
			},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.want != nil {
				b, err := yaml.Marshal(tt.want)
				if err != nil {
					t.Fatal(err)
				}

				if err := os.WriteFile(tt.want.path, b, 0o644); err != nil {
					t.Fatal(err)
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

	os.WriteFile(path, nil, 0o644)

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

func TestConfig_Persist(t *testing.T) {
	fname := filepath.Join(os.TempDir(), "andcli_test_config.yaml")
	defer os.RemoveAll(fname)

	cfg := &Config{File: "test.json", Type: "aegis", ClipboardCmd: "/usr/bin/test", path: fname, dirty: true}
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
	if err := yaml.Unmarshal(b, cfg2); err != nil {
		t.Fatal(err)
	}
	cfg2.path = fname
	cfg2.dirty = true

	if !reflect.DeepEqual(cfg, cfg2) {
		t.Errorf("Config.Persist() cfg2 = %v, want %v", cfg2, cfg)
	}
}

func TestConfig_Persist_preservesComments(t *testing.T) {
	fname := filepath.Join(os.TempDir(), "andcli_test_config_comments.yaml")
	defer os.RemoveAll(fname)

	original := `# This is a comment at the top
file: /path/to/vault.json # inline comment
type: aegis
# Comment before options
options:
  show_usernames: true # another inline
  show_tokens: false
clipboard_cmd: ""
# Comment before theme
theme:
  base: "#39A02E"
  green: "#39A02E"
  yellow: "#DB9F1F"
  red: "#f10000"
  grey: "#424242"
  black: "#000000"
  white: "#FFFFFF"
`

	if err := os.WriteFile(fname, []byte(original), 0o600); err != nil {
		t.Fatal(err)
	}

	cfg := &Config{
		File:         "/new/vault.json",
		Type:         vaults.Type("2fas"),
		ClipboardCmd: "pbcopy",
		Options: &Opts{
			ShowUsernames: true,
			ShowTokens:    true,
		},
		Theme: &Theme{
			Base:   "#111111",
			Green:  "#222222",
			Yellow: "#333333",
			Red:    "#444444",
			Grey:   "#555555",
			Black:  "#666666",
			White:  "#777777",
		},
		path:  fname,
		dirty: true,
	}

	if err := cfg.Persist(); err != nil {
		t.Fatalf("Config.Persist() error = %v", err)
	}

	b, err := os.ReadFile(fname)
	if err != nil {
		t.Fatal(err)
	}

	if !strings.Contains(string(b), "# This is a comment at the top") {
		t.Error("top comment was not preserved")
	}
	if !strings.Contains(string(b), "# inline comment") {
		t.Error("inline comment was not preserved")
	}
	if !strings.Contains(string(b), "# Comment before options") {
		t.Error("comment before options was not preserved")
	}
	if !strings.Contains(string(b), "# Comment before theme") {
		t.Error("comment before theme was not preserved")
	}

	if strings.Contains(string(b), "/path/to/vault.json") {
		t.Error("file path was not updated")
	}
	if strings.Contains(string(b), "aegis") {
		t.Error("type was not updated")
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

	// default config
	want := &Config{
		File:         abs,
		Type:         vaults.Type(*vtype),
		ClipboardCmd: "",
		Options: &Opts{
			ShowUsernames: true,
			ShowTokens:    false,
		},
		Theme:   &DefaultTheme,
		path:    filepath.Join(cfgDir, buildinfo.AppName, "config.yaml"),
		dirty:   true,
		timeout: 5,
	}

	if !reflect.DeepEqual(cfg, want) {
		t.Errorf("want: %#v, have: %#v", want, cfg)
	}
}

func TestConfig_Flags(t *testing.T) {
	args := os.Args
	defer func() { os.Args = args }()

	tmpFile, err := os.CreateTemp("", "dummy.vault")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name())
	absPath, _ := filepath.Abs(tmpFile.Name())

	tests := []struct {
		name  string
		args  []string
		check func(*Config)
	}{
		{
			"trims query",
			[]string{"andcli", "-q", "  myquery \n", "-t", "aegis", tmpFile.Name()},
			func(c *Config) {
				if c.Query() != "myquery" {
					t.Errorf("Query() = %q, want %q", c.Query(), "myquery")
				}
			},
		},
		{
			"sanitizes utf8",
			[]string{"andcli", "-q", " \xab  myquery \xff", "-t", "2fas", tmpFile.Name()},
			func(c *Config) {
				if c.Query() != "myquery" {
					t.Errorf("Query() = %q, want %q", c.Query(), "myquery")
				}
			},
		},
		{
			"reads query from -q",
			[]string{"andcli", "-q", "query", "-t", "aegis", tmpFile.Name()},
			func(c *Config) {
				if c.Query() != "query" {
					t.Errorf("Query() = %q, want %q", c.Query(), "query")
				}
			},
		},
		{
			"reads query from --query",
			[]string{"andcli", "--query", "query", "-t", "aegis", tmpFile.Name()},
			func(c *Config) {
				if c.Query() != "query" {
					t.Errorf("Query() = %q, want %q", c.Query(), "query")
				}
			},
		},
		{
			"reads --passwd-stdin",
			[]string{"andcli", "--passwd-stdin", "-t", "aegis", tmpFile.Name()},
			func(c *Config) {
				if !c.PasswdStdin() {
					t.Error("PasswdStdin() = false, want true")
				}
			},
		},
		{
			"sets file",
			[]string{"andcli", "-f", tmpFile.Name(), "-t", "aegis"},
			func(c *Config) {
				if c.File != absPath {
					t.Errorf("File = %q, want %q", c.File, absPath)
				}
			},
		},
		{
			"sets type",
			[]string{"andcli", "-t", "2fas", tmpFile.Name()},
			func(c *Config) {
				if c.Type != "2fas" {
					t.Errorf("Type = %q, want %q", c.Type, "2fas")
				}
			},
		},
		{
			"sets clipboardCmd",
			[]string{"andcli", "-c", "pbcopy", "-t", "aegis", tmpFile.Name()},
			func(c *Config) {
				if c.ClipboardCmd != "pbcopy" {
					t.Errorf("ClipboardCmd = %q, want %q", c.ClipboardCmd, "pbcopy")
				}
			},
		},
		{
			"sets file from arg[0]",
			[]string{"andcli", "-t", "aegis", tmpFile.Name()},
			func(c *Config) {
				if c.File != absPath {
					t.Errorf("File = %q, want %q", c.File, absPath)
				}
			},
		},
		{
			"arg[0] overrides file flag",
			[]string{"andcli", "-f", "other.vault", "-t", "aegis", tmpFile.Name()},
			func(c *Config) {
				if c.File != absPath {
					t.Errorf("File = %q, want %q", c.File, absPath)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			cfg := &Config{
				Options: &Opts{},
				Theme:   &DefaultTheme,
			}
			if err := cfg.parseFlags(); err != nil {
				t.Fatalf("parseFlags() failed: %v", err)
			}
			tt.check(cfg)
		})
	}
}
