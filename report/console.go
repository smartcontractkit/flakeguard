package report

import (
	"fmt"
	"time"

	"github.com/rs/zerolog"
)

func writeToConsole(l zerolog.Logger, results []*TestResult) error {
	l.Debug().Msg("Writing report to console")
	start := time.Now()

	fmt.Println("Flakeguard report")
	fmt.Println("--------------------------------")
	if len(results) == 0 {
		fmt.Println("No tests were run!")
		return nil
	}

	failures := 0
	for _, result := range results {
		if result.Failures > 0 || result.Panic {
			fmt.Printf(
				"%s %s ran %d times, failed %d times\n",
				result.TestPackage,
				result.TestName,
				len(result.Outputs),
				result.Failures,
			)
			failures += result.Failures
		}
	}

	if failures == 0 {
		fmt.Println("No flaky tests found")
	} else {
		fmt.Printf("Found %d failures in %d test runs\n", failures, len(results))
	}

	l.Debug().Dur("duration", time.Since(start)).Msg("Report written to console")
	return nil
}
