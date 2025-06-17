// Package cmd is the home of the cobra commands for flakeguard.
package cmd

import (
	"errors"
	"fmt"
	"os"
	"runtime"

	"github.com/go-git/go-git/v5"
	"github.com/rs/zerolog"
	"github.com/spf13/cobra"

	"github.com/smartcontractkit/flakeguard/exit"
	fg_git "github.com/smartcontractkit/flakeguard/git"
	fg_github "github.com/smartcontractkit/flakeguard/github"
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

	// GitHub
	// Flag for GitHub token
	githubToken string
	// Client for GitHub API
	githubClient *fg_github.Client

	// Reporting
	splunkURL        string
	splunkToken      string
	splunkIndex      string
	splunkSourceType string

	dxWebhookURL string

	slackWebhookURL string
)

var rootCmd = &cobra.Command{
	Use:   "flakeguard [detect | guard] [flakeguard-flags] [-- gotestsum-flags] [-- go-test-flags]",
	Short: "Detect and prevent flaky tests from disrupting CI/CD pipelines",
	Long: `Flakeguard helps you detect and prevent flaky tests from disrupting CI/CD pipelines.
It wraps gotestsum, so you can pass through all the flags you're used to using.

Examples:
  flakeguard -c -- --format testname -- ./pkg/...
  flakeguard --runs 10 -- --format dots -- -v -run TestMyFunction`,
	SilenceUsage: true,
	PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
		// Setup logging
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

		githubClient, err = fg_github.NewClient(logger, githubToken, nil)
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
		StringVar(&logFile, "log-file", "flakeguard.log.json", "File to store flakeguard logs")
	rootCmd.PersistentFlags().
		StringVarP(&logLevel, "log-level", "l", "info", "Log level to use")
	rootCmd.PersistentFlags().
		BoolVarP(&enableConsoleLogs, "enable-console-logs", "c", false, "Enable console logs for flakeguard")

	// Run behavior
	rootCmd.PersistentFlags().
		IntVarP(&runs, "runs", "r", 5, "Number of times to run each test in detect mode, or the number of times to retry a test in guard mode")
	rootCmd.PersistentFlags().
		StringVarP(&outputDir, "output-dir", "o", "./flakeguard-output", "Directory to store flakeguard outputs")
	rootCmd.PersistentFlags().
		BoolVarP(&dryRun, "dry-run", "d", false, "Disables making any changes to the codebase and prevents reporting results to outside services (Splunk, Slack, etc.)")

	// GitHub
	rootCmd.PersistentFlags().
		StringVarP(&githubToken, "github-token", "t", "", "GitHub token to use for GitHub API requests, if not provided, the GITHUB_TOKEN environment variable will be used")

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

// Execute executes the root flakeguard command.
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
	gotestsumFlags = make([]string, 0, len(args))
	goTestFlags = make([]string, 0, len(args))
	usingGotestsumFlags := true
	for _, arg := range args {
		if arg == "--" {
			usingGotestsumFlags = false
			continue
		}
		if usingGotestsumFlags {
			gotestsumFlags = append(gotestsumFlags, arg)
		} else {
			goTestFlags = append(goTestFlags, arg)
		}
	}

	return gotestsumFlags, goTestFlags
}

func testRunInfo(
	l zerolog.Logger,
	githubClient *fg_github.Client,
	repoPath string,
) (report.TestRunInfo, error) {
	repoInfo, err := fg_git.ReadBasicRepoInfo(l, repoPath)
	if errors.Is(err, git.ErrRepositoryNotExists) {
		return report.TestRunInfo{}, nil
	} else if err != nil {
		return report.TestRunInfo{}, fmt.Errorf("failed to get repo info: %w", err)
	}

	t := report.TestRunInfo{
		RepoURL:    repoInfo.URL,
		RepoOwner:  repoInfo.Owner,
		RepoName:   repoInfo.Name,
		HeadBranch: repoInfo.HeadBranch,
		HeadCommit: repoInfo.HeadCommit,
	}

	// Get GitHub Actions data if available
	githubEnv, err := fg_github.GetActionsEnv()
	if err != nil && !errors.Is(err, fg_github.ErrNotInActions) {
		return t, fmt.Errorf("failed to get GitHub Actions environment variables: %w", err)
	}
	t.GitHubEvent = githubEnv.EventName
	t.BaseBranch = githubEnv.BaseRef

	// Get GitHub Repo data if available
	err = fg_github.RepoInfo(githubClient, repoInfo.Owner, repoInfo.Name)
	if err != nil {
		return t, fmt.Errorf("failed to get GitHub repo info: %w", err)
	}
	t.GitHubEvent = githubEnv.EventName
	t.BaseBranch = githubEnv.BaseRef

	return t, nil
}
