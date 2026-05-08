package volumeleaders

import "context"

// Browser endpoint paths for volume leaderboard APIs used by VolumeLeaders.
const (
	InstitutionalVolumeGetInstitutionalVolumePath     = "/InstitutionalVolume/GetInstitutionalVolume"
	AHInstitutionalVolumeGetAHInstitutionalVolumePath = "/AHInstitutionalVolume/GetAHInstitutionalVolume"
	TotalVolumeGetTotalVolumePath                     = "/TotalVolume/GetTotalVolume"
)

// VolumeRequest contains DataTables paging and optional endpoint filters for
// volume leaderboard endpoints.
type VolumeRequest = EndpointRequest

// GetInstitutionalVolume posts a typed DataTables request to
// /InstitutionalVolume/GetInstitutionalVolume.
func (c *Client) GetInstitutionalVolume(
	ctx context.Context,
	req VolumeRequest,
) (*DataTablesResponse[Trade], error) {
	return getEndpoint[Trade](c, ctx, InstitutionalVolumeGetInstitutionalVolumePath, req, InstitutionalVolumeColumns())
}

// GetAHInstitutionalVolume posts a typed DataTables request to
// /AHInstitutionalVolume/GetAHInstitutionalVolume.
func (c *Client) GetAHInstitutionalVolume(
	ctx context.Context,
	req VolumeRequest,
) (*DataTablesResponse[Trade], error) {
	return getEndpoint[Trade](c, ctx, AHInstitutionalVolumeGetAHInstitutionalVolumePath, req, AHInstitutionalVolumeColumns())
}

// GetTotalVolume posts a typed DataTables request to
// /TotalVolume/GetTotalVolume.
func (c *Client) GetTotalVolume(
	ctx context.Context,
	req VolumeRequest,
) (*DataTablesResponse[Trade], error) {
	return getEndpoint[Trade](c, ctx, TotalVolumeGetTotalVolumePath, req, TotalVolumeColumns())
}

// GetInstitutionalVolumeLimit fetches up to limit trades by paging through
// GetInstitutionalVolume. A zero or negative limit fetches all available
// records.
func (c *Client) GetInstitutionalVolumeLimit(
	ctx context.Context,
	req VolumeRequest,
	limit int,
) ([]Trade, error) {
	return getEndpointLimit(ctx, req, limit, c.GetInstitutionalVolume)
}

// GetAHInstitutionalVolumeLimit fetches up to limit trades by paging through
// GetAHInstitutionalVolume. A zero or negative limit fetches all available
// records.
func (c *Client) GetAHInstitutionalVolumeLimit(
	ctx context.Context,
	req VolumeRequest,
	limit int,
) ([]Trade, error) {
	return getEndpointLimit(ctx, req, limit, c.GetAHInstitutionalVolume)
}

// GetTotalVolumeLimit fetches up to limit trades by paging through
// GetTotalVolume. A zero or negative limit fetches all available records.
func (c *Client) GetTotalVolumeLimit(
	ctx context.Context,
	req VolumeRequest,
	limit int,
) ([]Trade, error) {
	return getEndpointLimit(ctx, req, limit, c.GetTotalVolume)
}

// volumeColumns builds a column set for volume leaderboard endpoints. All three
// share the same leading (Ticker, Price, Sector, Industry) and trailing
// (LastTradeDate) columns; only the three middle metric columns differ.
func volumeColumns(volume, dollars, rank string) []DataTablesColumn {
	return []DataTablesColumn{
		{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true},
		{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true},
		{Data: columnPrice, Name: columnPrice, Searchable: true, Orderable: true},
		{Data: columnSector, Name: columnSector, Searchable: true, Orderable: true},
		{Data: columnIndustry, Name: columnIndustry, Searchable: true, Orderable: true},
		{Data: volume, Name: volume, Searchable: true, Orderable: true},
		{Data: dollars, Name: dollars, Searchable: true, Orderable: true},
		{Data: rank, Name: rank, Searchable: true, Orderable: true},
		{Data: tradeColumnLastTradeDate, Name: tradeColumnLastTradeDate, Searchable: true, Orderable: true},
		{Data: tradeColumnLastTradeDate, Name: tradeColumnLastTradeDate, Searchable: true, Orderable: true},
	}
}

// InstitutionalVolumeColumns returns the DataTables columns used by the
// institutional volume leaderboard endpoints.
func InstitutionalVolumeColumns() []DataTablesColumn {
	return volumeColumns("TotalInstitutionalVolume", "TotalInstitutionalDollars", columnTotalInstitutionalDollarsRank)
}

// AHInstitutionalVolumeColumns returns the DataTables columns used by the
// after-hours institutional volume leaderboard endpoint.
func AHInstitutionalVolumeColumns() []DataTablesColumn {
	return volumeColumns("AHInstitutionalVolume", "AHInstitutionalDollars", "AHInstitutionalDollarsRank")
}

// TotalVolumeColumns returns the DataTables columns used by the total volume
// leaderboard endpoint.
func TotalVolumeColumns() []DataTablesColumn {
	return volumeColumns("TotalVolume", "TotalDollars", "TotalDollarsRank")
}

