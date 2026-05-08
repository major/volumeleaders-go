package volumeleaders

import "context"

// Browser endpoint paths for trade cluster APIs captured from VolumeLeaders.
const (
	TradeClustersGetTradeClustersPath         = "/TradeClusters/GetTradeClusters"
	TradeClusterBombsGetTradeClusterBombsPath = "/TradeClusterBombs/GetTradeClusterBombs"
)

// TradeClustersRequest contains DataTables paging and optional endpoint filters
// for /TradeClusters/GetTradeClusters.
type TradeClustersRequest = EndpointRequest

// TradeClusterBombsRequest contains DataTables paging and optional endpoint
// filters for /TradeClusterBombs/GetTradeClusterBombs.
type TradeClusterBombsRequest = EndpointRequest

// clusterBase holds the fields shared by TradeCluster and TradeClusterBomb.
type clusterBase struct {
	Date                    AspNetDate `json:"Date"`
	DateKey                 int        `json:"DateKey"`
	SecurityKey             int        `json:"SecurityKey"`
	Ticker                  string     `json:"Ticker"`
	Sector                  string     `json:"Sector"`
	Industry                string     `json:"Industry"`
	Name                    string     `json:"Name"`
	MinFullDateTime         string     `json:"MinFullDateTime"`
	MaxFullDateTime         string     `json:"MaxFullDateTime"`
	MinFullTimeString24     string     `json:"MinFullTimeString24"`
	MaxFullTimeString24     string     `json:"MaxFullTimeString24"`
	ClosePrice              float64    `json:"ClosePrice"`
	Dollars                 float64    `json:"Dollars"`
	AverageBlockSizeShares  int        `json:"AverageBlockSizeShares"`
	AverageBlockSizeDollars float64    `json:"AverageBlockSizeDollars"`
	Volume                  int        `json:"Volume"`
	TradeCount              int        `json:"TradeCount"`
	IPODate                 AspNetDate `json:"IPODate"`
	DollarsMultiplier       float64    `json:"DollarsMultiplier"`
	CumulativeDistribution  float64    `json:"CumulativeDistribution"`
	AverageDailyVolume      int        `json:"AverageDailyVolume"`
	EOM                     FlexBool   `json:"EOM"`
	EOQ                     FlexBool   `json:"EOQ"`
	EOY                     FlexBool   `json:"EOY"`
	OPEX                    FlexBool   `json:"OPEX"`
	VOLEX                   FlexBool   `json:"VOLEX"`
	InsideBar               FlexBool   `json:"InsideBar"`
	DoubleInsideBar         FlexBool   `json:"DoubleInsideBar"`
	TotalRows               int        `json:"TotalRows"`
	ExternalFeed            FlexBool   `json:"ExternalFeed"`
}

// TradeCluster represents a VolumeLeaders trade cluster row.
type TradeCluster struct {
	clusterBase

	Price                          float64    `json:"Price"`
	LastComparibleTradeClusterDate AspNetDate `json:"LastComparibleTradeClusterDate"`
	TradeClusterRank               int        `json:"TradeClusterRank"`
}

// TradeClusterBomb represents a VolumeLeaders trade cluster bomb row.
type TradeClusterBomb struct {
	clusterBase

	LastComparableTradeClusterBombDate AspNetDate `json:"LastComparableTradeClusterBombDate"`
	TradeClusterBombRank               int        `json:"TradeClusterBombRank"`
}

// GetTradeClusters posts a typed DataTables request to
// /TradeClusters/GetTradeClusters.
func (c *Client) GetTradeClusters(
	ctx context.Context,
	req TradeClustersRequest,
) (*DataTablesResponse[TradeCluster], error) {
	return getEndpoint[TradeCluster](ctx, c, TradeClustersGetTradeClustersPath, req, TradeClustersColumns())
}

// GetTradeClusterBombs posts a typed DataTables request to
// /TradeClusterBombs/GetTradeClusterBombs.
func (c *Client) GetTradeClusterBombs(
	ctx context.Context,
	req TradeClusterBombsRequest,
) (*DataTablesResponse[TradeClusterBomb], error) {
	return getEndpoint[TradeClusterBomb](
		ctx, c, TradeClusterBombsGetTradeClusterBombsPath, req, TradeClusterBombsColumns(),
	)
}

