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
