package volumeleaders

import (
	"context"
	"fmt"
	"net/http"
	"path"
	"regexp"
	"strings"
)

const (
	executiveSummaryPath      = "/ExecutiveSummary"
	volumeLeadersCookieDomain = "volumeleaders.com"
	sessionCookieName         = "ASP.NET_SessionId"
	formsAuthCookieName       = ".ASPXAUTH"
	requestVerificationCookie = "__RequestVerificationToken"
	xsrfHeaderName            = "X-Xsrf-Token"
)

var (
	xsrfInputPattern = regexp.MustCompile(`(?is)<input\b[^>]*\bname=["']__RequestVerificationToken["'][^>]*>`)
	xsrfValuePattern = regexp.MustCompile(`(?is)\bvalue=["']([^"']+)["']`)
)

func requiredCookieNames() []string {
	return []string{sessionCookieName, formsAuthCookieName}
}

// FetchXSRFToken retrieves the hidden ASP.NET request verification token from
// the authenticated ExecutiveSummary page.
func FetchXSRFToken(ctx context.Context, session Session, opts ...Option) (string, error) {
	client, err := NewClient(session, opts...)
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, client.resolve(executiveSummaryPath), nil)
	if err != nil {
		return "", fmt.Errorf("create XSRF token request: %w", err)
	}
	client.setHeaders(req)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	for _, cookie := range client.session.Cookies {
		req.AddCookie(cookie)
	}

	resp, err := client.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fetch XSRF token page: %w", err)
	}
	defer resp.Body.Close()

	if redirectedToLogin(resp) {
		return "", sessionExpiredError{redirectPath: safeResponsePath(resp)}
	}

	body, err := client.readBody(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read XSRF token page: %w", err)
	}
	if looksLikeLoginPage(resp, body) {
		return "", sessionExpiredError{redirectPath: safeResponsePath(resp)}
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", statusError(resp, body)
	}

	token, err := parseXSRFToken(body)
	if err != nil {
		return "", err
	}
	return token, nil
}

func parseXSRFToken(body []byte) (string, error) {
	input := xsrfInputPattern.Find(body)
	if input == nil {
		return "", ErrXSRFTokenNotFound
	}
	matches := xsrfValuePattern.FindSubmatch(input)
	if matches == nil || len(matches[1]) == 0 {
		return "", ErrXSRFTokenNotFound
	}
	return string(matches[1]), nil
}

func redirectedToLogin(resp *http.Response) bool {
	return normalizeResponsePath(safeResponsePath(resp)) == "/login"
}

func looksLikeLoginPage(resp *http.Response, body []byte) bool {
	if resp == nil || !strings.Contains(strings.ToLower(resp.Header.Get("Content-Type")), "text/html") {
		return false
	}
	lowerBody := strings.ToLower(string(body))
	return strings.Contains(lowerBody, "<html") && strings.Contains(lowerBody, "login")
}

func safeResponsePath(resp *http.Response) string {
	if resp == nil || resp.Request == nil || resp.Request.URL == nil {
		return "unknown redirect target"
	}
	if resp.Request.URL.EscapedPath() == "" {
		return "/"
	}
	return resp.Request.URL.EscapedPath()
}

func normalizeResponsePath(responsePath string) string {
	if responsePath == "" {
		return "/"
	}
	if !strings.HasPrefix(responsePath, "/") {
		responsePath = "/" + responsePath
	}
	return strings.ToLower(path.Clean(responsePath))
}
