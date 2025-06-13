package testhelpers

import (
	"os"
	"path/filepath"
	"testing"
)

// CopyDir recursively copies a directory from src to dst.
// Handy for copying files needed for testing into a temp directory.
func CopyDir(tb testing.TB, src, dst string) error {
	tb.Helper()

	err := os.MkdirAll(dst, 0750)
	if err != nil {
		return err
	}

	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)

		if info.IsDir() {
			return os.MkdirAll(dstPath, info.Mode())
		}

		srcFile, err := os.Open(path)
		if err != nil {
			return err
		}
		defer func() {
			if err := srcFile.Close(); err != nil {
				tb.Logf("Error closing file %s: %v", path, err)
			}
		}()

		dstFile, err := os.Create(dstPath)
		if err != nil {
			return err
		}
		defer func() {
			if err := dstFile.Close(); err != nil {
				tb.Logf("Error closing file %s: %v", dstPath, err)
			}
		}()

		_, err = srcFile.WriteTo(dstFile)
		return err
	})
}
