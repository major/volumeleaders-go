package volumeleaders

import (
	"context"
	"io"
	"mime"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const snapshotPriceDelta = 0.0001

func TestGetExhaustionScoresPostsJSONRequest(t *testing.T) {
	var capturedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "GetExhaustionScores() method")
		assert.Equal(t, ExecutiveSummaryGetExhaustionScoresPath, r.URL.Path, "GetExhaustionScores() path")
		assert.Equal(
			t,
			"application/json; charset=UTF-8",
			r.Header.Get("Content-Type"),
			"GetExhaustionScores() Content-Type",
		)
		assert.Equal(
			t,
			"XMLHttpRequest",
			r.Header.Get("X-Requested-With"),
			"GetExhaustionScores() X-Requested-With",
		)
		assert.Equal(t, "token-123", r.Header.Get("X-Xsrf-Token"), "GetExhaustionScores() X-Xsrf-Token")

		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(GetExhaustionScores body)") {
			return
		}
		capturedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"DateKey": 20260501,
			"ExhaustionScoreRank": 4,
			"ExhaustionScoreRank30Day": 8,
			"ExhaustionScoreRank90Day": 11,
			"ExhaustionScoreRank365Day": 22
		}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(
		Session{XSRFToken: "token-123"},
		WithBaseURL(server.URL),
		WithHTTPClient(server.Client()),
	)
	require.NoError(t, err, "NewClient()")

	resp, err := client.GetExhaustionScores(context.Background(), ExhaustionScoresRequest{Date: "2026-05-01"})
	require.NoError(t, err, "GetExhaustionScores()")
	assert.JSONEq(t, `{"Date":"2026-05-01"}`, capturedBody, "GetExhaustionScores() body")
	assert.Equal(t, 20260501, resp.DateKey, "GetExhaustionScores() DateKey")
	assert.Equal(t, 22, resp.ExhaustionScoreRank365Day, "GetExhaustionScores() 365-day rank")
}

func TestDeleteWatchListAcceptsNullResponse(t *testing.T) {
	var capturedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, WatchListConfigsDeleteWatchListPath, r.URL.Path, "DeleteWatchList() path")
		assert.Equal(
			t,
			"application/json; charset=UTF-8",
			r.Header.Get("Content-Type"),
			"DeleteWatchList() Content-Type",
		)
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(DeleteWatchList body)") {
			return
		}
		capturedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`null`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	err = client.DeleteWatchList(context.Background(), DeleteWatchListRequest{WatchListKey: 6282})
	require.NoError(t, err, "DeleteWatchList()")
	assert.JSONEq(t, `{"WatchListKey":6282}`, capturedBody, "DeleteWatchList() body")
}

func TestSaveAlertConfigPostsMultipartForm(t *testing.T) {
	var capturedFields map[string][]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "SaveAlertConfig() method")
		assert.Equal(t, AlertConfigPath, r.URL.Path, "SaveAlertConfig() path")
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if !assert.NoError(t, err, "ParseMediaType(SaveAlertConfig Content-Type)") {
			return
		}
		assert.Equal(t, "multipart/form-data", mediaType, "SaveAlertConfig() Content-Type")
		assert.Equal(t, navigationAccept, r.Header.Get("Accept"), "SaveAlertConfig() Accept")
		assert.Empty(t, r.Header.Get("X-Requested-With"), "SaveAlertConfig() X-Requested-With")
		assert.Empty(t, r.Header.Get(xsrfHeaderName), "SaveAlertConfig() X-XSRF-Token")

		if !assert.NoError(t, r.ParseMultipartForm(64<<10), "ParseMultipartForm(SaveAlertConfig body)") {
			return
		}
		capturedFields = r.MultipartForm.Value
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`ok`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{XSRFToken: "token-123"}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	err = client.SaveAlertConfig(context.Background(), SaveAlertConfigRequest{Fields: url.Values{
		"AlertConfigKey":  {"42089"},
		"Name":            {"Testing 2"},
		"TickerGroup":     {"AllTickers"},
		"Tickers":         {""},
		"TradeRankLTE":    {"0"},
		"OffsettingPrint": {"true", "false"},
	}})
	require.NoError(t, err, "SaveAlertConfig()")
	assert.Equal(t, []string{"42089"}, capturedFields["AlertConfigKey"], "SaveAlertConfig() AlertConfigKey")
	assert.Equal(t, []string{"Testing 2"}, capturedFields["Name"], "SaveAlertConfig() Name")
	assert.Equal(
		t,
		[]string{"true", "false"},
		capturedFields["OffsettingPrint"],
		"SaveAlertConfig() duplicate checkbox field",
	)
	assert.NotContains(t, capturedFields, "__RequestVerificationToken", "SaveAlertConfig() hidden XSRF field")
}

