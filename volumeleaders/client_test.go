package volumeleaders

import (
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBodyPreviewTruncatesLargeBodies(t *testing.T) {
	body := []byte(strings.Repeat("x", errorBodyPreviewLimit+1))

	preview := bodyPreview(body)

	assert.Len(t, preview, errorBodyPreviewLimit, "bodyPreview(large body)")
}

func TestResponseMethodHandlesMissingRequest(t *testing.T) {
	assert.Empty(t, responseMethod(nil), "responseMethod(nil)")
	assert.Empty(t, responseMethod(&http.Response{}), "responseMethod(response without request)")
	assert.Equal(
		t,
		http.MethodPost,
		responseMethod(&http.Response{Request: &http.Request{Method: http.MethodPost}}),
		"responseMethod(response with request)",
	)
}

func TestRedirectHelpersHandleLocations(t *testing.T) {
	loginResp := &http.Response{Header: http.Header{"Location": {"/LOGIN"}}}
	expectedResp := &http.Response{Header: http.Header{"Location": {"/alertconfigs"}}}
	invalidResp := &http.Response{Header: http.Header{"Location": {"http://[::1"}}}

	assert.True(t, redirectsToLoginLocation(loginResp), "redirectsToLoginLocation(login redirect)")
	assert.False(
		t,
		redirectsToLoginLocation(&http.Response{Header: http.Header{}}),
		"redirectsToLoginLocation(missing location)",
	)
	assert.False(t, redirectsToLoginLocation(invalidResp), "redirectsToLoginLocation(invalid location)")
	assert.True(
		t,
		redirectsToExpectedLocation(expectedResp, "/alertconfigs"),
		"redirectsToExpectedLocation(expected redirect)",
	)
	assert.False(t, redirectsToExpectedLocation(expectedResp, ""), "redirectsToExpectedLocation(empty expected path)")
	assert.False(
		t,
		redirectsToExpectedLocation(invalidResp, "/alertconfigs"),
		"redirectsToExpectedLocation(invalid location)",
	)
}

func TestWasRedirected(t *testing.T) {
	originalURL := &url.URL{Scheme: "https", Host: "volumeleaders.com", Path: "/Alerts/Save"}
	redirectURL := &url.URL{Scheme: "https", Host: "volumeleaders.com", Path: "/alertconfigs"}
	req := &http.Request{URL: originalURL}
	resp := &http.Response{Request: &http.Request{URL: redirectURL}}

	assert.True(t, wasRedirected(req, resp, "/alertconfigs"), "wasRedirected(redirected response)")
	assert.False(t, wasRedirected(req, resp, ""), "wasRedirected(empty expected path)")
	assert.False(t, wasRedirected(nil, resp, "/alertconfigs"), "wasRedirected(nil request)")
	assert.False(
		t,
		wasRedirected(req, &http.Response{Request: &http.Request{URL: originalURL}}, "/alertconfigs"),
		"wasRedirected(same URL)",
	)
}
