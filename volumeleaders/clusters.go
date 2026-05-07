package volumeleaders

import (
	"context"
	"net/url"
)

// Browser endpoint paths for trade cluster APIs captured from VolumeLeaders.
const (
	TradeClustersGetTradeClustersPath         = "/TradeClusters/GetTradeClusters"
	TradeClusterBombsGetTradeClusterBombsPath = "/TradeClusterBombs/GetTradeClusterBombs"
)

// TradeClustersRequest contains DataTables paging and optional endpoint filters
// for /TradeClusters/GetTradeClusters.
type TradeClustersRequest struct {
	DataTables DataTablesRequest
	Filters    url.Values
}

// TradeClusterBombsRequest contains DataTables paging and optional endpoint
// filters for /TradeClusterBombs/GetTradeClusterBombs.
type TradeClusterBombsRequest struct {
	DataTables DataTablesRequest
	Filters    url.Values
}

// TradeCluster represents a VolumeLeaders trade cluster row.
type TradeCluster struct {
	Date                           AspNetDate `json:"Date"`
	DateKey                        int        `json:"DateKey"`
	SecurityKey                    int        `json:"SecurityKey"`
	Ticker                         string     `json:"Ticker"`
	Sector                         string     `json:"Sector"`
	Industry                       string     `json:"Industry"`
	Name                           string     `json:"Name"`
	MinFullDateTime                string     `json:"MinFullDateTime"`
	MaxFullDateTime                string     `json:"MaxFullDateTime"`
	MinFullTimeString24            string     `json:"MinFullTimeString24"`
	MaxFullTimeString24            string     `json:"MaxFullTimeString24"`
	Price                          float64    `json:"Price"`
	ClosePrice                     float64    `json:"ClosePrice"`
	Dollars                        float64    `json:"Dollars"`
	AverageBlockSizeShares         int        `json:"AverageBlockSizeShares"`
	AverageBlockSizeDollars        float64    `json:"AverageBlockSizeDollars"`
	Volume                         int        `json:"Volume"`
	TradeCount                     int        `json:"TradeCount"`
	LastComparibleTradeClusterDate AspNetDate `json:"LastComparibleTradeClusterDate"`
	IPODate                        AspNetDate `json:"IPODate"`
	DollarsMultiplier              float64    `json:"DollarsMultiplier"`
	CumulativeDistribution         float64    `json:"CumulativeDistribution"`
	TradeClusterRank               int        `json:"TradeClusterRank"`
	AverageDailyVolume             int        `json:"AverageDailyVolume"`
	EOM                            FlexBool   `json:"EOM"`
	EOQ                            FlexBool   `json:"EOQ"`
	EOY                            FlexBool   `json:"EOY"`
	OPEX                           FlexBool   `json:"OPEX"`
	VOLEX                          FlexBool   `json:"VOLEX"`
	InsideBar                      FlexBool   `json:"InsideBar"`
	DoubleInsideBar                FlexBool   `json:"DoubleInsideBar"`
	TotalRows                      int        `json:"TotalRows"`
	ExternalFeed                   FlexBool   `json:"ExternalFeed"`
}

// TradeClusterBomb represents a VolumeLeaders trade cluster bomb row.
type TradeClusterBomb struct {
	Date                               AspNetDate `json:"Date"`
	DateKey                            int        `json:"DateKey"`
	SecurityKey                        int        `json:"SecurityKey"`
	Ticker                             string     `json:"Ticker"`
	Sector                             string     `json:"Sector"`
	Industry                           string     `json:"Industry"`
	Name                               string     `json:"Name"`
	MinFullDateTime                    string     `json:"MinFullDateTime"`
	MaxFullDateTime                    string     `json:"MaxFullDateTime"`
	MinFullTimeString24                string     `json:"MinFullTimeString24"`
	MaxFullTimeString24                string     `json:"MaxFullTimeString24"`
	ClosePrice                         float64    `json:"ClosePrice"`
	Dollars                            float64    `json:"Dollars"`
	AverageBlockSizeShares             int        `json:"AverageBlockSizeShares"`
	AverageBlockSizeDollars            float64    `json:"AverageBlockSizeDollars"`
	Volume                             int        `json:"Volume"`
	TradeCount                         int        `json:"TradeCount"`
	LastComparableTradeClusterBombDate AspNetDate `json:"LastComparableTradeClusterBombDate"`
	IPODate                            AspNetDate `json:"IPODate"`
	DollarsMultiplier                  float64    `json:"DollarsMultiplier"`
	CumulativeDistribution             float64    `json:"CumulativeDistribution"`
	TradeClusterBombRank               int        `json:"TradeClusterBombRank"`
	AverageDailyVolume                 int        `json:"AverageDailyVolume"`
	EOM                                FlexBool   `json:"EOM"`
	EOQ                                FlexBool   `json:"EOQ"`
	EOY                                FlexBool   `json:"EOY"`
	OPEX                               FlexBool   `json:"OPEX"`
	VOLEX                              FlexBool   `json:"VOLEX"`
	InsideBar                          FlexBool   `json:"InsideBar"`
	DoubleInsideBar                    FlexBool   `json:"DoubleInsideBar"`
	TotalRows                          int        `json:"TotalRows"`
	ExternalFeed                       FlexBool   `json:"ExternalFeed"`
}

