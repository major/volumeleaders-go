package volumeleaders

import (
	"context"
	"net/url"
)

const (
	columnCharts                             = "Charts"
	columnConditions                         = "Conditions"
	columnCumulativeDistribution             = "CumulativeDistribution"
	columnCurrent                            = "Current"
	columnDate                               = "Date"
	columnDates                              = "Dates"
	columnDollars                            = "Dollars"
	columnDollarsName                        = "$$"
	columnDollarsMultiplier                  = "DollarsMultiplier"
	columnFullDateTime                       = "FullDateTime"
	columnIndustry                           = "Industry"
	columnLastDateName                       = "Last Date"
	columnLastComparableTradeClusterBombDate = "LastComparableTradeClusterBombDate"
	columnLastComparibleTradeClusterDate     = "LastComparibleTradeClusterDate"
	columnLevelDateRange                     = "Level Date Range"
	columnLevelRank                          = "Level Rank"
	columnMinFullTimeString24                = "MinFullTimeString24"
	columnName                               = "Name"
	columnNearestTop10TradeLevel             = "NearestTop10TradeLevel"
	columnPercentName                        = "PCT"
	columnPrice                              = "Price"
	columnRankName                           = "Rank"
	columnRelativeSize                       = "RelativeSize"
	columnRelativeSizeName                   = "RS"
	columnRName                              = "R"
	columnSector                             = "Sector"
	columnSharesName                         = "Shares"
	columnShName                             = "Sh"
	columnTicker                             = "Ticker"
	columnTickers                            = "Tickers"
	columnTradeClusterBombRank               = "TradeClusterBombRank"
	columnTradeClusterRank                   = "TradeClusterRank"
	columnTradeCount                         = "TradeCount"
	columnTradeLevelRank                     = "TradeLevelRank"
	columnTradeRank                          = "TradeRank"
	columnTrades                             = "Trades"
	columnTradeShortName                     = "Trade"
	columnTotalInstitutionalDollarsRank      = "TotalInstitutionalDollarsRank"
)

// SaveConfigRequest contains the multipart form fields submitted when creating
// or editing a VolumeLeaders configuration (alert or watchlist).
type SaveConfigRequest struct {
	// Fields is a low-level captured-form escape hatch. Callers must provide the
	// exact browser field names and values accepted by VolumeLeaders.
	Fields url.Values
}

func colTicker() DataTablesColumn {
	return DataTablesColumn{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true}
}

func colSector() DataTablesColumn {
	return DataTablesColumn{Data: columnSector, Name: columnSector, Searchable: true, Orderable: true}
}

func colIndustry() DataTablesColumn {
	return DataTablesColumn{Data: columnIndustry, Name: columnIndustry, Searchable: true, Orderable: true}
}

func colPrice() DataTablesColumn {
	return DataTablesColumn{Data: columnPrice, Name: columnPrice, Searchable: true, Orderable: true}
}

func colDollars() DataTablesColumn {
	return DataTablesColumn{Data: columnDollars, Name: columnDollarsName, Searchable: true, Orderable: true}
}

func colShares() DataTablesColumn {
	return DataTablesColumn{Data: tradeColumnVolume, Name: columnShName, Searchable: true, Orderable: true}
}

func colMultiplier() DataTablesColumn {
	return DataTablesColumn{
		Data: columnDollarsMultiplier, Name: columnRelativeSizeName,
		Searchable: true, Orderable: true,
	}
}

func colCumulativePct() DataTablesColumn {
	return DataTablesColumn{
		Data: columnCumulativeDistribution, Name: columnPercentName,
		Searchable: true, Orderable: true,
	}
}

func colCurrent() DataTablesColumn {
	return DataTablesColumn{Data: columnCurrent, Name: columnCurrent, Searchable: true, Orderable: false}
}

func (c *Client) saveConfig(ctx context.Context, path string, fields url.Values) error {
	return c.postMultipartForm(ctx, path, mergeValues(fields, nil), nil)
}

func (c *Client) postDataTables(
	ctx context.Context,
	path string,
	req DataTablesRequest,
	filters url.Values,
	columns []DataTablesColumn,
	result any,
) error {
	if len(req.Columns) == 0 {
		req.Columns = columns
	}
	req.Extra = mergeValues(req.Extra, filters)
	return c.postForm(ctx, path, EncodeDataTablesRequest(req).Encode(), result)
}

// getEndpoint posts a typed DataTables request and decodes the response into a
// DataTablesResponse[T]. Every DataTables Get* endpoint method delegates here.
func getEndpoint[T any](
	ctx context.Context,
	c *Client,
	path string,
	req EndpointRequest,
	columns []DataTablesColumn,
) (*DataTablesResponse[T], error) {
	var result DataTablesResponse[T]
	if err := c.postDataTables(ctx, path, req.DataTables, req.Filters, columns, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// getEndpointLimit pages through a Get* endpoint to collect up to limit
// records. A zero or negative limit fetches all available records.
func getEndpointLimit[T any](
	ctx context.Context,
	req EndpointRequest,
	limit int,
	fetch func(context.Context, EndpointRequest) (*DataTablesResponse[T], error),
) ([]T, error) {
	return fetchLimit(
		ctx,
		limit,
		req.DataTables,
		func(ctx context.Context, dt DataTablesRequest) (*DataTablesResponse[T], error) {
			req.DataTables = dt
			return fetch(ctx, req)
		},
	)
}
