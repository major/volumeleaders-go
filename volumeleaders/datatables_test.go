package volumeleaders

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEncodeDataTablesRequestDefaults(t *testing.T) {
	values := EncodeDataTablesRequest(DataTablesRequest{})

	assert.Equal(t, "1", values.Get("draw"), "EncodeDataTablesRequest(default) draw")
	assert.Equal(t, "0", values.Get("start"), "EncodeDataTablesRequest(default) start")
	assert.Equal(t, "25", values.Get("length"), "EncodeDataTablesRequest(default) length")
	assert.Equal(t, "1", values.Get("order[0][column]"), "EncodeDataTablesRequest(default) order column")
	assert.Equal(t, "desc", values.Get("order[0][dir]"), "EncodeDataTablesRequest(default) order dir")
	assert.Empty(t, values.Get("search[value]"), "EncodeDataTablesRequest(default) search value")
}

func TestEncodeDataTablesRequestEncodesColumnsSearchOrderAndExtraValues(t *testing.T) {
	req := DataTablesRequest{
		Draw:   3,
		Start:  25,
		Length: 50,
		Columns: []DataTablesColumn{
			{Data: "Ticker", Name: "Ticker", Searchable: true, Orderable: true},
			{Data: "Volume", Name: "Sh", Searchable: false, Orderable: true},
		},
		Order: []DataTablesOrder{
			{Column: 1, Dir: "asc", Name: "Sh"},
			{Column: 0},
		},
		IncludeSearch: true,
		SearchValue:   "AXP",
		SearchRegex:   true,
		Extra: url.Values{
			"Tickers": {"AXP", "MSFT"},
			"MinSize": {"1000"},
		},
	}

	values := EncodeDataTablesRequest(req)

	checks := map[string]string{
		"draw":                      "3",
		"start":                     "25",
		"length":                    "50",
		"columns[0][data]":          "Ticker",
		"columns[0][name]":          "Ticker",
		"columns[0][searchable]":    "true",
		"columns[0][orderable]":     "true",
		"columns[0][search][value]": "",
		"columns[0][search][regex]": "false",
		"columns[1][data]":          "Volume",
		"columns[1][name]":          "Sh",
		"columns[1][searchable]":    "false",
		"columns[1][orderable]":     "true",
		"columns[1][search][value]": "",
		"columns[1][search][regex]": "false",
		"order[0][column]":          "1",
		"order[0][dir]":             "asc",
		"order[0][name]":            "Sh",
		"order[1][column]":          "0",
		"order[1][dir]":             "desc",
		"search[value]":             "AXP",
		"search[regex]":             "true",
		"MinSize":                   "1000",
	}
	for key, want := range checks {
		assert.Contains(t, values, key, "EncodeDataTablesRequest(req) key %q", key)
		assert.Equal(t, want, values.Get(key), "EncodeDataTablesRequest(req)[%q]", key)
	}
	assert.Equal(t, []string{"AXP", "MSFT"}, values["Tickers"], "EncodeDataTablesRequest(req) repeated extra values")
}

func TestMergeValuesCopiesBothSides(t *testing.T) {
	left := url.Values{"Tickers": {"AXP"}}
	right := url.Values{"Tickers": {"MSFT"}, "MinVolume": {"1000"}}

	merged := mergeValues(left, right)
	left.Set("Tickers", "changed")
	right.Set("MinVolume", "changed")

	assert.Equal(t, []string{"AXP", "MSFT"}, merged["Tickers"], "mergeValues() repeated values")
	assert.Equal(t, []string{"1000"}, merged["MinVolume"], "mergeValues() copied right values")
}
