package volumeleaders

import "context"

// EarningsGetEarningsPath is the browser endpoint path for earnings rows.
const EarningsGetEarningsPath = "/Earnings/GetEarnings"

// EarningsRequest contains DataTables paging and optional endpoint filters for
// /Earnings/GetEarnings.
type EarningsRequest = EndpointRequest

// Earning represents a VolumeLeaders earnings row.
type Earning struct {
	Date                  AspNetDate `json:"Date"`
	EarningsDate          AspNetDate `json:"EarningsDate"`
	Name                  string     `json:"Name"`
	Ticker                string     `json:"Ticker"`
	Current               float64    `json:"Current"`
	Sector                *string    `json:"Sector"`
	Industry              *string    `json:"Industry"`
	AfterMarketClose      bool       `json:"AfterMarketClose"`
	TradeCount            int        `json:"TradeCount"`
	TradeClusterCount     int        `json:"TradeClusterCount"`
	TradeClusterBombCount int        `json:"TradeClusterBombCount"`
	TotalRows             int        `json:"TotalRows"`
}

// GetEarnings posts a typed DataTables request to /Earnings/GetEarnings.
func (c *Client) GetEarnings(
	ctx context.Context,
	req EarningsRequest,
) (*DataTablesResponse[Earning], error) {
	return getEndpoint[Earning](ctx, c, EarningsGetEarningsPath, req, EarningsColumns())
}

// GetEarningsLimit fetches up to limit earnings by paging through GetEarnings.
// A zero or negative limit fetches all available records.
func (c *Client) GetEarningsLimit(
	ctx context.Context,
	req EarningsRequest,
	limit int,
) ([]Earning, error) {
	return getEndpointLimit(ctx, req, limit, c.GetEarnings)
}

// EarningsColumns returns the DataTables columns captured from the earnings
// table.
func EarningsColumns() []DataTablesColumn {
	return []DataTablesColumn{
		{Data: columnDate, Name: "Earnings Date", Searchable: true, Orderable: true},
		{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true},
		{Data: columnCurrent, Name: columnCurrent, Searchable: true, Orderable: false},
		{Data: columnSector, Name: columnSector, Searchable: true, Orderable: true},
		{Data: columnIndustry, Name: columnIndustry, Searchable: true, Orderable: true},
		{Data: columnTradeCount, Name: "Recent Top-100 Trades", Searchable: true, Orderable: true},
		{
			Data:       "TradeClusterCount",
			Name:       "Recent Top-100 Clusters",
			Searchable: true,
			Orderable:  true,
		},
		{
			Data:       "TradeClusterBombCount",
			Name:       "Recent Top-100 Bombs",
			Searchable: true,
			Orderable:  true,
		},
		{Data: columnTicker, Name: columnCharts, Searchable: true, Orderable: false},
	}
}
