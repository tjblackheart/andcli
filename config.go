package main

import (
	"fmt"
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type config struct {
	File string `yaml:"file"`
	Type string `yaml:"type"`
}

func configFromFile(vaultFile, vaultType string) *config {
	var c = &config{File: vaultFile, Type: vaultType}

	// create dir if not existing
	if _, err := os.Stat(cfgDir); err != nil {
		if os.IsNotExist(err) {
			if err := os.MkdirAll(cfgDir, 0755); err != nil {
				log.Fatal("create config dir:", err)
			}
		}
	}

	if _, err := os.Stat(cfgFile); err != nil {
		if os.IsNotExist(err) {
			return c
		}
		log.Fatal(err)
	}

	b, err := os.ReadFile(cfgFile)
	if err != nil {
		log.Fatal(err)
	}

	if err := yaml.Unmarshal(b, &c); err != nil {
		log.Fatal(err)
	}

	if vaultFile != "" {
		c.File = vaultFile
	}

	if vaultType != "" {
		c.Type = vaultType
	}

	fmt.Printf("Open file: %s (%s)\n", c.File, c.Type)

	return c
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
