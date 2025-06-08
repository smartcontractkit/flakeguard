package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

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
	logger.Info().Msg("Detecting flaky tests")
	// We're intentionally not doing these runs in parallel, we're also not using the native -count flag to run the tests multiple times.
	// We want to model real-world behavior as closely as possible, and -count is not a good model for that.
	// We only use -count=1 to disable test caching.
	gotestsumFlags, goTestFlags := parseArgs(args)
	goTestFlags = append(goTestFlags, "-count=1")
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

	err := gotestsumCmd.Run()
	return handleTestRunError(run, err)
}

func handleTestRunError(run int, err error) error {
	if err == nil {
		logger.Debug().Msgf("Run %d: All tests passed", run+1)
		return nil
	}

	// Check if it's an exit error (command ran but exited with non-zero code)
	if exitError, ok := err.(*exec.ExitError); ok {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			exitCode := status.ExitStatus()

			switch exitCode {
			case ExitCodeSuccess:
				logger.Info().Int("run", run+1).Msg("All tests passed")
				return nil

			case ExitCodeGoFailingTest:
				logger.Info().
					Int("run", run+1).
					Int("exit_code", exitCode).
					Msg("Some tests failed - continuing detection")
				// Test failures are expected in flaky test detection, so we continue
				return nil

			case ExitCodeGoBuildError:
				logger.Error().
					Int("run", run+1).
					Int("exit_code", exitCode).
					Msg("Build/compilation error - stopping detection")
				// Build errors are serious and should stop the detection process
				return NewExitError(ExitCodeGoBuildError, fmt.Errorf("build error on run %d: %w", run+1, err))

			default:
				logger.Warn().
					Int("run", run+1).
					Int("exit_code", exitCode).
					Msg("Unexpected exit code - continuing detection")
				// For other exit codes, log but continue
				return nil
			}
		}
	}

	// For other types of errors (like command not found), return the error
	logger.Error().Int("run", run+1).Err(err).Msg("Command execution error")
	return NewExitError(ExitCodeFlakeguardError, fmt.Errorf("command execution failed on run %d: %w", run+1, err))
}
