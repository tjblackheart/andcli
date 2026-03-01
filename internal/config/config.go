package config

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/goccy/go-yaml"
	"github.com/tjblackheart/andcli/v2/internal/buildinfo"
	"github.com/tjblackheart/andcli/v2/internal/vaults"
)

type (
	Config struct {
		File         string      `yaml:"file"`
		Type         vaults.Type `yaml:"type"`
		ClipboardCmd string      `yaml:"clipboard_cmd"`
		Options      *Opts       `yaml:"options"`
		Theme        *Theme      `yaml:"theme"`
		//
		path              string
		passwordFromStdin bool
	}

	Opts struct {
		ShowUsernames bool `yaml:"show_usernames"`
		ShowTokens    bool `yaml:"show_tokens"`
	}
)

// Returns a new application config. It merges a possibly existing config
// plus given flags into a current app config. Missing dirs apart
// from the default system config directory will be created in the process.
func Create() (*Config, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("unable to read user directory: %s", err)
	}
	return create(dir)
}

func create(dir string) (*Config, error) {
	path := filepath.Join(dir, buildinfo.AppName)
	if err := os.MkdirAll(path, os.ModePerm); err != nil {
		return nil, err
	}

	cfg := &Config{
		path: filepath.Join(path, "config.yaml"),
		Options: &Opts{
			ShowUsernames: true,
			ShowTokens:    false,
		},
		Theme: &DefaultTheme,
	}

	if err := cfg.mergeExisting(); err != nil {
		return nil, err
	}

	if err := cfg.parseFlags(); err != nil {
		return nil, err
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Writes the current configuration to a yaml file.
func (cfg Config) Persist() error {
	b, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}

	return os.WriteFile(cfg.path, b, 0o600)
}

// Returns true if the flag option "passwd-stdin" was set.
func (cfg Config) PasswdStdin() bool {
	return cfg.passwordFromStdin
}

// Reads an possibly existing config file and merges the content
// into the current config.
func (cfg *Config) mergeExisting() error {
	if _, err := os.Stat(cfg.path); os.IsNotExist(err) {
		return nil
	}

	b, err := os.ReadFile(cfg.path)
	if err != nil {
		return err
	}

	existing := new(Config)
	if err := yaml.Unmarshal(b, existing); err != nil {
		return err
	}

	cfg.File = existing.File
	cfg.Type = existing.Type
	cfg.ClipboardCmd = existing.ClipboardCmd

	if existing.Options != nil {
		cfg.Options = existing.Options
	}

	if existing.Theme != nil {
		cfg.Theme = existing.Theme
		cfg.Theme.validate()
	}

	return nil
}

// Validates the current configuration.
func (cfg *Config) validate() error {
	if cfg.File == "" {
		return errors.New("no vault file specified")
	}

	if cfg.Type == "" {
		return errors.New("no vault type specified")
	}

	var err error
	if cfg.File, err = filepath.Abs(cfg.File); err != nil {
		return fmt.Errorf("%s: %s", cfg.File, err)
	}

	fi, err := os.Stat(cfg.File)
	if err != nil {
		return fmt.Errorf("%s: %s", cfg.File, err)
	}

	if fi.IsDir() {
		return fmt.Errorf("%s: is a directory, not a vault file", cfg.File)
	}

	// if set, check if the basic clipboard cmd is available in system PATH.
	// the option parsing is done at a later time.
	if parts := strings.SplitN(cfg.ClipboardCmd, " ", 2); parts[0] != "" {
		path, err := exec.LookPath(parts[0])
		if err != nil {
			return fmt.Errorf("%s: %s", parts[0], err)
		}
		cfg.ClipboardCmd = strings.ReplaceAll(cfg.ClipboardCmd, parts[0], path)
	}

	return nil
}
