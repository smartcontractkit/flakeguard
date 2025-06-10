package report

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
)

// writeToTextFile writes a flakeguard report to a human-readable text file
func writeToTextFile(l zerolog.Logger, summary *reportSummary, results []*TestResult, file string) error {
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

	_, err = fmt.Fprintf(reportFile, "%s\n", summary.String())
	if err != nil {
		return fmt.Errorf("failed to write to report file: %w", err)
	}
	_, err = fmt.Fprintf(reportFile, "====================\n")
	if err != nil {
		return fmt.Errorf("failed to write to report file: %w", err)
	}

	for _, result := range results {
		if result.Failures > 0 || result.Panic {
			_, err := reportFile.WriteString("--------------------------------\n")
			if err != nil {
				return fmt.Errorf("failed to write to report file: %w", err)
			}
			_, err = fmt.Fprintf(reportFile, "%s\n", result.String())
			if err != nil {
				return fmt.Errorf("failed to write to report file: %w", err)
			}
			_, err = reportFile.WriteString("--------------------------------\n")
			if err != nil {
				return fmt.Errorf("failed to write to report file: %w", err)
			}
			for _, failingRunNum := range result.FailingRunNumbers {
				_, err := fmt.Fprintf(reportFile, "\nFailing run %d\n", failingRunNum)
				if err != nil {
					return fmt.Errorf("failed to write to report file: %w", err)
				}
				_, err = reportFile.WriteString("--------------------------------\n")
				if err != nil {
					return fmt.Errorf("failed to write to report file: %w", err)
				}
				_, err = reportFile.WriteString(strings.Join(result.Outputs[failingRunNum], ""))
				if err != nil {
					return fmt.Errorf("failed to write to report file: %w", err)
				}
			}
		}
	}

	l.Debug().Dur("duration", time.Since(start)).Msg("Report written to file")
	return nil
}

// writeToJSONFile writes a flakeguard report to a JSON file
func writeToJSONFile(l zerolog.Logger, summary *reportSummary, results []*TestResult, file string) error {
	l.Debug().Str("file", file).Msg("Writing report to JSON file")
	start := time.Now()

	type jsonReport struct {
		Summary *reportSummary `json:"summary"`
		Results []*TestResult  `json:"results"`
	}

	json, err := json.Marshal(jsonReport{
		Summary: summary,
		Results: results,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal report to JSON: %w", err)
	}

	//nolint:gosec // G304 we're not writing to a file that we don't control
	jsonFile, err := os.Create(file)
	if err != nil {
		return fmt.Errorf("failed to create report file: %w", err)
	}
	defer func() {
		err := jsonFile.Close()
		if err != nil {
			l.Error().Str("file", file).Err(err).Msg("Failed to close report file")
		}
	}()

	_, err = jsonFile.Write(json)
	if err != nil {
		return fmt.Errorf("failed to write to report file: %w", err)
	}

	l.Debug().Dur("duration", time.Since(start)).Msg("Report written to JSON file")
	return nil
}
