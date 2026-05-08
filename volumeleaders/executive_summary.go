package volumeleaders

import (
	"context"
	"fmt"
	"strconv"
	"strings"
)

const snapshotPartCount = 2

// Browser endpoint paths for executive-summary APIs captured from
// VolumeLeaders.
const (
	ExecutiveSummaryGetExhaustionScoresPath     = "/ExecutiveSummary/GetExhaustionScores"
	ExecutiveSummaryGetWelcomeTradesPath        = "/ExecutiveSummary/GetWelcomeTrades"
	ExecutiveSummaryGetWelcomeTradeClustersPath = "/ExecutiveSummary/GetWelcomeTradeClusters"
	TradesGetAllSnapshotsPath                   = "/Trades/GetAllSnapshots"
)

// ExhaustionScoresRequest contains the JSON payload for exhaustion scores.
type ExhaustionScoresRequest struct {
	Date string `json:"Date"`
}

// ExhaustionScores represents market exhaustion rank data from VolumeLeaders.
type ExhaustionScores struct {
	DateKey                   int `json:"DateKey"`
	ExhaustionScoreRank       int `json:"ExhaustionScoreRank"`
	ExhaustionScoreRank30Day  int `json:"ExhaustionScoreRank30Day"`
	ExhaustionScoreRank90Day  int `json:"ExhaustionScoreRank90Day"`
	ExhaustionScoreRank365Day int `json:"ExhaustionScoreRank365Day"`
}

// WelcomeTradesRequest contains DataTables paging and optional endpoint filters
// for /ExecutiveSummary/GetWelcomeTrades.
type WelcomeTradesRequest = EndpointRequest

// WelcomeTradeClustersRequest contains DataTables paging and optional endpoint
// filters for /ExecutiveSummary/GetWelcomeTradeClusters.
type WelcomeTradeClustersRequest = EndpointRequest

// GetExhaustionScores posts a JSON request to
// /ExecutiveSummary/GetExhaustionScores.
func (c *Client) GetExhaustionScores(
	ctx context.Context,
	req ExhaustionScoresRequest,
) (*ExhaustionScores, error) {
	var result ExhaustionScores
	if err := c.postJSON(ctx, ExecutiveSummaryGetExhaustionScoresPath, req, &result); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetWelcomeTrades posts a typed DataTables request to
// /ExecutiveSummary/GetWelcomeTrades.
func (c *Client) GetWelcomeTrades(
	ctx context.Context,
	req WelcomeTradesRequest,
) (*DataTablesResponse[Trade], error) {
	var result DataTablesResponse[Trade]
	if err := c.postDataTables(
		ctx,
		ExecutiveSummaryGetWelcomeTradesPath,
		req.DataTables,
		req.Filters,
		WelcomeTradesColumns(),
		&result,
	); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetWelcomeTradeClusters posts a typed DataTables request to
// /ExecutiveSummary/GetWelcomeTradeClusters.
func (c *Client) GetWelcomeTradeClusters(
	ctx context.Context,
	req WelcomeTradeClustersRequest,
) (*DataTablesResponse[TradeCluster], error) {
	var result DataTablesResponse[TradeCluster]
	if err := c.postDataTables(
		ctx,
		ExecutiveSummaryGetWelcomeTradeClustersPath,
		req.DataTables,
		req.Filters,
		WelcomeTradeClustersColumns(),
		&result,
	); err != nil {
		return nil, err
	}
	return &result, nil
}

// GetAllSnapshots posts a JSON null request to /Trades/GetAllSnapshots and
// returns ticker snapshot prices keyed by ticker symbol.
func (c *Client) GetAllSnapshots(ctx context.Context) (map[string]float64, error) {
	raw, err := c.GetAllSnapshotsString(ctx)
	if err != nil {
		return nil, err
	}
	snapshots, err := ParseSnapshots(raw)
	if err != nil {
		return nil, err
	}
	return snapshots, nil
}

// GetAllSnapshotsString posts a JSON null request to /Trades/GetAllSnapshots
// and returns the raw semicolon-delimited ticker snapshot string.
func (c *Client) GetAllSnapshotsString(ctx context.Context) (string, error) {
	var result string
	if err := c.postJSON(ctx, TradesGetAllSnapshotsPath, nil, &result); err != nil {
		return "", err
	}
	return result, nil
}

// WelcomeTradesColumns returns the DataTables columns captured from the welcome
// trades table.
func WelcomeTradesColumns() []DataTablesColumn {
	return []DataTablesColumn{
		{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true},
		{Data: columnTradeRank, Name: columnRName, Searchable: true, Orderable: true},
		{Data: columnDollarsMultiplier, Name: columnRelativeSizeName, Searchable: true, Orderable: true},
		{Data: columnCumulativeDistribution, Name: columnPercentName, Searchable: true, Orderable: true},
		{Data: tradeColumnLastTradeDate, Name: columnCharts, Searchable: true, Orderable: false},
	}
}

// WelcomeTradeClustersColumns returns the DataTables columns captured from the
// welcome trade clusters table.
func WelcomeTradeClustersColumns() []DataTablesColumn {
	return []DataTablesColumn{
		{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true},
		{Data: columnTradeClusterRank, Name: columnRName, Searchable: true, Orderable: true},
		{Data: columnDollarsMultiplier, Name: columnRelativeSizeName, Searchable: true, Orderable: true},
		{Data: columnCumulativeDistribution, Name: columnPercentName, Searchable: true, Orderable: true},
		{Data: columnLastComparibleTradeClusterDate, Name: columnCharts, Searchable: true, Orderable: false},
	}
}

// ParseSnapshots parses the semicolon-delimited ticker snapshot string returned
// by /Trades/GetAllSnapshots.
func ParseSnapshots(raw string) (map[string]float64, error) {
	snapshots := map[string]float64{}
	for item := range strings.SplitSeq(raw, ";") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		parts := strings.SplitN(item, ":", snapshotPartCount)
		if len(parts) != snapshotPartCount {
			return nil, fmt.Errorf("parse snapshot %q: missing separator", item)
		}
		ticker := strings.TrimSpace(parts[0])
		if ticker == "" {
			return nil, fmt.Errorf("parse snapshot %q: missing ticker", item)
		}
		price, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
		if err != nil {
			return nil, fmt.Errorf("parse snapshot price for %q: %w", ticker, err)
		}
		snapshots[ticker] = price
	}
	return snapshots, nil
}
