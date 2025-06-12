# Flakeguard

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Flakeguard helps you detect flaky tests, quarantine them, and guard your CI pipelines from them.

## Usage

Flakeguard sits on top of [gotestsum](https://github.com/gotestyourself/gotestsum) to run tests and organize output.

```sh
# Print help output
flakeguard -h

# Typical usage
flakeguard [detect | guard] [flakeguard-flags] [-- gotestsum-flags] [-- go-test-flags]
```

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

## Contributing

We use [golangci-lint v2](https://golangci-lint.run/) for linting and formatting, and [pre-commit](https://pre-commit.com/) for pre-commit and pre-push checks.

```sh
pre-commit install
```
