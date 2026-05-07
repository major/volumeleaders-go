# volumeleaders-go

[![CI](https://github.com/major/volumeleaders-go/actions/workflows/ci.yml/badge.svg)](https://github.com/major/volumeleaders-go/actions/workflows/ci.yml)
[![codecov](https://codecov.io/gh/major/volumeleaders-go/branch/main/graph/badge.svg)](https://codecov.io/gh/major/volumeleaders-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/major/volumeleaders-go)](https://goreportcard.com/report/github.com/major/volumeleaders-go)
[![Go Reference](https://pkg.go.dev/badge/github.com/major/volumeleaders-go/volumeleaders.svg)](https://pkg.go.dev/github.com/major/volumeleaders-go/volumeleaders)
[![OpenSSF Scorecard](https://api.scorecard.dev/projects/github.com/major/volumeleaders-go/badge)](https://scorecard.dev/viewer/?uri=github.com/major/volumeleaders-go)
[![License](https://img.shields.io/github/license/major/volumeleaders-go)](LICENSE)

`volumeleaders-go` is a small Go client for authenticated, browser-backed VolumeLeaders endpoints. It keeps session handling, VolumeLeaders form keys, DataTables request encoding, and response envelopes inside a typed Go API for higher-level tools to call safely.

> [!NOTE]
> This project is unofficial and is not affiliated with, sponsored by, or endorsed by VolumeLeaders or volumeleaders.com in any way.

## Features

- **Explicit sessions** - build clients from caller-provided VolumeLeaders browser cookies and ASP.NET XSRF tokens.
- **Browser-backed auth** - optional desktop helper reads supported local browser cookie stores when a user is already logged in.
- **Typed trades queries** - `ListTrades` and `ListTradesLimit` hide raw DataTables form fields behind `TradeQuery`.
- **Captured endpoints** - typed low-level APIs for trades, snapshots, executive summary, clusters, levels, watchlists, and earnings.
- **Structured errors** - predicates classify auth, status-code, invalid-query, body-limit, and unexpected-content failures.
- **Safety boundaries** - capped pagination prevents unbounded caller-triggered requests.

## Installation

```bash
go get github.com/major/volumeleaders-go/volumeleaders
```

Requires Go 1.26.2 or later.

## Quick start

```go
package main

import (
	"context"
	"fmt"

	"github.com/major/volumeleaders-go/volumeleaders"
)

func main() {
	session := volumeleaders.NewSession("session-id", "auth-cookie", "xsrf-token")
	client, err := volumeleaders.NewClient(session)
	if err != nil {
		panic(err)
	}

	page, err := client.ListTrades(context.Background(), volumeleaders.TradeQuery{
		Tickers: []string{"AXP"},
		Length:  50,
	})
	if err != nil {
		panic(err)
	}

	for _, trade := range page.Trades {
		fmt.Println(trade.Ticker, trade.TradeID)
	}
}
```

## Documentation

- [Usage guide](docs/usage.md) covers explicit sessions, browser-backed sessions, trades, capped pagination, error handling, implemented endpoints, and consumer boundaries.
- Full API documentation is available on [pkg.go.dev](https://pkg.go.dev/github.com/major/volumeleaders-go/volumeleaders).

## Contributing

Contributions are welcome. Please open an issue to discuss larger changes before submitting a pull request.

## License

[Apache License 2.0](LICENSE)
