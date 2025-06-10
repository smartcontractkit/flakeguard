// Package report contains the logic for creating a report from test output. It will then send the report to selected destinations.
package report

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/sync/errgroup"

	"github.com/smartcontractkit/flakeguard/git"
	"github.com/smartcontractkit/flakeguard/github"
)

// TestResult contains the results and outputs of a single test
type TestResult struct {
	// Identifying info on the test
	TestName    string   `json:"test_name"`
	TestPackage string   `json:"test_package"`
	TestPath    string   `json:"test_path,omitempty"`   // TODO: Get this
	CodeOwners  []string `json:"code_owners,omitempty"` // TODO: Get this

	TestRunInfo *TestRunInfo `json:"test_run_info,omitempty"`

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

// TestRunInfo details meta information about the code when the test was run
type TestRunInfo struct {
	RepoURL   string `json:"repo_url"`
	RepoOwner string `json:"repo_owner"`
	RepoName  string `json:"repo_name"`
	// The current branch and commit that the tests were run on
	HeadBranch string `json:"head_branch"`
	HeadCommit string `json:"head_commit"`
	// The base branch and commit that the tests were run on (only applicable for PRs or merge groups)
	BaseBranch string `json:"base_branch,omitempty"`
	BaseCommit string `json:"base_commit,omitempty"`
	// If the test was run in a GitHub Actions environment, this is the event that triggered the run
	GitHubEvent string `json:"github_event,omitempty"`
}

func (t *TestResult) String() string {
	return fmt.Sprintf(
		"TestPackage: %s, TestName: %s, TestPath: %s, PackagePanic: %t, Panic: %t, Timeout: %t, Race: %t, PassPercentage: %.2f, Runs: %d, Failures: %d, Successes: %d, Skips: %d",
		t.TestPackage,
		t.TestName,
		t.TestPath,
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
	Action  string  `json:"Action,omitempty"`
	Test    string  `json:"Test,omitempty"`
	Package string  `json:"Package,omitempty"`
	Output  string  `json:"Output,omitempty"`
	Elapsed float64 `json:"Elapsed,omitempty"` // Decimal value in seconds
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
	toConsole bool

	reportFile string

	jsonFile string

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

// ToConsole writes a concise report to the console
func ToConsole() Option {
	return func(o *reportOptions) {
		o.toConsole = true
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

	summary, results, err := analyzeTestOutput(l, lines)
	if err != nil {
		return fmt.Errorf("failed to analyze test output: %w", err)
	}

	eg := errgroup.Group{}
	if opts.toConsole {
		eg.Go(func() error {
			return writeToConsole(l, summary, results)
		})
	}

	if opts.reportFile != "" {
		eg.Go(func() error {
			return writeToTextFile(l, summary, results, opts.reportFile)
		})
	}

	if opts.jsonFile != "" {
		eg.Go(func() error {
			return writeToJSONFile(l, summary, results, opts.jsonFile)
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

func testRunInfo(l zerolog.Logger, repoPath string) (*TestRunInfo, error) {
	repoInfo, err := git.GetRepoInfo(l, repoPath)
	if err != nil {
		return nil, fmt.Errorf("failed to get repo info: %w", err)
	}

	t := &TestRunInfo{
		RepoURL:    repoInfo.URL,
		RepoOwner:  repoInfo.Owner,
		RepoName:   repoInfo.Name,
		HeadBranch: repoInfo.Branch,
		HeadCommit: repoInfo.Commit,
	}

	if github.IsActions() {
		t.GitHubEvent = github.Event()
		t.BaseBranch, t.BaseCommit = github.Branches()
	}

	return t, nil
}
