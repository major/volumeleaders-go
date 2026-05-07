package volumeleaders

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

const maxRedirects = 10

const (
	alertConfigsRedirectPath     = "/alertconfigs"
	jsonAccept                   = "application/json, text/javascript, */*; q=0.01"
	navigationAccept             = "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8"
	watchListConfigsRedirectPath = "/watchlistconfigs"
	xmlHTTPRequest               = "XMLHttpRequest"
)

// Client sends authenticated browser-session-backed requests to VolumeLeaders.
type Client struct {
	httpClient        *http.Client
	baseURL           *url.URL
	responseBodyLimit int64
	userAgent         string
	defaultHeaders    http.Header
	session           Session
}

// NewClient creates a Client from caller-supplied browser session material.
func NewClient(session Session, opts ...Option) (*Client, error) {
	cfg, err := defaultClientConfig()
	if err != nil {
		return nil, err
	}
	for _, opt := range opts {
		if opt != nil {
			opt.apply(&cfg)
		}
	}
	if cfg.OptionError != nil {
		return nil, cfg.OptionError
	}

	baseURL := *cfg.BaseURL
	httpClient := cloneHTTPClient(cfg.HTTPClient)
	return &Client{
		httpClient:        httpClient,
		baseURL:           &baseURL,
		responseBodyLimit: cfg.ResponseBodyLimit,
		userAgent:         cfg.UserAgent,
		defaultHeaders:    cloneHeader(cfg.DefaultHeaders),
		session:           session.clone(),
	}, nil
}

func cloneHTTPClient(client *http.Client) *http.Client {
	if client == nil {
		client = http.DefaultClient
	}
	callerRedirectPolicy := client.CheckRedirect
	cloned := *client
	cloned.CheckRedirect = func(req *http.Request, via []*http.Request) error {
		if err := sameOriginRedirectPolicy(req, via); err != nil {
			return err
		}
		if callerRedirectPolicy != nil {
			return callerRedirectPolicy(req, via)
		}
		return nil
	}
	return &cloned
}

func sameOriginRedirectPolicy(req *http.Request, via []*http.Request) error {
	if len(via) >= maxRedirects {
		return fmt.Errorf("stopped after %d redirects", maxRedirects)
	}
	if len(via) == 0 {
		return nil
	}
	first := via[0].URL
	if req.URL.Scheme != first.Scheme || req.URL.Host != first.Host {
		return ErrCrossOriginRedirect
	}
	return nil
}

func (c *Client) postForm(ctx context.Context, path string, form string, result any) error {
	return c.post(ctx, path, strings.NewReader(form), "application/x-www-form-urlencoded; charset=UTF-8", result)
}

func (c *Client) postJSON(ctx context.Context, path string, payload any, result any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("marshal JSON request: %w", err)
	}
	return c.post(ctx, path, bytes.NewReader(body), "application/json; charset=UTF-8", result)
}

func (c *Client) postMultipartForm(ctx context.Context, path string, fields url.Values, result any) error {
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for key, values := range fields {
		for _, value := range values {
			if err := writer.WriteField(key, value); err != nil {
				return fmt.Errorf("write multipart field %q: %w", key, err)
			}
		}
	}
	if err := writer.Close(); err != nil {
		return fmt.Errorf("close multipart writer: %w", err)
	}
	return c.postWithOptions(ctx, path, &body, writer.FormDataContentType(), result, postOptions{
		Accept:               navigationAccept,
		AllowRedirectSuccess: true,
		ExpectedRedirectPath: multipartSuccessRedirectPath(path),
	})
}

func (c *Client) post(ctx context.Context, path string, requestBody io.Reader, contentType string, result any) error {
	return c.postWithOptions(ctx, path, requestBody, contentType, result, postOptions{
		Accept:           jsonAccept,
		RequestedWith:    xmlHTTPRequest,
		IncludeXSRFToken: true,
	})
}

type postOptions struct {
	Accept               string
	RequestedWith        string
	IncludeXSRFToken     bool
	AllowRedirectSuccess bool
	ExpectedRedirectPath string
}

