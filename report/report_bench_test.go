package report

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/flakeguard/internal/testhelpers"
)

func BenchmarkReadTestOutput(b *testing.B) {
	logger := testhelpers.Logger(b, testhelpers.Silent())

	for b.Loop() {
		_, err := readTestOutput(logger, filepath.Join(testData, "example_all.log.json"))
		require.NoError(b, err)
	}
}