func TestGetAlertConfigsPostsDataTablesRequest(t *testing.T) {
	var capturedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, AlertConfigsGetAlertConfigsPath, r.URL.Path, "GetAlertConfigs() path")
		assert.Equal(
			t,
			"application/x-www-form-urlencoded; charset=UTF-8",
			r.Header.Get("Content-Type"),
			"GetAlertConfigs() Content-Type",
		)
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(GetAlertConfigs body)") {
			return
		}
		capturedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"draw": 2,
			"recordsTotal": 1,
			"recordsFiltered": 1,
			"data": [{"AlertConfigKey": 42088, "Name": "testing 2", "Tickers": "[ALL TICKERS]", "TradeConditions": null, "ClosingTradeConditions": null}]
		}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	resp, err := client.GetAlertConfigs(context.Background(), AlertConfigsRequest{
		DataTables: DataTablesRequest{Draw: 2, Start: 0, Length: 10},
	})
	require.NoError(t, err, "GetAlertConfigs()")
	require.Len(t, resp.Data, 1, "GetAlertConfigs() data")
	assert.Equal(t, 42088, resp.Data[0].AlertConfigKey, "GetAlertConfigs() AlertConfigKey")
	assert.Equal(t, "testing 2", resp.Data[0].Name, "GetAlertConfigs() Name")
	assert.Nil(t, resp.Data[0].TradeConditions, "GetAlertConfigs() TradeConditions")
	assert.Nil(t, resp.Data[0].ClosingTradeConditions, "GetAlertConfigs() ClosingTradeConditions")

	form, err := url.ParseQuery(capturedBody)
	require.NoError(t, err, "ParseQuery(GetAlertConfigs body)")
	checks := map[string]string{
		"draw":             "2",
		"length":           "10",
		"columns[0][data]": "Name",
		"columns[2][data]": "Tickers",
		"columns[3][data]": "Conditions",
	}
	for key, want := range checks {
		assert.Equal(t, want, form.Get(key), "GetAlertConfigs() form[%q]", key)
	}
}

func TestSaveAlertConfigAcceptsRedirectSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case AlertConfigPath:
			http.Redirect(w, r, "/AlertConfigs?ViewMode=Desktop", http.StatusFound)
		case "/AlertConfigs":
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			_, _ = w.Write([]byte(`<html><body><a href="/logout">login session controls</a></body></html>`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	err = client.SaveAlertConfig(context.Background(), SaveAlertConfigRequest{Fields: url.Values{
		"AlertConfigKey": {"42089"},
		"Name":           {"Testing 2"},
	}})
	require.NoError(t, err, "SaveAlertConfig() redirect success")
}

func TestSaveWatchListConfigAcceptsUnfollowedRedirectSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, WatchListConfigPath, r.URL.Path, "SaveWatchListConfig() path")
		w.Header().Set("Location", "/WatchListConfigs?ViewMode=Desktop")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(server.Close)
	noRedirectClient := server.Client()
	noRedirectClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(noRedirectClient))
	require.NoError(t, err, "NewClient()")

	err = client.SaveWatchListConfig(context.Background(), SaveWatchListConfigRequest{Fields: url.Values{
		"SearchTemplateKey": {"6307"},
		"Name":              {"Testing 3"},
	}})
	require.NoError(t, err, "SaveWatchListConfig() unfollowed redirect success")
}

func TestSaveAlertConfigRejectsLoginRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, AlertConfigPath, r.URL.Path, "SaveAlertConfig() path")
		w.Header().Set("Location", "/login")
		w.WriteHeader(http.StatusFound)
	}))
	t.Cleanup(server.Close)
	noRedirectClient := server.Client()
	noRedirectClient.CheckRedirect = func(_ *http.Request, _ []*http.Request) error {
		return http.ErrUseLastResponse
	}

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(noRedirectClient))
	require.NoError(t, err, "NewClient()")

	err = client.SaveAlertConfig(context.Background(), SaveAlertConfigRequest{Fields: url.Values{
		"AlertConfigKey": {"42089"},
		"Name":           {"Testing 2"},
	}})
	require.Error(t, err, "SaveAlertConfig() login redirect")
	assert.True(t, IsAuthError(err), "IsAuthError(SaveAlertConfig login redirect)")
}

func TestDeleteAlertConfigPostsJSONRequest(t *testing.T) {
	var capturedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, AlertConfigsDeleteAlertConfigPath, r.URL.Path, "DeleteAlertConfig() path")
		assert.Equal(
			t,
			"application/json; charset=UTF-8",
			r.Header.Get("Content-Type"),
			"DeleteAlertConfig() Content-Type",
		)
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(DeleteAlertConfig body)") {
			return
		}
		capturedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`42088`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	err = client.DeleteAlertConfig(context.Background(), DeleteAlertConfigRequest{AlertConfigKey: 42088})
	require.NoError(t, err, "DeleteAlertConfig()")
	assert.JSONEq(t, `{"AlertConfigKey":42088}`, capturedBody, "DeleteAlertConfig() body")
}

