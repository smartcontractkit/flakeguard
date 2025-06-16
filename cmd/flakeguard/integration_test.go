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

	"github.com/rogpeppe/go-internal/testscript"
	"golang.org/x/mod/modfile"

	"github.com/smartcontractkit/flakeguard/internal/testhelpers"
)

const (
	distDir                  = "dist"
	flakeguardCoveredBinary  = "flakeguard_covered"
	flakeguardRootModulePath = "github.com/smartcontractkit/flakeguard"
)

var flakeguardBinaryPath string

func TestMain(m *testing.M) {
	var err error
	flakeguardBinaryPath, err = findBinary()
	if err != nil {
		log.Fatalf("Hit error while looking for flakeguard binary for integration tests: %v", err)
	}

	// Check if a coverage-instrumented binary exists
	if flakeguardBinaryPath != "" {
		fmt.Printf("Using flakeguard binary '%s' for integration tests\n", flakeguardBinaryPath)
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

func TestIntegrationScripts(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration tests with -short")
	}
	l := testhelpers.Logger(t)

	testscript.Run(t, testscript.Params{
		Dir: "testscripts",
		Setup: func(env *testscript.Env) error {
			// Copy example_tests directory to the testscript working directory
			exampleTestsSource := filepath.Join("..", "..", "example_tests")
			exampleTestsDest := filepath.Join(env.WorkDir, "example_tests")

			if err := testhelpers.CopyDir(t, exampleTestsSource, exampleTestsDest); err != nil {
				return err
			}

			if flakeguardBinaryPath != "" {
				l.Debug().Str("path", flakeguardBinaryPath).Msg("Found flakeguard binary for integration tests")
				if err := os.Symlink(flakeguardBinaryPath, filepath.Join(env.WorkDir, "flakeguard")); err != nil {
					return err
				}
			} else {
				l.Debug().Msg("No flakeguard binary found, using in-process compilation for integration tests")
			}

			return nil
		},
	})
}

// findBinary looks for a flakeguard binary in the current directory and in dist/, then returns the path to the binary
// if the binary is not found, it returns an empty string
func findBinary() (string, error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// First check for a binary in the current directory
	flakeguardBinaryPath := filepath.Join(currentDir, flakeguardCoveredBinary)
	if _, err := os.Stat(flakeguardBinaryPath); err == nil {
		return flakeguardBinaryPath, nil
	}

	// Find the project root (where go.mod should be)
	projectRoot, err := findProjectRoot(currentDir)
	if err != nil {
		return "", fmt.Errorf("failed to find project root: %w", err)
	}

	// Look in the dist directory at project root
	distPath := filepath.Join(projectRoot, "dist")
	if _, err := os.Stat(distPath); os.IsNotExist(err) {
		return "", nil // No dist directory, return empty path
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

	return foundBinaryPath, err
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
