package testhelpers

import (
	"fmt"
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

// CopyFile copies a file from src to dst.
// Handy for copying files needed for testing into a temp directory.
func CopyFile(tb testing.TB, src, dst string) error {
	tb.Helper()

	if _, err := os.Stat(src); os.IsNotExist(err) {
		return fmt.Errorf("source file %s does not exist", src)
	}

	if err := os.MkdirAll(filepath.Dir(dst), 0750); err != nil {
		return fmt.Errorf("failed to create directory '%s' for destination file '%s': %w", filepath.Dir(dst), dst, err)
	}

	destInfo, err := os.Stat(dst)
	if err == nil {
		if destInfo.IsDir() {
			return fmt.Errorf("destination file '%s' is a directory", dst)
		}
		return fmt.Errorf("destination file '%s' already exists", dst)
	}

	srcFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("failed to open source file '%s': %w", src, err)
	}
	defer func() {
		if err := srcFile.Close(); err != nil {
			tb.Logf("Error closing file %s: %v", src, err)
		}
	}()

	dstFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("failed to create destination file '%s': %w", dst, err)
	}
	defer func() {
		if err := dstFile.Close(); err != nil {
			tb.Logf("Error closing file %s: %v", dst, err)
		}
	}()

	_, err = srcFile.WriteTo(dstFile)
	if err != nil {
		return fmt.Errorf("failed to copy file '%s' to '%s': %w", src, dst, err)
	}

	return nil
}
