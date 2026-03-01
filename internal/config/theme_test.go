package config

import (
	"bytes"
	"log"
	"strings"
	"testing"
)

func TestTheme_validate(t *testing.T) {
	tests := []struct {
		name         string
		theme        *Theme
		expectedLogs string
		fails        bool
	}{
		{
			"valid theme",
			&DefaultTheme,
			"",
			false,
		},
		{
			"invalid theme",
			&Theme{
				Base:   "xxx",
				Green:  "yyy",
				Yellow: "",
				Red:    "#ff0000",
				White:  "#fff",
				Grey:   "§$%a",
				Black:  "000",
			},
			"Invalid fields: base, green, yellow, grey",
			true,
		},
	}

	buf := new(bytes.Buffer)
	log.SetOutput(buf)

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.theme.validate()
			if tt.fails && !strings.Contains(buf.String(), tt.expectedLogs) {
				t.Errorf("theme.validate(): want %q, got %q", tt.expectedLogs, buf.String())
			}
		})
		buf.Reset()
	}
}
