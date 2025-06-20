//go:build examples

// Package flaky has multiple tests designed to flake at different rates.
// P(flake) = 1 - (0.1 * 0.25 * 0.5) = 0.9875 = 98.75%
package flaky

import (
	"fmt"
	"math/rand"
	"testing"
)

func TestFlakeTenPercent(t *testing.T) {
	t.Parallel()

	if rand.Intn(10) == 0 {
		t.Log("I flake 10% of the time")
		t.FailNow()
	}
}

func TestFlakeTwentyFivePercent(t *testing.T) {
	t.Parallel()

	if rand.Intn(4) == 0 {
		t.Log("I flake 25% of the time")
		t.FailNow()
	}
}

func TestFlakeFiftyPercent(t *testing.T) {
	t.Parallel()

	if rand.Intn(2) == 0 {
		t.Log("I flake 50% of the time")
		t.FailNow()
	}
}

func TestPass(t *testing.T) {
	t.Parallel()

	for i := range 10 {
		t.Run(fmt.Sprintf("pass-%d", i), func(t *testing.T) {
			t.Log("I pass")
		})
	}
}
