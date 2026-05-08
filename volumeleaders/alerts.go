package volumeleaders

import (
	"context"
	"net/url"
)

// Browser endpoint paths for alert configuration APIs captured from VolumeLeaders.
const (
	AlertConfigPath                             = "/AlertConfig"
	AlertConfigsGetAlertConfigsPath             = "/AlertConfigs/GetAlertConfigs"
	AlertConfigsDeleteAlertConfigPath           = "/AlertConfigs/DeleteAlertConfig"
	TradeAlertsGetTradeAlertsPath               = "/TradeAlerts/GetTradeAlerts"
	TradeClusterAlertsGetTradeClusterAlertsPath = "/TradeClusterAlerts/GetTradeClusterAlerts"
)

// AlertConfigsRequest contains DataTables paging and optional endpoint filters
// for /AlertConfigs/GetAlertConfigs.
type AlertConfigsRequest = EndpointRequest

// TradeAlertsRequest contains DataTables paging and optional endpoint filters
// for /TradeAlerts/GetTradeAlerts.
type TradeAlertsRequest = EndpointRequest

// TradeClusterAlertsRequest contains DataTables paging and optional endpoint
// filters for /TradeClusterAlerts/GetTradeClusterAlerts.
type TradeClusterAlertsRequest = EndpointRequest

// SaveAlertConfigRequest contains the multipart form fields submitted to
// /AlertConfig when creating or editing an alert.
type SaveAlertConfigRequest struct {
	// Fields is a low-level captured-form escape hatch. Callers must provide the
	// exact browser field names and values accepted by VolumeLeaders.
	Fields url.Values
}

// DeleteAlertConfigRequest contains the JSON payload for deleting an alert.
type DeleteAlertConfigRequest struct {
	AlertConfigKey int `json:"AlertConfigKey"`
}

// AlertConfig represents a VolumeLeaders saved alert configuration.
type AlertConfig struct {
	AlertConfigKey         int      `json:"AlertConfigKey"`
	UserKey                int      `json:"UserKey"`
	Name                   string   `json:"Name"`
	Tickers                string   `json:"Tickers"`
	TradeRankLTE           *int     `json:"TradeRankLTE"`
	TradeVCDGTE            *float64 `json:"TradeVCDGTE"`
	TradeMultGTE           *float64 `json:"TradeMultGTE"`
	TradeVolumeGTE         *int     `json:"TradeVolumeGTE"`
	TradeDollarsGTE        *float64 `json:"TradeDollarsGTE"`
	TradeConditions        *string  `json:"TradeConditions"`
	TradeClusterRankLTE    *int     `json:"TradeClusterRankLTE"`
	TradeClusterVCDGTE     *float64 `json:"TradeClusterVCDGTE"`
	TradeClusterMultGTE    *float64 `json:"TradeClusterMultGTE"`
	TradeClusterVolumeGTE  *int     `json:"TradeClusterVolumeGTE"`
	TradeClusterDollarsGTE *float64 `json:"TradeClusterDollarsGTE"`
	TotalRankLTE           *int     `json:"TotalRankLTE"`
	TotalVolumeGTE         *int     `json:"TotalVolumeGTE"`
	TotalDollarsGTE        *float64 `json:"TotalDollarsGTE"`
	AHRankLTE              *int     `json:"AHRankLTE"`
	AHVolumeGTE            *int     `json:"AHVolumeGTE"`
	AHDollarsGTE           *float64 `json:"AHDollarsGTE"`
	ClosingTradeRankLTE    *int     `json:"ClosingTradeRankLTE"`
	ClosingTradeVCDGTE     *float64 `json:"ClosingTradeVCDGTE"`
	ClosingTradeMultGTE    *float64 `json:"ClosingTradeMultGTE"`
	ClosingTradeVolumeGTE  *int     `json:"ClosingTradeVolumeGTE"`
	ClosingTradeDollarsGTE *float64 `json:"ClosingTradeDollarsGTE"`
	ClosingTradeConditions *string  `json:"ClosingTradeConditions"`
	OffsettingPrint        bool     `json:"OffsettingPrint"`
	PhantomPrint           bool     `json:"PhantomPrint"`
	Sweep                  bool     `json:"Sweep"`
	DarkPool               bool     `json:"DarkPool"`
}