func (c *Client) postWithOptions(
	ctx context.Context,
	path string,
	requestBody io.Reader,
	contentType string,
	result any,
	opts postOptions,
) error {
	requestURL := c.resolve(path)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, requestURL, requestBody)
	if err != nil {
		return fmt.Errorf("create POST request: %w", err)
	}
	c.setHeaders(req)
	req.Header.Set("Content-Type", contentType)
	if opts.Accept != "" {
		req.Header.Set("Accept", opts.Accept)
	}
	if opts.RequestedWith != "" {
		req.Header.Set("X-Requested-With", opts.RequestedWith)
	}
	if opts.IncludeXSRFToken && c.session.XSRFToken != "" {
		req.Header.Set(xsrfHeaderName, c.session.XSRFToken)
	}
	for _, cookie := range c.session.Cookies {
		req.AddCookie(cookie)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("post request: %w", err)
	}
	defer resp.Body.Close()
	if redirectedToLogin(resp) {
		return sessionExpiredError{redirectPath: safeResponsePath(resp)}
	}
	if handled, redirectErr := redirectSuccess(req, resp, result, opts); handled {
		return redirectErr
	}

	body, err := c.readBody(resp.Body)
	if err != nil {
		return fmt.Errorf("read response body: %w", err)
	}
	if looksLikeLoginPage(resp, body) {
		return sessionExpiredError{redirectPath: safeResponsePath(resp)}
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return statusError(resp, body)
	}
	if result == nil {
		return nil
	}
	decodeErr := json.Unmarshal(body, result)
	if decodeErr != nil {
		return contentError(resp, body, decodeErr)
	}
	return nil
}

func redirectSuccess(req *http.Request, resp *http.Response, result any, opts postOptions) (bool, error) {
	if !opts.AllowRedirectSuccess || result != nil {
		return false, nil
	}
	if resp.StatusCode >= 300 && resp.StatusCode < 400 {
		if redirectsToLoginLocation(resp) {
			return true, sessionExpiredError{redirectPath: resp.Header.Get("Location")}
		}
		return redirectsToExpectedLocation(resp, opts.ExpectedRedirectPath), nil
	}
	if resp.StatusCode >= 200 && resp.StatusCode <= 299 {
		return wasRedirected(req, resp, opts.ExpectedRedirectPath), nil
	}
	return false, nil
}

func (c *Client) setHeaders(req *http.Request) {
	copyHeaderInto(req.Header, c.defaultHeaders)
	copyHeaderInto(req.Header, c.session.Headers)
	req.Header.Set("User-Agent", c.userAgent)
}

func (c *Client) readBody(body io.Reader) ([]byte, error) {
	var buf bytes.Buffer
	limited := io.LimitReader(body, c.responseBodyLimit+1)
	_, err := io.Copy(&buf, limited)
	if err != nil {
		return nil, err
	}
	if int64(buf.Len()) > c.responseBodyLimit {
		return nil, &BodyLimitError{Limit: c.responseBodyLimit}
	}
	return buf.Bytes(), nil
}

func (c *Client) resolve(path string) string {
	u := *c.baseURL
	u.Path = strings.TrimRight(u.Path, "/") + "/" + strings.TrimLeft(path, "/")
	u.RawQuery = ""
	return u.String()
}

func multipartSuccessRedirectPath(path string) string {
	switch path {
	case AlertConfigPath:
		return alertConfigsRedirectPath
	case WatchListConfigPath:
		return watchListConfigsRedirectPath
	default:
		return ""
	}
}

func redirectsToLoginLocation(resp *http.Response) bool {
	location := resp.Header.Get("Location")
	if location == "" {
		return false
	}
	locationURL, err := url.Parse(location)
	if err != nil {
		return false
	}
	return normalizeResponsePath(locationURL.EscapedPath()) == loginPath
}

func redirectsToExpectedLocation(resp *http.Response, expectedPath string) bool {
	if expectedPath == "" {
		return false
	}
	location := resp.Header.Get("Location")
	if location == "" {
		return false
	}
	locationURL, err := url.Parse(location)
	if err != nil {
		return false
	}
	return normalizeResponsePath(locationURL.EscapedPath()) == expectedPath
}

func wasRedirected(req *http.Request, resp *http.Response, expectedPath string) bool {
	if expectedPath == "" {
		return false
	}
	if req == nil || req.URL == nil || resp == nil || resp.Request == nil || resp.Request.URL == nil {
		return false
	}
	if req.URL.String() == resp.Request.URL.String() {
		return false
	}
	return normalizeResponsePath(resp.Request.URL.EscapedPath()) == expectedPath
}

func statusError(resp *http.Response, body []byte) *StatusError {
	return &StatusError{
		StatusCode:  resp.StatusCode,
		Method:      responseMethod(resp),
		Path:        safeResponsePath(resp),
		ContentType: resp.Header.Get("Content-Type"),
		Header:      cloneHeader(resp.Header),
		RetryAfter:  resp.Header.Get("Retry-After"),
		Body:        bodyPreview(body),
	}
}

func contentError(resp *http.Response, body []byte, err error) *ContentError {
	return &ContentError{
		Method:      responseMethod(resp),
		Path:        safeResponsePath(resp),
		ContentType: resp.Header.Get("Content-Type"),
		Body:        bodyPreview(body),
		Err:         err,
	}
}

func bodyPreview(body []byte) string {
	if len(body) <= errorBodyPreviewLimit {
		return string(body)
	}
	return string(body[:errorBodyPreviewLimit])
}

func responseMethod(resp *http.Response) string {
	if resp == nil || resp.Request == nil {
		return ""
	}
	return resp.Request.Method
}
