package volumeleaders

import (
	"context"
	"net/url"
	"strconv"
)

// Browser endpoint paths for watchlist APIs captured from VolumeLeaders.
const (
	WatchListConfigPath                 = "/WatchListConfig"
	WatchListConfigsGetWatchListsPath   = "/WatchListConfigs/GetWatchLists"
	WatchListsGetWatchListTickersPath   = "/WatchLists/GetWatchListTickers"
	WatchListConfigsDeleteWatchListPath = "/WatchListConfigs/DeleteWatchList"
	Chart0UpdateWatchListPath           = "/Chart0/UpdateWatchList"
)

// WatchListsRequest contains DataTables paging and optional endpoint filters
// for /WatchListConfigs/GetWatchLists.
type WatchListsRequest = EndpointRequest

// WatchListTickersRequest contains DataTables paging and optional endpoint
// filters for /WatchLists/GetWatchListTickers.
type WatchListTickersRequest = EndpointRequest

// SaveWatchListConfigRequest contains the multipart form fields submitted to
// /WatchListConfig when creating or editing a watchlist.
type SaveWatchListConfigRequest struct {
	// Fields is a low-level captured-form escape hatch. Callers must provide the
	// exact browser field names and values accepted by VolumeLeaders.
	Fields url.Values
}

// AddTickerToWatchListRequest contains the form payload for adding a ticker to
// an existing watchlist from the chart page.
type AddTickerToWatchListRequest struct {
	WatchListKey int
	Ticker       string
}

// AddTickerToWatchListResponse is the JSON envelope returned by
// /Chart0/UpdateWatchList.
type AddTickerToWatchListResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// DeleteWatchListRequest contains the JSON payload for deleting a watchlist.
type DeleteWatchListRequest struct {
	WatchListKey int `json:"WatchListKey"`
}

// WatchListConfig represents a VolumeLeaders saved watchlist configuration.
type WatchListConfig struct {
	SearchTemplateKey           int     `json:"SearchTemplateKey"`
	UserKey                     int     `json:"UserKey"`
	SearchTemplateTypeKey       int     `json:"SearchTemplateTypeKey"`
	Name                        string  `json:"Name"`
	Tickers                     string  `json:"Tickers"`
	SortOrder                   *int    `json:"SortOrder"`
	MinVolume                   int     `json:"MinVolume"`
	MaxVolume                   int     `json:"MaxVolume"`
	MinDollars                  float64 `json:"MinDollars"`
	MaxDollars                  float64 `json:"MaxDollars"`
	MinPrice                    float64 `json:"MinPrice"`
	MaxPrice                    float64 `json:"MaxPrice"`
	RSIOverboughtHourly         *int    `json:"RSIOverboughtHourly"`
	RSIOverboughtDaily          *int    `json:"RSIOverboughtDaily"`
	RSIOversoldHourly           *int    `json:"RSIOversoldHourly"`
	RSIOversoldDaily            *int    `json:"RSIOversoldDaily"`
	RSIOverboughtHourlySelected *bool   `json:"RSIOverboughtHourlySelected"`
	RSIOverboughtDailySelected  *bool   `json:"RSIOverboughtDailySelected"`
	RSIOversoldHourlySelected   *bool   `json:"RSIOversoldHourlySelected"`
	RSIOversoldDailySelected    *bool   `json:"RSIOversoldDailySelected"`
	Conditions                  string  `json:"Conditions"`
	MinRelativeSize             int     `json:"MinRelativeSize"`
	MinRelativeSizeSelected     *bool   `json:"MinRelativeSizeSelected"`
	MaxTradeRank                int     `json:"MaxTradeRank"`
	SecurityTypeKey             int     `json:"SecurityTypeKey"`
	SecurityType                *string `json:"SecurityType"`
	MaxTradeRankSelected        *bool   `json:"MaxTradeRankSelected"`
	MinVCD                      float64 `json:"MinVCD"`
	NormalPrints                bool    `json:"NormalPrints"`
	NormalPrintsSelected        bool    `json:"NormalPrintsSelected"`
	SignaturePrints             bool    `json:"SignaturePrints"`
	SignaturePrintsSelected     bool    `json:"SignaturePrintsSelected"`
	LatePrints                  bool    `json:"LatePrints"`
	LatePrintsSelected          bool    `json:"LatePrintsSelected"`
	TimelyPrints                bool    `json:"TimelyPrints"`
	TimelyPrintsSelected        bool    `json:"TimelyPrintsSelected"`
	DarkPools                   bool    `json:"DarkPools"`
	DarkPoolsSelected           bool    `json:"DarkPoolsSelected"`
	LitExchanges                bool    `json:"LitExchanges"`
	LitExchangesSelected        bool    `json:"LitExchangesSelected"`
	Sweeps                      bool    `json:"Sweeps"`
	SweepsSelected              bool    `json:"SweepsSelected"`
	Blocks                      bool    `json:"Blocks"`
	BlocksSelected              bool    `json:"BlocksSelected"`
	PremarketTrades             bool    `json:"PremarketTrades"`
	PremarketTradesSelected     bool    `json:"PremarketTradesSelected"`
	RTHTrades                   bool    `json:"RTHTrades"`
	RTHTradesSelected           bool    `json:"RTHTradesSelected"`
	AHTrades                    bool    `json:"AHTrades"`
	AHTradesSelected            bool    `json:"AHTradesSelected"`
	OpeningTrades               bool    `json:"OpeningTrades"`
	OpeningTradesSelected       bool    `json:"OpeningTradesSelected"`
	ClosingTrades               bool    `json:"ClosingTrades"`
	ClosingTradesSelected       bool    `json:"ClosingTradesSelected"`
	PhantomTrades               bool    `json:"PhantomTrades"`
	PhantomTradesSelected       bool    `json:"PhantomTradesSelected"`
	OffsettingTrades            bool    `json:"OffsettingTrades"`
	OffsettingTradesSelected    bool    `json:"OffsettingTradesSelected"`
	SectorIndustry              *string `json:"SectorIndustry"`
	APIKey                      *string `json:"APIKey"`
}

