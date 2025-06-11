package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/smartcontractkit/flakeguard/exit"
	"github.com/smartcontractkit/flakeguard/git"
	"github.com/smartcontractkit/flakeguard/github"
	"github.com/smartcontractkit/flakeguard/logging"
	"github.com/smartcontractkit/flakeguard/report"
)

var (
	logger zerolog.Logger

	// Flag vars
	// Logging
	logFile           string
	logLevel          string
	enableConsoleLogs bool

	// Run behavior
	runs      int
	outputDir string
	dryRun    bool

	// Reporting
	splunkURL        string
	splunkToken      string
	splunkIndex      string
	splunkSourceType string

	dxWebhookURL string

	slackWebhookURL string
)

var rootCmd = &cobra.Command{
	Use:   "flakeguard [flags] [-- gotestsum-flags] [-- go-test-flags]",
	Short: "Detect and prevent flaky tests from disrupting CI/CD pipelines",
	Long: `Flakeguard helps you detect and prevent flaky tests from disrupting CI/CD pipelines.
It wraps gotestsum, so you can pass through all the flags you're used to using.

Examples:
  flakeguard -c -- --format testname -- ./pkg/...
  flakeguard --runs 10 -- --format dots -- -v -run TestMyFunction`,
	SilenceUsage: true,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		loggingOpts := []logging.Option{}
		if !enableConsoleLogs {
			loggingOpts = append(loggingOpts, logging.DisableConsoleLog())
		}
		if logLevel != "" {
			loggingOpts = append(loggingOpts, logging.WithLevel(logLevel))
		}
		if logFile != "" {
			loggingOpts = append(loggingOpts, logging.WithFileName(fmt.Sprintf("%s/%s", outputDir, logFile)))
		}

		var err error
		if err = os.MkdirAll(outputDir, 0750); err != nil {
			return exit.New(exit.CodeFlakeguardError, err)
		}
		logger, err = logging.New(loggingOpts...)
		if err != nil {
			return exit.New(exit.CodeFlakeguardError, err)
		}
		logger.Debug().
			Str("version", version).
			Str("commit", commit).
			Str("buildTime", buildTime).
			Str("builtBy", builtBy).
			Str("builtWith", builtWith).
			Str("goVersion", runtime.Version()).
			Str("os", runtime.GOOS).
			Str("arch", runtime.GOARCH).
			Str("logFile", logFile).
			Str("logLevel", logLevel).
			Bool("enableConsoleLogs", enableConsoleLogs).
			Int("runs", runs).
			Str("outputDir", outputDir).
			Msg("Run info")
		return nil
	},
}

func init() {
	// Logging
	rootCmd.PersistentFlags().
		StringVarP(&logFile, "log-file", "l", "flakeguard.log.json", "File to store flakeguard logs")
	rootCmd.PersistentFlags().StringVarP(&logLevel, "log-level", "L", "info", "Log level to use")
	rootCmd.PersistentFlags().
		BoolVarP(&enableConsoleLogs, "enable-console-logs", "c", false, "Enable console logs for flakeguard")

	// Run behavior
	rootCmd.PersistentFlags().
		IntVarP(&runs, "runs", "r", 5, "Number of times to run each test in detect mode, or the number of times to retry a test in guard mode")
	rootCmd.PersistentFlags().
		StringVarP(&outputDir, "output-dir", "o", "./flakeguard-output", "Directory to store flakeguard outputs")
	rootCmd.PersistentFlags().
		BoolVarP(&dryRun, "dry-run", "d", false, "Disables making any changes to the codebase and prevents reporting results to outside services (Splunk, Slack, etc.)")

	// Reporting
	// Splunk
	rootCmd.PersistentFlags().
		StringVar(&splunkURL, "splunk-url", "", "Splunk HTTP Event Collector URL")
	rootCmd.PersistentFlags().
		StringVar(&splunkToken, "splunk-token", "", "Splunk HTTP Event Collector token")
	rootCmd.PersistentFlags().
		StringVar(&splunkIndex, "splunk-index", "flakeguard_json", "Splunk index to send events to")
	rootCmd.PersistentFlags().
		StringVar(&splunkSourceType, "splunk-source-type", "flakeguard_json", "Splunk source type to send events to")

	// DX
	rootCmd.PersistentFlags().
		StringVar(&dxWebhookURL, "dx-webhook-url", "", "DX webhook URL to send events to")

	// Slack
	rootCmd.PersistentFlags().
		StringVar(&slackWebhookURL, "slack-webhook-url", "", "Slack webhook URL to send events to")

	// Disable flag parsing after -- to allow passing through to gotestsum
	rootCmd.Flags().SetInterspersed(false)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		logger.Error().Err(err).Msg("Failed to execute command")
		os.Exit(exit.GetCode(err))
	}
}

// parseArgs parses command line arguments to separate gotestsum flags from go test flags.
// flakeguard [flakeguard flags] -- [gotestsum flags] -- [go test flags]
func parseArgs(args []string) (gotestsumFlags []string, goTestFlags []string) {
	// Find the position of the first --
	dashPos := -1
	for i, arg := range args {
		if arg == "--" {
			dashPos = i
			break
		}
	}
	if dashPos != -1 {
		gotestsumFlags = args[:dashPos]
		goTestFlags = args[dashPos+1:]
	} else {
		gotestsumFlags = args
		goTestFlags = []string{}
	}

	logger.Debug().
		Strs("gotestsum_flags", gotestsumFlags).
		Strs("go_test_flags", goTestFlags).
		Msg("Parsed Flags")
	return gotestsumFlags, goTestFlags
}

func testRunInfo(l zerolog.Logger, repoPath string) (report.TestRunInfo, error) {
	repoInfo, err := git.GetRepoInfo(l, repoPath)
	if err != nil {
		return report.TestRunInfo{}, fmt.Errorf("failed to get repo info: %w", err)
	}

	t := report.TestRunInfo{
		RepoURL:         repoInfo.URL,
		RepoOwner:       repoInfo.Owner,
		RepoName:        repoInfo.Name,
		DefaultBranch:   repoInfo.DefaultBranch,
		OnDefaultBranch: repoInfo.CurrentBranch == repoInfo.DefaultBranch,
		HeadBranch:      repoInfo.CurrentBranch,
		HeadCommit:      repoInfo.CurrentCommit,
	}

	githubEnv, err := github.ActionsEnvVars()
	if err != nil && !errors.Is(err, github.ErrNotInActions) {
		return t, fmt.Errorf("failed to get GitHub Actions environment variables: %w", err)
	} else {
		t.GitHubEvent = githubEnv.EventName
		t.HeadBranch = githubEnv.HeadRef
		t.BaseBranch = githubEnv.BaseRef
	}

	return t, nil
}
