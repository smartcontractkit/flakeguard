# Ideal Developer Experiences

What the Flakeguard experience ideally looks like for a developer. This are human stories intended to drive the UX direction of the project as a whole, and aren't necessarily detailed design breakdowns.

## Making a PR With No New Flakes Introduced

I am a developer who is making a PR to `packageA`. My changes are minor, and don't change how flaky tests are, either through modifying the test, or the underlying code.

### Ideal Experience

1. Flakeguard runs all tests.
2. Flakeguard recognizes that `packageA` has changed, and runs all the `packageA` tests in a `detect` loop.
3. Flakeguard finds no changes in flakiness and passes the PR.

### Avoid

* Don't extend test runtime by any significant amount; I shouldn't be waiting 2-3 times longer on CI waiting for re-runs.
* Don't pollute the comment space on the PR. If there aren't any issues, I don't want to hear about it.

## Making a PR The Introduces a New, Non-Flaky Test

### Ideal Experience

1. Flakeguard runs all tests.
2. Flakeguard recognizes that `packageA` has changed, and runs all the `packageA` tests in a `detect` loop.
3. Flakeguard finds no changes in flakiness and passes the PR.

### Avoid
