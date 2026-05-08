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
	return getEndpoint[TradeLevel](ctx, c, TradeLevelsGetTradeLevelsPath, req, TradeLevelsColumns())
}

// GetChartTradeLevels posts a typed DataTables request to
// /Chart/GetTradeLevels.
func (c *Client) GetChartTradeLevels(
	ctx context.Context,
	req TradeLevelsRequest,
) (*DataTablesResponse[TradeLevel], error) {
	return getEndpoint[TradeLevel](ctx, c, ChartGetTradeLevelsPath, req, TradeLevelsColumns())
}

// GetChart0TradeLevels posts a typed DataTables request to
// /Chart0/GetTradeLevels.
func (c *Client) GetChart0TradeLevels(
	ctx context.Context,
	req TradeLevelsRequest,
) (*DataTablesResponse[TradeLevel], error) {
	return getEndpoint[TradeLevel](ctx, c, Chart0GetTradeLevelsPath, req, TradeLevelsColumns())
}

// GetTradeLevelTouches posts a typed DataTables request to
// /TradeLevelTouches/GetTradeLevelTouches.
func (c *Client) GetTradeLevelTouches(
	ctx context.Context,
	req TradeLevelTouchesRequest,
) (*DataTablesResponse[TradeLevel], error) {
	return getEndpoint[TradeLevel](ctx, c, TradeLevelTouchesGetTradeLevelTouchesPath, req, TradeLevelTouchesColumns())
}

// GetTradeLevelTouchesLimit fetches up to limit trade level touches by paging
// through GetTradeLevelTouches. A zero or negative limit fetches all available
// records.
func (c *Client) GetTradeLevelTouchesLimit(
	ctx context.Context,
	req TradeLevelTouchesRequest,
	limit int,
) ([]TradeLevel, error) {
	return getEndpointLimit(ctx, req, limit, c.GetTradeLevelTouches)
}

// TradeLevelsColumns returns the DataTables columns captured from the trade
// levels table.
func TradeLevelsColumns() []DataTablesColumn {
	return []DataTablesColumn{
		colPrice(),
		colDollars(),
		{Data: tradeColumnVolume, Name: columnSharesName, Searchable: true, Orderable: true},
		{Data: columnTrades, Name: columnTrades, Searchable: true, Orderable: true},
		{Data: columnRelativeSize, Name: columnRelativeSizeName, Searchable: true, Orderable: true},
		colCumulativePct(),
		{Data: columnTradeLevelRank, Name: columnLevelRank, Searchable: true, Orderable: true},
		{Data: columnDates, Name: columnLevelDateRange, Searchable: true, Orderable: false},
	}
}

// TradeLevelTouchesColumns returns the DataTables columns captured from the
// trade level touches table.
func TradeLevelTouchesColumns() []DataTablesColumn {
	return []DataTablesColumn{
		{Data: columnFullDateTime, Name: "Date/Time", Searchable: true, Orderable: true},
		colTicker(),
		colSector(),
		colIndustry(),
		colDollars(),
		{Data: tradeColumnVolume, Name: columnSharesName, Searchable: true, Orderable: true},
		{Data: columnTrades, Name: columnTrades, Searchable: true, Orderable: true},
		colPrice(),
		{Data: columnRelativeSize, Name: columnRelativeSizeName, Searchable: true, Orderable: true},
		colCumulativePct(),
		{Data: columnTradeLevelRank, Name: columnLevelRank, Searchable: true, Orderable: true},
		{Data: columnDates, Name: columnLevelDateRange, Searchable: true, Orderable: false},
		{Data: "", Name: "", Searchable: true, Orderable: false},
	}
}
