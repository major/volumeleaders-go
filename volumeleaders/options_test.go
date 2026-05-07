package volumeleaders

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDefaultClientConfig(t *testing.T) {
	cfg, err := defaultClientConfig()
	require.NoError(t, err, "defaultClientConfig()")

	assert.Equal(t, BaseURL, cfg.BaseURL.String(), "defaultClientConfig() BaseURL")
	assert.Equal(t, DefaultResponseBodyLimit, cfg.ResponseBodyLimit, "defaultClientConfig() ResponseBodyLimit")
	assert.Equal(t, defaultUserAgent, cfg.UserAgent, "defaultClientConfig() UserAgent")
	assert.Equal(t, defaultHTTPTimeout, cfg.HTTPClient.Timeout, "defaultClientConfig() HTTPClient.Timeout")
	assert.Equal(
		t,
		defaultBrowserHeaders().Get("Sec-Fetch-Site"),
		cfg.DefaultHeaders.Get("Sec-Fetch-Site"),
		"defaultClientConfig() DefaultHeaders",
	)

	cfg.DefaultHeaders.Set("Sec-Fetch-Site", "modified")
	assert.Equal(
		t,
		"same-origin",
		defaultBrowserHeaders().Get("Sec-Fetch-Site"),
		"defaultClientConfig() copied default headers",
	)
}

func TestClientOptions(t *testing.T) {
	client := &http.Client{Timeout: 5 * time.Second}
	headers := http.Header{
		"X-Custom": {"first", "second"},
	}

	cfg, err := defaultClientConfig()
	require.NoError(t, err, "defaultClientConfig()")
	for _, opt := range []Option{
		WithHTTPClient(client),
		WithBaseURL("https://example.test/base"),
		WithResponseBodyLimit(512),
		WithUserAgent("volumeleaders-test"),
		WithDefaultHeader("X-One", "1"),
		WithDefaultHeaders(headers),
	} {
		opt.apply(&cfg)
	}
	headers.Set("X-Custom", "changed")

	assert.Same(t, client, cfg.HTTPClient, "options HTTPClient")
	assert.Equal(t, "https://example.test/base", cfg.BaseURL.String(), "options BaseURL")
	assert.Equal(t, int64(512), cfg.ResponseBodyLimit, "options ResponseBodyLimit")
	assert.Equal(t, "volumeleaders-test", cfg.UserAgent, "options UserAgent")
	assert.Equal(t, "1", cfg.DefaultHeaders.Get("X-One"), "options default header")
	assert.Equal(
		t,
		[]string{"first", "second"},
		cfg.DefaultHeaders.Values("X-Custom"),
		"options copied default headers",
	)
}

func TestClientOptionsIgnoreEmptyValues(t *testing.T) {
	cfg, err := defaultClientConfig()
	require.NoError(t, err, "defaultClientConfig()")
	wantClient := cfg.HTTPClient
	wantBaseURL := cfg.BaseURL.String()
	wantLimit := cfg.ResponseBodyLimit
	wantUserAgent := cfg.UserAgent

	for _, opt := range []Option{
		WithHTTPClient(nil),
		WithResponseBodyLimit(0),
		WithUserAgent(""),
		WithDefaultHeader("", "ignored"),
	} {
		opt.apply(&cfg)
	}

	assert.Same(t, wantClient, cfg.HTTPClient, "empty options HTTPClient")
	assert.Equal(t, wantBaseURL, cfg.BaseURL.String(), "empty options BaseURL")
	assert.Equal(t, wantLimit, cfg.ResponseBodyLimit, "empty options ResponseBodyLimit")
	assert.Equal(t, wantUserAgent, cfg.UserAgent, "empty options UserAgent")
	assert.Empty(t, cfg.DefaultHeaders.Get(""), "empty options blank header")
}

func TestWithBaseURLReportsInvalidValues(t *testing.T) {
	tests := []struct {
		name   string
		rawURL string
	}{
		{name: "relative", rawURL: "/relative"},
		{name: "parse error", rawURL: "http://[::1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := defaultClientConfig()
			require.NoError(t, err, "defaultClientConfig()")

			WithBaseURL(tt.rawURL).apply(&cfg)

			require.Error(t, cfg.OptionError, "WithBaseURL(%q) error", tt.rawURL)
			assert.Contains(t, cfg.OptionError.Error(), "invalid base URL", "WithBaseURL(%q) error", tt.rawURL)
		})
	}
}
