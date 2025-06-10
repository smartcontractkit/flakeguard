package cmd

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseArgs(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name                   string
		args                   []string
		expectedGotestsumFlags []string
		expectedGoTestFlags    []string
	}{
		{
			name:                   "basic gotestsum and go test flags",
			args:                   []string{"--format", "testname", "--", "./pkg/..."},
			expectedGotestsumFlags: []string{"--format", "testname"},
			expectedGoTestFlags:    []string{"./pkg/..."},
		},
		{
			name:                   "no separator",
			args:                   []string{"--format", "testname", "./pkg/..."},
			expectedGotestsumFlags: []string{"--format", "testname", "./pkg/..."},
			expectedGoTestFlags:    []string{},
		},
		{
			name:                   "empty args",
			args:                   []string{},
			expectedGotestsumFlags: []string{},
			expectedGoTestFlags:    []string{},
		},
		{
			name:                   "only separator",
			args:                   []string{"--"},
			expectedGotestsumFlags: []string{},
			expectedGoTestFlags:    []string{},
		},
		{
			name:                   "separator at beginning",
			args:                   []string{"--", "./pkg/..."},
			expectedGotestsumFlags: []string{},
			expectedGoTestFlags:    []string{"./pkg/..."},
		},
		{
			name:                   "separator at end",
			args:                   []string{"--format", "testname", "--"},
			expectedGotestsumFlags: []string{"--format", "testname"},
			expectedGoTestFlags:    []string{},
		},
		{
			name:                   "multiple separators - only first one counts",
			args:                   []string{"--format", "testname", "--", "-v", "--", "./pkg/..."},
			expectedGotestsumFlags: []string{"--format", "testname"},
			expectedGoTestFlags:    []string{"-v", "--", "./pkg/..."},
		},
		{
			name:                   "complex gotestsum flags",
			args:                   []string{"--format", "dots", "--jsonfile", "test.json", "--", "./..."},
			expectedGotestsumFlags: []string{"--format", "dots", "--jsonfile", "test.json"},
			expectedGoTestFlags:    []string{"./..."},
		},
		{
			name: "complex go test flags",
			args: []string{
				"--format",
				"testname",
				"--",
				"-v",
				"-run",
				"TestMyFunction",
				"-count",
				"3",
				"./pkg/...",
			},
			expectedGotestsumFlags: []string{"--format", "testname"},
			expectedGoTestFlags:    []string{"-v", "-run", "TestMyFunction", "-count", "3", "./pkg/..."},
		},
		{
			name:                   "single gotestsum flag",
			args:                   []string{"--version", "--", "./..."},
			expectedGotestsumFlags: []string{"--version"},
			expectedGoTestFlags:    []string{"./..."},
		},
		{
			name:                   "single go test flag",
			args:                   []string{"--", "-v"},
			expectedGotestsumFlags: []string{},
			expectedGoTestFlags:    []string{"-v"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			gotestsumFlags, goTestFlags := parseArgs(test.args)
			require.Equal(t, test.expectedGotestsumFlags, gotestsumFlags)
			require.Equal(t, test.expectedGoTestFlags, goTestFlags)
		})
	}
}
