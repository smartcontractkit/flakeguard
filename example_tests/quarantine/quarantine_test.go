//go:build examples

package quarantine

import (
	"testing"

	"github.com/smartcontractkit/flakeguard"
)

func TestQuarantined(t *testing.T) {
	t.Parallel()

	flakeguard.Quarantine(t, "quarantine example")
	t.Log("I'm quarantined")
	t.FailNow()
}

func TestNotQuarantined(t *testing.T) {
	t.Parallel()

	t.Log("I'm not quarantined")
}
