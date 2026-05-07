# volumeleaders-go

[![CI](https://github.com/major/volumeleaders-go/actions/workflows/ci.yml/badge.svg)](https://github.com/major/volumeleaders-go/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/major/volumeleaders-go/branch/main/graph/badge.svg)](https://codecov.io/gh/major/volumeleaders-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/major/volumeleaders-go)](https://goreportcard.com/report/github.com/major/volumeleaders-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/major/volumeleaders-go/volumeleaders.svg)](https://pkg.go.dev/github.com/major/volumeleaders-go/volumeleaders)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/major/volumeleaders-go/badge)](https://scorecard.dev/viewer/?uri=github.com/major/volumeleaders-go)
[![License](https://img.shields.io/github/license/major/volumeleaders-go)](LICENSE)

`volumeleaders-go` provides a small Go client for the browser-backed VolumeLeaders endpoints used by authenticated sessions. The package is intended to be the API layer underneath higher-level porcelain, including LLM-facing tools that should not know VolumeLeaders form keys, headers, or response envelope details.

> [!NOTE]
> This project is unofficial and is not affiliated with, sponsored by, or endorsed by VolumeLeaders or volumeleaders.com in any way.

## Installation

```bash
go get github.com/major/volumeleaders-go/volumeleaders
```

## Explicit session usage

Use `NewClient` when the caller already has VolumeLeaders browser cookies and an ASP.NET XSRF token.

```go
session := volumeleaders.NewSession(sessionID, authCookie, xsrfToken)
if err := volumeleaders.ValidateSession(session); err != nil {
	return err
}
client, err := volumeleaders.NewClient(session)
if err != nil {
	return err
}
```

`ValidateSession` and `MissingSessionFields` are optional preflight helpers. `NewClient` still accepts partial sessions so tests and custom transports can exercise unauthenticated responses, but porcelain should validate before making user-facing calls.

## Browser-backed session usage

Use `browserauth.New` when the program is running on a desktop machine with a supported browser already logged in to VolumeLeaders.

```go
import "github.com/major/volumeleaders-go/volumeleaders/browserauth"

client, err := browserauth.New(ctx)
if err != nil {
	return err
}
```

This path reads local browser cookie stores with `github.com/browserutils/kooky`, fetches the current XSRF token from `/ExecutiveSummary`, and builds a `Client`. It is convenient for desktop automation, but service and container callers should prefer explicit sessions so browser-store access stays outside the core API layer. The root `volumeleaders` package does not import browser-store, SQLite, or desktop keyring packages.

## Trades

`ListTrades` wraps `/Trades/GetTrades` with a typed query so callers do not need to know DataTables field names or VolumeLeaders form keys.

```go
page, err := client.ListTrades(ctx, volumeleaders.TradeQuery{
	Tickers: []string{"AXP"},
	Length:  50,
})
if err != nil {
	return err
}
for _, trade := range page.Trades {
	fmt.Println(trade.Ticker, trade.TradeID)
}
```

Use `TradeQuery` for common filters such as tickers, date range, volume, price, dollars, relative size, rank, dark pools, sweeps, and session buckets. Query validation returns errors matched by `IsInvalidQuery`.

The lower-level `GetTrades` API remains available for compatibility and for experiments with newly discovered form fields. Keep those raw mappings inside Go code, not in prompts, scripts, or user-facing porcelain.

## Capped pagination

Use `ListTradesLimit` when porcelain needs a bounded result set. The explicit limit prevents unbounded LLM-triggered requests.

```go
trades, err := client.ListTradesLimit(ctx, volumeleaders.TradeQuery{
	Tickers: []string{"AXP"},
	Length:  100,
}, 250)
if err != nil {
	return err
}
```

When `TradeQuery.Length` is set, it controls page size up to the limit. When it is unset, the helper uses the smaller of the limit and the default DataTables page size.

## Error handling

The package exposes typed errors and predicates for the cases porcelain usually needs to classify.

```go
if volumeleaders.IsAuthError(err) {
	return fmt.Errorf("VolumeLeaders authentication needs attention: %w", err)
}
if volumeleaders.IsStatusCode(err, http.StatusTooManyRequests) {
	return fmt.Errorf("VolumeLeaders rate limited the request: %w", err)
}
if volumeleaders.IsBodyLimit(err) {
	return fmt.Errorf("VolumeLeaders returned too much data: %w", err)
}
if volumeleaders.IsUnexpectedContent(err) {
	return fmt.Errorf("VolumeLeaders returned an unexpected response: %w", err)
}
```

`IsAuthError` covers invalid sessions, missing browser cookies, expired sessions, missing XSRF tokens, HTTP 401 or 403 responses, POST redirects to login, and HTML login pages returned with HTTP 200. `StatusError` and `ContentError` include method, path, content type, and a response body preview when available.

## API coverage

The package implements the authenticated endpoints observed in the browser captures:

- Trades: `GetTrades`, `GetTradesLimit`, `ListTrades`, `ListTradesLimit`, and `GetAllSnapshots` for `/Trades/GetTrades` and `/Trades/GetAllSnapshots`.
- Executive summary: `GetExhaustionScores`, `GetWelcomeTrades`, and `GetWelcomeTradeClusters`.
- Trade clusters: `GetTradeClusters` and `GetTradeClusterBombs`.
- Trade levels: `GetTradeLevels` and `GetTradeLevelTouches`.
- Watchlists: `GetWatchLists`, `GetWatchListTickers`, and `DeleteWatchList`.
- Earnings: `GetEarnings` with the captured request shape. The response row model is inferred from the DataTables columns because the available earnings captures did not include a successful JSON response.

Each DataTables endpoint has a low-level `Get*` method that accepts typed paging metadata plus `url.Values` filters. `ListTrades` and `ListTradesLimit` add a higher-level query surface for the most commonly used trades endpoint.

Endpoints not present in the captures remain outside this library until their browser traffic is captured and modeled. That includes alert create/update flows, multipart watchlist update forms, and institutional volume leaderboard families.

## Consumer boundary

The API layer owns authenticated request construction, VolumeLeaders endpoint paths, DataTables encoding, response models, and error classification. Porcelain should own natural-language parsing, user prompts, output shaping, retry policy, and safety limits.
