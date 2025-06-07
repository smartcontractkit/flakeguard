package cmd

import (
	"log"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/smartcontractkit/flakeguard/logging"
)

var (
	logger zerolog.Logger

	// Flag vars
	outputDir string
)

var rootCmd = &cobra.Command{
	Use:   "flakeguard",
	Short: "Detect and prevent flaky tests from disrupting CI/CD pipelines",
	Long:  ``,
}

func init() {
	logger, err := logging.New(
		logging.WithLevel("info"),
		logging.WithFileName("flakeguard.log"),
	)
	if err != nil {
		log.Fatal(err)
	}

	logger.Info().Msg("Flakeguard initialized")

	// Add flags
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
