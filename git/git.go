// Package git provides information and ways to interact with git repositories.
package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/rs/zerolog"
)

type RepoInfo struct {
	Owner         string
	Name          string
	URL           string
	CurrentBranch string
	CurrentCommit string
	DefaultBranch string
}

func GetRepoInfo(l zerolog.Logger, path string) (*RepoInfo, error) {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return nil, err
	}

	head, err := repo.Head()
	if err != nil {
		return nil, err
	}

	info := &RepoInfo{
		CurrentBranch: head.Name().Short(),
		CurrentCommit: head.Hash().String(),
	}

	remotes, err := repo.Remotes()
	if err != nil {
		return nil, fmt.Errorf("failed to get remotes: %w", err)
	}

	if len(remotes) == 0 {
		l.Warn().Str("path", path).Msg("No remotes found for repo")
		return info, nil
	}
	remote := remotes[0]

	info.URL = remote.Config().URLs[0]
	info.Name = remote.Config().Name

	return info, nil
}
