package buildinfo

import (
	"fmt"
	"runtime/debug"
)

// build vars
var (
	AppName   = "andcli"
	Commit    = "none"
	BuildDate = "right now"
)

// Returns a formatted build info string
func Long() string {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return fmt.Sprintf(
			"%s (%s): error reading debug information",
			AppName, Commit,
		)
	}

	return fmt.Sprintf(
		"%s %s built on %s, %s",
		AppName, info.Main.Version, BuildDate, info.GoVersion,
	)
}

func Short() string {
	version := "?"
	if info, ok := debug.ReadBuildInfo(); ok {
		version = info.Main.Version
	}

	return fmt.Sprintf("%s %s", AppName, version)
}