// GetTradeClustersLimit fetches up to limit trade clusters by paging through
// GetTradeClusters. A zero or negative limit fetches all available records.
func (c *Client) GetTradeClustersLimit(
	ctx context.Context,
	req TradeClustersRequest,
	limit int,
) ([]TradeCluster, error) {
	return getEndpointLimit(ctx, req, limit, c.GetTradeClusters)
}

// GetTradeClusterBombsLimit fetches up to limit trade cluster bombs by paging
// through GetTradeClusterBombs. A zero or negative limit fetches all available
// records.
func (c *Client) GetTradeClusterBombsLimit(
	ctx context.Context,
	req TradeClusterBombsRequest,
	limit int,
) ([]TradeClusterBomb, error) {
	return getEndpointLimit(ctx, req, limit, c.GetTradeClusterBombs)
}

// TradeClustersColumns returns the DataTables columns captured from the trade
// clusters table.
func TradeClustersColumns() []DataTablesColumn {
	return []DataTablesColumn{
		{Data: columnMinFullTimeString24, Name: "", Searchable: true, Orderable: false},
		{
			Data:       columnMinFullTimeString24,
			Name:       columnMinFullTimeString24,
			Searchable: true,
			Orderable:  true,
		},
		{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true},
		{Data: columnTradeCount, Name: columnTrades, Searchable: true, Orderable: true},
		{Data: columnCurrent, Name: columnCurrent, Searchable: true, Orderable: false},
		{Data: "Cluster", Name: "Cluster", Searchable: true, Orderable: false},
		{Data: columnSector, Name: columnSector, Searchable: true, Orderable: true},
		{Data: columnIndustry, Name: columnIndustry, Searchable: true, Orderable: true},
		{Data: tradeColumnVolume, Name: columnShName, Searchable: true, Orderable: true},
		{Data: columnDollars, Name: columnDollarsName, Searchable: true, Orderable: true},
		{
			Data:       columnDollarsMultiplier,
			Name:       columnRelativeSizeName,
			Searchable: true,
			Orderable:  true,
		},
		{
			Data:       columnCumulativeDistribution,
			Name:       columnPercentName,
			Searchable: true,
			Orderable:  true,
		},
		{Data: columnTradeClusterRank, Name: columnRankName, Searchable: true, Orderable: true},
		{
			Data:       columnLastComparibleTradeClusterDate,
			Name:       columnLastDateName,
			Searchable: true,
			Orderable:  true,
		},
		{
			Data:       columnLastComparibleTradeClusterDate,
			Name:       columnLastDateName,
			Searchable: true,
			Orderable:  false,
		},
	}
}

// TradeClusterBombsColumns returns the DataTables columns captured from the
// trade cluster bombs table.
func TradeClusterBombsColumns() []DataTablesColumn {
	return []DataTablesColumn{
		{
			Data:       columnMinFullTimeString24,
			Name:       columnMinFullTimeString24,
			Searchable: true,
			Orderable:  true,
		},
		{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true},
		{Data: columnTradeCount, Name: columnTrades, Searchable: true, Orderable: true},
		{Data: columnSector, Name: columnSector, Searchable: true, Orderable: true},
		{Data: columnIndustry, Name: columnIndustry, Searchable: true, Orderable: true},
		{Data: tradeColumnVolume, Name: columnShName, Searchable: true, Orderable: true},
		{Data: columnDollars, Name: columnDollarsName, Searchable: true, Orderable: true},
		{
			Data:       columnDollarsMultiplier,
			Name:       columnRelativeSizeName,
			Searchable: true,
			Orderable:  true,
		},
		{
			Data:       columnCumulativeDistribution,
			Name:       columnPercentName,
			Searchable: true,
			Orderable:  true,
		},
		{Data: columnTradeClusterBombRank, Name: columnRankName, Searchable: true, Orderable: true},
		{
			Data:       columnLastComparableTradeClusterBombDate,
			Name:       columnLastDateName,
			Searchable: true,
			Orderable:  true,
		},
		{
			Data:       columnLastComparableTradeClusterBombDate,
			Name:       columnLastDateName,
			Searchable: true,
			Orderable:  false,
		},
	}
}
