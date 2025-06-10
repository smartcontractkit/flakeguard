package golang

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/flakeguard/internal/testhelpers"
)

// TODO: Improve testing
func TestPackages(t *testing.T) {
	t.Parallel()

	l := testhelpers.Logger(t)
	packages, err := Packages(l, ".")
	require.NoError(t, err)
	require.Greater(t, len(packages), 0)
}
