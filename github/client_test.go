package github

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/go-github/v72/github"
	gh_graphql "github.com/shurcooL/githubv4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/smartcontractkit/flakeguard/internal/testhelpers"
)

func TestNewClient(t *testing.T) {
	// Uses t.Setenv, so we can't run it in parallel.
	logger := testhelpers.Logger(t)

	tests := []struct {
		name        string
		token       string
		envToken    string
		expectError bool
	}{
		{
			name: "no token",
		},
		{
			name:     "token overrides env",
			token:    "arg-token",
			envToken: "env-token",
		},
		{
			name:     "only env token",
			envToken: "env-token",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Uses t.Setenv, so we can't run it in parallel.
			t.Setenv(TokenEnvVar, tt.envToken)

			client, err := NewClient(logger, tt.token, nil)

			if tt.expectError {
				require.Error(t, err, "expected error")
				return
			}

			require.NoError(t, err, "expected no error")
			require.NotNil(t, client)
			require.NotNil(t, client.Rest)
			require.NotNil(t, client.GraphQL)
			require.IsType(t, &github.Client{}, client.Rest)
			require.IsType(t, &gh_graphql.Client{}, client.GraphQL)

			switch {
			case tt.token != "":
				assert.Equal(t, tt.token, client.token, "expected arg token to be set")
			case tt.envToken != "":
				assert.Equal(t, tt.envToken, client.token, "expected env token to be set")
			default:
				assert.Empty(t, client.token, "expected empty token")
			}
		})
	}
}

func TestNewClientWithCustomTransport(t *testing.T) {
	t.Parallel()
	logger := testhelpers.Logger(t)

	// Create a mock transport
	mockTransport := &mockRoundTripper{
		responses: make(map[string]*http.Response),
	}

	client, err := NewClient(logger, "test-token", mockTransport)

	require.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.Rest)
	assert.NotNil(t, client.GraphQL)
}

func TestClientRoundTripper(t *testing.T) {
	t.Parallel()
	logger := testhelpers.Logger(t)

	// Test with nil next transport (should use default)
	transport := clientRoundTripper("TEST", logger, nil)
	assert.NotNil(t, transport)
	assert.IsType(t, &loggingTransport{}, transport)

	// Test with custom next transport
	mockTransport := &mockRoundTripper{
		responses: make(map[string]*http.Response),
	}
	transport = clientRoundTripper("TEST", logger, mockTransport)
	assert.NotNil(t, transport)
	assert.IsType(t, &loggingTransport{}, transport)

	lt := transport.(*loggingTransport)
	assert.Equal(t, "TEST", lt.clientType)
	assert.Equal(t, mockTransport, lt.transport)
}

func TestLoggingTransportRoundTrip(t *testing.T) {
	t.Parallel()
	logger := testhelpers.Logger(t)

	tests := []struct {
		name          string
		statusCode    int
		headers       map[string]string
		expectError   bool
		isRateLimited bool
		isMockRequest bool
	}{
		{
			name:       "successful request",
			statusCode: http.StatusOK,
			headers: map[string]string{
				"X-RateLimit-Remaining": "4999",
				"X-RateLimit-Limit":     "5000",
				"X-RateLimit-Used":      "1",
				"X-RateLimit-Reset":     fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix()),
			},
		},
		{
			name:       "rate limit warning",
			statusCode: http.StatusOK,
			headers: map[string]string{
				"X-RateLimit-Remaining": "50",
				"X-RateLimit-Limit":     "5000",
				"X-RateLimit-Used":      "4950",
				"X-RateLimit-Reset":     fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix()),
			},
		},
		{
			name:       "mock request (no warning)",
			statusCode: http.StatusOK,
			headers: map[string]string{
				"X-RateLimit-Remaining": "50",
				"X-RateLimit-Limit":     "5000",
				"X-RateLimit-Used":      "4950",
				"X-RateLimit-Reset":     fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix()),
			},
			isMockRequest: true,
		},
		{
			name:          "rate limited",
			statusCode:    http.StatusTooManyRequests,
			isRateLimited: true,
		},
		{
			name:       "missing headers (defaults to 0)",
			statusCode: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mockTransport := &mockRoundTripper{
				responses: make(map[string]*http.Response),
			}

			transport := &loggingTransport{
				transport:  mockTransport,
				logger:     logger,
				clientType: "TEST",
			}

			// Create test server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				for key, value := range tt.headers {
					w.Header().Set(key, value)
				}
				w.WriteHeader(tt.statusCode)
				_, err := w.Write([]byte(`{"message": "test"}`))
				if err != nil {
					t.Errorf("failed to write response: %v", err)
					return
				}
			}))
			defer server.Close()

			// Create request
			req, err := http.NewRequest("GET", server.URL, nil)
			require.NoError(t, err)

			if tt.isMockRequest {
				req.Header.Set("Authorization", "Bearer "+MockToken)
			} else {
				req.Header.Set("Authorization", "Bearer real-token")
			}
			req.Header.Set("User-Agent", "test-agent")

			// Mock the transport to return our test server response
			mockTransport.responses[req.URL.String()] = &http.Response{
				StatusCode: tt.statusCode,
				Header:     make(http.Header),
				Request:    req,
			}

			// Copy headers to response
			for key, value := range tt.headers {
				mockTransport.responses[req.URL.String()].Header.Set(key, value)
			}

			// Execute request
			resp, err := transport.RoundTrip(req)

			if tt.expectError {
				assert.Error(t, err)
				return
			}

			require.NoError(t, err)
			assert.NotNil(t, resp)
			assert.Equal(t, tt.statusCode, resp.StatusCode)
		})
	}
}

