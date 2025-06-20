// Package git enables basic git operations.
package git

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/go-git/go-git/v5"
	"github.com/rs/zerolog"
)

// BasicRepoInfo contains basic information about a repository, all retrievable without signing git operations.
type BasicRepoInfo struct {
	URL        string `json:"repo_url"`
	Owner      string `json:"repo_owner"`
	Name       string `json:"repo_name"`
	HeadBranch string `json:"head_branch"`
	HeadCommit string `json:"head_commit"`
}

var (
	sshURLRe   = regexp.MustCompile(`^[^:]+@[^:]+:([^/]+)/([^.]+)(\.git)?$`)
	httpsURLRe = regexp.MustCompile(`^https?://[^/]+/([^/]+)/([^.]+)(\.git)?$`)
)

// parseGitURL extracts owner and repo name from a git URL
func parseGitURL(url string) (owner, name string, err error) {
	// Handle SSH format: git@github.com:owner/repo.git
	if sshURLRe.MatchString(url) {
		// Split by ':' to get the path part
		parts := strings.Split(url, ":")
		if len(parts) != 2 {
			return "", "", fmt.Errorf("invalid SSH git URL format: %s", url)
		}
		path := parts[1]
		// Remove .git suffix if present
		path = strings.TrimSuffix(path, ".git")
		// Split by '/' to get owner/repo
		pathParts := strings.Split(path, "/")
		if len(pathParts) != 2 {
			return "", "", fmt.Errorf("expected path format 'owner/repo', got '%s'", path)
		}
		return pathParts[0], pathParts[1], nil
	}

	// Handle HTTPS format: https://github.com/owner/repo.git
	if httpsURLRe.MatchString(url) {
		// Find the domain part and remove it
		trimmedURL := strings.TrimPrefix(strings.TrimPrefix(url, "https://"), "http://")
		domainEndIndex := strings.Index(trimmedURL, "/")
		if domainEndIndex == -1 {
			return "", "", fmt.Errorf("invalid HTTPS git URL format: %s", url)
		}
		path := trimmedURL[domainEndIndex+1:] // Skip domain + "/"

		// Remove .git suffix if present
		path = strings.TrimSuffix(path, ".git")

		// Split by '/' to get owner/repo
		pathParts := strings.Split(path, "/")
		if len(pathParts) < 2 {
			return "", "", fmt.Errorf("expected path format 'owner/repo', got '%s'", path)
		}
		return pathParts[0], pathParts[1], nil
	}

	return "", "", fmt.Errorf("unsupported git URL format: %s", url)
}

// ReadBasicRepoInfo returns basic information about the repository at the given path.
// This info is used to gather more detailed information from GitHub if possible.
func ReadBasicRepoInfo(l zerolog.Logger, repoPath string) (BasicRepoInfo, error) {
	l.Trace().Str("repo_path", repoPath).Msg("Reading basic repository information")

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return BasicRepoInfo{}, err
	}

	head, err := repo.Head()
	if err != nil {
		return BasicRepoInfo{}, err
	}

	remotes, err := repo.Remotes()
	if err != nil {
		return BasicRepoInfo{}, err
	}
	if len(remotes) == 0 {
		return BasicRepoInfo{}, fmt.Errorf("no remotes found")
	}
	remote := remotes[0]
	remoteURL := remote.Config().URLs[0]

	owner, name, err := parseGitURL(remoteURL)
	if err != nil {
		return BasicRepoInfo{}, fmt.Errorf("failed to parse git URL '%s': %w", remoteURL, err)
	}

	return BasicRepoInfo{
		URL:        remoteURL,
		Owner:      owner,
		Name:       name,
		HeadBranch: head.Name().Short(),
		HeadCommit: head.Hash().String(),
	}, nil
}
