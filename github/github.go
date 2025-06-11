// Package github provides information and ways to interact with GitHub.
package github

import (
	"os"
)

// IsActions returns true if running in GitHub Actions environment
func IsActions() bool {
	return os.Getenv("GITHUB_ACTIONS") == "true"
}

func Event() string {
	return os.Getenv("GITHUB_EVENT_NAME")
}

// Branches returns the head and base branch names if in PR context
func Branches() (headBranch, baseBranch string) {
	if !IsActions() {
		return "", ""
	}

	return os.Getenv("GITHUB_HEAD_REF"), os.Getenv("GITHUB_BASE_REF")
}
