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

	// Get the default branch from the remote HEAD reference
	remote := remotes[0]
	refs, err := remote.List(&git.ListOptions{})
	if err != nil {
		l.Warn().Err(err).Msg("Failed to list remote references")
	} else {
		// Look for HEAD reference which points to the default branch
		for _, ref := range refs {
			if ref.Name().String() == "HEAD" && ref.Target() != "" {
				info.DefaultBranch = ref.Target().Short()
				break
			}
		}
	}

	info.URL = remote.Config().URLs[0]
	info.Name = remote.Config().Name

	return info, nil
}
