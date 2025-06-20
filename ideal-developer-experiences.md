# Ideal Developer Experiences

What the Flakeguard experience ideally looks like for a developer. This are human stories intended to drive the UX direction of the project as a whole, and aren't necessarily detailed design breakdowns.

## Making a PR to Code with Already Flaky Tests

I am a developer who is making a minor PR to a codebase that already has a lot of flaky tests. These tests aren't quarantined (maybe their flake-rate is under the threshold, or we determine the flakes are worth it for whatever reason), but I don't want them to block my PR.

### Ideal Experience

1. Flakeguard runs all tests.
2. Flakeguard notices the failures, and realizes they were already flaky before this PR.
3. Flakeguard notices that the flake rates are mostly the same before and after the PR, so the user did not make the flakes worse.
4. Flakeguard re-runs the tests x times in order to get them to pass.

### Avoid

* "Blaming" the already-present flakes on the new PR.
* Polluting the comment space on the PR. If there aren't any issues, I don't want to hear about it.

## Making a PR With No New Flakes Introduced

I am a developer who is making a PR to `packageA`. My changes are minor, and don't change how flaky tests are, either through modifying the test, or the underlying code. I don't modify any test directly.

### Ideal Experience

1. Flakeguard runs all tests.
2. Flakeguard recognizes that `packageA` has changed, and runs all the `packageA` tests in a `detect` loop.
3. Flakeguard finds no changes in flakiness and passes the PR.

### Avoid

* Extending test runtime by any significant amount; I shouldn't be waiting 2-3 times longer on CI waiting for re-runs to confirm flake status.
* Polluting the comment space on the PR. If there aren't any issues, I don't want to hear about it.

## Making a PR With a New/Modified, Non-Flaky Test

I am a developer who is making a PR to `packageA`. My changes include adding/modifying `TestA` directly, and it is in no way flaky.

### Ideal Experience

1. Flakeguard runs all tests.
2. Flakeguard recognizes that `packageA` has changed, and runs all the `packageA` tests in a `detect` loop.
3. Flakeguard recognizes that `TestA` has been directly changed, and runs extra cycles for `TestA` if deemed necessary.
4. No flakes are found in either `packageA` or `TestA` and CI passes.

### Avoid

* Extending test runtime by any significant amount; I shouldn't be waiting 2-3 times longer on CI waiting for re-runs to confirm flake status.
* Polluting the comment space on the PR. If there aren't any issues, I don't want to hear about it.

## Making a PR With a New/Modified, Flaky Test

I am a developer who is making a PR to `packageA`. My changes include adding/modifying `TestA` directly. This test is flaky.

### Ideal Experience

1. Flakeguard runs all tests.
2. Flakeguard recognizes that `packageA` has changed, and runs all the `packageA` tests in a `detect` loop.
3. Flakeguard recognizes that `TestA` has been directly changed, and runs extra cycles for `TestA` if deemed necessary.
4. Flakes are detected and the PR is blocked, commenting on the PR why.

### Avoid

* Letting the newly-flaky tests past CI

## Making a PR to Fix a Flaky Test

I'm a developer who is trying to fix a flaky test that has been quarantined. I make modifications to the test directly.

### Ideal Experience

1. Flakeguard runs all tests.
2. Flakeguard recognizes that a quarantined test has been modified, and runs it in a `detect` loop.
3. Flakeguard reports whether it suspects the test has been fixed, and potentially removes it from quarantine.

### Avoid

* Being too confident about a fix. If the test is flaky 1% of the time, we can't confidently call it fixed after running it 5 times.

## Integrating Flakeguard in CI

I am a developer who is integrating Flakeguard in my CI flow.

### Ideal Experience

1. The `flakeguard` command is a simple, drop-in replacement for `go test`.
2. It should only require changing my `go test` line, or perhaps importing a simple GitHub Action.

### Avoid

* Determining CI flow for the user. We only want to replace a command, not ask that they restructure CI in large ways.
* Having more than a single step.
