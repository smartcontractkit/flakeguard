package report

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/rs/zerolog"
)

// TestResult contains the results and outputs of a single test
type TestResult struct {
	// ReportID is the ID of the report this test result belongs to
	// used mostly for Splunk logging
	ReportID       string              `json:"report_id,omitempty"`
	TestName       string              `json:"test_name"`
	TestPackage    string              `json:"test_package"`
	PackagePanic   bool                `json:"package_panic"`
	Panic          bool                `json:"panic"`
	Timeout        bool                `json:"timeout"`
	Race           bool                `json:"race"`
	Skipped        bool                `json:"skipped"`
	PassRatio      float64             `json:"pass_ratio"`
	Runs           int                 `json:"runs"`
	Failures       int                 `json:"failures"`
	Successes      int                 `json:"successes"`
	Skips          int                 `json:"skips"`
	Outputs        map[string][]string `json:"-"`                        // Temporary storage for outputs during test run
	PassedOutputs  map[string][]string `json:"passed_outputs,omitempty"` // Outputs for passed runs
	FailedOutputs  map[string][]string `json:"failed_outputs,omitempty"` // Outputs for failed runs
	Durations      []time.Duration     `json:"durations"`
	PackageOutputs []string            `json:"package_outputs,omitempty"`
	TestPath       string              `json:"test_path,omitempty"`
	CodeOwners     []string            `json:"code_owners,omitempty"`
}

// TestOutputLine is a single line of test output from the go test -json output
type TestOutputLine struct {
	Action  string  `json:"Action,omitempty"`
	Test    string  `json:"Test,omitempty"`
	Package string  `json:"Package,omitempty"`
	Output  string  `json:"Output,omitempty"`
	Elapsed float64 `json:"Elapsed,omitempty"` // Decimal value in seconds
}

// ReadTestOutput reads the JSON output of a test suite run
func ReadTestOutput(l *zerolog.Logger, path string) ([]*TestOutputLine, error) {
	l.Debug().Str("path", path).Msg("Reading test output")
	start := time.Now()

	//nolint:gosec // G304 we're not reading from a network source
	jsonFile, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open test output file %s: %w", path, err)
	}
	defer func() {
		if err := jsonFile.Close(); err != nil {
			l.Error().Str("path", path).Err(err).Msg("Failed to close test output file")
		}
	}()

	lines := []*TestOutputLine{}
	decoder := json.NewDecoder(jsonFile)
	for decoder.More() {
		var line TestOutputLine
		if err := decoder.Decode(&line); err != nil {
			return nil, fmt.Errorf("error unmarshalling go test -json output: %w", err)
		}
		lines = append(lines, &line)
	}

	l.Debug().Int("lines", len(lines)).Str("duration", time.Since(start).String()).Msg("Read test output")
	return lines, nil
}
