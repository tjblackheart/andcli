package buildinfo

import (
	"fmt"
	"runtime/debug"
	"strings"
)

// build vars
var (
	AppName    = "andcli"
	BuildDate  = ""
	AppVersion = ""
	Commit     = ""
)

// Returns a formatted build info string
func Long() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fmt.Sprintf("%s: error reading build information", AppName)
	}

	for _, kv := range info.Settings {
		if kv.Key == "vcs.revision" && Commit == "" {
			Commit = kv.Value[:6]
		}

		if kv.Key == "vcs.time" && BuildDate == "" {
			BuildDate = kv.Value
		}
	}

	parts := []string{AppName, " ", AppVersion}
	if AppVersion == "" {
		parts[2] = info.Main.Version
	}

	if Commit != "" {
		parts = append(parts, " ", fmt.Sprintf("(%s)", Commit))
	}

	if BuildDate != "" {
		parts = append(parts, " ", fmt.Sprintf("built at %s", BuildDate))
	}

	parts = append(parts, ", ", info.GoVersion)

	return strings.Join(parts, "")
}

func Short() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		if AppVersion == "" {
			AppVersion = info.Main.Version
		}
	}

	return fmt.Sprintf("%s %s", AppName, AppVersion)
}