func TestSaveWatchListConfigPostsMultipartForm(t *testing.T) {
	var capturedFields map[string][]string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, http.MethodPost, r.Method, "SaveWatchListConfig() method")
		assert.Equal(t, WatchListConfigPath, r.URL.Path, "SaveWatchListConfig() path")
		mediaType, _, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
		if !assert.NoError(t, err, "ParseMediaType(SaveWatchListConfig Content-Type)") {
			return
		}
		assert.Equal(t, "multipart/form-data", mediaType, "SaveWatchListConfig() Content-Type")
		assert.Equal(t, navigationAccept, r.Header.Get("Accept"), "SaveWatchListConfig() Accept")
		assert.Empty(t, r.Header.Get("X-Requested-With"), "SaveWatchListConfig() X-Requested-With")
		assert.Empty(t, r.Header.Get(xsrfHeaderName), "SaveWatchListConfig() X-XSRF-Token")

		if !assert.NoError(t, r.ParseMultipartForm(64<<10), "ParseMultipartForm(SaveWatchListConfig body)") {
			return
		}
		capturedFields = r.MultipartForm.Value
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(`ok`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	err = client.SaveWatchListConfig(context.Background(), SaveWatchListConfigRequest{Fields: url.Values{
		"SearchTemplateKey":    {"6307"},
		"Name":                 {"Testing 3 1"},
		"Tickers":              {"SPY,AAPL"},
		"SecurityTypeKey":      {"-1"},
		"NormalPrintsSelected": {"true", "false"},
	}})
	require.NoError(t, err, "SaveWatchListConfig()")
	assert.Equal(t, []string{"6307"}, capturedFields["SearchTemplateKey"], "SaveWatchListConfig() SearchTemplateKey")
	assert.Equal(t, []string{"SPY,AAPL"}, capturedFields["Tickers"], "SaveWatchListConfig() Tickers")
	assert.Equal(
		t,
		[]string{"true", "false"},
		capturedFields["NormalPrintsSelected"],
		"SaveWatchListConfig() duplicate checkbox field",
	)
	assert.NotContains(t, capturedFields, "__RequestVerificationToken", "SaveWatchListConfig() hidden XSRF field")
}

func TestGetWatchListsParsesRSIFields(t *testing.T) {
	var capturedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, WatchListConfigsGetWatchListsPath, r.URL.Path, "GetWatchLists() path")
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(GetWatchLists body)") {
			return
		}
		capturedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"draw": 1,
			"recordsTotal": 1,
			"recordsFiltered": 1,
			"data": [{
				"SearchTemplateKey": 6307,
				"Name": "Testing 3",
				"Tickers": "SPY,AAPL",
				"RSIOverboughtHourly": 70,
				"RSIOverboughtDaily": null,
				"RSIOversoldHourly": 30,
				"RSIOversoldDaily": null,
				"RSIOverboughtHourlySelected": true,
				"RSIOverboughtDailySelected": null,
				"RSIOversoldHourlySelected": false,
				"RSIOversoldDailySelected": null,
				"Conditions": "IgnoreOBD,IgnoreOSH"
			}]
		}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	resp, err := client.GetWatchLists(context.Background(), WatchListsRequest{
		DataTables: DataTablesRequest{Draw: 1, Start: 0, Length: 10},
	})
	require.NoError(t, err, "GetWatchLists()")
	require.Len(t, resp.Data, 1, "GetWatchLists() data")
	assert.Equal(t, 70, *resp.Data[0].RSIOverboughtHourly, "GetWatchLists() RSIOverboughtHourly")
	assert.Nil(t, resp.Data[0].RSIOverboughtDaily, "GetWatchLists() RSIOverboughtDaily")
	assert.Equal(t, 30, *resp.Data[0].RSIOversoldHourly, "GetWatchLists() RSIOversoldHourly")
	assert.Nil(t, resp.Data[0].RSIOversoldDaily, "GetWatchLists() RSIOversoldDaily")
	assert.True(t, *resp.Data[0].RSIOverboughtHourlySelected, "GetWatchLists() RSIOverboughtHourlySelected")
	assert.Nil(t, resp.Data[0].RSIOverboughtDailySelected, "GetWatchLists() RSIOverboughtDailySelected")

	form, err := url.ParseQuery(capturedBody)
	require.NoError(t, err, "ParseQuery(GetWatchLists body)")
	assert.Equal(t, "Name", form.Get("columns[0][data]"), "GetWatchLists() columns[0][data]")
	assert.Equal(t, "Criteria", form.Get("columns[3][data]"), "GetWatchLists() columns[3][data]")
}

func TestGetWatchListTickersParsesCapturedSchema(t *testing.T) {
	var capturedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, WatchListsGetWatchListTickersPath, r.URL.Path, "GetWatchListTickers() path")
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(GetWatchListTickers body)") {
			return
		}
		capturedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"draw": 1,
			"recordsTotal": 1,
			"recordsFiltered": 1,
			"data": [{
				"WatchListKey": 0,
				"SecurityKey": 63,
				"Ticker": "AAPL",
				"Sector": "Technology",
				"Industry": "Hardware, Storage and Peripherals",
				"MarketCap": 4222761723560,
				"NearestTop10TradeDate": "/Date(1766102400000)/",
				"NearestTop10TradePrice": 273.67,
				"NearestTop10TradeVolume": 47350618,
				"NearestTop10TradeDollars": 12958443628.06,
				"NearestTop10TradeRank": 3,
				"NearestTop10TradeClusterDate": "/Date(1766102400000)/",
				"NearestTop10TradeClusterPrice": 273.7,
				"NearestTop10TradeClusterVolume": 81266351,
				"NearestTop10TradeClusterDollars": 22240171565.69,
				"NearestTop10TradeClusterRank": 4,
				"NearestTop10TradeLevelDate": "/Date(1762819200000)/",
				"NearestTop10TradeLevelPrice": 273.7,
				"NearestTop10TradeLevelVolume": 96872511,
				"NearestTop10TradeLevelDollars": 26511266627.7,
				"NearestTop10TradeLevelRank": 3
			}]
		}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	resp, err := client.GetWatchListTickers(context.Background(), WatchListTickersRequest{
		DataTables: DataTablesRequest{Draw: 1, Start: 0, Length: -1},
		Filters:    url.Values{"WatchListKey": {"6260"}},
	})
	require.NoError(t, err, "GetWatchListTickers()")
	require.Len(t, resp.Data, 1, "GetWatchListTickers() data")
	assert.Equal(t, 63, resp.Data[0].SecurityKey, "GetWatchListTickers() SecurityKey")
	assert.Equal(t, "AAPL", resp.Data[0].Ticker, "GetWatchListTickers() Ticker")
	assert.Equal(t, 3, resp.Data[0].NearestTop10TradeLevelRank, "GetWatchListTickers() level rank")

	form, err := url.ParseQuery(capturedBody)
	require.NoError(t, err, "ParseQuery(GetWatchListTickers body)")
	checks := map[string]string{
		"draw":             "1",
		"length":           "-1",
		"columns[0][data]": "Ticker",
		"columns[4][data]": "NearestTop10TradeLevel",
		"WatchListKey":     "6260",
	}
	for key, want := range checks {
		assert.Equal(t, want, form.Get(key), "GetWatchListTickers() form[%q]", key)
	}
}

