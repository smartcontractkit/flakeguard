package report

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-resty/resty/v2"
	"github.com/rs/zerolog"
)

// splunkTestResult is the full wrapper structure sent to Splunk for a single test result
type splunkTestResult struct {
	Event      splunkTestResultEvent `json:"event"`      // https://docs.splunk.com/Splexicon:Event
	SourceType string                `json:"sourcetype"` // https://docs.splunk.com/Splexicon:Sourcetype
	Index      string                `json:"index"`      // https://docs.splunk.com/Splexicon:Index
}

type splunkTestResultEvent struct {
	// SplunkTestResultEvent contains the actual meat of the Splunk test result event
	Event string     `json:"event"`
	Data  TestResult `json:"data"`
}

// Splunk reports results to Splunk via HTTP Event Collector
func Splunk(l zerolog.Logger, results []TestResult, reportOptions reportOptions) error {
	var (
		splunkURL        = reportOptions.splunkURL
		splunkToken      = reportOptions.splunkToken
		splunkIndex      = reportOptions.splunkIndex
		splunkSourceType = reportOptions.splunkSourceType
	)

	if splunkURL == "" || splunkToken == "" || splunkIndex == "" {
		return fmt.Errorf("splunkURL, splunkToken, and splunkIndex must be set to use Splunk reporting")
	}

	const (
		// Actual splunk limit is over 800MB for a single request, but sending that much risks hitting weird limits and timeouts
		splunkSizeLimitBytes = 100_000_000 // 100MB
	)

	l.Debug().
		Int("results", len(results)).
		Msg("Reporting results to Splunk")
	startTime := time.Now()

	client := resty.New().
		SetBaseURL(splunkURL).
		SetAuthScheme("Splunk").
		SetAuthToken(splunkToken).
		SetHeader("Content-Type", "application/json").
		SetRetryCount(3). // Retry failed requests 3 times
		SetRetryWaitTime(100 * time.Millisecond).
		SetRetryMaxWaitTime(1 * time.Second)

	splunkBody := bytes.Buffer{}
	for count, result := range results {
		result.Outputs = nil // Don't send our test output to Splunk, can be overkill
		err := writeSplunkEvent(splunkBody, result, splunkSourceType, splunkIndex)
		if err != nil {
			return fmt.Errorf("failed to write test result to buffer: %w", err)
		}

		if count == len(results)-1 ||
			splunkBody.Len() >= splunkSizeLimitBytes { // Last result or size limit hit, send to Splunk
			l.Trace().
				Int("batchSizeKB", splunkBody.Len()/1024).
				Msg("Sending batch of test results to Splunk")

			// Splunk doesn't accept a JSON array, it wants individual JSON objects
			// https://docs.splunk.com/Documentation/Splunk/latest/RESTREF/RESTinput#Bulk_Data_Input

			// Append to a file for dry run mode
			if reportOptions.dryRun {
				err := writeSplunkDryRunFile(l, splunkBody, reportOptions)
				if err != nil {
					return fmt.Errorf("failed to write Splunk dry run file: %w", err)
				}
				l.Debug().
					Str("file", filepath.Join(reportOptions.reportDir, "splunk_test_results.json")).
					Int("batchSizeKB", splunkBody.Len()/1024).
					Msg("Dry Run: Wrote Splunk batch to file")
				return nil
			}

			resp, err := client.R().
				SetBody(splunkBody).
				Post("")
			if err != nil {
				return fmt.Errorf("failed to send test results to Splunk: %w", err)
			}
			if resp.IsError() {
				return fmt.Errorf("failed to send test results to Splunk: %s", resp.String())
			}
			l.Trace().
				Int("batchSizeKB", splunkBody.Len()/1024).
				Msg("Sent batch of test results to Splunk")

			splunkBody.Reset()
		}
	}

	l.Debug().
		Str("duration", time.Since(startTime).String()).
		Msg("Sent test results to Splunk")
	return nil
}

func writeSplunkDryRunFile(l zerolog.Logger, splunkBody bytes.Buffer, reportOptions reportOptions) error {
	splunkFileName := filepath.Join(reportOptions.reportDir, "splunk_test_results.json")
	err := os.MkdirAll(reportOptions.reportDir, 0700)
	if err != nil {
		return fmt.Errorf("failed to create report directory: %w", err)
	}

	var splunkFile *os.File
	if _, err := os.Stat(splunkFileName); err == nil { // If the file exists, just append to it
		splunkFile, err = os.OpenFile(splunkFileName, os.O_APPEND|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("failed to open Splunk dry run file: %w", err)
		}
	} else {
		splunkFile, err = os.Create(splunkFileName)
		if err != nil {
			return fmt.Errorf("failed to create Splunk dry run file: %w", err)
		}
	}
	defer func() {
		err := splunkFile.Close()
		if err != nil {
			l.Error().Err(err).Msg("Failed to close Splunk dry run file")
		}
	}()

	_, err = splunkFile.WriteString("Batch:\n")
	if err != nil {
		return fmt.Errorf("failed to write Splunk dry run file: %w", err)
	}
	_, err = splunkFile.Write(splunkBody.Bytes())
	if err != nil {
		return fmt.Errorf("failed to write Splunk dry run file: %w", err)
	}
	_, err = splunkFile.WriteString("\n")
	if err != nil {
		return fmt.Errorf("failed to write Splunk dry run file: %w", err)
	}
	return nil
}

// writeSplunkEvent writes a single test result to the Splunk body to be sent to Splunk
func writeSplunkEvent(splunkBody bytes.Buffer, result TestResult, splunkSourceType, splunkIndex string) error {
	splunkEvent := splunkTestResult{
		SourceType: splunkSourceType,
		Index:      splunkIndex,
		Event: splunkTestResultEvent{
			Event: "flakeguard_test_result",
			Data:  result,
		},
	}
	json, err := json.Marshal(splunkEvent)
	if err != nil {
		return fmt.Errorf("failed to marshal test result: %w", err)
	}
	_, err = splunkBody.Write(json)
	if err != nil {
		return fmt.Errorf("failed to write test result to buffer: %w", err)
	}
	_, err = splunkBody.WriteString("\n")
	if err != nil {
		return fmt.Errorf("failed to write test result to buffer: %w", err)
	}
	return nil
}
