package config

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/tjblackheart/andcli/internal/buildinfo"
	"gopkg.in/yaml.v3"
)

type (
	Config struct {
		File         string `yaml:"file"`
		Type         string `yaml:"type"`
		ClipboardCmd string `yaml:"clipboard_cmd"`
		//
		path string `yaml:"-"`
	}
)

// Returns a new application config. It merges a possibly existing config
// plus given flags into a current app config. Missing dirs apart
// from the default system config directory will be created in the process.
func Create() (*Config, error) {

	userDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("config: can not read directory: %s", err)
	}

	return create(userDir)
}

func create(d string) (*Config, error) {
	path, err := createDirectories(filepath.Join(d, buildinfo.AppName))
	if err != nil {
		return nil, err
	}

	cfg := &Config{
		path: filepath.Join(path, "config.yaml"),
	}

	if err := cfg.mergeExisting(); err != nil {
		return nil, fmt.Errorf("config: %s", err)
	}

	if err := cfg.parseFlags(); err != nil {
		return nil, fmt.Errorf("config: %s", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config: %s", err)
	}

	return cfg, nil
}

// Writes the current configuration to a yaml file.
func (c Config) Persist() error {
	b, err := yaml.Marshal(c)
	if err != nil {
		return fmt.Errorf("config: %s", err)
	}

	return os.WriteFile(c.path, b, 0600)
}

// Reads an possibly existing config file and merges the content
// into the current config.
func (c *Config) mergeExisting() error {
	if _, err := os.Stat(c.path); os.IsNotExist(err) {
		return nil
	}

	b, err := os.ReadFile(c.path)
	if err != nil {
		return err
	}

	var existing = new(Config)
	if err := yaml.Unmarshal(b, &existing); err != nil {
		return err
	}

	c.File = existing.File
	c.Type = existing.Type
	c.ClipboardCmd = existing.ClipboardCmd

	return nil
}

// Validates the current configuration.
func (c *Config) validate() error {

	if c.File == "" {
		return errors.New("no vault file specified")
	}

	if c.Type == "" {
		return errors.New("no vault type specified")
	}

	var err error
	if c.File, err = filepath.Abs(c.File); err != nil {
		return fmt.Errorf("%s: %s", c.File, err)
	}

	fi, err := os.Stat(c.File)
	if err != nil {
		return fmt.Errorf("%s: %s", c.File, err)
	}

	if fi.IsDir() {
		return fmt.Errorf("%s: is a directory, not a vault file", c.File)
	}

	// if set, check if the basic clipboard cmd is available in system PATH.
	// the option parsing is done at a later time.
	if parts := strings.SplitN(c.ClipboardCmd, " ", 2); parts[0] != "" {
		path, err := exec.LookPath(parts[0])
		if err != nil {
			return fmt.Errorf("%s: %s", parts[0], err)
		}
		c.ClipboardCmd = strings.ReplaceAll(c.ClipboardCmd, parts[0], path)
	}

	return nil
}

func createDirectories(path string) (string, error) {

	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.MkdirAll(path, 0700); err != nil {
			return "", err
		}
	}

	return path, nil
}
