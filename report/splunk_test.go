package report

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/flakeguard/internal/testhelpers"
)

func TestSplunk(t *testing.T) {
	t.Parallel()

	l := testhelpers.Logger(t)
	opts := defaultOptions()
	opts.reportDir = t.TempDir()
	opts.dryRun = true
	opts.splunkURL = "https://splunk.test.com"
	opts.splunkToken = "test"
	opts.splunkIndex = "test"
	opts.splunkSourceType = "test"

	results := []TestResult{
		{
			TimeRun: time.Now(),
			Name:    "TestSplunk",
			Package: "test/package",
			Path:    "path/to/test",
			TestRunInfo: TestRunInfo{
				RepoURL:     "https://github.com/testowner/testrepo",
				RepoOwner:   "testowner",
				RepoName:    "testrepo",
				HeadBranch:  "main",
				HeadCommit:  "1234567890",
				BaseBranch:  "main",
				BaseCommit:  "1234567890",
				GitHubEvent: "test_event",
			},
			Runs:              5,
			Failures:          1,
			Successes:         4,
			Skips:             0,
			FailingRunNumbers: []int{1},
			Durations: []time.Duration{
				1 * time.Second,
				2 * time.Second,
				3 * time.Second,
				4 * time.Second,
				5 * time.Second,
			},
			Outputs: map[int][]string{
				1: {"test output"},
				2: {"test output"},
				3: {"test output"},
				4: {"test output"},
				5: {"test output"},
			},
			PackagePanic: false,
			Panic:        false,
			Timeout:      false,
			Race:         false,
			Skipped:      false,
			PassRatio:    0.8,
		},
	}
	err := Splunk(l, results, opts)
	require.NoError(t, err)

	splunkFile := filepath.Join(opts.reportDir, "splunk_test_results.json")
	require.FileExists(t, splunkFile)
	content, err := os.ReadFile(splunkFile)
	require.NoError(t, err)

	require.Contains(t, string(content), "Batch:")
	require.Contains(t, string(content), "test/package")
	require.Contains(t, string(content), "TestSplunk")
	require.Contains(t, string(content), "path/to/test")
	require.Contains(t, string(content), "testowner/testrepo")
	require.Contains(t, string(content), "main")
	require.Contains(t, string(content), "1234567890")
	require.Contains(t, string(content), "main")
	require.NotContains(t, string(content), "test output", "Splunk reporter should not include test output")
}
