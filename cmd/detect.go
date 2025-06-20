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
		Strs("entered_gotestsum_flags", originalGotestsumFlags).
		Strs("entered_go_test_flags", goTestFlags).
		Strs("entered_args", args).
		Msg("Detecting flaky tests")
	fmt.Println("Detecting flaky tests")

	if slices.Contains(originalGotestsumFlags, "--jsonfile") {
		return fmt.Errorf("jsonfile flag cannot be overridden while using flakeguard")
	}

	// Intentionally set -count=1 to avoid caching test results during detection
	for _, flag := range goTestFlags {
		if strings.HasPrefix(flag, "-count=") {
			return fmt.Errorf("-count flag in go test cannot be overridden while using flakeguard")
		}
	}
	goTestFlags = append(goTestFlags, "-count=1")

	testRunInfo, err := testRunInfo(logger, githubClient, ".")
	if err != nil {
		return fmt.Errorf("failed to get test run info: %w", err)
	}

	startTime := time.Now()
	detectFiles := []string{}
	for run := range runs {
		detectFile, err := runDetect(logger, run, originalGotestsumFlags, goTestFlags)
		if err != nil {
			return err
		}
		detectFiles = append(detectFiles, detectFile)
		if durationTarget > 0 {
			if time.Since(startTime) > durationTarget {
				logger.Warn().
					Str("duration_target", durationTarget.String()).
					Str("elapsed_time", time.Since(startTime).String()).
					Msg("Duration target hit, stopping detection")
				break
			}
		}
	}

	err = report.New(
		logger,
		testRunInfo,
		detectFiles,
		report.WithDir(outputDir),
	)
	if err != nil {
		return err
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
		Str("detect_results_file", detectFile).
		Strs("gotestsum_flags", originalGotestsumFlags).
		Strs("go_test_flags", goTestFlags).
		Logger()

	startDetectTime := time.Now()
	err = gotestsumCmd.Run("gotestsum", fullArgs)
	l.Debug().Err(err).Str("duration", time.Since(startDetectTime).String()).Msg("Detect run completed")
	if err != nil {
		exitCode := getExitCode(err)
		if exitCode != 1 { // Exit code 1 is expected when there are flaky tests
			return detectFile, exit.New(exitCode, err)
		}
	}
	return detectFile, nil
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
