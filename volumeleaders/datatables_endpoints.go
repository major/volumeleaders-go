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