// GetTradeClusters posts a typed DataTables request to
// /TradeClusters/GetTradeClusters.
func (c *Client) GetTradeClusters(
	ctx context.Context,
	req TradeClustersRequest,
) (*DataTablesResponse[TradeCluster], error) {
	var result DataTablesResponse[TradeCluster]
	if err := c.postDataTables(
		ctx,
		TradeClustersGetTradeClustersPath,
		req.DataTables,
		req.Filters,
		TradeClustersColumns(),
		&result,
	); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetTradeClusterBombs posts a typed DataTables request to
// /TradeClusterBombs/GetTradeClusterBombs.
func (c *Client) GetTradeClusterBombs(
	ctx context.Context,
	req TradeClusterBombsRequest,
) (*DataTablesResponse[TradeClusterBomb], error) {
	var result DataTablesResponse[TradeClusterBomb]
	if err := c.postDataTables(
		ctx,
		TradeClusterBombsGetTradeClusterBombsPath,
		req.DataTables,
		req.Filters,
		TradeClusterBombsColumns(),
		&result,
	); err != nil {
		return nil, err
	}
	return &result, nil
}

// TradeClustersColumns returns the DataTables columns captured from the trade
// clusters table.
func TradeClustersColumns() []DataTablesColumn {
	return []DataTablesColumn{
		{Data: columnMinFullTimeString24, Name: "", Searchable: true, Orderable: false},
		{Data: columnMinFullTimeString24, Name: columnMinFullTimeString24, Searchable: true, Orderable: true},
		{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true},
		{Data: columnTradeCount, Name: columnTrades, Searchable: true, Orderable: true},
		{Data: columnCurrent, Name: columnCurrent, Searchable: true, Orderable: false},
		{Data: "Cluster", Name: "Cluster", Searchable: true, Orderable: false},
		{Data: columnSector, Name: columnSector, Searchable: true, Orderable: true},
		{Data: columnIndustry, Name: columnIndustry, Searchable: true, Orderable: true},
		{Data: tradeColumnVolume, Name: columnShName, Searchable: true, Orderable: true},
		{Data: columnDollars, Name: columnDollarsName, Searchable: true, Orderable: true},
		{Data: columnDollarsMultiplier, Name: columnRelativeSizeName, Searchable: true, Orderable: true},
		{Data: columnCumulativeDistribution, Name: columnPercentName, Searchable: true, Orderable: true},
		{Data: columnTradeClusterRank, Name: columnRankName, Searchable: true, Orderable: true},
		{Data: columnLastComparibleTradeClusterDate, Name: columnLastDateName, Searchable: true, Orderable: true},
		{Data: columnLastComparibleTradeClusterDate, Name: columnLastDateName, Searchable: true, Orderable: false},
	}
}

// TradeClusterBombsColumns returns the DataTables columns captured from the
// trade cluster bombs table.
func TradeClusterBombsColumns() []DataTablesColumn {
	return []DataTablesColumn{
		{Data: columnMinFullTimeString24, Name: columnMinFullTimeString24, Searchable: true, Orderable: true},
		{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true},
		{Data: columnTradeCount, Name: columnTrades, Searchable: true, Orderable: true},
		{Data: columnSector, Name: columnSector, Searchable: true, Orderable: true},
		{Data: columnIndustry, Name: columnIndustry, Searchable: true, Orderable: true},
		{Data: tradeColumnVolume, Name: columnShName, Searchable: true, Orderable: true},
		{Data: columnDollars, Name: columnDollarsName, Searchable: true, Orderable: true},
		{Data: columnDollarsMultiplier, Name: columnRelativeSizeName, Searchable: true, Orderable: true},
		{Data: columnCumulativeDistribution, Name: columnPercentName, Searchable: true, Orderable: true},
		{Data: columnTradeClusterBombRank, Name: columnRankName, Searchable: true, Orderable: true},
		{Data: columnLastComparableTradeClusterBombDate, Name: columnLastDateName, Searchable: true, Orderable: true},
		{Data: columnLastComparableTradeClusterBombDate, Name: columnLastDateName, Searchable: true, Orderable: false},
	}
}
