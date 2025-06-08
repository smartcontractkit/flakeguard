package cmd

import (
	"fmt"
	"os"
	"runtime"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/smartcontractkit/flakeguard/logging"
)

const (
	ExitCodeSuccess         = 0
	ExitCodeGoFailingTest   = 1
	ExitCodeGoBuildError    = 2
	ExitCodeFlakeguardError = 3
)

// ExitError represents an error with a specific exit code
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string {
	if e.Err != nil {
		return e.Err.Error()
	}
	return fmt.Sprintf("exit code %d", e.Code)
}

func (e *ExitError) Unwrap() error {
	return e.Err
}

// NewExitError creates a new ExitError with the given code and error
func NewExitError(code int, err error) *ExitError {
	return &ExitError{Code: code, Err: err}
}

var (
	logger zerolog.Logger

	// Flag vars
	logFile           string
	logLevel          string
	enableConsoleLogs bool

	runs      int
	outputDir string
)

var rootCmd = &cobra.Command{
	Use:   "flakeguard [flags] [-- gotestsum-flags] [-- go-test-flags]",
	Short: "Detect and prevent flaky tests from disrupting CI/CD pipelines",
	Long: `Flakeguard helps you detect and prevent flaky tests from disrupting CI/CD pipelines.
It wraps gotestsum, so you can pass through all the flags you're used to using.

Examples:
  flakeguard -c -- --format testname -- ./pkg/...
  flakeguard --runs 10 -- --format dots -- -v -run TestMyFunction`,
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		loggingOpts := []logging.Option{}
		if !enableConsoleLogs {
			loggingOpts = append(loggingOpts, logging.DisableConsoleLog())
		}
		if logLevel != "" {
			loggingOpts = append(loggingOpts, logging.WithLevel(logLevel))
		}
		if logFile != "" {
			loggingOpts = append(loggingOpts, logging.WithFileName(logFile))
		}

		var err error
		logger, err = logging.New(loggingOpts...)
		if err != nil {
			return NewExitError(ExitCodeFlakeguardError, err)
		}
		if err := os.MkdirAll(outputDir, 0750); err != nil {
			return NewExitError(ExitCodeFlakeguardError, err)
		}
		logger.Debug().
			Str("version", version).
			Str("commit", commit).
			Str("buildTime", buildTime).
			Str("builtBy", builtBy).
			Str("builtWith", builtWith).
			Str("goVersion", runtime.Version()).
			Str("os", runtime.GOOS).
			Str("arch", runtime.GOARCH).
			Str("logFile", logFile).
			Str("logLevel", logLevel).
			Bool("enableConsoleLogs", enableConsoleLogs).
			Int("runs", runs).
			Str("outputDir", outputDir).
			Msg("Run info")
		return nil
	},
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&logFile, "log-file", "l", "flakeguard.log", "File to store flakeguard logs")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "L", "info", "Log level to use")
	rootCmd.PersistentFlags().
		BoolVarP(&enableConsoleLogs, "enable-console-logs", "c", false, "Enable console logs for flakeguard")

	rootCmd.PersistentFlags().
		IntVarP(&runs, "runs", "r", 5, "Number of times to run each test in detect mode, or the number of times to retry a test in guard mode")
	rootCmd.PersistentFlags().
		StringVarP(&outputDir, "output-dir", "o", "./flakeguard-reports", "Directory to store flakeguard reports")

	// Disable flag parsing after -- to allow passing through to gotestsum
	rootCmd.Flags().SetInterspersed(false)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error().Err(err).Msg("Failed to execute command")
		os.Exit(getExitCode(err))
	}
}

func getExitCode(err error) int {
	// Check if it's a custom error with exit code
	if exitErr, ok := err.(*ExitError); ok {
		return exitErr.Code
	}

	// Default to flakeguard error for unknown errors
	return ExitCodeFlakeguardError
}

// parseArgs parses command line arguments to separate gotestsum flags from go test flags.
// flakeguard [flakeguard flags] -- [gotestsum flags] -- [go test flags]
func parseArgs(args []string) (gotestsumFlags []string, goTestFlags []string) {
	// Find the positions of the double dashes
	firstDashPos := -1
	secondDashPos := -1

	for i, arg := range args {
		if arg == "--" {
			if firstDashPos == -1 {
				firstDashPos = i
			} else if secondDashPos == -1 {
				secondDashPos = i
				break
			}
		}
	}

	if firstDashPos != -1 {
		if secondDashPos != -1 {
			// Both dashes present: -- <gotestsum flags> -- <go test flags>
			gotestsumFlags = args[firstDashPos+1 : secondDashPos]
			goTestFlags = args[secondDashPos+1:]
		} else {
			// Only first dash present: -- <gotestsum flags>
			gotestsumFlags = args[firstDashPos+1:]
		}
	}

	return gotestsumFlags, goTestFlags
}
