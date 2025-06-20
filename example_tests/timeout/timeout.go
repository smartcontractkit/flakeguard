//go:build examples

package timeout

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestTimeout(t *testing.T) {
	t.Parallel()

	deadline, ok := t.Deadline()
	require.True(t, ok, "This test should have a deadline")
	if time.Minute > time.Until(deadline) {
		t.Logf(
			"WARNING: This test is supposed to trigger a timeout, but the `-timeout` value is %s, which means you'll be waiting for a while.",
			time.Until(deadline),
		)
	}

	t.Logf("This test will sleep %s in order to timeout", time.Until(deadline).String())
	time.Sleep(time.Until(deadline))
	t.Logf("This test should have timed out")
}

func TestPass(t *testing.T) {
	t.Parallel()

	t.Log("This test should pass")
}
