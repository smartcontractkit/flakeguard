package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var RunCommand = &cobra.Command{
	Use:   "run",
	Short: "Flakeguard will run your tests to detect any flaky tests, and can re-run them to stop them from disrupting your CI pipeline.",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Hello, World!")
	},
}
