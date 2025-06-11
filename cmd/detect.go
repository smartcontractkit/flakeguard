package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"syscall"

	"github.com/spf13/cobra"

	"github.com/smartcontractkit/flakeguard/exit"
	"github.com/smartcontractkit/flakeguard/report"
)

const detectFileOutput = "%s/detect-%d.json"

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect flaky tests",
	Long: `Detect flaky tests by running the full test suites multiple times under the same conditions.

Test results are analyzed to determine which tests are flaky, and results are reported to various destinations, if configured.`,
	RunE: detectFlakyTests,
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
	detectFiles := make([]string, 0, runs)
	for run := range runs {
		run := run + 1
		logger.Info().Msgf("Running flake detection %d of %d", run, runs)
		detectFile, err := runDetect(run, gotestsumFlags, goTestFlags)
		if err != nil {
			return fmt.Errorf("failed to run test %d of %d: %w", run, runs, err)
		}
		detectFiles = append(detectFiles, detectFile)
	}

	testRunInfo, err := testRunInfo(logger, ".")
	if err != nil {
		return fmt.Errorf("failed to get test run info: %w", err)
	}

	err = report.New(
		logger,
		testRunInfo,
		detectFiles,
		report.ReportDir(outputDir),
	)
	if err != nil {
		return fmt.Errorf("failed to create report: %w", err)
	}

	return nil
}

func runDetect(run int, gotestsumFlags []string, goTestFlags []string) (string, error) {
	fullArgs := []string{"tool", "gotestsum", "--jsonfile", fmt.Sprintf(detectFileOutput, outputDir, run)}

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

	gotestsumCmd := exec.Command("go", fullArgs...)
	gotestsumCmd.Stdout = os.Stdout
	gotestsumCmd.Stderr = os.Stderr

	err := gotestsumCmd.Run()
	return fmt.Sprintf(detectFileOutput, outputDir, run), handleTestRunError(run, err)
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
			case exit.CodeSuccess:
				logger.Info().Int("run", run+1).Msg("All tests passed")
				return nil

			case exit.CodeGoFailingTest:
				logger.Info().
					Int("run", run+1).
					Int("exit_code", exitCode).
					Msg("Found flaky tests")
				// Test failures are expected in flaky test detection, so we continue
				return nil

			case exit.CodeGoBuildError:
				logger.Error().
					Err(err).
					Int("run", run+1).
					Int("exit_code", exitCode).
					Msg("Build/compilation error")
				// Build errors are serious and should stop the detection process
				return exit.New(exit.CodeGoBuildError, fmt.Errorf("build error on run %d: %w", run+1, err))

			default:
				logger.Warn().
					Err(err).
					Int("run", run+1).
					Int("exit_code", exitCode).
					Msg("Unexpected exit code")
				// For other exit codes, log but continue
				return nil
			}
		}
	}

	// For other types of errors (like command not found), return the error
	logger.Error().Int("run", run+1).Err(err).Msg("Flakeguard encountered an error")
	return exit.New(exit.CodeFlakeguardError, fmt.Errorf("flakeguard encountered an error on run %d: %w", run+1, err))
}
