package buildinfo

import (
	"fmt"
	"runtime/debug"
	"testing"
	"time"
)

func TestFull(t *testing.T) {

	AppVersion = "2.0.0-test"
	Commit = "123456"
	BuildDate = time.Now().Format(time.RFC3339)

	info, ok := debug.ReadBuildInfo()
	if !ok {
		t.Fatal("could not read debug info")
	}

	tests := []struct{ name, want string }{
		{
			"prints full info",
			fmt.Sprintf(
				"%s %s (%s) built at %s, %s",
				AppName, AppVersion, Commit, BuildDate, info.GoVersion,
			),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Long(); got != tt.want {
				t.Errorf("Full() = %v want %v", got, tt.want)
			}
		})
	}
}

func TestSimple(t *testing.T) {
	AppVersion = "2.0.0-test"

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