func TestAddTickerToWatchListPostsCapturedForm(t *testing.T) {
	var capturedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, Chart0UpdateWatchListPath, r.URL.Path, "AddTickerToWatchList() path")
		assert.Equal(
			t,
			"application/x-www-form-urlencoded; charset=UTF-8",
			r.Header.Get("Content-Type"),
			"AddTickerToWatchList() Content-Type",
		)
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(AddTickerToWatchList body)") {
			return
		}
		capturedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"success":true,"message":"Watch List updated!"}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{XSRFToken: "token-123"}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	resp, err := client.AddTickerToWatchList(context.Background(), AddTickerToWatchListRequest{
		WatchListKey: 6276,
		Ticker:       "AMD",
	})
	require.NoError(t, err, "AddTickerToWatchList()")
	assert.True(t, resp.Success, "AddTickerToWatchList() Success")
	assert.Equal(t, "Watch List updated!", resp.Message, "AddTickerToWatchList() Message")

	form, err := url.ParseQuery(capturedBody)
	require.NoError(t, err, "ParseQuery(AddTickerToWatchList body)")
	assert.Equal(t, "6276", form.Get("WatchListKey"), "AddTickerToWatchList() WatchListKey")
	assert.Equal(t, "AMD", form.Get("Ticker"), "AddTickerToWatchList() Ticker")
}

func TestChartTradeLevelsUseCapturedPaths(t *testing.T) {
	paths := []string{}
	bodies := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method, "chart trade levels method")
		assert.Equal(
			t,
			"application/x-www-form-urlencoded; charset=UTF-8",
			r.Header.Get("Content-Type"),
			"chart trade levels Content-Type",
		)
		assert.Equal(t, xmlHTTPRequest, r.Header.Get("X-Requested-With"), "chart trade levels X-Requested-With")
		assert.Equal(t, "token-123", r.Header.Get(xsrfHeaderName), "chart trade levels X-XSRF-Token")
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(chart trade levels body)") {
			return
		}
		bodies = append(bodies, string(body))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"draw":1,"recordsTotal":0,"recordsFiltered":0,"data":[]}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{XSRFToken: "token-123"}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")
	request := TradeLevelsRequest{
		DataTables: DataTablesRequest{
			Draw:          1,
			Start:         0,
			Length:        -1,
			Order:         []DataTablesOrder{{Column: 0, Dir: "DESC", Name: columnPrice}},
			IncludeSearch: true,
		},
		Filters: url.Values{
			"StartDate": {"2026-05-07"},
			"EndDate":   {"2026-05-07"},
			"Ticker":    {"AMD"},
			"Levels":    {"5"},
		},
	}

	_, err = client.GetChart0TradeLevels(context.Background(), request)
	require.NoError(t, err, "GetChart0TradeLevels()")
	_, err = client.GetChartTradeLevels(context.Background(), request)
	require.NoError(t, err, "GetChartTradeLevels()")
	assert.Equal(t, []string{Chart0GetTradeLevelsPath, ChartGetTradeLevelsPath}, paths, "chart trade level paths")
	require.Len(t, bodies, 2, "chart trade levels request bodies")
	for _, body := range bodies {
		form, parseErr := url.ParseQuery(body)
		require.NoError(t, parseErr, "ParseQuery(chart trade levels body)")
		checks := map[string]string{
			"draw":             "1",
			"length":           "-1",
			"columns[0][data]": "Price",
			"columns[1][name]": "$$",
			"columns[7][data]": "Dates",
			"order[0][column]": "0",
			"order[0][dir]":    "DESC",
			"order[0][name]":   "Price",
			"search[value]":    "",
			"search[regex]":    "false",
			"StartDate":        "2026-05-07",
			"EndDate":          "2026-05-07",
			"Ticker":           "AMD",
			"Levels":           "5",
		}
		for key, want := range checks {
			assert.Equal(t, want, form.Get(key), "chart trade levels form[%q]", key)
		}
	}
}

