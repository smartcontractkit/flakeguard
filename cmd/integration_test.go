// This file is for integration tests of flakeguard commands against example test packages.
package cmd

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// isDevEnvironment checks if we're running in the development environment
// by checking for the existence of example_tests directory
func isDevEnvironment() bool {
	_, err := os.Stat("../example_tests")
	return err == nil
}

// buildFlakeguard builds the flakeguard binary for testing
func buildFlakeguard(t *testing.T) string {
	t.Helper()

	// Build flakeguard binary
	tmpDir := t.TempDir()
	binaryPath := filepath.Join(tmpDir, "flakeguard")

	cmd := exec.Command("go", "build", "-o", binaryPath, "../cmd/flakeguard/main.go")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "Failed to build flakeguard: %s", string(output))

	return binaryPath
}

// runFlakeguardCommand executes a flakeguard command and returns the result
func runFlakeguardCommand(
	t *testing.T,
	binaryPath, mode, packageName string,
	extraArgs ...string,
) (exitCode int, stdout, stderr string) {
	t.Helper()

	outputDir := filepath.Join("./", t.Name())
	t.Cleanup(func() {
		if t.Failed() {
			t.Logf("%s failed, leaving output directory for debugging: %s", t.Name(), outputDir)
			return
		}
		if err := os.RemoveAll(outputDir); err != nil {
			t.Logf("Failed to remove output directory for test %s: %s", t.Name(), err)
		}
	})

	// Base args for the command
	args := []string{mode, "-c", "-L", "debug", "-o", outputDir, "--"}
	args = append(args, extraArgs...)
	args = append(args, "--", fmt.Sprintf("../example_tests/%s/...", packageName), "-tags", "examples")

	// Add package-specific flags
	switch packageName {
	case "race":
		args = append(args, "-race")
	case "timeout":
		args = append(args, "-timeout", "1s")
	}

	cmd := exec.Command(binaryPath, args...)
	cmd.Dir = ".." // Run from project root
	cmdOutput := bytes.Buffer{}
	cmd.Stdout = &cmdOutput
	cmd.Stderr = &cmdOutput
	err := cmd.Run()

	exitCode = 0
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			exitCode = 1
		}
	}

	return exitCode, cmdOutput.String(), ""
}

func TestIntegration(t *testing.T) {
	t.Parallel()

	if !isDevEnvironment() {
		t.Skip("Skipping integration tests - not in development environment")
	}
	if testing.Short() {
		t.Skip("Skipping integration tests - short mode")
	}
	// TODO: Implement further, this is currently unfinished and a template for later work.
	// Worth looking into using testscript: https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript
	t.Skip("Skipping integration tests - not implemented")

	binaryPath := buildFlakeguard(t)

	testCases := []struct {
		name           string
		packageName    string
		mode           string
		expectFailure  bool
		expectedOutput []string // Substrings that should appear in output
	}{
		{
			name:           "detect_flaky",
			packageName:    "flaky",
			mode:           "detect",
			expectFailure:  true, // Flaky tests should be detected
			expectedOutput: []string{"flaky", "detected"},
		},
		{
			name:          "guard_flaky",
			packageName:   "flaky",
			mode:          "guard",
			expectFailure: true, // Flaky tests should fail in guard mode
		},
		{
			name:           "detect_race",
			packageName:    "race",
			mode:           "detect",
			expectFailure:  true, // Race conditions should be detected
			expectedOutput: []string{"race"},
		},
		{
			name:          "guard_race",
			packageName:   "race",
			mode:          "guard",
			expectFailure: true, // Race conditions should fail in guard mode
		},
		{
			name:           "detect_panic",
			packageName:    "panic",
			mode:           "detect",
			expectFailure:  true, // Panics should be detected
			expectedOutput: []string{"panic"},
		},
		{
			name:          "guard_panic",
			packageName:   "panic",
			mode:          "guard",
			expectFailure: true, // Panics should fail in guard mode
		},
		{
			name:           "detect_timeout",
			packageName:    "timeout",
			mode:           "detect",
			expectFailure:  true, // Timeouts should be detected
			expectedOutput: []string{"timeout"},
		},
		{
			name:          "guard_timeout",
			packageName:   "timeout",
			mode:          "guard",
			expectFailure: true, // Timeouts should fail in guard mode
		},
		{
			name:           "detect_pass",
			packageName:    "pass",
			mode:           "detect",
			expectFailure:  false, // Passing tests should succeed
			expectedOutput: []string{"PASS"},
		},
		{
			name:           "guard_pass",
			packageName:    "pass",
			mode:           "guard",
			expectFailure:  false, // Passing tests should succeed
			expectedOutput: []string{"PASS"},
		},
		{
			name:           "detect_fail",
			packageName:    "fail",
			mode:           "detect",
			expectFailure:  true, // Failing tests should fail
			expectedOutput: []string{"FAIL"},
		},
		{
			name:           "guard_fail",
			packageName:    "fail",
			mode:           "guard",
			expectFailure:  true, // Failing tests should fail
			expectedOutput: []string{"FAIL"},
		},
		{
			name:          "detect_quarantine",
			packageName:   "quarantine",
			mode:          "detect",
			expectFailure: false, // Quarantined tests might pass in detect mode
		},
		{
			name:          "guard_quarantine",
			packageName:   "quarantine",
			mode:          "guard",
			expectFailure: false, // Quarantined tests should be handled gracefully
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			t.Logf("Running %s mode on %s package", tc.mode, tc.packageName)

			exitCode, stdout, stderr := runFlakeguardCommand(t, binaryPath, tc.mode, tc.packageName, t.TempDir())

			t.Logf("Exit code: %d", exitCode)
			t.Logf("Stdout: %s", stdout)
			if stderr != "" {
				t.Logf("Stderr: %s", stderr)
			}

			if tc.expectFailure {
				assert.NotEqual(t, 0, exitCode, "Expected command to fail but it succeeded")
			} else {
				assert.Equal(t, 0, exitCode, "Expected command to succeed but it failed")
			}

			// Check for expected output strings
			for _, expected := range tc.expectedOutput {
				assert.Contains(t, stdout, expected, "Expected output to contain: %s", expected)
			}
		})
	}
}
