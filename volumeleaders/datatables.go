package volumeleaders

import (
	"context"
	"math"
	"net/url"
	"strconv"
)

const (
	defaultDataTablesLength = 25
	defaultOrderDir         = "desc"
	ascendingOrderDir       = "asc"
)

// DataTablesColumn describes one DataTables form column definition.
type DataTablesColumn struct {
	Data       string
	Name       string
	Searchable bool
	Orderable  bool
}

// DataTablesOrder describes one DataTables sort instruction.
type DataTablesOrder struct {
	Column int
	Dir    string
	Name   string
}

// DataTablesRequest describes a server-side DataTables form request.
type DataTablesRequest struct {
	Draw          int
	Start         int
	Length        int
	Columns       []DataTablesColumn
	Order         []DataTablesOrder
	SearchValue   string
	SearchRegex   bool
	IncludeSearch bool
	Extra         url.Values
}

// DataTablesResponse is the typed server-side DataTables JSON envelope.
type DataTablesResponse[T any] struct {
	Draw            int `json:"draw"`
	RecordsTotal    int `json:"recordsTotal"`
	RecordsFiltered int `json:"recordsFiltered"`
	Data            []T `json:"data"`
}

// EncodeDataTablesRequest returns form values for a server-side DataTables
// request using the bracketed field names expected by ASP.NET MVC.
func EncodeDataTablesRequest(req DataTablesRequest) url.Values {
	values := url.Values{}
	draw := req.Draw
	if draw == 0 {
		draw = 1
	}
	length := req.Length
	if length == 0 {
		length = defaultDataTablesLength
	}
	values.Set("draw", strconv.Itoa(draw))
	values.Set("start", strconv.Itoa(req.Start))
	values.Set("length", strconv.Itoa(length))

	for i, column := range req.Columns {
		prefix := "columns[" + strconv.Itoa(i) + "]"
		values.Set(prefix+"[data]", column.Data)
		values.Set(prefix+"[name]", column.Name)
		values.Set(prefix+"[searchable]", strconv.FormatBool(column.Searchable))
		values.Set(prefix+"[orderable]", strconv.FormatBool(column.Orderable))
		values.Set(prefix+"[search][value]", "")
		values.Set(prefix+"[search][regex]", "false")
	}

	orders := req.Order
	if len(orders) == 0 {
		orders = []DataTablesOrder{{Column: 1, Dir: defaultOrderDir}}
	}
	for i, order := range orders {
		dir := order.Dir
		if dir == "" {
			dir = defaultOrderDir
		}
		prefix := "order[" + strconv.Itoa(i) + "]"
		values.Set(prefix+"[column]", strconv.Itoa(order.Column))
		values.Set(prefix+"[dir]", dir)
		if order.Name != "" {
			values.Set(prefix+"[name]", order.Name)
		}
	}

	if req.IncludeSearch {
		values.Set("search[value]", req.SearchValue)
		values.Set("search[regex]", strconv.FormatBool(req.SearchRegex))
	}
	for key, items := range req.Extra {
		for _, item := range items {
			values.Add(key, item)
		}
	}
	return values
}

// fetchLimit pages through a DataTables endpoint, collecting up to limit
// records. When limit is zero or negative the helper fetches all available
// records using the server-reported RecordsFiltered count as the stop
// condition.
func fetchLimit[T any](
	ctx context.Context,
	limit int,
	dt DataTablesRequest,
	fetch func(ctx context.Context, dt DataTablesRequest) (*DataTablesResponse[T], error),
) ([]T, error) {
	if limit <= 0 {
		limit = math.MaxInt
	}

	pageLength := dt.Length
	if pageLength <= 0 || pageLength > limit {
		pageLength = min(limit, defaultDataTablesLength)
	}

	items := make([]T, 0, min(limit, pageLength))
	start := dt.Start
	draw := dt.Draw
	if draw <= 0 {
		draw = 1
	}

	for len(items) < limit {
		remaining := limit - len(items)
		length := min(pageLength, remaining)

		pageDT := dt
		pageDT.Start = start
		pageDT.Length = length
		pageDT.Draw = draw

		resp, err := fetch(ctx, pageDT)
		if err != nil {
			return nil, err
		}
		if len(resp.Data) == 0 {
			break
		}

		data := resp.Data
		if len(data) > remaining {
			data = data[:remaining]
		}
		items = append(items, data...)
		start += len(resp.Data)
		draw++

		if len(resp.Data) < length || (resp.RecordsFiltered > 0 && start >= resp.RecordsFiltered) {
			break
		}
	}
	return items, nil
}
