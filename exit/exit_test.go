package exit

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetCode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		underlyingError error
		err             error
		want            int
	}{
		{name: "nil success", underlyingError: nil, err: nil, want: CodeSuccess},
		{name: "success", underlyingError: nil, err: New(CodeSuccess, nil), want: CodeSuccess},
		{
			name:            "go failing test",
			underlyingError: errors.New("go failing test"),
			err:             New(CodeGoFailingTest, errors.New("go failing test")),
			want:            CodeGoFailingTest,
		},
		{
			name:            "go build error",
			underlyingError: errors.New("go build error"),
			err:             New(CodeGoBuildError, errors.New("go build error")),
			want:            CodeGoBuildError,
		},
		{
			name:            "flakeguard error",
			underlyingError: errors.New("flakeguard error"),
			err:             New(CodeFlakeguardError, errors.New("flakeguard error")),
			want:            CodeFlakeguardError,
		},
		{
			name:            "non-Error type returns CodeFlakeguardError",
			underlyingError: errors.New("some other error"),
			err:             errors.New("some other error"),
			want:            CodeFlakeguardError,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := GetCode(test.err)
			require.Equal(t, test.want, got, "GetCode(%v) = %d, want %d", test.err, got, test.want)

			// Only check unwrapping for *Error types
			if testErr, ok := test.err.(*Error); ok {
				require.Equal(t, test.want, testErr.ExitCode())
				require.Equal(t, test.underlyingError, errors.Unwrap(test.err))
				if test.underlyingError != nil {
					require.Equal(t, test.underlyingError.Error(), test.err.Error())
				} else {
					require.Equal(t, test.err.Error(), fmt.Sprintf("exit code %d", test.want))
				}
			} else if test.err != nil {
				// For non-*Error types, the error should be itself
				require.Equal(t, test.err, test.underlyingError)
			}
		})
	}
}