func TestGetTradeAlertsUsesAgentContract(t *testing.T) {
	var capturedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, TradeAlertsGetTradeAlertsPath, r.URL.Path, "GetTradeAlerts() path")
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(GetTradeAlerts body)") {
			return
		}
		capturedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"draw": 4,
			"recordsTotal": 1,
			"recordsFiltered": 1,
			"data": [{"Ticker": "AMD", "TradeID": 123456, "AlertType": "Trade", "Sector": null, "Industry": null, "Sweep": 1, "DarkPool": false}]
		}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	resp, err := client.GetTradeAlerts(context.Background(), TradeAlertsRequest{
		DataTables: DataTablesRequest{Draw: 4, Start: 0, Length: 25},
		Filters:    url.Values{"Date": {"2026-05-07"}},
	})
	require.NoError(t, err, "GetTradeAlerts()")
	require.Len(t, resp.Data, 1, "GetTradeAlerts() data")
	assert.Equal(t, int64(123456), resp.Data[0].TradeID, "GetTradeAlerts() TradeID")
	assert.Equal(t, "Trade", resp.Data[0].AlertType, "GetTradeAlerts() AlertType")
	assert.Nil(t, resp.Data[0].Sector, "GetTradeAlerts() Sector")
	assert.True(t, bool(resp.Data[0].Sweep), "GetTradeAlerts() Sweep")

	form, err := url.ParseQuery(capturedBody)
	require.NoError(t, err, "ParseQuery(GetTradeAlerts body)")
	checks := map[string]string{
		"draw":              "4",
		"columns[0][data]":  "FullTimeString24",
		"columns[4][data]":  "Trade",
		"columns[11][data]": "TradeRank",
		"Date":              "2026-05-07",
	}
	for key, want := range checks {
		assert.Equal(t, want, form.Get(key), "GetTradeAlerts() form[%q]", key)
	}
}

func TestGetTradeClusterAlertsUsesAgentContract(t *testing.T) {
	var capturedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, TradeClusterAlertsGetTradeClusterAlertsPath, r.URL.Path, "GetTradeClusterAlerts() path")
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(GetTradeClusterAlerts body)") {
			return
		}
		capturedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"draw": 5,
			"recordsTotal": 1,
			"recordsFiltered": 1,
			"data": [{"Ticker": "AMD", "TradeClusterRank": 8, "TradeCount": 4}]
		}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	resp, err := client.GetTradeClusterAlerts(context.Background(), TradeClusterAlertsRequest{
		DataTables: DataTablesRequest{Draw: 5, Start: 0, Length: 25},
		Filters:    url.Values{"Date": {"2026-05-07"}},
	})
	require.NoError(t, err, "GetTradeClusterAlerts()")
	require.Len(t, resp.Data, 1, "GetTradeClusterAlerts() data")
	assert.Equal(t, 8, resp.Data[0].TradeClusterRank, "GetTradeClusterAlerts() TradeClusterRank")

	form, err := url.ParseQuery(capturedBody)
	require.NoError(t, err, "ParseQuery(GetTradeClusterAlerts body)")
	checks := map[string]string{
		"draw":              "5",
		"columns[0][data]":  "MinFullTimeString24",
		"columns[12][data]": "TradeClusterRank",
		"Date":              "2026-05-07",
	}
	for key, want := range checks {
		assert.Equal(t, want, form.Get(key), "GetTradeClusterAlerts() form[%q]", key)
	}
}

