// Package report contains the logic for creating a report from test output. It will then send the report to selected destinations.
package report

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"
)

// TestResult contains the results and outputs of a single test
type TestResult struct {
	// Identifying info of the test
	TimeRun    time.Time `json:"time_run"`
	Name       string    `json:"name"`
	Package    string    `json:"package"`
	Path       string    `json:"path,omitempty"`        // TODO: Get this
	CodeOwners []string  `json:"code_owners,omitempty"` // TODO: Get this

	// Meta information about the test run
	TestRunInfo TestRunInfo `json:"test_run_info"`

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

// TestRunInfo details meta information about the code where the tests were run
type TestRunInfo struct {
	RepoURL   string `json:"repo_url"`
	RepoOwner string `json:"repo_owner"`
	RepoName  string `json:"repo_name"`
	// The default branch of the repository
	DefaultBranch string `json:"default_branch"`
	// If the tests were run on the default branch
	OnDefaultBranch bool `json:"on_default_branch"`
	// The current branch that the tests were run on, also the 'from' branch in a PR or merge
	HeadBranch string `json:"head_branch"`
	// The commit that the tests were run on, also the 'from' commit in a PR or merge
	HeadCommit string `json:"head_commit"`
	// The 'to' branch in a PR or merge
	BaseBranch string `json:"base_branch,omitempty"`
	// The 'to' commit in a PR or merge
	BaseCommit string `json:"base_commit,omitempty"`
	// If the test was run in a GitHub Actions environment, this is the event that triggered the run
	GitHubEvent string `json:"github_event,omitempty"`
	// If the test was run in a GitHub Actions environment, this is the workflow that triggered the run
	GitHubWorkflow string `json:"github_workflow,omitempty"`
	// If the test was run in a GitHub Actions environment, this is the run ID
	GitHubRunID string `json:"github_run_id,omitempty"`
	// If the test was run in a GitHub Actions environment, this is the run number
	GitHubRunNumber string `json:"github_run_number,omitempty"`
}

func (t *TestResult) String() string {
	return fmt.Sprintf(
		"TestPackage: %s, TestName: %s, TestPath: %s, PackagePanic: %t, Panic: %t, Timeout: %t, Race: %t, PassPercentage: %.2f, Runs: %d, Failures: %d, Successes: %d, Skips: %d",
		t.Package,
		t.Name,
		t.Path,
		t.PackagePanic,
		t.Panic,
		t.Timeout,
		t.Race,
		t.PassRatio*100,
		t.Runs,
		t.Failures,
		t.Successes,
		t.Skips,
	)
}

// testOutputLine is a single line of test output from the go test -json
type testOutputLine struct {
	Action  string    `json:"Action,omitempty"`
	Test    string    `json:"Test,omitempty"`
	Package string    `json:"Package,omitempty"`
	Output  string    `json:"Output,omitempty"`
	Elapsed float64   `json:"Elapsed,omitempty"` // Decimal value in seconds
	Time    time.Time `json:"Time,omitempty"`    // Time of the log
}

type reportSummary struct {
	UniqueTestsRun int
	TotalTestRuns  int
	Successes      int
	Failures       int
	Panics         int
	Races          int
	Timeouts       int
	Skips          int
}

func (s *reportSummary) String() string {
	return fmt.Sprintf(
		"UniqueTestsRun: %d, TotalTestRuns: %d, Successes: %d, Failures: %d, Panics: %d, Races: %d, Timeouts: %d, Skips: %d",
		s.UniqueTestsRun,
		s.TotalTestRuns,
		s.Successes,
		s.Failures,
		s.Panics,
		s.Races,
		s.Timeouts,
		s.Skips,
	)
}

type reportOptions struct {
	dryRun    bool
	reportDir string

	// Local reporting
	toConsole    bool
	reportFile   string
	jsonFile     string
	markdownFile string

	// Remote reporting
	// Splunk
	splunkURL        string
	splunkToken      string
	splunkIndex      string
	splunkSourceType string
}

func defaultOptions() reportOptions {
	return reportOptions{
		reportDir:    "./flakeguard-output",
		toConsole:    true,
		reportFile:   "flakeguard-report.txt",
		jsonFile:     "flakeguard-report.json",
		markdownFile: "flakeguard-report.md",
	}
}

type Option func(*reportOptions)

// DryRun disables reporting to outside services (Splunk, Slack, etc.), useful for debugging
func DryRun() Option {
	return func(o *reportOptions) {
		o.dryRun = true
	}
}

// ReportDir sets the directory to write reports files to. If not set, reports will be written to the current working directory.
func ReportDir(path string) Option {
	return func(o *reportOptions) {
		o.reportDir = path
	}
}

// SilenceConsole disables writing a concise report to the console
func SilenceConsole() Option {
	return func(o *reportOptions) {
		o.toConsole = false
	}
}

// ToFile writes the report to a human-readable text file, good for debugging
func ToFile(path string) Option {
	return func(o *reportOptions) {
		o.reportFile = path
	}
}

// ToJSON writes the report to a JSON file, good for ingesting into other programs
func ToJSON(path string) Option {
	return func(o *reportOptions) {
		o.jsonFile = path
	}
}

// ToSplunk sends the report to Splunk via HTTP Event Collector
func ToSplunk(url, token, index, sourceType string) Option {
	return func(o *reportOptions) {
		if url != "" {
			o.splunkURL = url
		}
		if token != "" {
			o.splunkToken = token
		}
		if index != "" {
			o.splunkIndex = index
		}
		if sourceType != "" {
			o.splunkSourceType = sourceType
		}
	}
}

// New creates a new report from scanning go test -json output. It will then send the report to selected destinations.
func New(l zerolog.Logger, testRunInfo TestRunInfo, files []string, options ...Option) error {
	opts := defaultOptions()
	for _, option := range options {
		option(&opts)
	}

	lines, err := readTestOutput(l, opts.reportDir, files...)
	if err != nil {
		return fmt.Errorf("failed to read test output: %w", err)
	}

	summary, results, err := analyzeTestOutput(l, lines)
	if err != nil {
		return fmt.Errorf("failed to analyze test output: %w", err)
	}

	for _, result := range results {
		result.TestRunInfo = testRunInfo
	}

	eg := errgroup.Group{}
	if opts.toConsole {
		eg.Go(func() error {
			return writeToConsole(l, summary, results)
		})
	}

	if opts.reportFile != "" {
		eg.Go(func() error {
			return writeToTextFile(l, summary, results, opts.reportDir, opts.reportFile)
		})
	}

	if opts.jsonFile != "" {
		eg.Go(func() error {
			return writeToJSONFile(l, summary, results, opts.reportDir, opts.jsonFile)
		})
	}

	if err := eg.Wait(); err != nil {
		return fmt.Errorf("failed to write report: %w", err)
	}

	return nil
}

// readTestOutput reads the JSON output of a test suite run into structs
func readTestOutput(l zerolog.Logger, dir string, files ...string) ([]*testOutputLine, error) {
	l.Debug().Strs("files", files).Msg("Reading test output")
	start := time.Now()

	lines := []*testOutputLine{}
	for _, file := range files {
		filePath := filepath.Join(dir, file)
		//nolint:gosec // we're reading from our own files
		jsonFile, err := os.Open(filePath)
		if err != nil {
			return nil, fmt.Errorf("failed to open test output file '%s': %w", filePath, err)
		}
		defer func() {
			if err := jsonFile.Close(); err != nil {
				l.Error().Str("path", filePath).Err(err).Msg("Failed to close test output file")
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
