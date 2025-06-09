// Package report contains the logic for creating a report from test output. It will then send the report to selected destinations.
package report

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
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
	RepoURL   string `json:"repo_url,omitempty"`
	RepoOwner string `json:"repo_owner,omitempty"`
	RepoName  string `json:"repo_name,omitempty"`
	// Head branch and commit are the branch and commit that the tests were run on.
	HeadBranch string `json:"head_branch,omitempty"`
	HeadCommit string `json:"head_commit,omitempty"`
	// Base branch is the branch that you want to merge code into, if the run was on a PR
	BaseBranch string `json:"base_branch,omitempty"`
	BaseCommit string `json:"base_commit,omitempty"`

	// Details of the test execution

	// If any test in the same package panics, this is true.
	// Same package panics can destroy the results of all other tests that were also running.
	PackagePanic      bool            `json:"package_panic"`
	Panic             bool            `json:"panic"`
	Timeout           bool            `json:"timeout"`
	Race              bool            `json:"race"`
	Skipped           bool            `json:"skipped"`
	PassRatio         float64         `json:"pass_ratio"`
	Runs              int             `json:"runs"`
	Failures          int             `json:"failures"`
	Successes         int             `json:"successes"`
	Skips             int             `json:"skips"`
	FailingRunNumbers []int           `json:"failing_runs,omitempty"`
	Durations         []time.Duration `json:"durations,omitempty"`
	// Run number -> outputs
	Outputs map[int][]string `json:"outputs,omitempty"`
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

	reportFile string

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

func ToFile(path string) Option {
	return func(o *reportOptions) {
		o.reportFile = path
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

	lines, err := readTestOutput(l, files...)
	if err != nil {
		return fmt.Errorf("failed to read test output: %w", err)
	}

	results, err := analyzeTestOutput(l, lines)
	if err != nil {
		return fmt.Errorf("failed to analyze test output: %w", err)
	}

	eg := errgroup.Group{}
	if opts.toConsole {
		eg.Go(func() error {
			return writeToConsole(l, results)
		})
	}

	if opts.reportFile != "" {
		eg.Go(func() error {
			return writeToFile(l, results, opts.reportFile)
		})
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
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
	l.Debug().Int("lines", len(lines)).Msg("Analyzing test output")
	start := time.Now()

	// package -> test_name -> TestResult
	results := map[string]map[string]*TestResult{}
	// package -> test_name -> current_run_number
	testRunNumber := map[string]map[string]int{}
	panickedPackages := []string{}

	for _, line := range lines {
		if _, ok := results[line.Package]; !ok {
			results[line.Package] = make(map[string]*TestResult)
		}
		if _, ok := testRunNumber[line.Package]; !ok {
			testRunNumber[line.Package] = make(map[string]int)
		}

		result, ok := results[line.Package][line.Test]
		if !ok {
			testRunNumber[line.Package][line.Test] = 1
			result = &TestResult{
				TestName:    line.Test,
				TestPackage: line.Package,
				Outputs:     make(map[int][]string),
				Durations:   []time.Duration{},
			}
			results[line.Package][line.Test] = result
		}

		result.Outputs[testRunNumber[line.Package][line.Test]] = append(
			result.Outputs[testRunNumber[line.Package][line.Test]],
			line.Output,
		)
		if line.Elapsed > 0 {
			result.Durations = append(result.Durations, time.Duration(line.Elapsed*1000000000))
		}
		switch line.Action {
		case "pass":
			result.Successes++

			testRunNumber[line.Package][line.Test]++
		case "fail":
			result.Failures++
			result.FailingRunNumbers = append(result.FailingRunNumbers, testRunNumber[line.Package][line.Test])

			testRunNumber[line.Package][line.Test]++
		case "skip":
			result.Skips++

			testRunNumber[line.Package][line.Test]++
		case "panic":
			result.Panic = true
			result.PackagePanic = true
			panickedPackages = append(panickedPackages, line.Package)
			result.FailingRunNumbers = append(result.FailingRunNumbers, testRunNumber[line.Package][line.Test])

			testRunNumber[line.Package][line.Test]++
		}
	}

	// Mark all test results in panicked packages as panicked
	for _, packageName := range panickedPackages {
		for _, result := range results[packageName] {
			result.PackagePanic = true
		}
	}

	resultSlice := make([]*TestResult, 0, len(results))
	for _, packageResults := range results {
		for _, result := range packageResults {
			resultSlice = append(resultSlice, result)
		}
	}

	l.Debug().Int("tests", len(resultSlice)).Str("duration", time.Since(start).String()).Msg("Analyzed test output")
	return resultSlice, nil
}
