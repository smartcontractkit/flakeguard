package cmd

import (
	"fmt"
	"os"
	"os/exec"
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
	Use:   "flakeguard",
	Short: "Detect and prevent flaky tests from disrupting CI/CD pipelines",
	Long:  ``,
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

	// Allow unknown flags to be passed through to gotestsum
	rootCmd.SetFlagErrorFunc(func(cmd *cobra.Command, err error) error {
		return nil
	})
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Fatal().Err(err).Msg("Failed to execute command")
	}
}

func runFlakeguard(cmd *cobra.Command, args []string) error {
	gotestsumArgs := []string{"tool", "gotestsum", "--jsonfile", testOutputFile}
	gotestsumArgs = append(gotestsumArgs, args...)

	// Print the full command that will be executed
	logger.Info().Msgf("Running command: go %s", strings.Join(gotestsumArgs, " "))

	//nolint:gosec // G204 we need to call out to gotestsum
	gotestsumCmd := exec.Command("go", gotestsumArgs...)
	gotestsumCmd.Stdout = os.Stdout
	gotestsumCmd.Stderr = os.Stderr
	err := gotestsumCmd.Run()
	if err != nil {
		return fmt.Errorf("gotestsum failed: %w", err)
	}
	return nil
}
