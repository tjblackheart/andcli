package clipboard

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"slices"
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

func TestDetectLinux(t *testing.T) {
	tools := []string{"termux-clipboard-set", "wl-copy", "xclip", "xsel"}
	tests := []struct {
		name  string
		env   map[string]string
		chain []string
	}{
		{
			"no display",
			map[string]string{"TERMUX_VERSION": "", "WAYLAND_DISPLAY": "", "DISPLAY": ""},
			[]string{"wl-copy", "xclip", "xsel"},
		},
		{
			"x11",
			map[string]string{"DISPLAY": ":0"},
			[]string{"xclip", "xsel"},
		},
		{
			"wayland",
			map[string]string{"WAYLAND_DISPLAY": "wayland-0"},
			[]string{"wl-copy"},
		},
		{
			"wayland and x11",
			map[string]string{"WAYLAND_DISPLAY": "wayland-0", "DISPLAY": ":0"},
			[]string{"wl-copy", "xclip", "xsel"},
		},
		{
			"termux",
			map[string]string{"TERMUX_VERSION": "1"},
			[]string{"termux-clipboard-set", "wl-copy", "xclip", "xsel"},
		},
		{
			"termux and x11",
			map[string]string{"TERMUX_VERSION": "1", "DISPLAY": ":0"},
			[]string{"termux-clipboard-set", "xclip", "xsel"},
		},
		{
			"termux and wayland",
			map[string]string{"TERMUX_VERSION": "1", "WAYLAND_DISPLAY": "wayland-0"},
			[]string{"termux-clipboard-set", "wl-copy"},
		},
		{
			"termux wayland and x11",
			map[string]string{"TERMUX_VERSION": "1", "WAYLAND_DISPLAY": "wayland-0", "DISPLAY": ":0"},
			[]string{"termux-clipboard-set", "wl-copy", "xclip", "xsel"},
		},
	}

	for _, tt := range tests {
		for mask := 0; mask < 1<<len(tools); mask++ {
			t.Run(fmt.Sprintf("%s/tools_%02d", tt.name, mask), func(t *testing.T) {
				available := make(map[string]bool, len(tools))
				for i, tool := range tools {
					available[tool] = mask&(1<<i) != 0
				}

				getenv := func(key string) string { return tt.env[key] }
				lookPath := func(name string) (string, bool) {
					if available[name] {
						return "/usr/bin/" + name, true
					}
					return "", false
				}

				wantCmd := ""
				wantArgs := []string{}
				for _, tool := range tt.chain {
					if !available[tool] {
						continue
					}
					wantCmd = "/usr/bin/" + tool
					switch tool {
					case "xclip":
						wantArgs = []string{"-selection", "clipboard"}
					case "xsel":
						wantArgs = []string{"-b"}
					}
					break
				}

				gotCmd, gotArgs := detect("linux", getenv, lookPath)
				if gotCmd != wantCmd {
					t.Errorf("detect() cmd = %q, want %q", gotCmd, wantCmd)
				}
				if !slices.Equal(gotArgs, wantArgs) {
					t.Errorf("detect() args = %v, want %v", gotArgs, wantArgs)
				}
			})
		}
	}
}

func TestDetectOtherSystems(t *testing.T) {
	tests := []struct {
		goos string
		tool string
	}{
		{"darwin", "pbcopy"},
		{"windows", "clip.exe"},
		{"android", "termux-clipboard-set"},
	}

	for _, tt := range tests {
		t.Run(tt.goos, func(t *testing.T) {
			getenv := func(string) string { return "" }
			lookPath := func(name string) (string, bool) {
				return "/usr/bin/" + name, name == tt.tool
			}

			cmd, args := detect(tt.goos, getenv, lookPath)
			want := "/usr/bin/" + tt.tool
			if cmd != want {
				t.Errorf("detect() cmd = %q, want %q", cmd, want)
			}
			if len(args) != 0 {
				t.Errorf("detect() args = %v, want none", args)
			}

			cmd, args = detect(tt.goos, getenv, func(string) (string, bool) { return "", false })
			if cmd != "" || len(args) != 0 {
				t.Errorf("detect() with missing tool = (%q, %v), want uninitialized", cmd, args)
			}
		})
	}

	t.Run("unsupported", func(t *testing.T) {
		cmd, args := detect(
			"plan9",
			func(string) string { return "" },
			func(name string) (string, bool) { return "/usr/bin/" + name, true },
		)
		if cmd != "" || len(args) != 0 {
			t.Errorf("detect() = (%q, %v), want uninitialized", cmd, args)
		}
	})
}

func TestClipboard_Set(t *testing.T) {
	t.Setenv("ANDCLI_CLIPBOARD_HELPER", "1")
	sink := filepath.Join(t.TempDir(), "clipboard")
	cb := Clipboard{
		cmd:  os.Args[0],
		args: []string{"-test.run=^TestClipboardHelperProcess$", "--", sink},
	}
	want := []byte("123456\n")

	if err := cb.Set(want); err != nil {
		t.Fatalf("Clipboard.Set() error = %v", err)
	}

	got, err := os.ReadFile(sink)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("Clipboard.Set() wrote %q, want %q", got, want)
	}
}

func TestClipboard_SetError(t *testing.T) {
	cb := Clipboard{cmd: filepath.Join(t.TempDir(), "missing")}
	if err := cb.Set([]byte("123456")); err == nil {
		t.Fatal("Clipboard.Set() error = nil, want an execution error")
	}
}

func TestClipboardHelperProcess(t *testing.T) {
	if os.Getenv("ANDCLI_CLIPBOARD_HELPER") != "1" {
		return
	}

	separator := -1
	for i, arg := range os.Args {
		if arg == "--" {
			separator = i
			break
		}
	}
	if separator < 0 || separator+1 >= len(os.Args) {
		os.Exit(2)
	}

	b, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(2)
	}
	if err := os.WriteFile(os.Args[separator+1], b, 0o600); err != nil {
		os.Exit(2)
	}
	os.Exit(0)
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
