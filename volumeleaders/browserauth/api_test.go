package browserauth

import (
	"context"
	"errors"
	"net/http"
	"os/exec"
	"strings"
	"testing"

	"github.com/browserutils/kooky"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/major/volumeleaders-go/volumeleaders"
)

func TestFindSessionAcceptsBrowserauthLocalOptions(t *testing.T) {
	requireFindSessionSignature(t, FindSession)

	options := []Option{
		WithBrowser("firefox"),
		WithProfile("default"),
		WithoutValidation(),
		WithClientOptions(
			volumeleaders.WithBaseURL("https://example.test"),
			volumeleaders.WithHTTPClient(http.DefaultClient),
		),
	}

	assert.Len(t, options, 4, "FindSession options contract should include browser, profile, validation, and client bridges")
}

func TestFindSessionWithoutValidationReturnsCookiesWithoutXSRFToken(t *testing.T) {
	stubReadBrowserCookies(t, func(context.Context, ...kooky.Filter) (kooky.Cookies, error) {
		return requiredKookyCookies(), nil
	})

	session, err := FindSession(context.Background(), WithoutValidation())
	require.NoError(t, err, "FindSession(WithoutValidation()) returned unexpected error")
	assert.Empty(t, session.XSRFToken, "FindSession(WithoutValidation()) XSRFToken must be empty")
	assertSessionCookieValues(t, session, map[string]string{
		volumeleaders.SessionCookieName:             "session-id",
		volumeleaders.FormsAuthCookieName:           "forms-auth",
		volumeleaders.RequestVerificationCookieName: "verification",
	})
}

func TestFindSessionBrowserauthErrorContract(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{name: "browser unavailable", err: ErrBrowserUnavailable, want: ErrBrowserUnavailable},
		{name: "profile unavailable", err: ErrProfileUnavailable, want: ErrProfileUnavailable},
		{name: "required cookie missing", err: ErrRequiredCookieMissing, want: ErrRequiredCookieMissing},
		{name: "request verification token missing", err: ErrRequestVerificationTokenMissing, want: ErrRequestVerificationTokenMissing},
		{name: "session expired", err: volumeleaders.ErrSessionExpired, want: volumeleaders.ErrSessionExpired},
		{name: "validation request failed", err: ErrValidationFailed, want: ErrValidationFailed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := errors.Join(tt.err)
			assert.ErrorIs(t, wrapped, tt.want, "FindSession error %q must be matchable with errors.Is", tt.name)
		})
	}
}

func TestFindSessionValidationFailureSupportsErrorsAs(t *testing.T) {
	validationErr := &ValidationError{Err: errors.New("validation request failed")}

	var got *ValidationError
	assert.ErrorAs(t, validationErr, &got, "FindSession validation request failures must be matchable with errors.As")
	assert.ErrorIs(t, validationErr, ErrValidationFailed, "FindSession validation request failures must wrap ErrValidationFailed")
}

func TestRootPackageDoesNotExposeBrowserauthOptions(t *testing.T) {
	output := goDocVolumeleadersRoot(t)

	for _, symbol := range []string{"WithBrowser", "WithProfile", "WithoutValidation"} {
		t.Run(symbol, func(t *testing.T) {
			assert.NotContains(t, output, symbol, "root volumeleaders package must not expose browserauth option symbol %q", symbol)
		})
	}
}

func requireFindSessionSignature(
	t *testing.T,
	find func(context.Context, ...Option) (volumeleaders.Session, error),
) {
	t.Helper()
	require.NotNil(t, find, "FindSession must accept browserauth.Option, not volumeleaders.Option")
}

func goDocVolumeleadersRoot(t *testing.T) string {
	t.Helper()

	cmd := exec.Command("go", "doc", "github.com/major/volumeleaders-go/volumeleaders")
	output, err := cmd.CombinedOutput()
	require.NoError(t, err, "go doc volumeleaders root failed: %s", strings.TrimSpace(string(output)))
	return string(output)
}
