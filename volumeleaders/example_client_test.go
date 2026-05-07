package volumeleaders_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	volumeleaders "github.com/major/volumeleaders-go/volumeleaders"
)

func ExampleNewClient() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != volumeleaders.TradesGetTradesPath {
			http.NotFound(w, r)
			return
		}
		_, _ = w.Write(
			[]byte(
				`{"draw":1,"recordsTotal":1,"recordsFiltered":1,"data":[{"Ticker":"AXP","TradeID":71774613157188}]}`,
			),
		)
	}))
	defer server.Close()

	session := volumeleaders.NewSession("example-session", "example-auth-cookie", "example-xsrf-token")
	client, err := volumeleaders.NewClient(
		session,
		volumeleaders.WithBaseURL(server.URL),
		volumeleaders.WithHTTPClient(server.Client()),
	)
	if err != nil {
		panic(err)
	}

	page, err := client.ListTrades(context.Background(), volumeleaders.TradeQuery{})
	if err != nil {
		panic(err)
	}

	fmt.Printf("records=%d ticker=%s\n", page.RecordsFiltered, page.Trades[0].Ticker)
	// Output:
	// records=1 ticker=AXP
}

func ExampleClient_ListTrades() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("X-Xsrf-Token"); got != "example-xsrf-token" {
			http.Error(w, "missing token", http.StatusUnauthorized)
			return
		}
		if _, err := r.Cookie(volumeleaders.SessionCookieName); err != nil {
			http.Error(w, "missing session cookie", http.StatusUnauthorized)
			return
		}
		if _, err := r.Cookie(volumeleaders.FormsAuthCookieName); err != nil {
			http.Error(w, "missing auth cookie", http.StatusUnauthorized)
			return
		}
		if got := r.FormValue("Tickers"); got != "AXP" {
			http.Error(w, "missing ticker filter", http.StatusBadRequest)
			return
		}
		_, _ = w.Write(
			[]byte(
				`{"draw":7,"recordsTotal":465,"recordsFiltered":1,"data":[{"Ticker":"AXP","TradeID":71774613157188}]}`,
			),
		)
	}))
	defer server.Close()

	client, err := volumeleaders.NewClient(
		volumeleaders.NewSession("example-session", "example-auth-cookie", "example-xsrf-token"),
		volumeleaders.WithBaseURL(server.URL),
		volumeleaders.WithHTTPClient(server.Client()),
	)
	if err != nil {
		panic(err)
	}

	page, err := client.ListTrades(context.Background(), volumeleaders.TradeQuery{
		Draw:    7,
		Length:  50,
		Search:  "AXP",
		Tickers: []string{"AXP"},
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("filtered=%d tradeID=%d\n", page.RecordsFiltered, page.Trades[0].TradeID)
	// Output:
	// filtered=1 tradeID=71774613157188
}

func ExampleEncodeDataTablesRequest() {
	values := volumeleaders.EncodeDataTablesRequest(volumeleaders.DataTablesRequest{
		Draw:          2,
		Start:         25,
		Length:        50,
		Columns:       volumeleaders.TradesColumns(),
		IncludeSearch: true,
		SearchValue:   "AXP",
		Extra:         url.Values{"Tickers": {"AXP"}},
	})

	keys := []string{
		"draw",
		"start",
		"length",
		"columns[2][data]",
		"order[0][column]",
		"order[0][dir]",
		"search[value]",
		"Tickers",
	}
	for _, key := range keys {
		fmt.Printf("%s=%s\n", key, strings.Join(values[key], ","))
	}
	// Output:
	// draw=2
	// start=25
	// length=50
	// columns[2][data]=Ticker
	// order[0][column]=1
	// order[0][dir]=desc
	// search[value]=AXP
	// Tickers=AXP
}

func ExampleIsStatusCode() {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "login required", http.StatusUnauthorized)
	}))
	defer server.Close()

	client, err := volumeleaders.NewClient(
		volumeleaders.Session{},
		volumeleaders.WithBaseURL(server.URL),
		volumeleaders.WithHTTPClient(server.Client()),
	)
	if err != nil {
		panic(err)
	}

	_, err = client.GetTrades(context.Background(), volumeleaders.TradesRequest{})
	if volumeleaders.IsStatusCode(err, http.StatusUnauthorized) {
		fmt.Println("login required")
	}
	// Output:
	// login required
}
