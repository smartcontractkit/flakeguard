package cmd

import "github.com/spf13/cobra"

var detectCmd = &cobra.Command{
	Use:   "detect",
	Short: "Detect flaky tests",
	RunE:  detectFlakyTests,
}

func init() {
	rootCmd.AddCommand(detectCmd)
}

func detectFlakyTests(cmd *cobra.Command, args []string) error {
	logger.Info().Msg("Detecting flaky tests")
	return nil
}
