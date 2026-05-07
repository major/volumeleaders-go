package volumeleaders

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const tradesFixturePath = "testdata/trades_get_trades_response.json"

func TestGetTradesPostsDataTablesXHRRequest(t *testing.T) {
	fixture := mustReadFixture(t)
	var capturedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "GetTrades() method")
		assert.Equal(t, TradesGetTradesPath, r.URL.Path, "GetTrades() path")
		assert.Equal(
			t,
			"application/x-www-form-urlencoded; charset=UTF-8",
			r.Header.Get("Content-Type"),
			"GetTrades() Content-Type",
		)
		assert.Equal(t, "XMLHttpRequest", r.Header.Get("X-Requested-With"), "GetTrades() X-Requested-With")
		assert.Equal(t, "token-123", r.Header.Get("X-Xsrf-Token"), "GetTrades() X-Xsrf-Token")
		assert.Equal(t, "test-agent", r.Header.Get("User-Agent"), "GetTrades() User-Agent")

		browserHeaders := map[string]string{
			"Sec-Ch-Ua":          `"Chromium";v="147", "Not A(Brand";v="24", "Google Chrome";v="147"`,
			"Sec-Ch-Ua-Mobile":   "?0",
			"Sec-Ch-Ua-Platform": `"Windows"`,
			"Sec-Fetch-Dest":     "empty",
			"Sec-Fetch-Mode":     "cors",
			"Sec-Fetch-Site":     "same-origin",
			"Accept-Language":    "en-US,en;q=0.9",
		}
		for name, want := range browserHeaders {
			assert.Equal(t, want, r.Header.Get(name), "GetTrades() %s", name)
		}
		assert.NotContains(t, r.Header.Get("Accept-Encoding"), "br", "GetTrades() Accept-Encoding")

		cookie, err := r.Cookie("ASP.NET_SessionId")
		if !assert.NoError(t, err, "GetTrades() cookie") {
			return
		}
		assert.Equal(t, "session-123", cookie.Value, "GetTrades() ASP.NET_SessionId cookie")

		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(request body)") {
			return
		}
		capturedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	t.Cleanup(server.Close)

	session := Session{
		Cookies:   []*http.Cookie{{Name: "ASP.NET_SessionId", Value: "session-123"}},
		XSRFToken: "token-123",
	}
	client, err := NewClient(
		session,
		WithBaseURL(server.URL),
		WithUserAgent("test-agent"),
		WithHTTPClient(server.Client()),
	)
	require.NoError(t, err, "NewClient()")

	resp, err := client.GetTrades(context.Background(), TradesRequest{
		DataTables: DataTablesRequest{Draw: 7, Start: 25, Length: 50, IncludeSearch: true, SearchValue: "AXP"},
		Filters:    url.Values{"Tickers": {"AXP"}, "MinVolume": {"1000"}},
	})
	require.NoError(t, err, "GetTrades()")
	assert.Equal(t, 465, resp.RecordsFiltered, "GetTrades() RecordsFiltered")
	require.Len(t, resp.Data, 2, "GetTrades() data")
	assert.Equal(t, "AXP", resp.Data[0].Ticker, "GetTrades() first trade ticker")
	assert.Equal(t, int64(71774613157188), resp.Data[0].TradeID, "GetTrades() first trade ID")

	form, err := url.ParseQuery(capturedBody)
	require.NoError(t, err, "ParseQuery(GetTrades body)")
	checks := map[string]string{
		"draw":                       "7",
		"start":                      "25",
		"length":                     "50",
		"search[value]":              "AXP",
		"columns[0][data]":           "FullTimeString24",
		"columns[0][orderable]":      "false",
		"columns[7][name]":           "Sh",
		"order[0][column]":           "1",
		"order[0][dir]":              "desc",
		"Tickers":                    "AXP",
		"MinVolume":                  "1000",
		"columns[14][search][regex]": "false",
	}
	for key, want := range checks {
		assert.Equal(t, want, form.Get(key), "GetTrades() form[%q]", key)
	}
}

