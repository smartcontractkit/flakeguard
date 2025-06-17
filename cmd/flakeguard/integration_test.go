package main

import (
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/rogpeppe/go-internal/testscript"
	"golang.org/x/mod/modfile"

	"github.com/smartcontractkit/flakeguard/internal/testhelpers"
)

const (
	distDir                    = "dist"
	flakeguardCoveredBinary    = "flakeguard_covered"
	flakeguardRootModulePath   = "github.com/smartcontractkit/flakeguard"
	integrationTestCoverageDir = "integration"
	// Environment variable to set the coverage directory for integration tests
	// We don't use the default GOCOVERDIR because testscript sets its own GOCOVERDIR
	CoverDirEnvVar = "FLAKEGUARD_GOCOVERDIR"
)

var (
	sourceFlakeguardBinaryPath string
	sourceFlakeguardBuiltTime  time.Time
)

func TestMain(m *testing.M) {
	var err error
	sourceFlakeguardBinaryPath, sourceFlakeguardBuiltTime, err = findBinary()
	if err != nil {
		log.Printf("Hit error while looking for flakeguard binary for integration tests: %v", err)
		os.Exit(1)
	}

	// Check if a coverage-instrumented binary exists
	if sourceFlakeguardBinaryPath != "" {
		// Check when binary was built
		// Coverage binary exists, don't include "flakeguard" in the map so testscript will look for external binary
		testscript.Main(m, map[string]func(){})
	} else {
		fmt.Println("Using in-process flakeguard binary for integration tests")
		// No binary, use in-process function
		testscript.Main(m, map[string]func(){
			"flakeguard": main,
		})
	}
}

// TestIntegrationDetect tests the detect command
func TestIntegrationDetect(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration tests with -short")
	}

	testscript.Run(t, testscript.Params{
		Dir:   "testscripts/detect",
		Setup: setupTestscript(t),
	})
}

// TestIntegrationGuard tests the guard command
func TestIntegrationGuard(t *testing.T) {
	t.Parallel()
	if testing.Short() {
		t.Skip("skipping integration tests with -short")
	}
	t.Skip("guard command not ready for integration tests yet")

	testscript.Run(t, testscript.Params{
		Dir:   "testscripts/guard",
		Setup: setupTestscript(t),
	})
}

// setupTestscript sets up the testscript environment for the integration tests
// It copies the example_tests directory to the testscript working directory
// and sets up the flakeguard binary and coverage collection
func setupTestscript(t *testing.T) func(env *testscript.Env) error {
	t.Helper()

	return func(env *testscript.Env) error {
		l := testhelpers.Logger(t)

		// Copy example_tests directory to the testscript working directory
		exampleTestsSource := filepath.Join("..", "..", "example_tests")
		exampleTestsDest := filepath.Join(env.WorkDir, "example_tests")

		if err := testhelpers.CopyDir(t, exampleTestsSource, exampleTestsDest); err != nil {
			return err
		}

		// Set the initial working directory to example_tests
		env.Cd = exampleTestsDest

		// Set up Go cache and module directories within the test work directory
		// to avoid issues with read-only or non-existent HOME directory
		goCacheDir := filepath.Join(env.WorkDir, ".gocache")
		goModCacheDir := filepath.Join(env.WorkDir, ".gomodcache")

		// Create the cache directories
		if err := os.MkdirAll(goCacheDir, 0750); err != nil {
			return fmt.Errorf("failed to create GOCACHE directory: %w", err)
		}
		if err := os.MkdirAll(goModCacheDir, 0750); err != nil {
			return fmt.Errorf("failed to create GOMODCACHE directory: %w", err)
		}

		// Set Go environment variables
		env.Setenv("GOCACHE", goCacheDir)
		env.Setenv("GOMODCACHE", goModCacheDir)
		env.Setenv("HOME", env.WorkDir) // Override the testscript default of /no-home

		if sourceFlakeguardBinaryPath != "" {
			// Copy the binary to the working directory
			destFlakeguardBinaryPath := filepath.Join(env.WorkDir, "flakeguard")
			l.Debug().
				Str("source", sourceFlakeguardBinaryPath).
				Str("dest", destFlakeguardBinaryPath).
				Msg("Copying flakeguard binary")
			if err := testhelpers.CopyFile(t, sourceFlakeguardBinaryPath, destFlakeguardBinaryPath); err != nil {
				fmt.Println("Error copying flakeguard binary, printing $WORKDIR contents for debugging")
				if err := testhelpers.ShowDirContents(t, env.WorkDir); err != nil {
					fmt.Println("Error showing directory contents: ", err)
				}
				return fmt.Errorf("failed to copy flakeguard binary: %w", err)
			}

			// Make sure the binary is executable
			//nolint:gosec // G302: we want to allow execution of the binary
			if err := os.Chmod(destFlakeguardBinaryPath, 0755); err != nil {
				return err
			}

			// Add the working directory to PATH so testscripts can find 'flakeguard' directly
			currentPath := env.Getenv("PATH")
			newPath := env.WorkDir + string(os.PathListSeparator) + currentPath
			env.Setenv("PATH", newPath)

			// Set up coverage collection for the binary
			if err := setupCoverageCollection(env); err != nil {
				l.Warn().Err(err).Msg("Failed to setup coverage collection, continuing without coverage")
			}
			l.Info().
				Str("GOCOVERDIR", env.Getenv("GOCOVERDIR")).
				Str("GOCACHE", env.Getenv("GOCACHE")).
				Str("GOMODCACHE", env.Getenv("GOMODCACHE")).
				Str("HOME", env.Getenv("HOME")).
				Str("WORKDIR", env.WorkDir).
				Str("sourceFlakeguardBinary", sourceFlakeguardBinaryPath).
				Str("destFlakeguardBinary", destFlakeguardBinaryPath).
				Time("flakeguardBuiltTime", sourceFlakeguardBuiltTime).
				Str("flakeguardBinaryAge", time.Since(sourceFlakeguardBuiltTime).String()).
				Msg("Running integration tests with flakeguard binary")
			if time.Since(sourceFlakeguardBuiltTime) > time.Minute {
				l.Warn().
					Str("hintIfInCI", "If you're running these tests in CI, it's common and likely harmless. The timestamp given to the binary is often weird in CI.").
					Time("flakeguardBuiltTime", sourceFlakeguardBuiltTime).
					Str("flakeguardBinaryAge", time.Since(sourceFlakeguardBuiltTime).String()).
					Msg("flakeguard binary is older than 1 minute, consider rebuilding")
			}
		} else {
			l.Info().Msg("No flakeguard binary found, using in-process compilation for integration tests")
		}

		return nil
	}
}

