// Package golang enables ways to inspect and edit Go code.
package golang

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"golang.org/x/tools/go/packages"
)

// Absolute path to root directory -> PackageInfo
var (
	packagesCache      = map[string][]PackageInfo{}
	packagesCacheMutex = sync.RWMutex{}

	ErrTestNotFound = errors.New("test not found")
)

// TODO: Implement to add the quarantine call to a test function
func QuarantineTest(packageName, testName string) error {
	// TODO: See if we can borrow gotestsum's approach: https://github.com/gotestyourself/gotestsum/tree/v1.12.2/cmd/tool/slowest
	return errors.New("not implemented")
}

// TestLocation contains information about where a test function is located
type TestLocation struct {
	FilePath   string // Relative path to the file containing the test
	LineNumber int    // Line number where the test function is defined
}

// FindTestLocation finds the location of a test function in a package
func FindTestLocation(l zerolog.Logger, rootDir, pkgImportPath, testName string) (*TestLocation, error) {
	l = l.With().Str("rootDir", rootDir).Str("pkgImportPath", pkgImportPath).Str("testName", testName).Logger()
	l.Trace().Msg("Finding test location")
	start := time.Now()

	pkgs, err := Packages(l, rootDir)
	if err != nil {
		return nil, err
	}

	var testLocation *TestLocation

	for _, pkg := range pkgs {
		if pkg.ImportPath == pkgImportPath {
			for _, testFile := range pkg.TestGoFiles {
				testLocation, err = findTestInFile(testFile, testName)
				if err != nil {
					return nil, fmt.Errorf("error finding test '%s' in file '%s': %w", testName, testFile, err)
				}
				if testLocation != nil {
					break
				}
			}
		}
	}

	if testLocation == nil {
		l.Warn().Msg("Could not find test location")
		return nil, fmt.Errorf("%w: looking for test '%s' in package '%s'", ErrTestNotFound, testName, pkgImportPath)
	}

	l.Trace().
		Str("testFile", testLocation.FilePath).
		Int("lineNumber", testLocation.LineNumber).
		Str("duration", time.Since(start).String()).
		Msg("Found test location")
	return testLocation, nil
}

// findTestInFile finds the location of a test function in a file
// It returns a TestLocation if the test is found, otherwise it returns a nil location
func findTestInFile(testFile, testName string) (*TestLocation, error) {
	fset := token.NewFileSet()
	fileAst, err := parser.ParseFile(fset, testFile, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("error parsing file '%s': %w", testFile, err)
	}

	var testLocation *TestLocation
	ast.Inspect(fileAst, func(n ast.Node) bool {
		// Look for function declarations
		fn, ok := n.(*ast.FuncDecl)
		if !ok || fn.Recv != nil || !isTestFunc(fn.Name.Name) {
			return true
		}
		if fn.Name.Name == testName {
			testLocation = &TestLocation{
				FilePath:   testFile,
				LineNumber: fset.Position(fn.Pos()).Line,
			}
			return false
		}
		return true
	})

	// If we didn't find the test, return a nil location
	return testLocation, nil
}

func isTestFunc(name string) bool {
	return strings.HasPrefix(name, "Test") || strings.HasPrefix(name, "Example") ||
		strings.HasPrefix(name, "Benchmark") ||
		strings.HasPrefix(name, "Fuzz")
}

// PackageInfo contains comprehensive information about a Go package
type PackageInfo struct {
	ImportPath   string   // Package import path (e.g., "github.com/user/repo/pkg")
	Name         string   // Package name
	Dir          string   // Directory containing the package
	GoFiles      []string // .go source files
	TestGoFiles  []string // _test.go files
	XTestGoFiles []string // _test.go files with different package names
	Module       string   // Module path
	IsCommand    bool     // True if this is a main package
}

// Packages finds all Go packages in the given directory and subdirectories
func Packages(l zerolog.Logger, rootDir string) ([]PackageInfo, error) {
	absRootDir, err := filepath.Abs(rootDir)
	if err != nil {
		return nil, err
	}

	packagesCacheMutex.RLock()
	cachedPackages, ok := packagesCache[absRootDir]
	packagesCacheMutex.RUnlock()
	if ok {
		return cachedPackages, nil
	}

	l = l.With().Str("rootDir", rootDir).Str("absRootDir", absRootDir).Logger()
	l.Trace().Msg("Loading packages")
	start := time.Now()
	config := &packages.Config{
		Mode:  packages.NeedName | packages.NeedModule | packages.NeedFiles,
		Dir:   rootDir,
		Tests: true,
	}

	// Use "./..." pattern to find all packages recursively
	pkgs, err := packages.Load(config, "./...")
	if err != nil {
		return nil, err
	}

	var result []PackageInfo
	for _, pkg := range pkgs {
		if len(pkg.Errors) > 0 {
			l.Error().Err(pkg.Errors[0]).Msg("Error loading package")
			// Skip packages with errors, but continue processing others
			continue
		}

		info := PackageInfo{
			ImportPath: pkg.PkgPath,
			Name:       pkg.Name,
			IsCommand:  pkg.Name == "main",
		}

		if pkg.Module != nil {
			info.Module = pkg.Module.Path
		}

		// Separate regular files from test files
		for _, file := range pkg.GoFiles {
			if strings.HasSuffix(file, "_test.go") {
				info.TestGoFiles = append(info.TestGoFiles, file)
			} else {
				info.GoFiles = append(info.GoFiles, file)
			}
		}

		// Get directory from first file
		if len(pkg.GoFiles) > 0 {
			info.Dir = filepath.Dir(pkg.GoFiles[0])
		}

		result = append(result, info)
	}

	sort.Slice(result, func(i, j int) bool {
		return result[i].ImportPath < result[j].ImportPath
	})

	packagesCacheMutex.Lock()
	packagesCache[absRootDir] = result
	packagesCacheMutex.Unlock()

	for _, pkg := range result {
		l.Trace().
			Strs("files", pkg.GoFiles).
			Strs("testFiles", pkg.TestGoFiles).
			Strs("xTestFiles", pkg.XTestGoFiles).
			Str("name", pkg.Name).
			Str("importPath", pkg.ImportPath).
			Str("module", pkg.Module).
			Bool("isCommand", pkg.IsCommand).
			Str("pkgDir", pkg.Dir).
			Msg("Loaded package")
	}

	l.Trace().Str("duration", time.Since(start).String()).Msg("Loaded packages")
	return result, nil
}
