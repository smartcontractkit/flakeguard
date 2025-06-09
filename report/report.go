// Package report contains the logic for creating a report from test output. It will then send the report to selected destinations.
package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"time"

	"github.com/rs/zerolog"
)

var (
	startPanicRe = regexp.MustCompile(`^panic:`)
	startRaceRe  = regexp.MustCompile(`^WARNING: DATA RACE`)
)

// TestResult contains the results and outputs of a single test
type TestResult struct {
	// Identifying info on the test
	TestName    string   `json:"test_name"`
	TestPackage string   `json:"test_package"`
	TestPath    string   `json:"test_path,omitempty"`
	CodeOwners  []string `json:"code_owners,omitempty"`

	// Details of the code when the test was executed by flakeguard
	RepoURL    string `json:"repo_url,omitempty"`
	RepoOwner  string `json:"repo_owner,omitempty"`
	RepoName   string `json:"repo_name,omitempty"`
	BaseBranch string `json:"base_branch,omitempty"`
	BaseCommit string `json:"base_commit,omitempty"`
	// Target
	TargetBranch string `json:"target_branch,omitempty"`
	TargetCommit string `json:"target_commit,omitempty"`

	// Details of the test execution

	// If any test in the same package panics, this is true.
	// Same package panics can destroy the results of all other tests that were also running.
	PackagePanic bool            `json:"package_panic"`
	Panic        bool            `json:"panic"`
	Timeout      bool            `json:"timeout"`
	Race         bool            `json:"race"`
	Skipped      bool            `json:"skipped"`
	PassRatio    float64         `json:"pass_ratio"`
	Runs         int             `json:"runs"`
	Failures     int             `json:"failures"`
	Successes    int             `json:"successes"`
	Skips        int             `json:"skips"`
	Durations    []time.Duration `json:"durations"`
	// Run number -> outputs
	Outputs map[int][]string `json:"outputs"`
}

// testOutputLine is a single line of test output from the go test -json
type testOutputLine struct {
	Action  string  `json:"Action,omitempty"`
	Test    string  `json:"Test,omitempty"`
	Package string  `json:"Package,omitempty"`
	Output  string  `json:"Output,omitempty"`
	Elapsed float64 `json:"Elapsed,omitempty"` // Decimal value in seconds
}

type reportOptions struct {
	toConsole bool

	splunkURL            string
	splunkToken          string
	splunkIndex          string
	splunkSource         string
	splunkSourceType     string
	splunkSourceHost     string
	splunkSourceHostName string
	splunkSourceHostIP   string

	dxWebhookURL string

	slackWebhookURL string
}

type Option func(*reportOptions)

func ToConsole() Option {
	return func(o *reportOptions) {
		o.toConsole = true
	}
}

// TODO: Add support for sending to Splunk via HTTP Event Collector
func ToSplunk(url, token, index, source, sourceType, sourceHost, sourceHostName, sourceHostIP string) Option {
	return func(o *reportOptions) {
		o.splunkURL = url
		o.splunkToken = token
		o.splunkIndex = index
		o.splunkSource = source
		o.splunkSourceType = sourceType
		o.splunkSourceHost = sourceHost
		o.splunkSourceHostName = sourceHostName
		o.splunkSourceHostIP = sourceHostIP
	}
}

// TODO: Add support for sending to DX via webhook
func ToDX(webhookURL string) Option {
	return func(o *reportOptions) {
		o.dxWebhookURL = webhookURL
	}
}

// TODO: Add support for sending to Slack via webhook
func ToSlack(webhookURL string) Option {
	return func(o *reportOptions) {
		o.slackWebhookURL = webhookURL
	}
}

// New creates a new report from scanning go test -json output. It will then send the report to selected destinations.
func New(l zerolog.Logger, files []string, options ...Option) error {
	opts := reportOptions{}
	for _, option := range options {
		option(&opts)
	}

	// TODO: analyze the test output and create a report
	_, err := readTestOutput(l, files...)
	if err != nil {
		return fmt.Errorf("failed to read test output: %w", err)
	}
	return nil
}

// readTestOutput reads the JSON output of a test suite run into structs
func readTestOutput(l zerolog.Logger, files ...string) ([]*testOutputLine, error) {
	l.Debug().Strs("files", files).Msg("Reading test output")
	start := time.Now()

	lines := []*testOutputLine{}
	for _, file := range files {
		//nolint:gosec // we're reading from our own files
		jsonFile, err := os.Open(file)
		if err != nil {
			return nil, fmt.Errorf("failed to open test output file %s: %w", file, err)
		}
		defer func() {
			if err := jsonFile.Close(); err != nil {
				l.Error().Str("path", file).Err(err).Msg("Failed to close test output file")
			}
		}()

		decoder := json.NewDecoder(jsonFile)
		for decoder.More() {
			var line testOutputLine
			if err := decoder.Decode(&line); err != nil {
				return nil, fmt.Errorf("error unmarshalling go test -json output: %w", err)
			}
			lines = append(lines, &line)
		}
	}

	l.Debug().
		Int("lines", len(lines)).
		Strs("files", files).
		Str("duration", time.Since(start).String()).
		Msg("Read test output")
	return lines, nil
}

func analyzeTestOutput(l zerolog.Logger, lines []*testOutputLine) ([]*TestResult, error) {
	l.Debug().Msg("Analyzing test output")
	start := time.Now()

	// package/test_name -> TestResult
	results := map[string]*TestResult{}

	for _, line := range lines {
		testKey := filepath.Join(line.Package, line.Test)
		result, ok := results[testKey]
		if !ok {
			result = &TestResult{
				TestName:    line.Test,
				TestPackage: line.Package,
			}
			results[testKey] = result
		}

		switch line.Action {
		case "pass":
			result.Successes++
		case "fail":
			result.Failures++
		case "skip":
			result.Skips++
		case "panic":
			result.Panic = true

		}
	}

	// Convert map to slice
	resultSlice := make([]*TestResult, 0, len(results))
	for _, result := range results {
		resultSlice = append(resultSlice, result)
	}

	l.Debug().Int("tests", len(resultSlice)).Str("duration", time.Since(start).String()).Msg("Analyzed test output")
	return resultSlice, nil
}
