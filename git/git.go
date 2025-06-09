package git

import (
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/rs/zerolog"
)

type RepoInfo struct {
	Owner  string
	Name   string
	URL    string
	Branch string
	Commit string
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
		Branch: head.Name().Short(),
		Commit: head.Hash().String(),
	}

	remotes, err := repo.Remotes()
	if err != nil {
		return nil, fmt.Errorf("failed to get remotes: %w", err)
	}

	if len(remotes) == 0 {
		l.Warn().Str("path", path).Msg("No remotes found for repo")
		return info, nil
	}

	info.URL = remotes[0].Config().URLs[0]
	info.Name = remotes[0].Config().Name

	return info, nil
}