func TestVolumeLeaderboardsUseAgentContracts(t *testing.T) {
	paths := []string{}
	bodies := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(volume leaderboard body)") {
			return
		}
		bodies = append(bodies, string(body))
		w.Header().Set("Content-Type", "application/json")
		responseData := `{"Ticker": "AMD", "TotalInstitutionalVolume": 1000, "TotalInstitutionalDollars": 12345.67, "TotalInstitutionalDollarsRank": 3}`
		switch r.URL.Path {
		case AHInstitutionalVolumeGetAHInstitutionalVolumePath:
			responseData = `{"Ticker": "AMD", "AHInstitutionalVolume": 500, "AHInstitutionalDollars": 6789.01, "AHInstitutionalDollarsRank": 4}`
		case TotalVolumeGetTotalVolumePath:
			responseData = `{"Ticker": "AMD", "TotalVolume": 2000, "TotalDollars": 12345.67, "TotalDollarsRank": 2}`
		}
		_, _ = w.Write([]byte(`{
			"draw": 1,
			"recordsTotal": 1,
			"recordsFiltered": 1,
			"data": [` + responseData + `]
		}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")
	request := VolumeRequest{
		DataTables: DataTablesRequest{
			Draw:   1,
			Start:  0,
			Length: 100,
			Order:  []DataTablesOrder{{Column: 1, Dir: ascendingOrderDir}},
		},
		Filters: url.Values{"Date": {"2026-05-07"}, "Tickers": {"AMD"}},
	}

	inst, err := client.GetInstitutionalVolume(context.Background(), request)
	require.NoError(t, err, "GetInstitutionalVolume()")
	ah, err := client.GetAHInstitutionalVolume(context.Background(), request)
	require.NoError(t, err, "GetAHInstitutionalVolume()")
	total, err := client.GetTotalVolume(context.Background(), request)
	require.NoError(t, err, "GetTotalVolume()")
	assert.Equal(t, 1000, inst.Data[0].TotalInstitutionalVolume, "GetInstitutionalVolume() volume")
	assert.Equal(t, 500, ah.Data[0].AHInstitutionalVolume, "GetAHInstitutionalVolume() volume")
	assert.Equal(t, 2, total.Data[0].TotalDollarsRank, "GetTotalVolume() rank")
	assert.Equal(t, []string{
		InstitutionalVolumeGetInstitutionalVolumePath,
		AHInstitutionalVolumeGetAHInstitutionalVolumePath,
		TotalVolumeGetTotalVolumePath,
	}, paths, "volume leaderboard paths")
	require.Len(t, bodies, 3, "volume leaderboard request bodies")

	for i, body := range bodies {
		form, parseErr := url.ParseQuery(body)
		require.NoError(t, parseErr, "ParseQuery(volume leaderboard body %d)", i)
		assert.Equal(t, "Ticker", form.Get("columns[0][data]"), "volume form %d columns[0][data]", i)
		assert.Equal(t, "Price", form.Get("columns[2][data]"), "volume form %d columns[2][data]", i)
		assert.Equal(t, "2026-05-07", form.Get("Date"), "volume form %d Date", i)
		assert.Equal(t, "AMD", form.Get("Tickers"), "volume form %d Tickers", i)
	}
	instForm, err := url.ParseQuery(bodies[0])
	require.NoError(t, err, "ParseQuery(institutional volume body)")
	assert.Equal(t, "TotalInstitutionalVolume", instForm.Get("columns[5][data]"), "institutional columns[5][data]")
	assert.Equal(t, "TotalInstitutionalDollars", instForm.Get("columns[6][data]"), "institutional columns[6][data]")
	assert.Equal(t, "TotalInstitutionalDollarsRank", instForm.Get("columns[7][data]"), "institutional columns[7][data]")
	ahForm, err := url.ParseQuery(bodies[1])
	require.NoError(t, err, "ParseQuery(AH institutional volume body)")
	assert.Equal(t, "AHInstitutionalVolume", ahForm.Get("columns[5][data]"), "AH institutional columns[5][data]")
	assert.Equal(t, "AHInstitutionalDollars", ahForm.Get("columns[6][data]"), "AH institutional columns[6][data]")
	assert.Equal(t, "AHInstitutionalDollarsRank", ahForm.Get("columns[7][data]"), "AH institutional columns[7][data]")
	totalForm, err := url.ParseQuery(bodies[2])
	require.NoError(t, err, "ParseQuery(total volume body)")
	assert.Equal(t, "TotalVolume", totalForm.Get("columns[5][data]"), "total columns[5][data]")
	assert.Equal(t, "TotalDollars", totalForm.Get("columns[6][data]"), "total columns[6][data]")
	assert.Equal(t, "TotalDollarsRank", totalForm.Get("columns[7][data]"), "total columns[7][data]")
}

func TestGetTradeClustersPostsCapturedDataTablesColumns(t *testing.T) {
	var capturedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, TradeClustersGetTradeClustersPath, r.URL.Path, "GetTradeClusters() path")
		assert.Equal(
			t,
			"application/x-www-form-urlencoded; charset=UTF-8",
			r.Header.Get("Content-Type"),
			"GetTradeClusters() Content-Type",
		)
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(GetTradeClusters body)") {
			return
		}
		capturedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"draw": 3,
			"recordsTotal": 1,
			"recordsFiltered": 1,
			"data": [{"Ticker": "AAPL", "TradeClusterRank": 7, "TradeCount": 3}]
		}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	resp, err := client.GetTradeClusters(context.Background(), TradeClustersRequest{
		DataTables: DataTablesRequest{Draw: 3, Start: 10, Length: 5},
		Filters:    url.Values{"Tickers": {"AAPL"}, "TradeClusterRank": {"10"}},
	})
	require.NoError(t, err, "GetTradeClusters()")
	require.Len(t, resp.Data, 1, "GetTradeClusters() data")
	assert.Equal(t, "AAPL", resp.Data[0].Ticker, "GetTradeClusters() ticker")
	assert.Equal(t, 7, resp.Data[0].TradeClusterRank, "GetTradeClusters() rank")

	form, err := url.ParseQuery(capturedBody)
	require.NoError(t, err, "ParseQuery(GetTradeClusters body)")
	checks := map[string]string{
		"draw":              "3",
		"start":             "10",
		"length":            "5",
		"columns[0][data]":  "MinFullTimeString24",
		"columns[3][data]":  "TradeCount",
		"columns[12][data]": "TradeClusterRank",
		"order[0][column]":  "1",
		"order[0][dir]":     "desc",
		"Tickers":           "AAPL",
		"TradeClusterRank":  "10",
	}
	for key, want := range checks {
		assert.Equal(t, want, form.Get(key), "GetTradeClusters() form[%q]", key)
	}
}

func TestGetAllSnapshotsParsesTickerPrices(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, TradesGetAllSnapshotsPath, r.URL.Path, "GetAllSnapshots() path")
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(GetAllSnapshots body)") {
			return
		}
		assert.Equal(t, `null`, string(body), "GetAllSnapshots() body")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`"A:114.52;AA:62.67;"`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	snapshots, err := client.GetAllSnapshots(context.Background())
	require.NoError(t, err, "GetAllSnapshots()")
	assert.InDelta(t, 114.52, snapshots["A"], snapshotPriceDelta, "GetAllSnapshots() A price")
	assert.InDelta(t, 62.67, snapshots["AA"], snapshotPriceDelta, "GetAllSnapshots() AA price")
}

func TestParseSnapshotsReportsMalformedItems(t *testing.T) {
	_, err := ParseSnapshots("A:114.52;broken")
	require.Error(t, err, "ParseSnapshots(malformed)")
	assert.Contains(t, err.Error(), "missing separator", "ParseSnapshots() error")
}

func TestGetEarningsPostsCapturedDataTablesColumns(t *testing.T) {
	var capturedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, EarningsGetEarningsPath, r.URL.Path, "GetEarnings() path")
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(GetEarnings body)") {
			return
		}
		capturedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"draw": 6,
			"recordsTotal": 1,
			"recordsFiltered": 1,
			"data": [{"Ticker": "AMD", "Current": 220.25, "TradeCount": 9}]
		}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	resp, err := client.GetEarnings(context.Background(), EarningsRequest{
		DataTables: DataTablesRequest{Draw: 6, Start: 5, Length: 15},
		Filters:    url.Values{"Date": {"2026-05-07"}, "Tickers": {"AMD"}},
	})
	require.NoError(t, err, "GetEarnings()")
	require.Len(t, resp.Data, 1, "GetEarnings() data")
	assert.Equal(t, "AMD", resp.Data[0].Ticker, "GetEarnings() Ticker")
	assert.InDelta(t, 220.25, resp.Data[0].Current, snapshotPriceDelta, "GetEarnings() Current")

	form, err := url.ParseQuery(capturedBody)
	require.NoError(t, err, "ParseQuery(GetEarnings body)")
	checks := map[string]string{
		"draw":             "6",
		"start":            "5",
		"length":           "15",
		"columns[0][data]": "Date",
		"columns[1][data]": "Ticker",
		"columns[5][name]": "Recent Top-100 Trades",
		"columns[8][name]": "Charts",
		"Date":             "2026-05-07",
		"Tickers":          "AMD",
	}
	for key, want := range checks {
		assert.Equal(t, want, form.Get(key), "GetEarnings() form[%q]", key)
	}
}

func TestWelcomeDataTablesEndpointsPostCapturedColumns(t *testing.T) {
	paths := []string{}
	bodies := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(welcome endpoint body)") {
			return
		}
		bodies = append(bodies, string(body))
		w.Header().Set("Content-Type", "application/json")
		responseData := `{"Ticker":"AMD","TradeRank":11}`
		if r.URL.Path == ExecutiveSummaryGetWelcomeTradeClustersPath {
			responseData = `{"Ticker":"AMD","TradeClusterRank":7}`
		}
		_, _ = w.Write([]byte(`{
			"draw": 7,
			"recordsTotal": 1,
			"recordsFiltered": 1,
			"data": [` + responseData + `]
		}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")
	tradeReq := WelcomeTradesRequest{
		DataTables: DataTablesRequest{Draw: 7, Start: 0, Length: 10},
		Filters:    url.Values{"Date": {"2026-05-07"}},
	}
	clusterReq := WelcomeTradeClustersRequest{
		DataTables: DataTablesRequest{Draw: 8, Start: 10, Length: 5},
		Filters:    url.Values{"Date": {"2026-05-08"}},
	}

	trades, err := client.GetWelcomeTrades(context.Background(), tradeReq)
	require.NoError(t, err, "GetWelcomeTrades()")
	clusters, err := client.GetWelcomeTradeClusters(context.Background(), clusterReq)
	require.NoError(t, err, "GetWelcomeTradeClusters()")
	assert.Equal(t, 11, trades.Data[0].TradeRank, "GetWelcomeTrades() TradeRank")
	assert.Equal(t, 7, clusters.Data[0].TradeClusterRank, "GetWelcomeTradeClusters() TradeClusterRank")
	assert.Equal(t, []string{
		ExecutiveSummaryGetWelcomeTradesPath,
		ExecutiveSummaryGetWelcomeTradeClustersPath,
	}, paths, "welcome endpoint paths")
	require.Len(t, bodies, 2, "welcome endpoint request bodies")

	tradeForm, err := url.ParseQuery(bodies[0])
	require.NoError(t, err, "ParseQuery(GetWelcomeTrades body)")
	assert.Equal(t, "TradeRank", tradeForm.Get("columns[1][data]"), "GetWelcomeTrades() columns[1][data]")
	assert.Equal(t, "LastComparibleTradeDate", tradeForm.Get("columns[4][data]"), "GetWelcomeTrades() columns[4][data]")
	assert.Equal(t, "2026-05-07", tradeForm.Get("Date"), "GetWelcomeTrades() Date")

	clusterForm, err := url.ParseQuery(bodies[1])
	require.NoError(t, err, "ParseQuery(GetWelcomeTradeClusters body)")
	assert.Equal(t, "8", clusterForm.Get("draw"), "GetWelcomeTradeClusters() draw")
	assert.Equal(
		t,
		"TradeClusterRank",
		clusterForm.Get("columns[1][data]"),
		"GetWelcomeTradeClusters() columns[1][data]",
	)
	assert.Equal(
		t,
		"LastComparibleTradeClusterDate",
		clusterForm.Get("columns[4][data]"),
		"GetWelcomeTradeClusters() columns[4][data]",
	)
	assert.Equal(t, "2026-05-08", clusterForm.Get("Date"), "GetWelcomeTradeClusters() Date")
}

func TestGetTradeLevelEndpointsPostCapturedColumns(t *testing.T) {
	paths := []string{}
	bodies := []string{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(trade level endpoint body)") {
			return
		}
		bodies = append(bodies, string(body))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"draw": 9,
			"recordsTotal": 1,
			"recordsFiltered": 1,
			"data": [{"Ticker":"AMD","Price":220.25,"TradeLevelRank":6}]
		}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")
	levelsReq := TradeLevelsRequest{
		DataTables: DataTablesRequest{Draw: 9, Start: 0, Length: 10},
		Filters:    url.Values{"Tickers": {"AMD"}},
	}
	touchesReq := TradeLevelTouchesRequest{
		DataTables: DataTablesRequest{Draw: 10, Start: 5, Length: 15},
		Filters:    url.Values{"Tickers": {"AMD"}, "StartDate": {"2026-05-07"}},
	}

	levels, err := client.GetTradeLevels(context.Background(), levelsReq)
	require.NoError(t, err, "GetTradeLevels()")
	touches, err := client.GetTradeLevelTouches(context.Background(), touchesReq)
	require.NoError(t, err, "GetTradeLevelTouches()")
	assert.Equal(t, 6, levels.Data[0].TradeLevelRank, "GetTradeLevels() TradeLevelRank")
	assert.Equal(t, 6, touches.Data[0].TradeLevelRank, "GetTradeLevelTouches() TradeLevelRank")
	assert.Equal(
		t,
		[]string{TradeLevelsGetTradeLevelsPath, TradeLevelTouchesGetTradeLevelTouchesPath},
		paths,
		"trade level paths",
	)
	require.Len(t, bodies, 2, "trade level request bodies")

	levelsForm, err := url.ParseQuery(bodies[0])
	require.NoError(t, err, "ParseQuery(GetTradeLevels body)")
	assert.Equal(t, "Price", levelsForm.Get("columns[0][data]"), "GetTradeLevels() columns[0][data]")
	assert.Equal(t, "TradeLevelRank", levelsForm.Get("columns[6][data]"), "GetTradeLevels() columns[6][data]")
	assert.Equal(t, "AMD", levelsForm.Get("Tickers"), "GetTradeLevels() Tickers")

	touchesForm, err := url.ParseQuery(bodies[1])
	require.NoError(t, err, "ParseQuery(GetTradeLevelTouches body)")
	assert.Equal(t, "10", touchesForm.Get("draw"), "GetTradeLevelTouches() draw")
	assert.Equal(t, "FullDateTime", touchesForm.Get("columns[0][data]"), "GetTradeLevelTouches() columns[0][data]")
	assert.Empty(t, touchesForm.Get("columns[12][data]"), "GetTradeLevelTouches() trailing action column")
	assert.Equal(t, "2026-05-07", touchesForm.Get("StartDate"), "GetTradeLevelTouches() StartDate")
}

