//go:build examples

// Package subtests contains common subtest and table test scenarios.
// The main purpose is to check that Flakeguard can accurately detect and quarantine subtests, a tricky edge case for code parsing.
package subtests

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPass(t *testing.T) {
	t.Parallel()

	t.Run("passing subtest", func(t *testing.T) {
		t.Log("passing subtest")
	})

	t.Run(fmt.Sprintf("passing subtest with dynamic name %d", 1), func(t *testing.T) {
		t.Log("passing subtest with dynamic name")
	})
}

func TestFail(t *testing.T) {
	t.Parallel()

	t.Run("failing subtest", func(t *testing.T) {
		t.Log("failing subtest")
		t.Fail()
	})

	t.Run(fmt.Sprintf("failing subtest with dynamic name %d", 1), func(t *testing.T) {
		t.Log("failing subtest with dynamic name")
		t.Fail()
	})
}

func TestTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		want int
	}{
		{name: "test passing", want: 1},
		{name: "test failing", want: 2},
	}

	for _, test := range tests {
		t.Run(fmt.Sprintf("%s %d", test.name, test.want), func(t *testing.T) {
			t.Log(test.name)
			require.Equal(t, test.want, 1)
		})
	}
}
