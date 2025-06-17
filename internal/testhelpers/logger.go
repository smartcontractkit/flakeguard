// Package testhelpers provides helpers for testing.
package testhelpers

import (
	"io"
	"os"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/flakeguard/logging"
)

const (
	testLogLevelEnvVar = "FLAKEGUARD_TEST_LOG_LEVEL"
)

// Option is a function that sets an option for the test-specific logger.
type Option func(*options)

// options holds the options for the test-specific logger.
type options struct {
	logLevel   string
	writers    []io.Writer
	soleWriter io.Writer
}

// WithWriters sets additional writers to use for logging.
// This is useful for testing logging output that you also want written to default writers.
func WithWriters(writers ...io.Writer) Option {
	return func(o *options) {
		o.writers = writers
	}
}

// WithSoleWriter sets a custom writer to use instead of the default writers.
// This is useful for testing logging output that you don't want written anywhere else.
func WithSoleWriter(writer io.Writer) Option {
	return func(o *options) {
		o.soleWriter = writer
	}
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
		logging.WithWriters(opts.writers...),
	}
	log, err := logging.New(loggingOpts...)
	if opts.soleWriter != nil {
		log = log.Output(opts.soleWriter)
	}
	require.NoError(tb, err, "error setting up logging")
	log = log.With().Str("running_test", tb.Name()).Logger()
	return log
}