func TestGetTradesReturnsStatusError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "30")
		http.Error(w, "not authenticated", http.StatusUnauthorized)
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	_, err = client.GetTrades(context.Background(), TradesRequest{})
	require.Error(t, err, "GetTrades()")
	var statusErr *StatusError
	require.ErrorAs(t, err, &statusErr, "GetTrades() status error")
	assert.Equal(t, http.StatusUnauthorized, statusErr.StatusCode, "GetTrades() status")
	assert.Equal(t, http.MethodPost, statusErr.Method, "GetTrades() status method")
	assert.Equal(t, TradesGetTradesPath, statusErr.Path, "GetTrades() status path")
	assert.Equal(t, "30", statusErr.RetryAfter, "GetTrades() Retry-After")
	assert.Equal(t, "30", statusErr.Header.Get("Retry-After"), "GetTrades() header Retry-After")
	assert.Contains(t, statusErr.Body, "not authenticated", "GetTrades() body preview")
	assert.True(t, IsStatusCode(err, http.StatusUnauthorized), "IsStatusCode(GetTrades error, 401)")
}

func TestGetTradesReportsSessionExpiredRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == TradesGetTradesPath {
			http.Redirect(w, r, "/login", http.StatusFound)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body>login required</body></html>`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	_, err = client.GetTrades(context.Background(), TradesRequest{})

	require.Error(t, err, "GetTrades() login redirect")
	assert.True(t, IsSessionExpired(err), "IsSessionExpired(GetTrades login redirect error)")
	assert.True(t, IsAuthError(err), "IsAuthError(GetTrades login redirect error)")
}

func TestGetTradesReportsSessionExpiredLoginPage(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body>login required</body></html>`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	_, err = client.GetTrades(context.Background(), TradesRequest{})

	require.Error(t, err, "GetTrades() login page")
	assert.True(t, IsSessionExpired(err), "IsSessionExpired(GetTrades login page error)")
	assert.True(t, IsAuthError(err), "IsAuthError(GetTrades login page error)")
}

func TestGetTradesReportsUnexpectedContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<html><body>not JSON</body></html>`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	_, err = client.GetTrades(context.Background(), TradesRequest{})

	require.Error(t, err, "GetTrades() HTML response")
	assert.True(t, IsUnexpectedContent(err), "IsUnexpectedContent(GetTrades HTML response error)")
	assert.False(t, IsAuthError(err), "IsAuthError(GetTrades HTML response error)")
	var contentErr *ContentError
	require.ErrorAs(t, err, &contentErr, "GetTrades() content error")
	assert.Equal(t, http.MethodPost, contentErr.Method, "ContentError.Method")
	assert.Equal(t, TradesGetTradesPath, contentErr.Path, "ContentError.Path")
	assert.Contains(t, contentErr.ContentType, "text/html", "ContentError.ContentType")
	assert.Contains(t, contentErr.Body, "not JSON", "ContentError.Body")
}

func TestGetTradesReturnsBodyLimitError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"draw":1,"recordsTotal":1,"recordsFiltered":1,"data":[]}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(
		Session{},
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
		WithResponseBodyLimit(8),
	)
	require.NoError(t, err, "NewClient()")

	_, err = client.GetTrades(context.Background(), TradesRequest{})
	require.Error(t, err, "GetTrades()")
	var limitErr *BodyLimitError
	require.ErrorAs(t, err, &limitErr, "GetTrades() body limit error")
	assert.Equal(t, int64(8), limitErr.Limit, "BodyLimitError.Limit")
	assert.True(t, IsBodyLimit(err), "IsBodyLimit(GetTrades error)")
}

func TestGetTradesLimitFetchesBoundedPages(t *testing.T) {
	allTrades := []Trade{
		{Ticker: "AXP", TradeID: 1},
		{Ticker: "MSFT", TradeID: 2},
		{Ticker: "NVDA", TradeID: 3},
		{Ticker: "AMD", TradeID: 4},
		{Ticker: "TSLA", TradeID: 5},
	}
	var starts []string
	var lengths []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(GetTradesLimit request body)") {
			return
		}
		form, err := url.ParseQuery(string(body))
		if !assert.NoError(t, err, "ParseQuery(GetTradesLimit request body)") {
			return
		}
		starts = append(starts, form.Get("start"))
		lengths = append(lengths, form.Get("length"))

		start, err := strconv.Atoi(form.Get("start"))
		if !assert.NoError(t, err, "Atoi(GetTradesLimit start)") {
			return
		}
		length, err := strconv.Atoi(form.Get("length"))
		if !assert.NoError(t, err, "Atoi(GetTradesLimit length)") {
			return
		}
		end := min(start+length, len(allTrades))
		if start > end {
			start = end
		}

		w.Header().Set("Content-Type", "application/json")
		_, err = fmt.Fprintf(
			w,
			`{"draw":1,"recordsTotal":%d,"recordsFiltered":%d,"data":%s}`,
			len(allTrades),
			len(allTrades),
			mustMarshalJSON(t, allTrades[start:end]),
		)
		assert.NoError(t, err, "Fprintf(GetTradesLimit response)")
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	trades, err := client.GetTradesLimit(context.Background(), TradesRequest{
		DataTables: DataTablesRequest{Length: 2},
	}, 3)

	require.NoError(t, err, "GetTradesLimit(limit 3)")
	assert.Equal(t, []string{"0", "2"}, starts, "GetTradesLimit(limit 3) starts")
	assert.Equal(t, []string{"2", "1"}, lengths, "GetTradesLimit(limit 3) lengths")
	require.Len(t, trades, 3, "GetTradesLimit(limit 3) trades")
	assert.Equal(
		t,
		[]string{"AXP", "MSFT", "NVDA"},
		[]string{trades[0].Ticker, trades[1].Ticker, trades[2].Ticker},
		"GetTradesLimit(limit 3) tickers",
	)
}

func TestGetTradesLimitRequiresPositiveLimit(t *testing.T) {
	client, err := NewClient(Session{})
	require.NoError(t, err, "NewClient()")

	trades, err := client.GetTradesLimit(context.Background(), TradesRequest{}, 0)

	assert.Nil(t, trades, "GetTradesLimit(limit 0) trades")
	assert.Error(t, err, "GetTradesLimit(limit 0) error")
}

func TestListTradesMapsTypedQuery(t *testing.T) {
	fixture := mustReadFixture(t)
	includePremarket := true
	darkPools := false
	var capturedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(ListTrades request body)") {
			return
		}
		capturedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(fixture)
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	page, err := client.ListTrades(context.Background(), TradeQuery{
		Draw:             9,
		Start:            10,
		Length:           20,
		Search:           "AXP",
		Tickers:          []string{" AXP ", "MSFT"},
		Sectors:          []string{"Financial Services"},
		Industries:       []string{"Credit Services"},
		SecurityTypes:    []string{"Stock"},
		TradeConditions:  []string{"Regular"},
		StartDate:        time.Date(2026, time.May, 1, 12, 0, 0, 0, time.UTC),
		EndDate:          time.Date(2026, time.May, 2, 12, 0, 0, 0, time.UTC),
		MinVolume:        1000,
		MaxVolume:        2000,
		MinPrice:         10.5,
		MaxPrice:         99.75,
		MinDollars:       50000,
		MaxDollars:       250000,
		MinRelativeSize:  1.5,
		MaxTradeRank:     30,
		DarkPools:        &darkPools,
		IncludePremarket: &includePremarket,
		Sort:             []TradeSort{{Field: TradeSortVolume, Desc: true}},
	})
	require.NoError(t, err, "ListTrades()")
	require.NotNil(t, page, "ListTrades() page")
	assert.Equal(t, 1, page.Draw, "ListTrades() Draw")
	assert.Equal(t, 10, page.Start, "ListTrades() Start")
	assert.Equal(t, 20, page.Length, "ListTrades() Length")
	assert.Equal(t, 465, page.RecordsFiltered, "ListTrades() RecordsFiltered")
	require.Len(t, page.Trades, 2, "ListTrades() Trades")

	form, err := url.ParseQuery(capturedBody)
	require.NoError(t, err, "ParseQuery(ListTrades body)")
	checks := map[string]string{
		"draw":             "9",
		"start":            "10",
		"length":           "20",
		"search[value]":    "AXP",
		"Tickers":          "AXP,MSFT",
		"Sectors":          "Financial Services",
		"Industries":       "Credit Services",
		"SecurityTypes":    "Stock",
		"TradeConditions":  "Regular",
		"StartDate":        "2026-05-01",
		"EndDate":          "2026-05-02",
		"MinVolume":        "1000",
		"MaxVolume":        "2000",
		"MinPrice":         "10.5",
		"MaxPrice":         "99.75",
		"MinDollars":       "50000",
		"MaxDollars":       "250000",
		"RelativeSize":     "1.5",
		"TradeRank":        "30",
		"DarkPools":        "false",
		"IncludePremarket": "true",
		"columns[0][data]": "FullTimeString24",
		"order[0][column]": "7",
		"order[0][dir]":    "desc",
		"order[0][name]":   "Volume",
		"search[regex]":    "false",
	}
	for key, want := range checks {
		assert.Equal(t, want, form.Get(key), "ListTrades() form[%q]", key)
	}
}

func TestListTradesRejectsInvalidQuery(t *testing.T) {
	client, err := NewClient(Session{})
	require.NoError(t, err, "NewClient()")

	_, err = client.ListTrades(context.Background(), TradeQuery{Start: -1})

	require.Error(t, err, "ListTrades(invalid query)")
	require.ErrorIs(t, err, ErrInvalidQuery, "ListTrades(invalid query) error")
	assert.True(t, IsInvalidQuery(err), "IsInvalidQuery(ListTrades error)")
}

func TestListTradesLimitUsesTypedQuery(t *testing.T) {
	allTrades := []Trade{{Ticker: "AXP", TradeID: 1}, {Ticker: "MSFT", TradeID: 2}}
	var tickerFilters []string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(ListTradesLimit request body)") {
			return
		}
		form, err := url.ParseQuery(string(body))
		if !assert.NoError(t, err, "ParseQuery(ListTradesLimit request body)") {
			return
		}
		tickerFilters = append(tickerFilters, form.Get("Tickers"))
		w.Header().Set("Content-Type", "application/json")
		_, err = fmt.Fprintf(
			w,
			`{"draw":1,"recordsTotal":%d,"recordsFiltered":%d,"data":%s}`,
			len(allTrades),
			len(allTrades),
			mustMarshalJSON(t, allTrades),
		)
		assert.NoError(t, err, "Fprintf(ListTradesLimit response)")
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	trades, err := client.ListTradesLimit(context.Background(), TradeQuery{Tickers: []string{"AXP"}, Length: 2}, 2)

	require.NoError(t, err, "ListTradesLimit()")
	assert.Equal(t, []string{"AXP"}, tickerFilters, "ListTradesLimit() ticker filters")
	assert.Equal(t, []string{"AXP", "MSFT"}, []string{trades[0].Ticker, trades[1].Ticker}, "ListTradesLimit() tickers")
}

func TestListTradesLimitRejectsInvalidLimit(t *testing.T) {
	client, err := NewClient(Session{})
	require.NoError(t, err, "NewClient()")

	_, err = client.ListTradesLimit(context.Background(), TradeQuery{}, 0)

	require.Error(t, err, "ListTradesLimit(invalid limit)")
	require.ErrorIs(t, err, ErrInvalidQuery, "ListTradesLimit(invalid limit) error")
	assert.True(t, IsInvalidQuery(err), "IsInvalidQuery(ListTradesLimit error)")
}

func TestClientBlocksCrossOriginRedirects(t *testing.T) {
	redirectTargetHit := false
	target := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		redirectTargetHit = true
		assert.Empty(t, r.Header.Get("X-Xsrf-Token"), "redirect target X-Xsrf-Token")
		assert.Empty(t, r.Header.Get("X-Custom-Secret"), "redirect target X-Custom-Secret")
	}))
	t.Cleanup(target.Close)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusFound)
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(
		Session{XSRFToken: "token-123", Headers: http.Header{"X-Custom-Secret": {"secret-123"}}},
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
	)
	require.NoError(t, err, "NewClient()")

	_, err = client.GetTrades(context.Background(), TradesRequest{})
	require.ErrorIs(t, err, ErrCrossOriginRedirect, "GetTrades() redirect error")
	assert.False(t, redirectTargetHit, "redirect target was hit")
}

func TestTradesFixtureDecodes(t *testing.T) {
	fixture := mustReadFixture(t)
	var resp DataTablesResponse[Trade]
	require.NoError(t, json.Unmarshal(fixture, &resp), "Unmarshal(trades fixture)")
	require.Len(t, resp.Data, 2, "Unmarshal(trades fixture) data")

	trade := resp.Data[0]
	require.True(t, trade.Date.Valid, "Trade.Date.Valid")
	wantDate := time.Date(2026, time.May, 1, 0, 0, 0, 0, time.UTC)
	assert.True(t, trade.Date.Time.Equal(wantDate), "Trade.Date = %s, want %s", trade.Date.Time, wantDate)
	assert.False(t, trade.TD30.Valid, "Trade.TD30.Valid for sentinel date")
	assert.False(t, trade.OffsettingTradeDate.Valid, "Trade.OffsettingTradeDate.Valid for 1900 sentinel date")
	assert.True(t, bool(trade.DarkPool), "Trade.DarkPool")
	assert.False(t, bool(trade.Cancelled), "Trade.Cancelled")
	assert.False(
		t,
		resp.Data[1].PhantomPrintFulfillmentDate.Valid,
		"Trade.PhantomPrintFulfillmentDate.Valid for empty date",
	)
}

func TestSessionIsDefensivelyCopied(t *testing.T) {
	fixture := mustReadFixture(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "before", r.Header.Get("X-Custom"), "GetTrades() X-Custom")
		cookie, err := r.Cookie("auth")
		if !assert.NoError(t, err, "GetTrades() auth cookie") {
			return
		}
		assert.Equal(t, "before", cookie.Value, "GetTrades() auth cookie")
		_, _ = w.Write(fixture)
	}))
	t.Cleanup(server.Close)

	cookie := &http.Cookie{Name: "auth", Value: "before"}
	headers := http.Header{"X-Custom": {"before"}}
	client, err := NewClient(
		Session{Cookies: []*http.Cookie{cookie}, Headers: headers},
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
	)
	require.NoError(t, err, "NewClient()")
	cookie.Value = "after"
	headers.Set("X-Custom", "after")

	_, err = client.GetTrades(context.Background(), TradesRequest{})
	require.NoError(t, err, "GetTrades()")
}

func mustReadFixture(t *testing.T) []byte {
	t.Helper()
	data, err := os.ReadFile(tradesFixturePath)
	require.NoError(t, err, "ReadFile(%q)", tradesFixturePath)
	return data
}

func mustMarshalJSON(t *testing.T, value any) string {
	t.Helper()
	data, err := json.Marshal(value)
	require.NoError(t, err, "Marshal(%T)", value)
	return string(data)
}
