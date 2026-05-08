package volumeleaders

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"time"
)

const (
	dateTimeMinEpochMillis = -62135596800000
	date1900EpochMillis    = -2208988800000
	jsonNull               = "null"
	jsonTrue               = "true"
	jsonFalse              = "false"
)

var aspNetDatePattern = regexp.MustCompile(`^/Date\((-?\d+)\)/$`)

// AspNetDate is a nullable time that parses ASP.NET /Date(epoch_ms)/ JSON
// strings returned by VolumeLeaders.
type AspNetDate struct {
	time.Time

	Valid bool
}

// UnmarshalJSON parses ASP.NET JSON date strings into UTC times.
func (d *AspNetDate) UnmarshalJSON(data []byte) error {
	if string(data) == jsonNull {
		*d = AspNetDate{}
		return nil
	}

	var value string
	if err := json.Unmarshal(data, &value); err != nil {
		return fmt.Errorf("unmarshal ASP.NET date string: %w", err)
	}
	if value == "" {
		*d = AspNetDate{}
		return nil
	}

	matches := aspNetDatePattern.FindStringSubmatch(value)
	if matches == nil {
		return fmt.Errorf("invalid ASP.NET date format %q", value)
	}
	epochMillis, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return fmt.Errorf("parse ASP.NET date epoch milliseconds: %w", err)
	}
	if epochMillis == dateTimeMinEpochMillis || epochMillis == date1900EpochMillis {
		*d = AspNetDate{}
		return nil
	}

	*d = AspNetDate{Time: time.UnixMilli(epochMillis).UTC(), Valid: true}
	return nil
}

// MarshalJSON serializes valid dates as RFC3339 strings and invalid dates as
// null.
func (d AspNetDate) MarshalJSON() ([]byte, error) {
	if !d.Valid {
		return []byte(jsonNull), nil
	}
	data, err := json.Marshal(d.Time.UTC().Format(time.RFC3339))
	if err != nil {
		return nil, fmt.Errorf("marshal ASP.NET date: %w", err)
	}
	return data, nil
}

// FlexBool handles JSON booleans that arrive as 0/1 integers or true/false.
// The VolumeLeaders API returns boolean fields as numeric 0 and 1.
type FlexBool bool

// UnmarshalJSON accepts JSON true, false, 0, 1, and null.
func (b *FlexBool) UnmarshalJSON(data []byte) error {
	switch string(data) {
	case jsonTrue, "1":
		*b = true
	case jsonFalse, "0", jsonNull:
		*b = false
	default:
		return fmt.Errorf("cannot unmarshal %s into bool", string(data))
	}
	return nil
}

// MarshalJSON writes a standard JSON boolean.
func (b FlexBool) MarshalJSON() ([]byte, error) {
	if b {
		return []byte(jsonTrue), nil
	}
	return []byte(jsonFalse), nil
}

