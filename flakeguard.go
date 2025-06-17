// Package flakeguard allows you to quarantine tests that have been identified as flaky.
package flakeguard

import (
	"os"
	"testing"
)

// RunQuarantinedTestsEnvVar is the environment variable that controls whether to run quarantined tests.
// Set to true to run quarantined tests, false to skip them.
// This is helpful to keep track of tests' flakiness even when they are quarantined.
const RunQuarantinedTestsEnvVar = "FLAKEGUARD_RUN_QUARANTINED_TESTS"

// Quarantine a test so that it is skipped during your CI/CD pipelines.
// You can still make the test run by setting FLAKEGUARD_RUN_QUARANTINED_TESTS to true.
func Quarantine(t *testing.T, quarantineMessage string) {
	if os.Getenv(RunQuarantinedTestsEnvVar) != "true" {
		t.Skip(quarantineMessage)
	}
}
