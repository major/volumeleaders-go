package volumeleaders

import (
	"context"
	"net/url"
)

// EarningsGetEarningsPath is the browser endpoint path for earnings rows.
const EarningsGetEarningsPath = "/Earnings/GetEarnings"

// EarningsRequest contains DataTables paging and optional endpoint filters for
// /Earnings/GetEarnings.
type EarningsRequest struct {
	DataTables DataTablesRequest
	Filters    url.Values
}

// Earning represents a VolumeLeaders earnings row.
//
// The available earnings captures included the request columns but no
// successful response body, so this model follows the captured DataTables
// column names until a successful earnings payload can confirm extra fields.
type Earning struct {
	Date                  AspNetDate `json:"Date"`
	Ticker                string     `json:"Ticker"`
	Current               float64    `json:"Current"`
	Sector                string     `json:"Sector"`
	Industry              string     `json:"Industry"`
	TradeCount            int        `json:"TradeCount"`
	TradeClusterCount     int        `json:"TradeClusterCount"`
	TradeClusterBombCount int        `json:"TradeClusterBombCount"`
	TotalRows             int        `json:"TotalRows"`
}

// GetEarnings posts a typed DataTables request to /Earnings/GetEarnings.
func (c *Client) GetEarnings(ctx context.Context, req EarningsRequest) (*DataTablesResponse[Earning], error) {
	var result DataTablesResponse[Earning]
	if err := c.postDataTables(
		ctx,
		EarningsGetEarningsPath,
		req.DataTables,
		req.Filters,
		EarningsColumns(),
		&result,
	); err != nil {
		return nil, err
	}
	return &result, nil
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
		{Data: "TradeClusterCount", Name: "Recent Top-100 Clusters", Searchable: true, Orderable: true},
		{Data: "TradeClusterBombCount", Name: "Recent Top-100 Bombs", Searchable: true, Orderable: true},
		{Data: columnTicker, Name: columnCharts, Searchable: true, Orderable: false},
	}
}
