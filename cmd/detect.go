package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/spf13/cobra"
	gotestsumCmd "gotest.tools/gotestsum/cmd"

	"github.com/smartcontractkit/flakeguard/exit"
	"github.com/smartcontractkit/flakeguard/report"
)

const detectFileOutput = "detect-test-output-%d.json"

var (
	// Detect specific flags
	durationTarget time.Duration
)

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect flaky tests",
	Long: `Detect flaky tests by running the full test suites multiple times under the same conditions.

Test results are analyzed to determine which tests are flaky, and results are reported to various destinations, if configured.`,
	RunE: runDetectCmd,
}

func runDetectCmd(_ *cobra.Command, args []string) error {
	originalGotestsumFlags, goTestFlags := parseArgs(args)
	logger.Info().
		Int("runs", runs).
		Strs("enteredGotestsumFlags", originalGotestsumFlags).
		Strs("enteredGoTestFlags", goTestFlags).
		Strs("enteredArgs", args).
		Msg("Detecting flaky tests")

	if slices.Contains(originalGotestsumFlags, "--jsonfile") {
		return fmt.Errorf("jsonfile flag cannot be overridden while using flakeguard")
	}

	for _, flag := range goTestFlags {
		if strings.HasPrefix(flag, "-count=") {
			return fmt.Errorf("-count flag in go test cannot be overridden while using flakeguard")
		}
	}

	// Intentionally set -count=1 to avoid caching test results
	goTestFlags = append(goTestFlags, "-count=1")
	detectFiles := make([]string, 0, runs)
	startTime := time.Now()
	for run := range runs {
		detectFile, err := runDetect(run, originalGotestsumFlags, goTestFlags)
		if err := handleTestRunError(run, err); err != nil {
			return err
		}
		detectFiles = append(detectFiles, detectFile)
		if durationTarget > 0 {
			if time.Since(startTime) > durationTarget {
				logger.Warn().
					Str("durationTarget", durationTarget.String()).
					Str("elapsedTime", time.Since(startTime).String()).
					Msg("Duration target hit, stopping detection")
				break
			}
		}
	}

	testRunInfo, err := testRunInfo(logger, githubClient, ".")
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
		return fmt.Errorf("failed to create flakeguard report: %w", err)
	}

	return nil
}

// runDetect runs a single detect run.
func runDetect(
	run int,
	originalGotestsumFlags []string,
	goTestFlags []string,
) (detectFile string, err error) {
	detectFile = fmt.Sprintf(detectFileOutput, run+1)
	//nolint:gocritic // The slice appends are needed to avoid modifying the original slices
	gotestsumFlags := append(originalGotestsumFlags, "--jsonfile", filepath.Join(outputDir, detectFile))
	//nolint:gocritic // The slice appends are needed to avoid modifying the original slices
	fullArgs := append(gotestsumFlags, "--")
	fullArgs = append(fullArgs, goTestFlags...)
	logger.Info().
		Strs("gotestsumFlags", gotestsumFlags).
		Strs("goTestFlags", goTestFlags).
		Strs("fullArgs", fullArgs).
		Int("run", run+1).
		Msg("Detect run")

	return detectFile, gotestsumCmd.Run("gotestsum", fullArgs)
}

func init() {
	rootCmd.AddCommand(detectCmd)
	detectCmd.Flags().
		DurationVar(&durationTarget, "duration-target", 0, "Target duration for the full detection run. If set, detect will attempt to stop as soon as this duration is hit. This is a soft-limit, and will not abort in the middle of a run.")
}

// handleTestRunError handles the error from a test run, exiting with the appropriate code.
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
