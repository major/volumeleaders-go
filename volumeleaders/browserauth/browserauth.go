// Package browserauth discovers VolumeLeaders sessions from local browser cookie stores.
package browserauth

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/browserutils/kooky"
	_ "github.com/browserutils/kooky/browser/all" // Register supported browser cookie stores.

	"github.com/major/volumeleaders-go/volumeleaders"
)

//nolint:gochecknoglobals // Test seam for substituting kooky.ReadCookies without touching browser stores.
var readBrowserCookies = kooky.ReadCookies

// New creates a client from the active VolumeLeaders browser session.
//
// The function reads cookies from local browser stores, fetches the current
// ASP.NET XSRF token from /ExecutiveSummary, and then builds a client with the
// resulting session. Callers must already be logged in at volumeleaders.com in a
// supported local browser.
func New(ctx context.Context, opts ...volumeleaders.Option) (*volumeleaders.Client, error) {
	session, err := Session(ctx, opts...)
	if err != nil {
		return nil, err
	}
	return volumeleaders.NewClient(session, opts...)
}

// Session extracts browser cookies and fetches the XSRF token needed for
// authenticated VolumeLeaders XHR endpoints.
func Session(ctx context.Context, opts ...volumeleaders.Option) (volumeleaders.Session, error) {
	cookies, err := ExtractCookies(ctx)
	if err != nil {
		return volumeleaders.Session{}, err
	}

	session := volumeleaders.SessionFromCookies(cookies, "", nil)
	token, err := volumeleaders.FetchXSRFToken(ctx, session, opts...)
	if err != nil {
		return volumeleaders.Session{}, fmt.Errorf("fetch XSRF token: %w", err)
	}
	return volumeleaders.SessionFromCookies(cookies, token, nil), nil
}

// FindSession discovers a VolumeLeaders session from local browser cookies.
//
// By default, FindSession reads valid VolumeLeaders cookies from supported local
// browser stores and validates the session by fetching the ASP.NET request
// verification token. Use WithoutValidation when callers only need the browser
// cookies and can tolerate an empty XSRF token.
func FindSession(ctx context.Context, opts ...Option) (volumeleaders.Session, error) {
	cfg := findConfig{}
	for _, opt := range opts {
		if opt != nil {
			opt.apply(&cfg)
		}
	}

	browserCookies, err := readBrowserCookies(ctx, kooky.Valid, kooky.DomainHasSuffix(volumeleaders.CookieDomain))
	if err != nil && len(browserCookies) == 0 {
		return volumeleaders.Session{}, fmt.Errorf("%w: read browser cookies", volumeleaders.ErrBrowserCookiesUnavailable)
	}

	selected := volumeLeadersCookies(filterBrowserCookies(browserCookies, cfg))
	if missing := volumeleaders.MissingCookieFields(selected); len(missing) != 0 {
		return volumeleaders.Session{}, missingCookieError(cfg, missing)
	}

	if cfg.skipValidation {
		return volumeleaders.SessionFromCookies(selected, "", nil), nil
	}

	session := volumeleaders.SessionFromCookies(selected, "", nil)
	token, err := volumeleaders.FetchXSRFToken(ctx, session, cfg.clientOpts...)
	if err != nil {
		return volumeleaders.Session{}, &ValidationError{Err: err}
	}
	return volumeleaders.SessionFromCookies(selected, token, nil), nil
}

// ExtractCookies reads VolumeLeaders authentication cookies from local browser
// stores.
//
// Browser store read errors are tolerated as long as the required cookies are
// found in at least one supported browser.
func ExtractCookies(ctx context.Context) ([]*http.Cookie, error) {
	validCookies, validErr := readBrowserCookies(ctx, kooky.Valid, kooky.DomainHasSuffix(volumeleaders.CookieDomain))
	selected := volumeLeadersCookies(validCookies)
	if len(volumeleaders.MissingCookieFields(selected)) == 0 {
		return selected, nil
	}

	allCookies, allErr := readBrowserCookies(ctx, kooky.DomainHasSuffix(volumeleaders.CookieDomain))
	return nil, fmt.Errorf(
		"%w: missing %s; valid cookies found: %d; browser stores with matching cookies: %d; browser read errors: %w",
		volumeleaders.ErrBrowserCookiesUnavailable,
		strings.Join(volumeleaders.MissingCookieFields(selected), ", "),
		len(validCookies),
		browserStoreCount(allCookies),
		errors.Join(validErr, allErr),
	)
}

func volumeLeadersCookies(cookies kooky.Cookies) []*http.Cookie {
	selected := make([]*http.Cookie, 0, len(volumeleaders.MissingCookieFields(nil)))
	for _, cookie := range cookies {
		if cookie == nil {
			continue
		}
		switch cookie.Name {
		case volumeleaders.SessionCookieName,
			volumeleaders.FormsAuthCookieName,
			volumeleaders.RequestVerificationCookieName:
			// #nosec G124 - these cookies come from a browser store and are sent back to VolumeLeaders only.
			selected = append(selected, &http.Cookie{Name: cookie.Name, Value: cookie.Value})
		}
	}
	return selected
}

func filterBrowserCookies(cookies kooky.Cookies, cfg findConfig) kooky.Cookies {
	if cfg.browser == "" && cfg.profile == "" {
		return cookies
	}

	filtered := make(kooky.Cookies, 0, len(cookies))
	for _, cookie := range cookies {
		if cookie == nil || cookie.Browser == nil {
			continue
		}
		if cfg.browser != "" && cookie.Browser.Browser() != cfg.browser {
			continue
		}
		if cfg.profile != "" && cookie.Browser.Profile() != cfg.profile {
			continue
		}
		filtered = append(filtered, cookie)
	}
	return filtered
}

func missingCookieError(cfg findConfig, missing []string) error {
	var errs []error
	if cfg.browser != "" {
		errs = append(errs, ErrBrowserUnavailable)
	}
	if cfg.profile != "" {
		errs = append(errs, ErrProfileUnavailable)
	}
	errs = append(errs, volumeleaders.ErrBrowserCookiesUnavailable, ErrRequiredCookieMissing)
	return fmt.Errorf("%w: missing %s", errors.Join(errs...), strings.Join(missing, ", "))
}

func browserStoreCount(cookies kooky.Cookies) int {
	stores := make(map[string]struct{})
	for _, cookie := range cookies {
		if cookie == nil || cookie.Browser == nil {
			stores["unknown"] = struct{}{}
			continue
		}
		stores[cookie.Browser.Browser()+":"+cookie.Browser.Profile()] = struct{}{}
	}
	return len(stores)
}