// findBinary looks for a flakeguard binary in the current directory, then the project root, then in dist/ (generated by goreleaser),
// then returns the path to the binary and the time it was built for the current OS and ARCH.
// If the binary is not found, it returns an empty string and time.Time{}
func findBinary() (path string, builtTime time.Time, err error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", time.Time{}, err
	}

	// First check for a binary in the current directory
	flakeguardBinaryPath := filepath.Join(currentDir, flakeguardCoveredBinary)
	if _, err := os.Stat(flakeguardBinaryPath); err == nil {
		return flakeguardBinaryPath, time.Time{}, nil
	}

	// Find the project root (where go.mod should be)
	projectRoot, err := findProjectRoot(currentDir)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to find project root: %w", err)
	}

	// Look in the project root for the binary
	possibleBinaryPath := filepath.Join(projectRoot, flakeguardCoveredBinary)
	if _, err := os.Stat(possibleBinaryPath); err == nil {
		return possibleBinaryPath, time.Time{}, nil
	}

	// Look in the dist directory at project root
	distPath := filepath.Join(projectRoot, "dist")
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		return "", time.Time{}, nil // No dist directory, return empty path
	}

	var foundBinaryPath string
	err = filepath.WalkDir(distPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() == flakeguardCoveredBinary {
			// Check the folder name against our current platform
			platform := fmt.Sprintf("%s_%s", runtime.GOOS, runtime.GOARCH)
			if !strings.Contains(filepath.Dir(path), platform) {
				return nil
			}
			foundBinaryPath = path
			return fs.SkipDir // Stop walking the directory
		}
		return nil
	})
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to walk dist directory: %w", err)
	}

	if foundBinaryPath == "" {
		return "", time.Time{}, nil
	}
	flakeguardBinaryInfo, err := os.Stat(foundBinaryPath)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to stat flakeguard binary: %w", err)
	}

	return foundBinaryPath, flakeguardBinaryInfo.ModTime(), err
}

// findProjectRoot walks up the directory tree to find the project root (where go.mod is located)
func findProjectRoot(startDir string) (string, error) {
	dir := startDir
	for {
		// Check if go.mod exists in current directory
		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			// Parse the go.mod file to verify it's the correct module
			goModContent, err := os.ReadFile(goModPath)
			if err != nil {
				return "", fmt.Errorf("failed to read go.mod file: %w", err)
			}

			modFile, err := modfile.Parse(goModPath, goModContent, nil)
			if err != nil {
				return "", fmt.Errorf("failed to parse go.mod file: %w", err)
			}

			if modFile.Module.Mod.Path == flakeguardRootModulePath {
				return dir, nil
			}
		}

		// Move up one directory
		parent := filepath.Dir(dir)

		// If we've reached the filesystem root, stop
		if parent == dir {
			break
		}

		dir = parent
	}

	return "", fmt.Errorf("could not find go.mod file starting from %s", startDir)
}

// setupCoverageCollection sets up the GOCOVERDIR for testscript to use
func setupCoverageCollection(env *testscript.Env) error {
	var (
		coverageDir     string
		userCoverageDir = os.Getenv(CoverDirEnvVar)
	)

	// Check for a user-specified coverage directory first. This should be a global dir that is shared by all test runs.
	if userCoverageDir == "" {
		return fmt.Errorf(
			"no coverage directory specified, integration tests will not collect coverage data. Set the environment variable '%s' to enable coverage collection",
			CoverDirEnvVar,
		)
	}

	if !strings.HasSuffix(userCoverageDir, integrationTestCoverageDir) {
		coverageDir = filepath.Join(userCoverageDir, integrationTestCoverageDir)
	} else {
		coverageDir = userCoverageDir
	}

	if err := os.MkdirAll(coverageDir, 0750); err != nil {
		return fmt.Errorf("failed to create coverage directory: %w", err)
	}

	// Override Go's temporary GOCOVERDIR with our desired directory
	env.Setenv("GOCOVERDIR", coverageDir)
	return nil
}
