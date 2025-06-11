package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var guardCmd = &cobra.Command{
	Use:   "guard [flakeguard flags] -- [gotestsum flags] -- [go test flags]",
	Short: "Guard your tests",
	Long: `Guard your tests by running them multiple times and retrying them if they fail.

Examples:
  flakeguard guard -- --format testname -- ./pkg/...
  flakeguard guard --runs 10 -- --format dots -- -v -run TestMyFunction`,
	RunE: guardTests,
}

func init() {
	rootCmd.AddCommand(guardCmd)
}

func guardTests(_ *cobra.Command, args []string) error {
	fullArgs := []string{
		"tool",
		"gotestsum",
		"--jsonfile",
		fmt.Sprintf("%s/guard.json", outputDir),
		fmt.Sprintf("--rerun-fails=%d", runs),
		"--packages=./...",
	}

	gotestsumFlags, goTestFlags := parseArgs(args)

	// Add gotestsum flags first
	fullArgs = append(fullArgs, gotestsumFlags...)

	// Add go test flags if present
	if len(goTestFlags) > 0 {
		fullArgs = append(fullArgs, "--")
		fullArgs = append(fullArgs, goTestFlags...)
	}

	command := fmt.Sprintf("go %s", strings.Join(fullArgs, " "))
	fmt.Println(command)
	logger.Info().Msgf("Running command: %s", command)

	//nolint:gosec // G204 we need to call out to gotestsum
	gotestsumCmd := exec.Command("go", fullArgs...)
	gotestsumCmd.Stdout = os.Stdout
	gotestsumCmd.Stderr = os.Stderr
	return gotestsumCmd.Run()

}
