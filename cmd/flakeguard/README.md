# Flakeguard Integration Tests

We use the [testscript package](https://pkg.go.dev/github.com/rogpeppe/go-internal/testscript) to run integration tests on Flakeguard commands.

## Setup and Code Coverage

1. Build a binary with the `-cover` flag.
2. Copy the binary over into the `testscript` environment and use it to run integration tests (instead of just hooking into the main function).
3. Collect all coverage in a universal directory specified by the `FLAKEGUARD_GOCOVERDIR` env var. (We can't just use `GOCOVERDIR` as testscript creates its own and things get weird.)
4. Print out coverage for both unit and integration tests, then combine them and print out full coverage stats using the `test_coverage_report` command in our [Makefile](../../Makefile).
