package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var guardCmd = &cobra.Command{
	Use:   "guard [flakeguard flags] -- [gotestsum flags] -- [go test flags]",
	Short: "Guard your tests",
	Long: `Guard your CI/CD pipeline by running your tests multiple times and retrying them if they fail.

Examples:
  flakeguard guard -- --format testname -- ./pkg/...
  flakeguard guard --runs 10 -- --format dots -- -v -run TestMyFunction`,
	RunE: guardTests,
}

func init() {
	rootCmd.AddCommand(guardCmd)
}

func guardTests(_ *cobra.Command, args []string) error {
	return fmt.Errorf("guard not implemented")
}
