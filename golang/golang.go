package golang

import (
	"errors"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/packages"
)

// TODO: Implement
func QuarantineTest(testName string) error {
	return errors.New("not implemented")
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
func Packages(rootDir string) ([]PackageInfo, error) {
	config := &packages.Config{
		Mode:  packages.NeedName | packages.NeedModule,
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

	return result, nil
}
