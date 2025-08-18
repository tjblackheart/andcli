package buildinfo

import (
	"fmt"
	"runtime/debug"
)

// build vars
var (
	AppName    = "andcli"
	BuildDate  = ""
	AppVersion = ""
	Commit     = ""
	GoVersion  = ""
)

// Returns a formatted build info string
func Long() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fmt.Sprintf("%s: error reading debug information", AppName)
	}

	if GoVersion == "" {
		GoVersion = info.GoVersion
	}

	if AppVersion == "" {
		AppVersion = info.Main.Version
	}

	for _, kv := range info.Settings {
		if kv.Key == "vcs.revision" && Commit == "" {
			Commit = kv.Value[:6]
		}

		if kv.Key == "vcs.time" && BuildDate == "" {
			BuildDate = kv.Value
		}
	}

	return fmt.Sprintf(
		"%s %s (%s) built %s, %s",
		AppName, AppVersion, Commit, BuildDate, GoVersion,
	)
}

func Short() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		if AppVersion == "" {
			AppVersion = info.Main.Version
		}
	}

	return fmt.Sprintf("%s %s", AppName, AppVersion)
}
