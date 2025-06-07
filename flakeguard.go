package flakeguard

import (
	"os"
	"testing"
)

const RunQuarantinedTestsEnvVar = "FLAKEGUARD_RUN_QUARANTINED_TESTS"

// Quarantine a test so that it is skipped during your CI/CD pipelines.
// You can still make the test run by setting FLAKEGUARD_RUN_QUARANTINED_TESTS to true.
func Quarantine(t *testing.T, quarantineMessage string) {
	if os.Getenv(RunQuarantinedTestsEnvVar) != "true" {
		t.Skip(quarantineMessage)
	}
}
