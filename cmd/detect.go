package cmd

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/rs/zerolog"
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
		detectFile, err := runDetect(logger, run, originalGotestsumFlags, goTestFlags)
		if err != nil {
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
		report.WithDir(outputDir),
	)
	if err != nil {
		return fmt.Errorf("failed to create flakeguard report: %w", err)
	}

	return nil
}

// runDetect runs a single detect run.
func runDetect(
	l zerolog.Logger,
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
	l = l.With().
		Int("run", run+1).
		Str("detectResultsFile", detectFile).
		Strs("gotestsumFlags", originalGotestsumFlags).
		Strs("goTestFlags", goTestFlags).
		Logger()

	startDetectTime := time.Now()
	err = gotestsumCmd.Run("gotestsum", fullArgs)
	exitCode := getExitCode(err)
	l.Debug().Err(err).Int("exitCode", exitCode).Msg("Detect run completed")
	return detectFile, handleTestRunError(l, run, err, startDetectTime)
}

func init() {
	rootCmd.AddCommand(detectCmd)
	detectCmd.Flags().
		DurationVar(&durationTarget, "duration-target", 0, "Target duration for the full detection run. If set, detect will attempt to stop as soon as this duration is hit. This is a soft-limit, and will not abort in the middle of a run.")
}

// getExitCode extracts the exit code from an error returned by exec.Cmd
func getExitCode(err error) int {
	if err == nil {
		return 0
	}
	if exitError, ok := err.(*exec.ExitError); ok {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			return status.ExitStatus()
		}
	}
	return -1 // Unknown exit code
}

// handleTestRunError handles the error from a test run, exiting with the appropriate code.
func handleTestRunError(l zerolog.Logger, run int, err error, startRunTime time.Time) error {
	l = l.With().Int("run", run+1).Str("duration", time.Since(startRunTime).String()).Logger()
	if err == nil {
		l.Info().
			Msg("All tests passed!")
		return nil
	}

	// Check if it's an exit error (command ran but exited with non-zero code)
	if exitError, ok := err.(*exec.ExitError); ok {
		if status, ok := exitError.Sys().(syscall.WaitStatus); ok {
			exitCode := status.ExitStatus()

			switch exitCode {
			case exit.CodeSuccess:
				l.Info().Msg("All tests passed")
				return nil

			case exit.CodeGoFailingTest:
				l.Warn().Err(err).Int("exit_code", exitCode).Msg("Found flaky tests")
				// Test failures are expected in flaky test detection, so we continue
				return nil

			case exit.CodeGoBuildError:
				l.Error().
					Err(err).
					Int("exit_code", exitCode).
					Msg("Build/compilation error")
				// Build errors are serious and should stop the detection process
				return exit.New(exit.CodeGoBuildError, fmt.Errorf("build error on run %d: %w", run+1, err))

			default:
				l.Error().
					Err(err).
					Int("exit_code", exitCode).
					Msg("Unexpected exit code")
				// For other exit codes, log but continue
				return nil
			}
		}
	}

	// For other types of errors (like command not found), return the error
	l.Error().Err(err).Msg("Flakeguard encountered an error")
	return exit.New(exit.CodeFlakeguardError, fmt.Errorf("flakeguard encountered an error on run %d: %w", run+1, err))
}
