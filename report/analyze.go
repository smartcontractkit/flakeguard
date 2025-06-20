package report

import (
	"fmt"
	"regexp"
	"sort"
	"time"

	"github.com/rs/zerolog"

	"github.com/smartcontractkit/flakeguard/exit"
)

var (
	timeoutRe = regexp.MustCompile(`^panic: test timed out after (.*)`)
	panicRe   = regexp.MustCompile(`^panic:`)
	raceRe    = regexp.MustCompile(`^WARNING: DATA RACE`)
)

func analyzeTestOutput(l zerolog.Logger, lines []*testOutputLine) (*reportSummary, []*TestResult, error) {
	l.Trace().Int("lines", len(lines)).Msg("Analyzing test output")
	start := time.Now()

	if len(lines) == 0 {
		return nil, nil, exit.New(exit.CodeFlakeguardError, fmt.Errorf("no tests run"))
	}

	summary := &reportSummary{
		UniqueTestsRun: 0,
		TotalTestRuns:  0,
		Successes:      0,
		Failures:       0,
		Panics:         0,
		Races:          0,
		Timeouts:       0,
		Skips:          0,
	}

	// package -> test_name -> TestResult
	results := map[string]map[string]*TestResult{}
	// package -> test_name -> current_run_number
	testRunNumber := map[string]map[string]int{}
	panickedPackages := []string{}

	for _, line := range lines {
		if line.Action == "build-fail" {
			return nil, nil, exit.New(exit.CodeGoBuildError, fmt.Errorf("go test build failed"))
		}

		if _, ok := results[line.Package]; !ok {
			results[line.Package] = make(map[string]*TestResult)
		}
		if _, ok := testRunNumber[line.Package]; !ok {
			testRunNumber[line.Package] = make(map[string]int)
		}
		if line.Test == "" { // This is a package summary line, not a test result
			continue
		}

		result, ok := results[line.Package][line.Test]
		if !ok {
			summary.UniqueTestsRun++
			testRunNumber[line.Package][line.Test] = 1
			result = &TestResult{
				TimeRun:   line.Time,
				Name:      line.Test,
				Package:   line.Package,
				Outputs:   make(map[int][]string),
				Durations: []time.Duration{},
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

		// Panics and races will often lie in JSON output, so that the attached line.Test isn't the actual test that panicked.
		// This is a limitation of how go test output works. There are some tricks where you can better attribute the panic to the correct test,
		// but they're full of their own edge cases and limitations.
		// We'll just use the line.Test as a best guess.

		if timeoutRe.MatchString(line.Output) { // Timeouts are a special kind of panic
			result.Timeout = true
			summary.Timeouts++
			result.Runs++
			summary.TotalTestRuns++
			panickedPackages = append(panickedPackages, line.Package)
			result.FailingRunNumbers = append(result.FailingRunNumbers, testRunNumber[line.Package][line.Test])

			testRunNumber[line.Package][line.Test]++
			continue
		}

		if panicRe.MatchString(line.Output) {
			result.Panic = true
			result.PackagePanic = true
			summary.Panics++
			result.Runs++
			summary.TotalTestRuns++
			panickedPackages = append(panickedPackages, line.Package)
			result.FailingRunNumbers = append(result.FailingRunNumbers, testRunNumber[line.Package][line.Test])

			testRunNumber[line.Package][line.Test]++
			continue
		}

		if raceRe.MatchString(line.Output) {
			result.Race = true
			summary.Races++
			result.Runs++
			summary.TotalTestRuns++
			result.FailingRunNumbers = append(result.FailingRunNumbers, testRunNumber[line.Package][line.Test])

			testRunNumber[line.Package][line.Test]++
			continue
		}

		switch line.Action {
		case "build-fail":
			return nil, nil, exit.New(exit.CodeGoBuildError, fmt.Errorf("build failed for package %s", line.Package))
		case "pass":
			result.Successes++
			summary.Successes++
			result.Runs++
			summary.TotalTestRuns++

			testRunNumber[line.Package][line.Test]++
		case "fail":
			result.Failures++
			summary.Failures++
			result.Runs++
			summary.TotalTestRuns++
			result.FailingRunNumbers = append(result.FailingRunNumbers, testRunNumber[line.Package][line.Test])

			testRunNumber[line.Package][line.Test]++
		case "skip":
			result.Skips++
			summary.Skips++

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

	// Sort by package and name for easier reading
	sort.Slice(resultSlice, func(i, j int) bool {
		if resultSlice[i].Package == resultSlice[j].Package {
			return resultSlice[i].Name < resultSlice[j].Name
		}
		return resultSlice[i].Package < resultSlice[j].Package
	})

	if summary.UniqueTestsRun == 0 {
		return nil, nil, exit.New(exit.CodeFlakeguardError, fmt.Errorf("no tests run"))
	}

	l.Trace().Int("tests", len(resultSlice)).Str("duration", time.Since(start).String()).Msg("Analyzed test output")
	return summary, resultSlice, nil
}
