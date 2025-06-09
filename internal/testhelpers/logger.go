package testhelpers

import (
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/flakeguard/logging"
)

const (
	testLogLevelEnvVar = "FLAKEGUARD_TEST_LOG_LEVEL"
)

type Option func(*options)

type options struct {
	logLevel string
}

// Silent disables tests logging to console.
// Can also be set via the FLAKEGUARD_TEST_LOG_LEVEL environment variable.
// This option takes precedence over the FLAKEGUARD_TEST_LOG_LEVEL environment variable.
func Silent() Option {
	return func(o *options) {
		o.logLevel = "disabled"
	}
}

// LogLevel sets the log level for the test.
// Can also be set via the FLAKEGUARD_TEST_LOG_LEVEL environment variable.
// This option takes precedence over the FLAKEGUARD_TEST_LOG_LEVEL environment variable.
func LogLevel(level string) Option {
	return func(o *options) {
		o.logLevel = level
	}
}

func defaultOptions() *options {
	return &options{
		logLevel: "debug",
	}
}

// Logger creates a new logger for the test.
func Logger(tb testing.TB, options ...Option) zerolog.Logger {
	tb.Helper()

	opts := defaultOptions()
	envLogLevel := os.Getenv(testLogLevelEnvVar)
	if envLogLevel != "" {
		opts.logLevel = envLogLevel
	}
	for _, opt := range options {
		opt(opts)
	}

	loggingOpts := []logging.Option{
		logging.DisableFileLogging(),
		logging.WithLevel(opts.logLevel),
	}
	log, err := logging.New(loggingOpts...)
	require.NoError(tb, err, "error setting up logging")
	log = log.With().Str("test_name", tb.Name()).Logger()
	return log
}
