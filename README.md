# 🧪 Flakeguard

[![Go Report Card](https://goreportcard.com/badge/github.com/smartcontractkit/flakeguard)](https://goreportcard.com/report/github.com/smartcontractkit/flakeguard)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

Flakeguard is a powerful tool that extends [gotestsum](https://github.com/gotestsuite/gotestsum) to detect and prevent flaky tests in your Go projects. It runs your test suite multiple times and analyzes the results to identify tests that pass sometimes and fail other times, helping you maintain a more reliable test suite.

## ✨ Features

- **🔄 Multiple Test Runs**: Automatically runs your test suite multiple times to detect inconsistent behavior
- **📊 Comprehensive Reporting**: Generates detailed JSON and human-readable reports
- **⚙️ Configurable Thresholds**: Set custom pass rate thresholds to define what constitutes a "flaky" test
- **🔧 Gotestsum Integration**: Seamlessly integrates with gotestsum, passing through all flags and arguments
- **📈 Pass Rate Analysis**: Calculates pass rates for each test across multiple runs
- **🎯 Configuration Management**: Support for JSON configuration files
- **📝 Detailed Output**: Provides clear, actionable information about flaky tests

## 🚀 Installation

### From Source

```bash
go install github.com/smartcontractkit/flakeguard@latest
```

### Build from Repository

```bash
git clone https://github.com/smartcontractkit/flakeguard.git
cd flakeguard
go build -o flakeguard
```

## 📋 Prerequisites

Flakeguard requires [gotestsum](https://github.com/gotestsuite/gotestsum) to be installed and available in your PATH:

```bash
go install gotest.tools/gotestsum@latest
```

## 🔧 Usage

### Basic Usage

Run flakeguard on your entire test suite:

```bash
flakeguard ./...
```

Run flakeguard on specific packages:

```bash
flakeguard ./pkg/mypackage
```

### Configuration Options

```bash
# Run tests 5 times instead of the default 3
flakeguard --runs 5 ./...

# Set flaky threshold to 90% (tests passing less than 90% are considered flaky)
flakeguard --threshold 0.9 ./...

# Specify custom output directory
flakeguard --output-dir ./test-reports ./...

# Enable verbose logging
flakeguard --verbose ./...

# Pass arguments through to gotestsum
flakeguard --gotestsum-args="--format,testname" ./...
```

### Configuration File

Create a configuration file for consistent settings:

```bash
# Generate a default configuration file
flakeguard config init

# Show current configuration
flakeguard config show

# Use a specific configuration file
flakeguard --config ./custom-config.json ./...
```

Example `flakeguard.json`:

```json
{
  "runs": 5,
  "threshold": 0.85,
  "output_dir": "./flakeguard-reports",
  "gotestsum_args": ["--format", "testname", "--junitfile", "junit.xml"],
  "ignore_patterns": ["*_integration_test.go"],
  "slow_test_threshold": "30s"
}
```

## 📊 Understanding the Output

### Console Output

```
🧪 Flakeguard Analysis Complete
================================
Total Runs: 3
Total Tests: 12

⚠️  Flaky Tests Detected: 2

  🔄 example.TestFlaky
     Pass Rate: 66.7% (2 passes, 1 failures in 3 runs)

  🔄 example.TestVeryFlaky
     Pass Rate: 33.3% (1 passes, 2 failures in 3 runs)

📊 Reports saved to: ./flakeguard-reports
```

### Generated Reports

Flakeguard generates several types of reports:

1. **`flakeguard-report.json`**: Detailed JSON report with all test results and metadata
2. **`flakeguard-summary.txt`**: Human-readable summary report
3. **`run_N.json`**: Individual gotestsum JSON output for each run

### Report Structure

The JSON report includes:
- **Test Summary**: Pass rates, failure counts, and flaky status for each test
- **Run Results**: Detailed results from each individual test run
- **Flaky Tests**: List of tests identified as flaky based on your threshold
- **Metadata**: Configuration used, timestamps, and analysis parameters

## 🛠️ Advanced Usage

### CI/CD Integration

Use flakeguard in your CI pipeline to catch flaky tests before they reach production:

```yaml
# GitHub Actions example
- name: Run Flaky Test Detection
  run: |
    flakeguard --runs 5 --threshold 0.95 ./...

# Fail the build if flaky tests are detected
- name: Check for Flaky Tests
  run: |
    if [ $(jq '.flaky_tests | length' flakeguard-reports/flakeguard-report.json) -gt 0 ]; then
      echo "Flaky tests detected!"
      exit 1
    fi
```

### Custom Gotestsum Arguments

Pass any gotestsum arguments through flakeguard:

```bash
# Use specific test format and output
flakeguard --gotestsum-args="--format,dots,--junitfile,junit.xml" ./...

# Run with race detection
flakeguard -- -race ./...

# Run specific tests with verbose output
flakeguard -- -v -run TestSpecific ./...
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [gotestsum](https://github.com/gotestsuite/gotestsum) for the excellent test runner that powers this tool
- The Go testing community for inspiration and best practices

## 🐛 Issues and Support

If you encounter any issues or have questions:

1. Check the [existing issues](https://github.com/smartcontractkit/flakeguard/issues)
2. Create a new issue with detailed information about your problem
3. Include your configuration, command used, and any relevant output

---

**Happy Testing! 🧪**