func TestGetTradeClusterBombsPostsCapturedColumns(t *testing.T) {
	var capturedBody string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, TradeClusterBombsGetTradeClusterBombsPath, r.URL.Path, "GetTradeClusterBombs() path")
		body, err := io.ReadAll(r.Body)
		if !assert.NoError(t, err, "ReadAll(GetTradeClusterBombs body)") {
			return
		}
		capturedBody = string(body)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"draw": 11,
			"recordsTotal": 1,
			"recordsFiltered": 1,
			"data": [{"Ticker":"AMD","TradeClusterBombRank":4,"TradeCount":12}]
		}`))
	}))
	t.Cleanup(server.Close)

	client, err := NewClient(Session{}, WithBaseURL(server.URL), WithHTTPClient(server.Client()))
	require.NoError(t, err, "NewClient()")

	resp, err := client.GetTradeClusterBombs(context.Background(), TradeClusterBombsRequest{
		DataTables: DataTablesRequest{Draw: 11, Start: 20, Length: 25},
		Filters:    url.Values{"Tickers": {"AMD"}, "TradeClusterBombRank": {"5"}},
	})
	require.NoError(t, err, "GetTradeClusterBombs()")
	require.Len(t, resp.Data, 1, "GetTradeClusterBombs() data")
	assert.Equal(t, 4, resp.Data[0].TradeClusterBombRank, "GetTradeClusterBombs() TradeClusterBombRank")

	form, err := url.ParseQuery(capturedBody)
	require.NoError(t, err, "ParseQuery(GetTradeClusterBombs body)")
	checks := map[string]string{
		"draw":                 "11",
		"start":                "20",
		"columns[0][data]":     "MinFullTimeString24",
		"columns[9][data]":     "TradeClusterBombRank",
		"columns[11][data]":    "LastComparableTradeClusterBombDate",
		"Tickers":              "AMD",
		"TradeClusterBombRank": "5",
	}
	for key, want := range checks {
		assert.Equal(t, want, form.Get(key), "GetTradeClusterBombs() form[%q]", key)
	}
}
