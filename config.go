package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type config struct {
	File string `yaml:"file"`
	Type string `yaml:"type"`
}

func configFromFile(vaultFile, vaultType string) (*config, error) {
	var err error

	if vaultFile != "" {
		vaultFile, err = filepath.Abs(vaultFile)
		if err != nil {
			return nil, err
		}
	}

	cfg := &config{File: vaultFile, Type: vaultType}

	// create dir if not existing
	if _, err = os.Stat(cfgDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(cfgDir, 0755); err != nil {
				return nil, err
			}
		}
	}

	if _, err = os.Stat(cfgFile); err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return nil, err
	}

	b, err := os.ReadFile(cfgFile)
	if err != nil {
		return nil, err
	}

	if err := yaml.Unmarshal(b, &cfg); err != nil {
		return nil, err
	}

	// override saved state with given params, if any.
	if vaultFile != "" {
		cfg.File = vaultFile
	}

	if vaultType != "" {
		cfg.Type = vaultType
	}

	fmt.Printf("Open file: %s (%s)\n", cfg.File, cfg.Type)

	return cfg, nil
}

func (c *config) save() error {
	b, err := yaml.Marshal(c)
	if err != nil {
		log.Fatal("error saving config:", err)
	}

	if err := os.WriteFile(cfgFile, b, 0644); err != nil {
		log.Fatal("error saving config:", err)
	}

	return nil
}
