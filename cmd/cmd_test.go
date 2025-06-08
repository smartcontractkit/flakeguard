package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		args                   []string
		expectedGotestsumFlags []string
		expectedGoTestFlags    []string
	}{
		{
			args:                   []string{"--format", "testname", "--", "./pkg/..."},
			expectedGotestsumFlags: []string{"--format", "testname"},
			expectedGoTestFlags:    []string{"./pkg/..."},
		},
	}

	for _, test := range tests {
		gotestsumFlags, goTestFlags := parseArgs(test.args)
		require.Equal(t, test.expectedGotestsumFlags, gotestsumFlags)
		require.Equal(t, test.expectedGoTestFlags, goTestFlags)
	}
}
