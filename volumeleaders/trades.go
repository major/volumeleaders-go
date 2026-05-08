package volumeleaders

import (
	"context"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// TradesGetTradesPath is the browser endpoint path for institutional trades.
const TradesGetTradesPath = "/Trades/GetTrades"

const (
	tradeColumnTime          = "FullTimeString24"
	tradeColumnVolume        = "Volume"
	tradeColumnLastTradeDate = "LastComparibleTradeDate"
)

// TradeSortField identifies a sortable trades table column.
type TradeSortField string

// Supported high-level trade sort fields.
const (
	TradeSortTime                   TradeSortField = tradeColumnTime
	TradeSortTicker                 TradeSortField = columnTicker
	TradeSortSector                 TradeSortField = columnSector
	TradeSortIndustry               TradeSortField = columnIndustry
	TradeSortVolume                 TradeSortField = tradeColumnVolume
	TradeSortDollars                TradeSortField = columnDollars
	TradeSortDollarsMultiplier      TradeSortField = columnDollarsMultiplier
	TradeSortCumulativeDistribution TradeSortField = columnCumulativeDistribution
	TradeSortRank                   TradeSortField = columnTradeRank
	TradeSortRelativeSize           TradeSortField = columnRelativeSize
	TradeSortLastTradeDate          TradeSortField = tradeColumnLastTradeDate
)

// TradeSort describes one typed sort instruction for ListTrades.
type TradeSort struct {
	Field TradeSortField
	Desc  bool
}

// TradesRequest contains DataTables paging and optional endpoint filters for
// /Trades/GetTrades.
type TradesRequest = EndpointRequest

// TradeQuery contains high-level filters for /Trades/GetTrades.
//
// Zero-valued filter fields are omitted. Boolean pointer fields are omitted when
// nil so callers can distinguish "do not send this filter" from false.
type TradeQuery struct {
	Draw   int    // optional DataTables draw counter; defaults to 1.
	Start  int    // zero-based row offset.
	Length int    // page size; zero uses the package default.
	Search string // global trades table search text.

	Tickers         []string
	Sectors         []string
	Industries      []string
	SecurityTypes   []string
	TradeConditions []string

	StartDate time.Time // inclusive trade date lower bound.
	EndDate   time.Time // inclusive trade date upper bound.

	MinVolume  int     // minimum share volume.
	MaxVolume  int     // maximum share volume.
	MinPrice   float64 // minimum trade price.
	MaxPrice   float64 // maximum trade price.
	MinDollars float64 // minimum notional dollars.
	MaxDollars float64 // maximum notional dollars.

	MinRelativeSize float64 // minimum relative size filter.
	MaxTradeRank    int     // maximum trade rank, where lower ranks are larger trades.

	DarkPools        *bool // optional dark-pool filter.
	Sweeps           *bool // optional sweep filter.
	LatePrints       *bool // optional late-print filter.
	IncludePremarket *bool // optional premarket session bucket.
	IncludeRTH       *bool // optional regular-trading-hours session bucket.
	IncludeAH        *bool // optional after-hours session bucket.
	IncludeOpening   *bool // optional opening-trade session bucket.
	IncludeClosing   *bool // optional closing-trade session bucket.

	Sort []TradeSort
}

// TradePage contains one typed page of trades and the matching pagination
// metadata returned by VolumeLeaders.
type TradePage struct {
	Trades          []Trade
	RecordsTotal    int
	RecordsFiltered int
	Draw            int
	Start           int
	Length          int
}

// GetTrades posts a typed DataTables request to /Trades/GetTrades.
func (c *Client) GetTrades(
	ctx context.Context,
	req TradesRequest,
) (*DataTablesResponse[Trade], error) {
	return getEndpoint[Trade](ctx, c, TradesGetTradesPath, req, TradesColumns())
}

// ListTrades fetches one typed trade page without exposing DataTables or raw
// VolumeLeaders form fields to callers.
func (c *Client) ListTrades(ctx context.Context, query TradeQuery) (*TradePage, error) {
	req, err := query.tradesRequest()
	if err != nil {
		return nil, err
	}
	resp, err := c.GetTrades(ctx, req)
	if err != nil {
		return nil, err
	}
	return &TradePage{
		Trades:          resp.Data,
		RecordsTotal:    resp.RecordsTotal,
		RecordsFiltered: resp.RecordsFiltered,
		Draw:            resp.Draw,
		Start:           req.DataTables.Start,
		Length:          req.DataTables.Length,
	}, nil
}

// GetTradesLimit fetches up to limit trades by making one or more paged
// /Trades/GetTrades requests.
//
// A positive limit caps the total number of returned records. When limit is
// zero or negative, all available records are fetched. When
// req.DataTables.Length is positive it controls the page size (up to limit);
// otherwise the helper uses the smaller of limit and the DataTables default
// page size.
func (c *Client) GetTradesLimit(
	ctx context.Context,
	req TradesRequest,
	limit int,
) ([]Trade, error) {
	return getEndpointLimit(ctx, req, limit, c.GetTrades)
}

// ListTradesLimit fetches up to limit trades using the high-level TradeQuery
// filter surface. A zero or negative limit fetches all available records.
func (c *Client) ListTradesLimit(
	ctx context.Context,
	query TradeQuery,
	limit int,
) ([]Trade, error) {
	req, err := query.tradesRequest()
	if err != nil {
		return nil, err
	}
	return c.GetTradesLimit(ctx, req, limit)
}

// TradesColumns returns the DataTables columns captured from the trades table.
func TradesColumns() []DataTablesColumn {
	return []DataTablesColumn{
		{Data: tradeColumnTime, Name: "", Searchable: true, Orderable: false},
		{Data: tradeColumnTime, Name: tradeColumnTime, Searchable: true, Orderable: true},
		{Data: columnTicker, Name: columnTicker, Searchable: true, Orderable: true},
		{Data: columnCurrent, Name: columnCurrent, Searchable: true, Orderable: false},
		{
			Data:       columnTradeShortName,
			Name:       columnTradeShortName,
			Searchable: true,
			Orderable:  false,
		},
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
		{Data: columnTradeRank, Name: columnRName, Searchable: true, Orderable: true},
		{Data: columnRelativeSize, Name: columnRelativeSize, Searchable: true, Orderable: true},
		{
			Data:       tradeColumnLastTradeDate,
			Name:       tradeColumnLastTradeDate,
			Searchable: true,
			Orderable:  true,
		},
		{
			Data:       tradeColumnLastTradeDate,
			Name:       tradeColumnLastTradeDate,
			Searchable: true,
			Orderable:  false,
		},
	}
}

func (q TradeQuery) tradesRequest() (TradesRequest, error) {
	if q.Start < 0 {
		return TradesRequest{}, fmt.Errorf(
			"%w: start must be non-negative: %d",
			ErrInvalidQuery,
			q.Start,
		)
	}
	if q.Length < 0 {
		return TradesRequest{}, fmt.Errorf(
			"%w: length must be non-negative: %d",
			ErrInvalidQuery,
			q.Length,
		)
	}
	if !q.StartDate.IsZero() && !q.EndDate.IsZero() && q.StartDate.After(q.EndDate) {
		return TradesRequest{}, fmt.Errorf(
			"%w: start date %s is after end date %s",
			ErrInvalidQuery,
			tradeDate(q.StartDate),
			tradeDate(q.EndDate),
		)
	}
	if q.MinVolume < 0 || q.MaxVolume < 0 || q.MaxTradeRank < 0 {
		return TradesRequest{}, fmt.Errorf(
			"%w: numeric trade filters must be non-negative",
			ErrInvalidQuery,
		)
	}
	if hasNegativeFloat(q.MinPrice, q.MaxPrice, q.MinDollars, q.MaxDollars, q.MinRelativeSize) {
		return TradesRequest{}, fmt.Errorf(
			"%w: decimal trade filters must be non-negative",
			ErrInvalidQuery,
		)
	}
	if q.MinVolume > 0 && q.MaxVolume > 0 && q.MinVolume > q.MaxVolume {
		return TradesRequest{}, fmt.Errorf(
			"%w: minimum volume %d exceeds maximum volume %d",
			ErrInvalidQuery,
			q.MinVolume,
			q.MaxVolume,
		)
	}
	if q.MinPrice > 0 && q.MaxPrice > 0 && q.MinPrice > q.MaxPrice {
		return TradesRequest{}, fmt.Errorf(
			"%w: minimum price %.2f exceeds maximum price %.2f",
			ErrInvalidQuery,
			q.MinPrice,
			q.MaxPrice,
		)
	}
	if q.MinDollars > 0 && q.MaxDollars > 0 && q.MinDollars > q.MaxDollars {
		return TradesRequest{}, fmt.Errorf(
			"%w: minimum dollars %.2f exceeds maximum dollars %.2f",
			ErrInvalidQuery,
			q.MinDollars,
			q.MaxDollars,
		)
	}

	filters := url.Values{}
	addStringList(filters, "Tickers", q.Tickers)
	addStringList(filters, "Sectors", q.Sectors)
	addStringList(filters, "Industries", q.Industries)
	addStringList(filters, "SecurityTypes", q.SecurityTypes)
	addStringList(filters, "TradeConditions", q.TradeConditions)
	addDate(filters, "StartDate", q.StartDate)
	addDate(filters, "EndDate", q.EndDate)
	addPositiveInt(filters, "MinVolume", q.MinVolume)
	addPositiveInt(filters, "MaxVolume", q.MaxVolume)
	addPositiveFloat(filters, "MinPrice", q.MinPrice)
	addPositiveFloat(filters, "MaxPrice", q.MaxPrice)
	addPositiveFloat(filters, "MinDollars", q.MinDollars)
	addPositiveFloat(filters, "MaxDollars", q.MaxDollars)
	addPositiveFloat(filters, "RelativeSize", q.MinRelativeSize)
	addPositiveInt(filters, "TradeRank", q.MaxTradeRank)
	addBool(filters, "DarkPools", q.DarkPools)
	addBool(filters, "Sweeps", q.Sweeps)
	addBool(filters, "LatePrints", q.LatePrints)
	addBool(filters, "IncludePremarket", q.IncludePremarket)
	addBool(filters, "IncludeRTH", q.IncludeRTH)
	addBool(filters, "IncludeAH", q.IncludeAH)
	addBool(filters, "IncludeOpening", q.IncludeOpening)
	addBool(filters, "IncludeClosing", q.IncludeClosing)

	return TradesRequest{
		DataTables: DataTablesRequest{
			Draw:          q.Draw,
			Start:         q.Start,
			Length:        q.Length,
			Order:         q.dataTablesOrder(),
			IncludeSearch: q.Search != "",
			SearchValue:   q.Search,
		},
		Filters: filters,
	}, nil
}

func (q TradeQuery) dataTablesOrder() []DataTablesOrder {
	if len(q.Sort) == 0 {
		return nil
	}
	columns := TradesColumns()
	orders := make([]DataTablesOrder, 0, len(q.Sort))
	for _, sort := range q.Sort {
		column := tradeSortColumn(columns, sort.Field)
		if column < 0 {
			continue
		}
		dir := ascendingOrderDir
		if sort.Desc {
			dir = defaultOrderDir
		}
		orders = append(orders, DataTablesOrder{Column: column, Dir: dir, Name: string(sort.Field)})
	}
	return orders
}

func tradeSortColumn(columns []DataTablesColumn, field TradeSortField) int {
	for i, column := range columns {
		if column.Orderable && column.Data == string(field) {
			return i
		}
	}
	return -1
}

func mergeValues(left, right url.Values) url.Values {
	merged := url.Values{}
	for key, values := range left {
		for _, value := range values {
			merged.Add(key, value)
		}
	}
	for key, values := range right {
		for _, value := range values {
			merged.Add(key, value)
		}
	}
	return merged
}

func addStringList(values url.Values, key string, items []string) {
	cleaned := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item != "" {
			cleaned = append(cleaned, item)
		}
	}
	if len(cleaned) > 0 {
		values.Set(key, strings.Join(cleaned, ","))
	}
}

func addDate(values url.Values, key string, value time.Time) {
	if !value.IsZero() {
		values.Set(key, tradeDate(value))
	}
}

func addPositiveInt(values url.Values, key string, value int) {
	if value > 0 {
		values.Set(key, strconv.Itoa(value))
	}
}

func addPositiveFloat(values url.Values, key string, value float64) {
	if value > 0 {
		values.Set(key, strconv.FormatFloat(value, 'f', -1, 64))
	}
}

func addBool(values url.Values, key string, value *bool) {
	if value != nil {
		values.Set(key, strconv.FormatBool(*value))
	}
}

func tradeDate(value time.Time) string {
	return value.Format("2006-01-02")
}

func hasNegativeFloat(values ...float64) bool {
	for _, value := range values {
		if value < 0 {
			return true
		}
	}
	return false
}
