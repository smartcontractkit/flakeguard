package git

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseGitURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		url       string
		wantOwner string
		wantName  string
		wantErr   bool
	}{
		// Valid SSH URLs
		{
			name:      "SSH URL with .git suffix",
			url:       "git@github.com:owner/repo.git",
			wantOwner: "owner",
			wantName:  "repo",
			wantErr:   false,
		},
		{
			name:      "SSH URL without .git suffix",
			url:       "git@github.com:owner/repo",
			wantOwner: "owner",
			wantName:  "repo",
			wantErr:   false,
		},
		{
			name:      "SSH URL with GitLab",
			url:       "git@gitlab.com:owner/repo.git",
			wantOwner: "owner",
			wantName:  "repo",
			wantErr:   false,
		},
		{
			name:      "SSH URL with custom domain",
			url:       "git@custom-git.company.com:owner/repo.git",
			wantOwner: "owner",
			wantName:  "repo",
			wantErr:   false,
		},
		// Valid HTTPS URLs
		{
			name:      "HTTPS URL with .git suffix",
			url:       "https://github.com/owner/repo.git",
			wantOwner: "owner",
			wantName:  "repo",
			wantErr:   false,
		},
		{
			name:      "HTTPS URL without .git suffix",
			url:       "https://github.com/owner/repo",
			wantOwner: "owner",
			wantName:  "repo",
			wantErr:   false,
		},
		{
			name:      "HTTP URL with .git suffix",
			url:       "http://github.com/owner/repo.git",
			wantOwner: "owner",
			wantName:  "repo",
			wantErr:   false,
		},
		{
			name:      "HTTPS URL with GitLab",
			url:       "https://gitlab.com/owner/repo.git",
			wantOwner: "owner",
			wantName:  "repo",
			wantErr:   false,
		},
		{
			name:      "HTTPS URL with custom domain",
			url:       "https://custom-git.company.com/owner/repo.git",
			wantOwner: "owner",
			wantName:  "repo",
			wantErr:   false,
		},
		{
			name:      "HTTPS URL with nested path",
			url:       "https://github.com/owner/repo/path/to/something.git",
			wantOwner: "owner",
			wantName:  "repo",
			wantErr:   false,
		},
		// Invalid SSH URLs
		{
			name:    "SSH URL with invalid format - no colon",
			url:     "git@github.com/owner/repo.git",
			wantErr: true,
		},
		{
			name:    "SSH URL with too many colons",
			url:     "git@github.com:owner:repo.git",
			wantErr: true,
		},
		{
			name:    "SSH URL with invalid path - no slash",
			url:     "git@github.com:ownerrepo.git",
			wantErr: true,
		},
		{
			name:    "SSH URL with empty owner",
			url:     "git@github.com:/repo.git",
			wantErr: true,
		},
		{
			name:    "SSH URL with empty repo",
			url:     "git@github.com:owner/.git",
			wantErr: true,
		},
		// Invalid HTTPS URLs
		{
			name:    "HTTPS URL with no path",
			url:     "https://github.com/",
			wantErr: true,
		},
		{
			name:    "HTTPS URL with only owner",
			url:     "https://github.com/owner",
			wantErr: true,
		},
		{
			name:    "HTTPS URL with empty owner",
			url:     "https://github.com//repo.git",
			wantErr: true,
		},
		{
			name:    "HTTPS URL with empty repo",
			url:     "https://github.com/owner/.git",
			wantErr: true,
		},
		{
			name:    "HTTPS URL with invalid format - no domain",
			url:     "https:///owner/repo.git",
			wantErr: true,
		},
		// Completely invalid URLs
		{
			name:    "empty URL",
			url:     "",
			wantErr: true,
		},
		{
			name:    "invalid protocol",
			url:     "ftp://github.com/owner/repo.git",
			wantErr: true,
		},
		{
			name:    "local file path",
			url:     "/path/to/local/repo",
			wantErr: true,
		},
		{
			name:    "malformed URL",
			url:     "not-a-url",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			owner, name, err := parseGitURL(tt.url)

			if tt.wantErr {
				require.Error(t, err, "parseGitURL(%q) should return an error", tt.url)
				require.Empty(t, owner, "owner should be empty when error occurs")
				require.Empty(t, name, "name should be empty when error occurs")
			} else {
				require.NoError(t, err, "parseGitURL(%q) should not return an error", tt.url)
				require.Equal(t, tt.wantOwner, owner, "owner mismatch for URL %q", tt.url)
				require.Equal(t, tt.wantName, name, "name mismatch for URL %q", tt.url)
			}
		})
	}
}

func FuzzParseGitURL(f *testing.F) {
	f.Add("git@github.com:owner/repo.git")
	f.Add("https://github.com/owner/repo.git")
	f.Add("invalid-url")

	f.Fuzz(func(t *testing.T, url string) {
		owner, name, err := parseGitURL(url)
		// The function should never panic
		// If no error, owner and name should be non-empty
		if err == nil {
			if owner == "" || name == "" {
				t.Errorf("parseGitURL(%q) returned empty owner/name without error: owner=%q, name=%q", url, owner, name)
			}
		}
		// If error, owner and name should be empty
		if err != nil && (owner != "" || name != "") {
			t.Errorf(
				"parseGitURL(%q) returned non-empty owner/name with error: owner=%q, name=%q, err=%v",
				url,
				owner,
				name,
				err,
			)
		}
	})
}
