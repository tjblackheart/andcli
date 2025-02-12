package clipboard

import (
	"reflect"
	"testing"
)

func TestNew(t *testing.T) {

	tests := []struct {
		name, arg string
		want      *Clipboard
	}{
		{"inits user", "test", &Clipboard{cmd: "test", args: []string{}}},
		{"parses user args", "test -a -b -c", &Clipboard{cmd: "test", args: []string{"-a", "-b", "-c"}}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.arg); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClipboard_IsInitialized(t *testing.T) {

	tests := []struct {
		name string
		cb   *Clipboard
		want bool
	}{
		{"true if set", &Clipboard{cmd: "test"}, true},
		{"false if unset", &Clipboard{cmd: ""}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cb.IsInitialized(); got != tt.want {
				t.Errorf("Clipboard.IsInitialized() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClipboard_String(t *testing.T) {
	tests := []struct {
		name string
		cb   *Clipboard
		want string
	}{
		{
			"stringer #1",
			&Clipboard{cmd: "test"},
			"test",
		},
		{
			"stringer #2",
			&Clipboard{cmd: "/usr/bin/test", args: []string{"-a", "-b"}},
			"/usr/bin/test -a -b",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cb.String(); got != tt.want {
				t.Errorf("Clipboard.String() = %q, want %q", got, tt.want)
			}
		})
	}
}
