package main

import (
	"path/filepath"
	"testing"

	"github.com/rogpeppe/go-internal/testscript"

	"github.com/smartcontractkit/flakeguard/internal/testhelpers"
)

func TestMain(m *testing.M) {
	testscript.Main(m, map[string]func(){
		"flakeguard": main,
	})
}

func TestIntegrationScripts(t *testing.T) {
	t.Parallel()

	if testing.Short() {
		t.Skip("skipping integration tests with -short")
	}

	testscript.Run(t, testscript.Params{
		Dir: "testscripts",
		Setup: func(env *testscript.Env) error {
			// Copy example_tests directory to the testscript working directory
			exampleTestsSource := filepath.Join("..", "..", "example_tests")
			exampleTestsDest := filepath.Join(env.WorkDir, "example_tests")

			return testhelpers.CopyDir(t, exampleTestsSource, exampleTestsDest)
		},
	})
}
