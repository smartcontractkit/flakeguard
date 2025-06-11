# Flakeguard Design

An overview of how Flakeguard is designed and how it interacts with other services. This is a living doc and can change as the scope, plans, and existing structure of Flakeguard changes.

## Flakeguard `detect`

The `detect` command is used to run test suites over and over and detect if any of them are flaky.

## Flakeguard `guard`

The `guard` command will guard your CI/CD systems from suffering the consequences of flaky tests.

## Flakeguard's Outside Systems

Flakeguard (optionally, but ideally) relies on a few different outside systems in order to make it a more scalable solution.

* `Reporters`, systems like [Splunk](https://www.splunk.com/) and [DX](https://getdx.com/), are used to store and retrieve data on the status of your flaky tests (e.g. how flaky has TestX been in the past 7 days).
* `Ticketers`, systems like [Jira](https://jira.atlassian.com/), are used to create tickets that assign work to fix tests identified as flakes. Flakeguard scans for tickets that already exist to add more detail to them, or closed tickets for the same test, so that it can attach context.

```mermaid
flowchart LR
  reps[(Reporters)] --> Flakeguard
  ticks[/Ticketers/] --> Flakeguard
  Flakeguard --> reps
  Flakeguard --> ticks

```

## Flakeguard's Usage of [gotestsum](https://github.com/gotestyourself/gotestsum)

Flakeguard uses gotestsum to execute tests, and reads the JSON output after execution is complete. To do so, we call out to gotestsum and provide it with args for it and for those to pass on to `go test`. This lets us leverage `gotestsum`'s handy tools for console output and re-running failures.

### Why Not Use `--post-run-command` or `gotestsum tool`?

We might find reason to do this in the future, but for now it's not feasible if we wish to emulate a real test running environment. Especially for the `detect` command, we want to re-run test suites multiple times in a setup that emulates how they would run in a typical flow. If we exclusively use a post-run hook, we lose the ability to do this cleanly. Using `-count=n` doesn't accurately emulate how tests would actually run in a real environment multiple times.
