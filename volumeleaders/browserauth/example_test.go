package browserauth_test

import (
	"context"
	"fmt"

	"github.com/major/volumeleaders-go/volumeleaders"
	"github.com/major/volumeleaders-go/volumeleaders/browserauth"
)

// ExampleFindSession shows the default discovery path for desktop automation.
// FindSession reads cookies from all supported local browser stores, validates
// the session against VolumeLeaders, and returns a Session ready for NewClient.
func ExampleFindSession() {
	ctx := context.Background()

	session, err := browserauth.FindSession(ctx)
	if err != nil {
		fmt.Println("session discovery failed:", err)
		return
	}

	client, err := volumeleaders.NewClient(session)
	if err != nil {
		fmt.Println("client creation failed:", err)
		return
	}

	_ = client
	// Output:
}

// ExampleFindSession_withBrowser shows explicit browser selection when multiple
// supported browsers are installed. Only cookies from the named browser are
// considered; FindSession returns ErrBrowserUnavailable if that browser has no
// matching VolumeLeaders cookies.
func ExampleFindSession_withBrowser() {
	ctx := context.Background()

	session, err := browserauth.FindSession(ctx, browserauth.WithBrowser("firefox"))
	if err != nil {
		fmt.Println("session discovery failed:", err)
		return
	}

	client, err := volumeleaders.NewClient(session)
	if err != nil {
		fmt.Println("client creation failed:", err)
		return
	}

	_ = client
	// Output:
}

// ExampleFindSession_withoutValidation shows skipping the XSRF token fetch.
// The returned Session contains the discovered browser cookies but has an empty
// XSRFToken field. Use this when the caller will supply the token separately or
// when the validation request itself is not needed.
func ExampleFindSession_withoutValidation() {
	ctx := context.Background()

	session, err := browserauth.FindSession(ctx, browserauth.WithoutValidation())
	if err != nil {
		fmt.Println("session discovery failed:", err)
		return
	}

	// XSRFToken is empty when WithoutValidation is used.
	fmt.Println("xsrf empty:", session.XSRFToken == "")

	client, err := volumeleaders.NewClient(session)
	if err != nil {
		fmt.Println("client creation failed:", err)
		return
	}

	_ = client
	// Output:
	// xsrf empty: true
}
