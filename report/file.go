package report

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// writeToFile writes a flakeguard report to a file
func writeToFile(l zerolog.Logger, results []*TestResult, file string) error {
	l.Debug().Str("file", file).Msg("Writing report to file")
	start := time.Now()

	//nolint:gosec // G304 we're not writing to a file that we don't control
	reportFile, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer func() {
		err := reportFile.Close()
		if err != nil {
			l.Error().Str("file", file).Err(err).Msg("Failed to close report file")
		}
	}()

	for _, result := range results {
		if result.Failures > 0 || result.Panic {
			_, err := reportFile.WriteString("--------------------------------\n")
			if err != nil {
				return fmt.Errorf("failed to write to report file: %w", err)
			}
			_, err = fmt.Fprintf(
				reportFile,
				"%s %s ran %d times, failed %d times\n",
				result.TestPackage,
				result.TestName,
				len(result.Outputs),
				result.Failures,
			)
			if err != nil {
				return fmt.Errorf("failed to write to report file: %w", err)
			}
			_, err = reportFile.WriteString("--------------------------------\n")
			if err != nil {
				return fmt.Errorf("failed to write to report file: %w", err)
			}
			for _, failingRunNum := range result.FailingRunNumbers {
				_, err := fmt.Fprintf(reportFile, "Failing run %d:\n", failingRunNum)
				if err != nil {
					return fmt.Errorf("failed to write to report file: %w", err)
				}
				_, err = reportFile.WriteString(strings.Join(result.Outputs[failingRunNum], "\n"))
				if err != nil {
					return fmt.Errorf("failed to write to report file: %w", err)
				}
			}
		}
	}

	l.Debug().Dur("duration", time.Since(start)).Msg("Report written to file")
	return nil
}
