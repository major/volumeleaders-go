package volumeleaders

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// BaseURL is the default VolumeLeaders web application origin.
const BaseURL = "https://www.volumeleaders.com"

// DefaultResponseBodyLimit is the maximum response body size read by default.
const DefaultResponseBodyLimit int64 = 10 << 20

const (
	defaultUserAgent   = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/147.0.0.0 Safari/537.36"
	defaultHTTPTimeout = 60 * time.Second
)

func defaultBrowserHeaders() http.Header {
	return http.Header{
		"Sec-Ch-Ua":          {`"Chromium";v="147", "Not A(Brand";v="24", "Google Chrome";v="147"`},
		"Sec-Ch-Ua-Mobile":   {"?0"},
		"Sec-Ch-Ua-Platform": {`"Windows"`},
		"Sec-Fetch-Dest":     {"empty"},
		"Sec-Fetch-Mode":     {"cors"},
		"Sec-Fetch-Site":     {"same-origin"},
		"Accept-Language":    {"en-US,en;q=0.9"},
	}
}

// ClientConfig holds configuration for a VolumeLeaders API client.
type ClientConfig struct {
	HTTPClient        *http.Client
	BaseURL           *url.URL
	ResponseBodyLimit int64
	UserAgent         string
	DefaultHeaders    http.Header
	OptionError       error
}

// Option configures a Client.
type Option interface {
	apply(*ClientConfig)
}

type optionFunc func(*ClientConfig)

func (f optionFunc) apply(cfg *ClientConfig) {
	f(cfg)
}

// WithHTTPClient sets the HTTP client used for requests. A nil client is
// ignored.
//
// NewClient copies the supplied client before using it. If client.CheckRedirect
// is nil, NewClient installs a same-origin redirect policy to keep session
// headers and XSRF tokens from crossing origins. If client.CheckRedirect is not
// nil, the caller's redirect policy is preserved.
func WithHTTPClient(client *http.Client) Option {
	return optionFunc(func(cfg *ClientConfig) {
		if client != nil {
			cfg.HTTPClient = client
		}
	})
}

// WithBaseURL overrides the default VolumeLeaders origin.
func WithBaseURL(rawURL string) Option {
	return optionFunc(func(cfg *ClientConfig) {
		u, err := url.Parse(rawURL)
		if err != nil {
			cfg.OptionError = errors.Join(cfg.OptionError, fmt.Errorf("invalid base URL %q: %w", rawURL, err))
			return
		}
		if u.Scheme == "" || u.Host == "" {
			cfg.OptionError = errors.Join(
				cfg.OptionError,
				fmt.Errorf("invalid base URL %q: absolute URL required", rawURL),
			)
			return
		}
		cfg.BaseURL = u
	})
}

// WithResponseBodyLimit sets the maximum response body size read by the client.
// Non-positive limits are ignored.
func WithResponseBodyLimit(limit int64) Option {
	return optionFunc(func(cfg *ClientConfig) {
		if limit > 0 {
			cfg.ResponseBodyLimit = limit
		}
	})
}

// WithUserAgent sets the User-Agent header sent with each request.
func WithUserAgent(userAgent string) Option {
	return optionFunc(func(cfg *ClientConfig) {
		if userAgent != "" {
			cfg.UserAgent = userAgent
		}
	})
}

// WithDefaultHeader sets a default request header. Empty names are ignored.
func WithDefaultHeader(name, value string) Option {
	return optionFunc(func(cfg *ClientConfig) {
		if name == "" {
			return
		}
		cfg.DefaultHeaders.Set(name, value)
	})
}

// WithDefaultHeaders sets default request headers. Header values are copied.
func WithDefaultHeaders(headers http.Header) Option {
	return optionFunc(func(cfg *ClientConfig) {
		copyHeaderInto(cfg.DefaultHeaders, headers)
	})
}

func defaultClientConfig() (ClientConfig, error) {
	baseURL, err := url.Parse(BaseURL)
	if err != nil {
		return ClientConfig{}, fmt.Errorf("parse default base URL: %w", err)
	}
	return ClientConfig{
		HTTPClient:        &http.Client{Timeout: defaultHTTPTimeout},
		BaseURL:           baseURL,
		ResponseBodyLimit: DefaultResponseBodyLimit,
		UserAgent:         defaultUserAgent,
		DefaultHeaders:    cloneHeader(defaultBrowserHeaders()),
	}, nil
}
