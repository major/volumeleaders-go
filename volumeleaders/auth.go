package volumeleaders

import (
	"context"
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
	loginPath                 = "/login"
	unknownRedirectTarget     = "unknown redirect target"
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
	body, err := client.getHTML(ctx, executiveSummaryPath)
	if err != nil {
		return "", err
	}
	return parseXSRFToken(body)
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
	return normalizeResponsePath(safeResponsePath(resp)) == loginPath
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
		return unknownRedirectTarget
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
