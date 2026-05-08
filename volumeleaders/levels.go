package volumeleaders

import "context"

// Browser endpoint paths for trade level APIs captured from VolumeLeaders.
const (
	ChartGetTradeLevelsPath                   = "/Chart/GetTradeLevels"
	Chart0GetTradeLevelsPath                  = "/Chart0/GetTradeLevels"
	TradeLevelsGetTradeLevelsPath             = "/TradeLevels/GetTradeLevels"
	TradeLevelTouchesGetTradeLevelTouchesPath = "/TradeLevelTouches/GetTradeLevelTouches"
)

// TradeLevelsRequest contains DataTables paging and optional endpoint filters
// for /TradeLevels/GetTradeLevels.
type TradeLevelsRequest = EndpointRequest

// TradeLevelTouchesRequest contains DataTables paging and optional endpoint
// filters for /TradeLevelTouches/GetTradeLevelTouches.
type TradeLevelTouchesRequest = EndpointRequest

// TradeLevel represents a VolumeLeaders trade level or level touch row.
type TradeLevel struct {
	Ticker                 string     `json:"Ticker"`
	Sector                 *string    `json:"Sector"`
	Industry               *string    `json:"Industry"`
	Name                   *string    `json:"Name"`
	Date                   AspNetDate `json:"Date"`
	MinDate                AspNetDate `json:"MinDate"`
	MaxDate                AspNetDate `json:"MaxDate"`
	FullDateTime           *string    `json:"FullDateTime"`
	FullTimeString24       *string    `json:"FullTimeString24"`
	Dates                  string     `json:"Dates"`
	Price                  float64    `json:"Price"`
	Dollars                float64    `json:"Dollars"`
	Volume                 int        `json:"Volume"`
	Trades                 int        `json:"Trades"`
	CumulativeDistribution float64    `json:"CumulativeDistribution"`
	TradeLevelRank         int        `json:"TradeLevelRank"`
	TotalRows              int        `json:"TotalRows"`
	TradeLevelTouches      int        `json:"TradeLevelTouches"`
	RelativeSize           float64    `json:"RelativeSize"`
}

// GetTradeLevels posts a typed DataTables request to
// /TradeLevels/GetTradeLevels.
func (c *Client) GetTradeLevels(
	ctx context.Context,
	req TradeLevelsRequest,
) (*DataTablesResponse[TradeLevel], error) {
	return c.getTradeLevels(ctx, TradeLevelsGetTradeLevelsPath, req)
}

// GetChartTradeLevels posts a typed DataTables request to
// /Chart/GetTradeLevels.
func (c *Client) GetChartTradeLevels(
	ctx context.Context,
	req TradeLevelsRequest,
) (*DataTablesResponse[TradeLevel], error) {
	return c.getTradeLevels(ctx, ChartGetTradeLevelsPath, req)
}

// GetChart0TradeLevels posts a typed DataTables request to
// /Chart0/GetTradeLevels.
func (c *Client) GetChart0TradeLevels(
	ctx context.Context,
	req TradeLevelsRequest,
) (*DataTablesResponse[TradeLevel], error) {
	return c.getTradeLevels(ctx, Chart0GetTradeLevelsPath, req)
}

func (c *Client) getTradeLevels(
	ctx context.Context,
	path string,
	req TradeLevelsRequest,
) (*DataTablesResponse[TradeLevel], error) {
	var result DataTablesResponse[TradeLevel]
	if err := c.postDataTables(
		ctx,
		path,
		req.DataTables,
		req.Filters,
		TradeLevelsColumns(),
		&result,
	); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTradeLevelTouches posts a typed DataTables request to
// /TradeLevelTouches/GetTradeLevelTouches.
func (c *Client) GetTradeLevelTouches(
	ctx context.Context,
	req TradeLevelTouchesRequest,
) (*DataTablesResponse[TradeLevel], error) {
	var result DataTablesResponse[TradeLevel]
	if err := c.postDataTables(
		ctx,
		TradeLevelTouchesGetTradeLevelTouchesPath,
		req.DataTables,
		req.Filters,
		TradeLevelTouchesColumns(),
		&result,
	); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTradeLevelTouchesLimit fetches up to limit trade level touches by paging
// through GetTradeLevelTouches. A zero or negative limit fetches all available
// records.
func (c *Client) GetTradeLevelTouchesLimit(
	ctx context.Context,
	req TradeLevelTouchesRequest,
	limit int,
) ([]TradeLevel, error) {
	return fetchLimit(
		ctx,
		limit,
		req.DataTables,
		func(ctx context.Context, dt DataTablesRequest) (*DataTablesResponse[TradeLevel], error) {
			req.DataTables = dt
			return c.GetTradeLevelTouches(ctx, req)
		},
	)
}

// TradeLevelsColumns returns the DataTables columns captured from the trade
// levels table.
func TradeLevelsColumns() []DataTablesColumn {
	return []DataTablesColumn{
		{Data: columnPrice, Name: columnPrice, Searchable: true, Orderable: true},
		{Data: columnDollars, Name: columnDollarsName, Searchable: true, Orderable: true},
		{Data: tradeColumnVolume, Name: columnSharesName, Searchable: true, Orderable: true},
		{Data: columnTrades, Name: columnTrades, Searchable: true, Orderable: true},
		{Data: columnRelativeSize, Name: columnRelativeSizeName, Searchable: true, Orderable: true},
		{
			Data:       columnCumulativeDistribution,
			Name:       columnPercentName,
			Searchable: true,
			Orderable:  true,
		},
		{Data: columnTradeLevelRank, Name: columnLevelRank, Searchable: true, Orderable: true},
		{Data: columnDates, Name: columnLevelDateRange, Searchable: true, Orderable: false},
	}
}

// TradeLevelTouchesColumns returns the DataTables columns captured from the
// trade level touches table.
func TradeLevelTouchesColumns() []DataTablesColumn {
	return []DataTablesColumn{
		{Data: columnFullDateTime, Name: "Date/Time", Searchable: true, Orderable: true},
		{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true},
		{Data: columnSector, Name: columnSector, Searchable: true, Orderable: true},
		{Data: columnIndustry, Name: columnIndustry, Searchable: true, Orderable: true},
		{Data: columnDollars, Name: columnDollarsName, Searchable: true, Orderable: true},
		{Data: tradeColumnVolume, Name: columnSharesName, Searchable: true, Orderable: true},
		{Data: columnTrades, Name: columnTrades, Searchable: true, Orderable: true},
		{Data: columnPrice, Name: columnPrice, Searchable: true, Orderable: true},
		{Data: columnRelativeSize, Name: columnRelativeSizeName, Searchable: true, Orderable: true},
		{
			Data:       columnCumulativeDistribution,
			Name:       columnPercentName,
			Searchable: true,
			Orderable:  true,
		},
		{Data: columnTradeLevelRank, Name: columnLevelRank, Searchable: true, Orderable: true},
		{Data: columnDates, Name: columnLevelDateRange, Searchable: true, Orderable: false},
		{Data: "", Name: "", Searchable: true, Orderable: false},
	}
}
