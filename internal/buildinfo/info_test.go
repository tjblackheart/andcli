package buildinfo

import (
	"fmt"
	"testing"
)

func TestFull(t *testing.T) {
	AppVersion = "2.0.0-test"
	Commit = "test"
	GoVersion = "1.x"

	tests := []struct{ name, want string }{
		{
			"prints full info",
			fmt.Sprintf(
				"%s %s (%s) built %s, %s",
				AppName, AppVersion, Commit, BuildDate, GoVersion,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Long(); got != tt.want {
				t.Errorf("Full() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSimple(t *testing.T) {
	tests := []struct{ name, want string }{
		{"prints simple info", fmt.Sprintf("%s %s", AppName, AppVersion)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Short(); got != tt.want {
				t.Errorf("Simple() = %v, want %v", got, tt.want)
			}
		})
	}
}
