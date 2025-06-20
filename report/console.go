package report

import (
	"fmt"
	"strings"
)

func writeToConsole(summary *reportSummary, results []*TestResult) error {
	summaryStr := summary.String()
	fmt.Println(strings.Repeat("-", len(summaryStr)))
	fmt.Println(summaryStr)
	fmt.Println(strings.Repeat("-", len(summaryStr)))

	for _, result := range results {
		if result.Failures > 0 || result.Panic {
			fmt.Println(result.String())
		}
	}

	return nil
}
