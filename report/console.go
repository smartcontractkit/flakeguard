package report

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

func writeToConsole(l zerolog.Logger, summary *reportSummary, results []*TestResult) error {
	l.Trace().Msg("Writing report to console")
	start := time.Now()

	fmt.Println("Flakeguard report")
	fmt.Println("--------------------------------")
	fmt.Println(summary.String())
	fmt.Println("--------------------------------")

	for _, result := range results {
		if result.Failures > 0 || result.Panic {
			fmt.Println(result.String())
		}
	}

	l.Trace().Dur("duration", time.Since(start)).Msg("Report written to console")
	return nil
}