func TestLoggingTransportInvalidHeaders(t *testing.T) {
	t.Parallel()
	logger := testhelpers.Logger(t)

	mockTransport := &mockRoundTripper{
		responses: make(map[string]*http.Response),
	}

	transport := &loggingTransport{
		transport:  mockTransport,
		logger:     logger,
		clientType: "TEST",
	}

	// Test with invalid header values
	invalidHeaders := []struct {
		name    string
		headers map[string]string
	}{
		{
			name: "invalid remaining",
			headers: map[string]string{
				"X-RateLimit-Remaining": "invalid",
				"X-RateLimit-Limit":     "5000",
				"X-RateLimit-Used":      "1",
				"X-RateLimit-Reset":     "1234567890",
			},
		},
		{
			name: "invalid limit",
			headers: map[string]string{
				"X-RateLimit-Remaining": "4999",
				"X-RateLimit-Limit":     "invalid",
				"X-RateLimit-Used":      "1",
				"X-RateLimit-Reset":     "1234567890",
			},
		},
		{
			name: "invalid used",
			headers: map[string]string{
				"X-RateLimit-Remaining": "4999",
				"X-RateLimit-Limit":     "5000",
				"X-RateLimit-Used":      "invalid",
				"X-RateLimit-Reset":     "1234567890",
			},
		},
		{
			name: "invalid reset",
			headers: map[string]string{
				"X-RateLimit-Remaining": "4999",
				"X-RateLimit-Limit":     "5000",
				"X-RateLimit-Used":      "1",
				"X-RateLimit-Reset":     "invalid",
			},
		},
	}

	for _, tt := range invalidHeaders {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			req, err := http.NewRequest("GET", "https://api.github.com/test", nil)
			require.NoError(t, err)

			// Mock response with invalid headers
			mockResp := &http.Response{
				StatusCode: http.StatusOK,
				Header:     make(http.Header),
				Request:    req,
			}

			for key, value := range tt.headers {
				mockResp.Header.Set(key, value)
			}

			mockTransport.responses[req.URL.String()] = mockResp

			// This should return an error due to invalid header conversion
			_, err = transport.RoundTrip(req)
			require.Error(t, err)
			assert.Contains(t, err.Error(), "failed to convert")
		})
	}
}

// mockRoundTripper is a mock implementation of http.RoundTripper for testing
type mockRoundTripper struct {
	responses map[string]*http.Response
	err       error
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	if m.err != nil {
		return nil, m.err
	}

	if resp, ok := m.responses[req.URL.String()]; ok {
		return resp, nil
	}

	// Default response
	return &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Request:    req,
	}, nil
}

func TestConstants(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "GITHUB_TOKEN", TokenEnvVar)
	assert.Equal(t, "mock-github-token", MockToken)
}

func TestMockTokenDetection(t *testing.T) {
	t.Parallel()
	logger := testhelpers.Logger(t)

	mockTransport := &mockRoundTripper{
		responses: make(map[string]*http.Response),
	}

	transport := &loggingTransport{
		transport:  mockTransport,
		logger:     logger,
		clientType: "TEST",
	}

	// Test with mock token - should not trigger rate limit warning
	req, err := http.NewRequest("GET", "https://api.github.com/test", nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+MockToken)

	mockResp := &http.Response{
		StatusCode: http.StatusOK,
		Header:     make(http.Header),
		Request:    req,
	}
	mockResp.Header.Set("X-RateLimit-Remaining", "10") // This would normally trigger warning
	mockResp.Header.Set("X-RateLimit-Limit", "5000")
	mockResp.Header.Set("X-RateLimit-Used", "4990")
	mockResp.Header.Set("X-RateLimit-Reset", fmt.Sprintf("%d", time.Now().Add(time.Hour).Unix()))

	mockTransport.responses[req.URL.String()] = mockResp

	resp, err := transport.RoundTrip(req)

	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}
