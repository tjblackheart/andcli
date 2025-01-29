package buildinfo

import (
	"fmt"
	"testing"
)

func TestFull(t *testing.T) {
	tests := []struct{ name, want string }{
		{
			"prints full info",
			fmt.Sprintf(
				"%s %s (%s) built %s with Go %s",
				AppName, "test", Commit, BuildDate, "",
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
		{"prints simple info", fmt.Sprintf("%s %s", AppName, "AppVersion")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Short(); got != tt.want {
				t.Errorf("Simple() = %v, want %v", got, tt.want)
			}
		})
	}
}