// WatchListTicker represents a ticker row inside a VolumeLeaders watchlist.
type WatchListTicker struct {
	WatchListKey                    int        `json:"WatchListKey"`
	SecurityKey                     int        `json:"SecurityKey"`
	Ticker                          string     `json:"Ticker"`
	Sector                          *string    `json:"Sector"`
	Industry                        *string    `json:"Industry"`
	MarketCap                       *float64   `json:"MarketCap"`
	Price                           *float64   `json:"Price"`
	NearestTop10TradeDate           AspNetDate `json:"NearestTop10TradeDate"`
	NearestTop10TradePrice          float64    `json:"NearestTop10TradePrice"`
	NearestTop10TradeVolume         int        `json:"NearestTop10TradeVolume"`
	NearestTop10TradeDollars        float64    `json:"NearestTop10TradeDollars"`
	NearestTop10TradeRank           int        `json:"NearestTop10TradeRank"`
	NearestTop10TradeClusterDate    AspNetDate `json:"NearestTop10TradeClusterDate"`
	NearestTop10TradeClusterPrice   float64    `json:"NearestTop10TradeClusterPrice"`
	NearestTop10TradeClusterVolume  int        `json:"NearestTop10TradeClusterVolume"`
	NearestTop10TradeClusterDollars float64    `json:"NearestTop10TradeClusterDollars"`
	NearestTop10TradeClusterRank    int        `json:"NearestTop10TradeClusterRank"`
	NearestTop10TradeLevel          *float64   `json:"NearestTop10TradeLevel"`
	NearestTop10TradeLevelDate      AspNetDate `json:"NearestTop10TradeLevelDate"`
	NearestTop10TradeLevelPrice     float64    `json:"NearestTop10TradeLevelPrice"`
	NearestTop10TradeLevelVolume    int        `json:"NearestTop10TradeLevelVolume"`
	NearestTop10TradeLevelDollars   float64    `json:"NearestTop10TradeLevelDollars"`
	NearestTop10TradeLevelRank      int        `json:"NearestTop10TradeLevelRank"`
}

// GetWatchLists posts a typed DataTables request to
// /WatchListConfigs/GetWatchLists.
func (c *Client) GetWatchLists(
	ctx context.Context,
	req WatchListsRequest,
) (*DataTablesResponse[WatchListConfig], error) {
	return getEndpoint[WatchListConfig](c, ctx, WatchListConfigsGetWatchListsPath, req, WatchListsColumns())
}

// GetWatchListTickers posts a typed DataTables request to
// /WatchLists/GetWatchListTickers.
func (c *Client) GetWatchListTickers(
	ctx context.Context,
	req WatchListTickersRequest,
) (*DataTablesResponse[WatchListTicker], error) {
	return getEndpoint[WatchListTicker](c, ctx, WatchListsGetWatchListTickersPath, req, WatchListTickersColumns())
}

// SaveWatchListConfig posts a multipart create or edit request to
// /WatchListConfig.
func (c *Client) SaveWatchListConfig(ctx context.Context, req SaveWatchListConfigRequest) error {
	return c.postMultipartForm(ctx, WatchListConfigPath, mergeValues(req.Fields, nil), nil)
}

// AddTickerToWatchList posts a chart-page watchlist update request that adds a
// ticker to an existing watchlist.
func (c *Client) AddTickerToWatchList(
	ctx context.Context,
	req AddTickerToWatchListRequest,
) (*AddTickerToWatchListResponse, error) {
	values := url.Values{}
	values.Set("WatchListKey", strconv.Itoa(req.WatchListKey))
	values.Set("Ticker", req.Ticker)

	var result AddTickerToWatchListResponse
	if err := c.postForm(ctx, Chart0UpdateWatchListPath, values.Encode(), &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// DeleteWatchList posts a JSON delete request to
// /WatchListConfigs/DeleteWatchList.
func (c *Client) DeleteWatchList(ctx context.Context, req DeleteWatchListRequest) error {
	return c.postJSON(ctx, WatchListConfigsDeleteWatchListPath, req, nil)
}

// WatchListsColumns returns the DataTables columns captured from the watchlists
// table.
func WatchListsColumns() []DataTablesColumn {
	return []DataTablesColumn{
		{Data: columnName, Name: "", Searchable: true, Orderable: false},
		{Data: columnName, Name: columnName, Searchable: true, Orderable: true},
		{Data: columnTickers, Name: columnTickers, Searchable: true, Orderable: false},
		{Data: "Criteria", Name: "Criteria", Searchable: true, Orderable: false},
	}
}

// WatchListTickersColumns returns the DataTables columns captured from the
// watchlist tickers table.
func WatchListTickersColumns() []DataTablesColumn {
	return []DataTablesColumn{
		{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true},
		{Data: columnPrice, Name: columnPrice, Searchable: true, Orderable: true},
		{Data: "NearestTop10TradeDate", Name: "NearestTop10TradeDate", Searchable: true, Orderable: true},
		{Data: "NearestTop10TradeClusterDate", Name: "NearestTop10TradeClusterDate", Searchable: true, Orderable: true},
		{Data: columnNearestTop10TradeLevel, Name: columnNearestTop10TradeLevel, Searchable: true, Orderable: true},
		{Data: columnTicker, Name: columnCharts, Searchable: true, Orderable: false},
	}
}
