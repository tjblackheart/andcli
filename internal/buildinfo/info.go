package buildinfo

import (
	"fmt"
	"runtime/debug"
)

// build vars
var (
	AppName    = "andcli"
	Commit     = "dev"
	BuildDate  = "now"
	AppVersion = "v2.0.x"
	GoVersion  = ""
)

// Returns a formatted build info string
func Long() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fmt.Sprintf(
			"%s (%s): error reading debug information",
			AppName, AppVersion,
		)
	}

	if info.Main.Version != "" {
		AppVersion = info.Main.Version
	}

	if GoVersion == "" {
		GoVersion = info.GoVersion
	}

	return fmt.Sprintf(
		"%s %s (%s) built %s, %s",
		AppName, AppVersion, Commit, BuildDate, GoVersion,
	)
}

func Short() string {
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" {
		AppVersion = info.Main.Version
	}

	return fmt.Sprintf("%s %s", AppName, AppVersion)
}