// TradeAlert represents a VolumeLeaders trade alert row with alert metadata and
// trade details.
type TradeAlert struct {
	Date                         AspNetDate `json:"Date"`
	StartDate                    AspNetDate `json:"StartDate"`
	EndDate                      AspNetDate `json:"EndDate"`
	FullTimeString24             string     `json:"FullTimeString24"`
	DateKey                      int        `json:"DateKey"`
	SecurityKey                  int        `json:"SecurityKey"`
	TimeKey                      int        `json:"TimeKey"`
	TradeID                      int64      `json:"TradeID"`
	SequenceNumber               int        `json:"SequenceNumber"`
	UserKey                      int        `json:"UserKey"`
	UserKeys                     *string    `json:"UserKeys"`
	Sent                         bool       `json:"Sent"`
	Email                        *string    `json:"Email"`
	Emails                       *string    `json:"Emails"`
	Ticker                       string     `json:"Ticker"`
	Sector                       *string    `json:"Sector"`
	Industry                     *string    `json:"Industry"`
	Name                         string     `json:"Name"`
	AlertType                    string     `json:"AlertType"`
	Price                        float64    `json:"Price"`
	TradeRank                    int        `json:"TradeRank"`
	VolumeCumulativeDistribution float64    `json:"VolumeCumulativeDistribution"`
	DollarsMultiplier            float64    `json:"DollarsMultiplier"`
	Volume                       int        `json:"Volume"`
	Dollars                      float64    `json:"Dollars"`
	LastComparibleTradeDateKey   int        `json:"LastComparibleTradeDateKey"`
	LastComparibleTradeDate      AspNetDate `json:"LastComparibleTradeDate"`
	OffsettingTradeDate          AspNetDate `json:"OffsettingTradeDate"`
	PhantomPrintFulfillmentDate  AspNetDate `json:"PhantomPrintFulfillmentDate"`
	FullDateTime                 string     `json:"FullDateTime"`
	IPODate                      AspNetDate `json:"IPODate"`
	RSIHour                      float64    `json:"RSIHour"`
	RSIDay                       float64    `json:"RSIDay"`
	InProcess                    bool       `json:"InProcess"`
	Complete                     bool       `json:"Complete"`
	Sweep                        FlexBool   `json:"Sweep"`
	DarkPool                     FlexBool   `json:"DarkPool"`
	LatePrint                    FlexBool   `json:"LatePrint"`
	ClosingTrade                 FlexBool   `json:"ClosingTrade"`
	SignaturePrint               FlexBool   `json:"SignaturePrint"`
	PhantomPrint                 FlexBool   `json:"PhantomPrint"`
}

// TradeClusterAlert represents a trade cluster alert row. The endpoint returns
// the same shape as trade cluster rows.
type TradeClusterAlert = TradeCluster

// GetAlertConfigs posts a typed DataTables request to
// /AlertConfigs/GetAlertConfigs.
func (c *Client) GetAlertConfigs(
	ctx context.Context,
	req AlertConfigsRequest,
) (*DataTablesResponse[AlertConfig], error) {
	return getEndpoint[AlertConfig](c, ctx, AlertConfigsGetAlertConfigsPath, req, AlertConfigsColumns())
}

// GetTradeAlerts posts a typed DataTables request to
// /TradeAlerts/GetTradeAlerts.
func (c *Client) GetTradeAlerts(
	ctx context.Context,
	req TradeAlertsRequest,
) (*DataTablesResponse[TradeAlert], error) {
	return getEndpoint[TradeAlert](c, ctx, TradeAlertsGetTradeAlertsPath, req, TradeAlertsColumns())
}

// GetTradeClusterAlerts posts a typed DataTables request to
// /TradeClusterAlerts/GetTradeClusterAlerts.
func (c *Client) GetTradeClusterAlerts(
	ctx context.Context,
	req TradeClusterAlertsRequest,
) (*DataTablesResponse[TradeClusterAlert], error) {
	return getEndpoint[TradeClusterAlert](c, ctx, TradeClusterAlertsGetTradeClusterAlertsPath, req, TradeClusterAlertsColumns())
}

// GetAlertConfigsLimit fetches up to limit alert configs by paging through
// GetAlertConfigs. A zero or negative limit fetches all available records.
func (c *Client) GetAlertConfigsLimit(
	ctx context.Context,
	req AlertConfigsRequest,
	limit int,
) ([]AlertConfig, error) {
	return getEndpointLimit(ctx, req, limit, c.GetAlertConfigs)
}

// GetTradeAlertsLimit fetches up to limit trade alerts by paging through
// GetTradeAlerts. A zero or negative limit fetches all available records.
func (c *Client) GetTradeAlertsLimit(
	ctx context.Context,
	req TradeAlertsRequest,
	limit int,
) ([]TradeAlert, error) {
	return getEndpointLimit(ctx, req, limit, c.GetTradeAlerts)
}

// GetTradeClusterAlertsLimit fetches up to limit trade cluster alerts by
// paging through GetTradeClusterAlerts. A zero or negative limit fetches all
// available records.
func (c *Client) GetTradeClusterAlertsLimit(
	ctx context.Context,
	req TradeClusterAlertsRequest,
	limit int,
) ([]TradeClusterAlert, error) {
	return getEndpointLimit(ctx, req, limit, c.GetTradeClusterAlerts)
}

// SaveAlertConfig posts a multipart create or edit request to /AlertConfig.
func (c *Client) SaveAlertConfig(ctx context.Context, req SaveAlertConfigRequest) error {
	return c.postMultipartForm(ctx, AlertConfigPath, mergeValues(req.Fields, nil), nil)
}

// DeleteAlertConfig posts a JSON delete request to
// /AlertConfigs/DeleteAlertConfig.
func (c *Client) DeleteAlertConfig(ctx context.Context, req DeleteAlertConfigRequest) error {
	return c.postJSON(ctx, AlertConfigsDeleteAlertConfigPath, req, nil)
}

// AlertConfigsColumns returns the DataTables columns captured from the alert
// configurations table.
func AlertConfigsColumns() []DataTablesColumn {
	return []DataTablesColumn{
		{Data: columnName, Name: "", Searchable: true, Orderable: false},
		{Data: columnName, Name: columnName, Searchable: true, Orderable: true},
		{Data: columnTickers, Name: columnTickers, Searchable: true, Orderable: true},
		{Data: columnConditions, Name: columnConditions, Searchable: true, Orderable: false},
	}
}

// TradeAlertsColumns returns the DataTables columns used by the trade alerts
// table.
func TradeAlertsColumns() []DataTablesColumn {
	return TradesColumns()
}

// TradeClusterAlertsColumns returns the DataTables columns used by the trade
// cluster alerts table.
func TradeClusterAlertsColumns() []DataTablesColumn {
	return TradeClustersColumns()
}