// Trade represents a VolumeLeaders institutional trade row.
type Trade struct {
	Date                          AspNetDate `json:"Date"`
	StartDate                     AspNetDate `json:"StartDate"`
	EndDate                       AspNetDate `json:"EndDate"`
	TD30                          AspNetDate `json:"TD30"`
	TD90                          AspNetDate `json:"TD90"`
	TD1CY                         AspNetDate `json:"TD1CY"`
	DateKey                       int        `json:"DateKey"`
	TimeKey                       int        `json:"TimeKey"`
	SecurityKey                   int        `json:"SecurityKey"`
	TradeID                       int64      `json:"TradeID"`
	SequenceNumber                int        `json:"SequenceNumber"`
	EOM                           FlexBool   `json:"EOM"`
	EOQ                           FlexBool   `json:"EOQ"`
	EOY                           FlexBool   `json:"EOY"`
	OPEX                          FlexBool   `json:"OPEX"`
	VOLEX                         FlexBool   `json:"VOLEX"`
	Ticker                        string     `json:"Ticker"`
	Sector                        string     `json:"Sector"`
	Industry                      string     `json:"Industry"`
	Name                          string     `json:"Name"`
	FullDateTime                  string     `json:"FullDateTime"`
	FullTimeString24              string     `json:"FullTimeString24"`
	Price                         float64    `json:"Price"`
	Bid                           float64    `json:"Bid"`
	Ask                           float64    `json:"Ask"`
	Dollars                       float64    `json:"Dollars"`
	AverageBlockSizeDollars       float64    `json:"AverageBlockSizeDollars"`
	AverageBlockSizeShares        int        `json:"AverageBlockSizeShares"`
	DollarsMultiplier             float64    `json:"DollarsMultiplier"`
	Volume                        int        `json:"Volume"`
	AverageDailyVolume            int        `json:"AverageDailyVolume"`
	PercentDailyVolume            float64    `json:"PercentDailyVolume"`
	RelativeSize                  float64    `json:"RelativeSize"`
	LastComparibleTradeDate       AspNetDate `json:"LastComparibleTradeDate"`
	IPODate                       AspNetDate `json:"IPODate"`
	OffsettingTradeDate           AspNetDate `json:"OffsettingTradeDate"`
	PhantomPrintFulfillmentDate   AspNetDate `json:"PhantomPrintFulfillmentDate"`
	PhantomPrintFulfillmentDays   *int       `json:"PhantomPrintFulfillmentDays"`
	TradeCount                    int        `json:"TradeCount"`
	CumulativeDistribution        float64    `json:"CumulativeDistribution"`
	TradeRank                     int        `json:"TradeRank"`
	TradeRankSnapshot             int        `json:"TradeRankSnapshot"`
	LatePrint                     FlexBool   `json:"LatePrint"`
	Sweep                         FlexBool   `json:"Sweep"`
	DarkPool                      FlexBool   `json:"DarkPool"`
	OpeningTrade                  FlexBool   `json:"OpeningTrade"`
	ClosingTrade                  FlexBool   `json:"ClosingTrade"`
	PhantomPrint                  FlexBool   `json:"PhantomPrint"`
	InsideBar                     FlexBool   `json:"InsideBar"`
	DoubleInsideBar               FlexBool   `json:"DoubleInsideBar"`
	SignaturePrint                FlexBool   `json:"SignaturePrint"`
	NewPosition                   FlexBool   `json:"NewPosition"`
	AHInstitutionalDollars        float64    `json:"AHInstitutionalDollars"`
	AHInstitutionalDollarsRank    int        `json:"AHInstitutionalDollarsRank"`
	AHInstitutionalVolume         int        `json:"AHInstitutionalVolume"`
	TotalInstitutionalDollars     float64    `json:"TotalInstitutionalDollars"`
	TotalInstitutionalDollarsRank int        `json:"TotalInstitutionalDollarsRank"`
	TotalInstitutionalVolume      int        `json:"TotalInstitutionalVolume"`
	ClosingTradeDollars           float64    `json:"ClosingTradeDollars"`
	ClosingTradeDollarsRank       int        `json:"ClosingTradeDollarsRank"`
	ClosingTradeVolume            int        `json:"ClosingTradeVolume"`
	TotalDollars                  float64    `json:"TotalDollars"`
	TotalDollarsRank              int        `json:"TotalDollarsRank"`
	TotalVolume                   int        `json:"TotalVolume"`
	ClosePrice                    float64    `json:"ClosePrice"`
	RSIHour                       float64    `json:"RSIHour"`
	RSIDay                        float64    `json:"RSIDay"`
	TotalRows                     int        `json:"TotalRows"`
	TradeConditions               *string    `json:"TradeConditions"`
	FrequencyLast30TD             int        `json:"FrequencyLast30TD"`
	FrequencyLast90TD             int        `json:"FrequencyLast90TD"`
	FrequencyLast1CY              int        `json:"FrequencyLast1CY"`
	Cancelled                     FlexBool   `json:"Cancelled"` //nolint:misspell // VolumeLeaders API spells this field "Cancelled".
	TotalTrades                   int        `json:"TotalTrades"`
	ExternalFeed                  FlexBool   `json:"ExternalFeed"`
}

// SimilarTradeCountLast30Days returns how often VolumeLeaders saw trades of
// this size in the 30-day lookback window.
//
// This helps callers distinguish isolated institutional prints from trades
// that repeat often enough to suggest a recurring pattern.
func (t Trade) SimilarTradeCountLast30Days() int {
	return t.FrequencyLast30TD
}

// SimilarTradeCountLast90Days returns how often VolumeLeaders saw trades of
// this size in the 90-day lookback window.
//
// This broader window gives callers more context for prints that may be rare
// month-to-month but still recur across a quarter.
func (t Trade) SimilarTradeCountLast90Days() int {
	return t.FrequencyLast90TD
}

// SimilarTradeCountLastYear returns how often VolumeLeaders saw trades of this
// size in the one-calendar-year lookback window.
//
// This longest frequency window helps callers spot whether a print is unusual
// relative to the ticker's yearly institutional trade history.
func (t Trade) SimilarTradeCountLastYear() int {
	return t.FrequencyLast1CY
}

// PercentileRank returns the trade's percentile rank relative to all other
// VolumeLeaders trades for the same ticker.
//
// VolumeLeaders labels this value as cumulative distribution, or VCD, in some
// request and alert fields. Higher values indicate trades that rank farther out
// in the ticker's historical trade-size distribution.
func (t Trade) PercentileRank() float64 {
	return t.CumulativeDistribution
}

// IsPhantomPrint reports whether VolumeLeaders marked the trade as a print far
// from the current price.
func (t Trade) IsPhantomPrint() bool {
	return bool(t.PhantomPrint)
}

// IsOptionsExpiration reports whether the trade occurred during monthly
// options expiration.
func (t Trade) IsOptionsExpiration() bool {
	return bool(t.OPEX)
}

// IsVolatilityExpiration reports whether the trade occurred during VIX
// contract expiration.
func (t Trade) IsVolatilityExpiration() bool {
	return bool(t.VOLEX)
}

// IsCancelled reports whether the exchange marked the trade as canceled after
// it hit the tape.
//
// Canceled trades should usually be ignored by analysis because the reported
// print did not represent a real completed order.
func (t Trade) IsCancelled() bool {
	return bool(t.Cancelled)
}
