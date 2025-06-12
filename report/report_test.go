package report

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/flakeguard/internal/testhelpers"
)

const testData = "testdata"

func TestReadTestOutput(t *testing.T) {
	t.Parallel()

	logger := testhelpers.Logger(t)
	lines, err := readTestOutput(
		logger,
		testData,
		"example_all.log.json",
		"example_flaky.log.json",
	)
	require.NoError(t, err)
	// TODO: Better validation
	require.Len(t, lines, 638)
}
