package vaulttest

import (
	"os"
	"path/filepath"
	"testing"
)

func Seed(f *testing.F, dir string) {
	f.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		f.Fatal(err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}

		b, err := os.ReadFile(filepath.Join(dir, e.Name()))
		if err != nil {
			f.Fatal(err)
		}

		f.Add(b)
	}
}
