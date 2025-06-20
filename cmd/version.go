package cmd

import (
	"runtime"
	"runtime/debug"
	"time"
)

// These variables are set at build time and describe the version and build of the application
var (
	version   = "dev"
	commit    = "dev"
	buildTime = time.Now().Format("2006-01-02T15:04:05.000")
	builtBy   = "local"
	builtWith = runtime.Version()
)

// getGoTestSumVersion returns the version of gotestsum that we're using
func getGoTestSumVersion() (string, error) {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, dep := range info.Deps {
			if dep.Path == "gotest.tools/gotestsum" {
				return dep.Version, nil
			}
		}
	}
	return "unknown", nil
}
