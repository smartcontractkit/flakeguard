package github

import (
	"errors"
	"fmt"
	"os"

	"github.com/caarlos0/env/v11"
)

var (
	ErrNotInActions = errors.New("not in GitHub Actions environment")
)

// Tracks GitHub Actions environment variables
// https://docs.github.com/en/actions/writing-workflows/choosing-what-your-workflow-does/store-information-in-variables#default-environment-variables
type githubActionsEnvVars struct {
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
	Job string `json:"GITHUB_JOB,omitempty"                 env:"GITHUB_JOB"`
	// JobName is a custom env var set by octometrics-action and describes the name of the job on the runner so we can match it with the API.
	// There is currently no native way to do this in GitHub Actions.
	// https://github.com/actions/toolkit/issues/550
	JobName           string `json:"GITHUB_JOB_NAME,omitempty"            env:"GITHUB_JOB_NAME"`
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

// ActionsEnvVars returns the GitHub Actions environment variables if running in a GitHub Actions environment
// Otherwise, it returns the error errNotInActions
func ActionsEnvVars() (githubActionsEnvVars, error) {
	if os.Getenv("GITHUB_ACTIONS") != "true" {
		return githubActionsEnvVars{}, ErrNotInActions
	}

	var envVars githubActionsEnvVars
	if err := env.Parse(&envVars); err != nil {
		return githubActionsEnvVars{}, fmt.Errorf("unable to parse GitHub Actions environment variables: %w", err)
	}

	return envVars, nil
}
