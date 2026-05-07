package volumeleaders

import (
	"context"
	"net/url"
)

// Browser endpoint paths for volume leaderboard APIs used by VolumeLeaders.
const (
	InstitutionalVolumeGetInstitutionalVolumePath     = "/InstitutionalVolume/GetInstitutionalVolume"
	AHInstitutionalVolumeGetAHInstitutionalVolumePath = "/AHInstitutionalVolume/GetAHInstitutionalVolume"
	TotalVolumeGetTotalVolumePath                     = "/TotalVolume/GetTotalVolume"
)

// VolumeRequest contains DataTables paging and optional endpoint filters for
// volume leaderboard endpoints.
type VolumeRequest struct {
	DataTables DataTablesRequest
	Filters    url.Values
}

// GetInstitutionalVolume posts a typed DataTables request to
// /InstitutionalVolume/GetInstitutionalVolume.
func (c *Client) GetInstitutionalVolume(
	ctx context.Context,
	req VolumeRequest,
) (*DataTablesResponse[Trade], error) {
	return c.getVolume(ctx, InstitutionalVolumeGetInstitutionalVolumePath, req, InstitutionalVolumeColumns())
}

// GetAHInstitutionalVolume posts a typed DataTables request to
// /AHInstitutionalVolume/GetAHInstitutionalVolume.
func (c *Client) GetAHInstitutionalVolume(
	ctx context.Context,
	req VolumeRequest,
) (*DataTablesResponse[Trade], error) {
	return c.getVolume(ctx, AHInstitutionalVolumeGetAHInstitutionalVolumePath, req, AHInstitutionalVolumeColumns())
}

// GetTotalVolume posts a typed DataTables request to
// /TotalVolume/GetTotalVolume.
func (c *Client) GetTotalVolume(ctx context.Context, req VolumeRequest) (*DataTablesResponse[Trade], error) {
	return c.getVolume(ctx, TotalVolumeGetTotalVolumePath, req, TotalVolumeColumns())
}

// InstitutionalVolumeColumns returns the DataTables columns used by the
// institutional volume leaderboard endpoints.
func InstitutionalVolumeColumns() []DataTablesColumn {
	return []DataTablesColumn{
		{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true},
		{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true},
		{Data: columnPrice, Name: columnPrice, Searchable: true, Orderable: true},
		{Data: columnSector, Name: columnSector, Searchable: true, Orderable: true},
		{Data: columnIndustry, Name: columnIndustry, Searchable: true, Orderable: true},
		{Data: "TotalInstitutionalVolume", Name: "TotalInstitutionalVolume", Searchable: true, Orderable: true},
		{Data: "TotalInstitutionalDollars", Name: "TotalInstitutionalDollars", Searchable: true, Orderable: true},
		{Data: "TotalInstitutionalDollarsRank", Name: "TotalInstitutionalDollarsRank", Searchable: true, Orderable: true},
		{Data: tradeColumnLastTradeDate, Name: tradeColumnLastTradeDate, Searchable: true, Orderable: true},
		{Data: tradeColumnLastTradeDate, Name: tradeColumnLastTradeDate, Searchable: true, Orderable: true},
	}
}

// AHInstitutionalVolumeColumns returns the DataTables columns used by the
// after-hours institutional volume leaderboard endpoint.
func AHInstitutionalVolumeColumns() []DataTablesColumn {
	return []DataTablesColumn{
		{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true},
		{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true},
		{Data: columnPrice, Name: columnPrice, Searchable: true, Orderable: true},
		{Data: columnSector, Name: columnSector, Searchable: true, Orderable: true},
		{Data: columnIndustry, Name: columnIndustry, Searchable: true, Orderable: true},
		{Data: "AHInstitutionalVolume", Name: "AHInstitutionalVolume", Searchable: true, Orderable: true},
		{Data: "AHInstitutionalDollars", Name: "AHInstitutionalDollars", Searchable: true, Orderable: true},
		{Data: "AHInstitutionalDollarsRank", Name: "AHInstitutionalDollarsRank", Searchable: true, Orderable: true},
		{Data: tradeColumnLastTradeDate, Name: tradeColumnLastTradeDate, Searchable: true, Orderable: true},
		{Data: tradeColumnLastTradeDate, Name: tradeColumnLastTradeDate, Searchable: true, Orderable: true},
	}
}

// TotalVolumeColumns returns the DataTables columns used by the total volume
// leaderboard endpoint.
func TotalVolumeColumns() []DataTablesColumn {
	return []DataTablesColumn{
		{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true},
		{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true},
		{Data: columnPrice, Name: columnPrice, Searchable: true, Orderable: true},
		{Data: columnSector, Name: columnSector, Searchable: true, Orderable: true},
		{Data: columnIndustry, Name: columnIndustry, Searchable: true, Orderable: true},
		{Data: "TotalVolume", Name: "TotalVolume", Searchable: true, Orderable: true},
		{Data: "TotalDollars", Name: "TotalDollars", Searchable: true, Orderable: true},
		{Data: "TotalDollarsRank", Name: "TotalDollarsRank", Searchable: true, Orderable: true},
		{Data: tradeColumnLastTradeDate, Name: tradeColumnLastTradeDate, Searchable: true, Orderable: true},
		{Data: tradeColumnLastTradeDate, Name: tradeColumnLastTradeDate, Searchable: true, Orderable: true},
	}
}

func (c *Client) getVolume(
	ctx context.Context,
	path string,
	req VolumeRequest,
	columns []DataTablesColumn,
) (*DataTablesResponse[Trade], error) {
	var result DataTablesResponse[Trade]
	if err := c.postDataTables(ctx, path, req.DataTables, req.Filters, columns, &result); err != nil {
		return nil, err
	}
	return &result, nil
}
