package git

import (
	"fmt"
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

	remote, err := repo.Remote("origin")
	if err != nil {
		return BasicRepoInfo{}, err
	}

	splitName := strings.Split(remote.Config().Name, "/")
	if len(splitName) != 2 {
		return BasicRepoInfo{}, fmt.Errorf(
			"expected remote name to be in the format 'owner/repo', got '%s'",
			remote.Config().Name,
		)
	}

	owner := splitName[0]
	name := splitName[1]

	return BasicRepoInfo{
		URL:        remote.Config().URLs[0],
		Owner:      owner,
		Name:       name,
		HeadBranch: head.Name().Short(),
		HeadCommit: head.Hash().String(),
	}, nil
}
