---
name: volumeleaders-go
description: |
  Go API client for authenticated VolumeLeaders browser-backed endpoints. Use it when writing Go code that needs typed trades queries, captured DataTables endpoints, explicit VolumeLeaders sessions, optional browser cookie discovery, or VolumeLeaders error classification.
metadata:
  author: major
  version: dev
---

# volumeleaders-go

## Use this library for

- Typed trades queries against `/Trades/GetTrades` through `Client.ListTrades` and `Client.ListTradesLimit`.
- Typed low-level APIs for captured VolumeLeaders DataTables endpoints, including executive summary, clusters, levels, watchlists, earnings, and snapshots.
- Explicit authenticated sessions with `volumeleaders.NewSession` or `volumeleaders.SessionFromCookies`.
- Optional local browser cookie discovery through `volumeleaders/browserauth`.
- Error classification with predicates such as `IsAuthError`, `IsStatusCode`, `IsInvalidQuery`, `IsBodyLimit`, and `IsUnexpectedContent`.

## Do not expose raw protocol details

Keep VolumeLeaders form keys, DataTables field names, request headers, and endpoint paths inside Go code. Higher-level porcelain and LLM prompts should call typed APIs such as `TradeQuery` instead of building `url.Values` directly.

## Authentication guidance

Prefer explicit sessions in services and containers:

```go
session := volumeleaders.NewSession(sessionID, authCookie, xsrfToken)
client, err := volumeleaders.NewClient(session)
```

Use `browserauth.New(ctx)` only for desktop automation where a user is already logged in through a supported browser. The root `volumeleaders` package must stay free of browser-store, SQLite, desktop keyring, and `kooky` dependencies.

## Current API coverage

Implemented high-level API:

- `ListTrades(ctx, TradeQuery)`.
- `ListTradesLimit(ctx, TradeQuery, limit)`.

Compatibility and low-level helpers:

- `GetTrades(ctx, TradesRequest)`.
- `GetTradesLimit(ctx, TradesRequest, limit)`.
- `GetAllSnapshots(ctx)` and `GetAllSnapshotsString(ctx)`.
- `GetExhaustionScores(ctx, ExhaustionScoresRequest)`.
- `GetWelcomeTrades(ctx, WelcomeTradesRequest)` and `GetWelcomeTradeClusters(ctx, WelcomeTradeClustersRequest)`.
- `GetTradeClusters(ctx, TradeClustersRequest)` and `GetTradeClusterBombs(ctx, TradeClusterBombsRequest)`.
- `GetTradeLevels(ctx, TradeLevelsRequest)` and `GetTradeLevelTouches(ctx, TradeLevelTouchesRequest)`.
- `GetWatchLists(ctx, WatchListsRequest)`, `GetWatchListTickers(ctx, WatchListTickersRequest)`, and `DeleteWatchList(ctx, DeleteWatchListRequest)`.
- `GetEarnings(ctx, EarningsRequest)` with an inferred row model because captured earnings responses did not include a successful JSON body.
- `EncodeDataTablesRequest` and DataTables model types.

Not implemented yet:

- Trade alerts and trade cluster alerts.
- Alert configuration create, update, delete, and multipart form submission.
- Institutional, after-hours institutional, and total volume leaderboards.
- Watchlist update and multipart form submission.

Add typed Go APIs before exposing these endpoint families through prompts or porcelain.

## Recovery guidance

- `IsAuthError(err)`: ask the caller to refresh the VolumeLeaders browser session or provide fresh cookies and XSRF token.
- `IsInvalidQuery(err)`: fix the typed query values before retrying.
- `IsStatusCode(err, http.StatusTooManyRequests)`: back off or reduce request volume.
- `IsBodyLimit(err)`: narrow the query or reduce page size.
- `IsUnexpectedContent(err)`: inspect `ContentError` metadata because VolumeLeaders likely returned HTML, a schema mismatch, or another non-JSON response.
