# Flakeguard

[![Go Reference](https://pkg.go.dev/badge/github.com/smartcontractkit/flakeguard.svg)](https://pkg.go.dev/github.com/smartcontractkit/flakeguard)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Report Card](https://goreportcard.com/badge/github.com/smartcontractkit/flakeguard)](https://goreportcard.com/report/github.com/smartcontractkit/flakeguard)


Flakeguard helps you detect flaky tests, quarantine them, and guard your CI pipelines from them.

## Usage

```sh
# Print help output
flakeguard -h
```

Once you find a flaky test, take a look at our [Fixing Flaky Tests Guide](./fixing-flaky-tests-guide.md) for tips and processes to help you debug and narrow down the source of flakes.

There are two modes for running flakeguard, `detect` and `guard`.

### `detect`

Run tests over and over to figure out which ones are flaky. Best used in a nightly cron job to regularly checkup on your test suite health.

```sh
flakeguard detect -h
```

### `guard`

Guard your CI pipelines from being affected by flaky tests. This will attempt to re-run any failing tests and make them pass so that your PRs and merge queues aren't decimated by getting unlucky with flakes. It will also look for newly added and modified tests and try to determine if your code changes are introducing new flakes.

```sh
flakeguard guard -h
```

## Design

For detailed technical design diagrams and decisions, see the [Flakeguard Design Doc](./design.md). For guiding principles for UX, see the [Ideal Flakeguard Developer Experiences](./ideal-developer-experiences.md) page.

## Contributing

We use [golangci-lint v2](https://golangci-lint.run/) for linting and formatting, and [pre-commit](https://pre-commit.com/) for pre-commit and pre-push checks.

```sh
pre-commit install # Install our pre-commit scripts
```

See the [Makefile](./Makefile) for helpful commands for local development.

```sh
make build            # Build binaries, results placed in dist/

make lint             # Lint and format code

make test_short       # Run only short tests
make test_unit        # Run only unit tests
make test_integration # Run only integration tests
make test_full        # Run all tests with extensive coverage stats
make test_full_race   # Run all tests with extensive coverage stats and race detection
```

### Cursor

If you use [Cursor](https://www.cursor.com/), you can utilize the [Cursor rules](https://docs.cursor.com/context/rules#workflow-automation) included in [.cursor/rules](./.cursor/rules/) to guide suggestions.

### Test

* `FLAKEGUARD_TEST_LOG_LEVEL` Sets the logging level for tests to use, use `FLAKEGUARD_TEST_LOG_LEVEL=trace` if you're chasing down confusing bugs.
* `FLAKEGUARD_GOCOVERDIR` Sets the coverage dir for integration tests to utilize. This is handled automatically in most `make` commands.

A handy debugging setup is available already in the included [.vscode](.vscode/) folder as "Debug Flakeguard".
