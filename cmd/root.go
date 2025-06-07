package cmd

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/smartcontractkit/flakeguard/logging"
)

var (
	logger zerolog.Logger

	// Flag vars
	logFile           string
	logLevel          string
	enableConsoleLogs bool

	testOutputFile string
	runs           int
	outputDir      string
)

var rootCmd = &cobra.Command{
	Use:   "flakeguard [flags] [-- gotestsum-flags] [-- go-test-flags]",
	Short: "Detect and prevent flaky tests from disrupting CI/CD pipelines",
	Long: `Flakeguard wraps gotestsum to detect and prevent flaky tests.

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
			return fmt.Errorf("failed to create logger: %w", err)
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
			Str("testOutputFile", testOutputFile).
			Int("runs", runs).
			Str("outputDir", outputDir).
			Msg("Run info")
		return nil
	},
	RunE: runFlakeguard,
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&logFile, "log-file", "l", "flakeguard.log", "File to store flakeguard logs")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "L", "info", "Log level to use")
	rootCmd.PersistentFlags().
		BoolVarP(&enableConsoleLogs, "enable-console-logs", "c", false, "Enable console logs for flakeguard")

	rootCmd.PersistentFlags().
		StringVarP(&testOutputFile, "test-output-file", "t", "test-output.json", "File to store test output")
	rootCmd.PersistentFlags().IntVarP(&runs, "runs", "r", 5, "Number of times to run the tests")
	rootCmd.PersistentFlags().
		StringVarP(&outputDir, "output-dir", "o", "./flakeguard-reports", "Directory to store flakeguard reports")

	// Disable flag parsing after -- to allow passing through to gotestsum
	rootCmd.Flags().SetInterspersed(false)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error().Err(err).Msg("Failed to execute command")
		os.Exit(1)
	}
}

func runFlakeguard(cmd *cobra.Command, args []string) error {
	fullArgs := []string{"tool", "gotestsum", "--jsonfile", testOutputFile}

	command := fmt.Sprintf("go %s", strings.Join(fullArgs, " "))
	fmt.Println(command)
	logger.Info().Msgf("Running command: %s", command)

	//nolint:gosec // G204 we need to call out to gotestsum
	gotestsumCmd := exec.Command("go", fullArgs...)
	gotestsumCmd.Stdout = os.Stdout
	gotestsumCmd.Stderr = os.Stderr
	return gotestsumCmd.Run()
}
