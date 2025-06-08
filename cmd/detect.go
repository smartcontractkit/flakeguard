package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cobra"
)

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect flaky tests",
	RunE:  detectFlakyTests,
}

func init() {
	rootCmd.AddCommand(detectCmd)
}

func detectFlakyTests(_ *cobra.Command, args []string) error {
	logger.Info().Msgf("Detecting flaky tests")
	gotestsumFlags, goTestFlags := parseArgs(args)
	goTestFlags = append(goTestFlags, "-count=1")
	// We're intentionally not doing these runs in parallel, we're also not using the native -count flag to run the tests multiple times.
	// We want to model real-world behavior as closely as possible, and -count is not a good model for that.
	// We only use -count=1 to disable test caching.
	for run := range runs {
		logger.Info().Msgf("Running flake detection %d of %d", run+1, runs)
		if err := runDetect(run, gotestsumFlags, goTestFlags); err != nil {
			return fmt.Errorf("failed to run test %d of %d: %w", run+1, runs, err)
		}
	}
	return nil
}

func runDetect(run int, gotestsumFlags []string, goTestFlags []string) error {
	fullArgs := []string{"tool", "gotestsum", "--jsonfile", fmt.Sprintf("%s/detect-%d.json", outputDir, run)}

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
