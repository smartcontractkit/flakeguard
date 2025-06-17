// Package github provides information and ways to interact with GitHub.
package github

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"

	"github.com/caarlos0/env/v11"
)

var (
	// ErrNotInActions is returned when the code is not running in a GitHub Actions environment.
	ErrNotInActions = errors.New("not in GitHub Actions environment")
)

// ActionsEnv tracks GitHub Actions environment variables
// https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/store-information-in-variables#default-environment-variables
type ActionsEnv struct {
	Action           string `json:"GITHUB_ACTION,omitempty"              env:"GITHUB_ACTION"`
	ActionPath       string `json:"GITHUB_ACTION_PATH,omitempty"         env:"GITHUB_ACTION_PATH"`
	ActionRepository string `json:"GITHUB_ACTION_REPOSITORY,omitempty"   env:"GITHUB_ACTION_REPOSITORY"`
	Actor            string `json:"GITHUB_ACTOR,omitempty"               env:"GITHUB_ACTOR"`
	ActorID          string `json:"GITHUB_ACTOR_ID,omitempty"            env:"GITHUB_ACTOR_ID"`
	APIURL           string `json:"GITHUB_API_URL,omitempty"             env:"GITHUB_API_URL"`
	BaseRef          string `json:"GITHUB_BASE_REF,omitempty"            env:"GITHUB_BASE_REF"`
	Env              string `json:"GITHUB_ENV,omitempty"                 env:"GITHUB_ENV"`
	EventName        string `json:"GITHUB_EVENT_NAME,omitempty"          env:"GITHUB_EVENT_NAME"`
	EventPath        string `json:"GITHUB_EVENT_PATH,omitempty"          env:"GITHUB_EVENT_PATH"`
	GraphQLURL       string `json:"GITHUB_GRAPHQL_URL,omitempty"         env:"GITHUB_GRAPHQL_URL"`
	HeadRef          string `json:"GITHUB_HEAD_REF,omitempty"            env:"GITHUB_HEAD_REF"`
	// Job is the github-context job_id. This in no way matches to the numerical Job ID returned by the API, nor the name of the job.
	Job               string `json:"GITHUB_JOB,omitempty"                 env:"GITHUB_JOB"`
	Output            string `json:"GITHUB_OUTPUT,omitempty"              env:"GITHUB_OUTPUT"`
	Path              string `json:"GITHUB_PATH,omitempty"                env:"GITHUB_PATH"`
	Ref               string `json:"GITHUB_REF,omitempty"                 env:"GITHUB_REF"`
	RefName           string `json:"GITHUB_REF_NAME,omitempty"            env:"GITHUB_REF_NAME"`
	Repository        string `json:"GITHUB_REPOSITORY,omitempty"          env:"GITHUB_REPOSITORY"`
	RepositoryID      string `json:"GITHUB_REPOSITORY_ID,omitempty"       env:"GITHUB_REPOSITORY_ID"`
	RepositoryOwner   string `json:"GITHUB_REPOSITORY_OWNER,omitempty"    env:"GITHUB_REPOSITORY_OWNER"`
	RepositoryOwnerID string `json:"GITHUB_REPOSITORY_OWNER_ID,omitempty" env:"GITHUB_REPOSITORY_OWNER_ID"`
	RetentionDays     string `json:"GITHUB_RETENTION_DAYS,omitempty"      env:"GITHUB_RETENTION_DAYS"`
	RunAttempt        string `json:"GITHUB_RUN_ATTEMPT,omitempty"         env:"GITHUB_RUN_ATTEMPT"`
	// RunID refers to the workflow run ID
	RunID       int64  `json:"GITHUB_RUN_ID,omitempty"              env:"GITHUB_RUN_ID"`
	RunNumber   int    `json:"GITHUB_RUN_NUMBER,omitempty"          env:"GITHUB_RUN_NUMBER"`
	ServerURL   string `json:"GITHUB_SERVER_URL,omitempty"          env:"GITHUB_SERVER_URL"`
	SHA         string `json:"GITHUB_SHA,omitempty"                 env:"GITHUB_SHA"`
	StepSummary string `json:"GITHUB_STEP_SUMMARY,omitempty"        env:"GITHUB_STEP_SUMMARY"`
	// Token isn't guaranteed to be set as an env var, but it's a standard process, especially for the octometrics-action.
	Token             string `json:"GITHUB_TOKEN,omitempty"               env:"GITHUB_TOKEN"`
	TriggeringActor   string `json:"GITHUB_TRIGGERING_ACTOR,omitempty"    env:"GITHUB_TRIGGERING_ACTOR"`
	Workflow          string `json:"GITHUB_WORKFLOW,omitempty"            env:"GITHUB_WORKFLOW"`
	WorkflowRef       string `json:"GITHUB_WORKFLOW_REF,omitempty"        env:"GITHUB_WORKFLOW_REF"`
	WorkflowSHA       string `json:"GITHUB_WORKFLOW_SHA,omitempty"        env:"GITHUB_WORKFLOW_SHA"`
	Workspace         string `json:"GITHUB_WORKSPACE,omitempty"           env:"GITHUB_WORKSPACE"`
	RunnerArch        string `json:"RUNNER_ARCH,omitempty"                env:"RUNNER_ARCH"`
	RunnerDebug       string `json:"RUNNER_DEBUG,omitempty"               env:"RUNNER_DEBUG"`
	RunnerEnvironment string `json:"RUNNER_ENVIRONMENT,omitempty"         env:"RUNNER_ENVIRONMENT"`
	RunnerName        string `json:"RUNNER_NAME,omitempty"                env:"RUNNER_NAME"`
	RunnerOS          string `json:"RUNNER_OS,omitempty"                  env:"RUNNER_OS"`
	RunnerTemp        string `json:"RUNNER_TEMP,omitempty"                env:"RUNNER_TEMP"`
	RunnerToolCache   string `json:"RUNNER_TOOL_CACHE,omitempty"          env:"RUNNER_TOOL_CACHE"`
}

// GetActionsEnv returns the GitHub Actions environment variables if running in a GitHub Actions environment
// Otherwise, it returns ErrNotInActions
func GetActionsEnv() (ActionsEnv, error) {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		return ActionsEnv{}, ErrNotInActions
	}

	var envVars ActionsEnv
	if err := env.Parse(&envVars); err != nil {
		return ActionsEnv{}, fmt.Errorf("unable to parse GitHub Actions environment variables: %w", err)
	}

	return envVars, nil
}

// RepoInfo gets the repository information from the GitHub API.
func RepoInfo(client *Client, repoOwner, repoName string) error {
	repo, resp, err := client.Rest.Repositories.Get(context.Background(), repoOwner, repoName)
	if err != nil {
		return fmt.Errorf("failed to get GitHub repo info: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to get GitHub repo info: %s", resp.Status)
	}
	repo.GetDefaultBranch()

	return nil
}
