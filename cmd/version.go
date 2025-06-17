package cmd

import (
	"fmt"
	"runtime"
	"runtime/debug"
	"time"

	"github.com/spf13/cobra"
)

// These variables are set at build time and describe the version and build of the application
var (
	version   = "dev"
	commit    = "dev"
	buildTime = time.Now().Format("2006-01-02T15:04:05.000")
	builtBy   = "local"
	builtWith = runtime.Version()
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print Flakeguard version info",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf(
			"flakeguard version %s built with %s from commit %s at %s by %s\n",
			version,
			builtWith,
			commit,
			buildTime,
			builtBy,
		)
		foundGotestsumVersion, err := getGoTestSumVersion()
		if err != nil {
			logger.Error().Err(err).Msg("Failed to get gotestsum version")
			fmt.Printf("error getting gotestsum version\n")
		} else {
			fmt.Printf("gotestsum version: %s\n", foundGotestsumVersion)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

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
