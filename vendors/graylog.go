package vendors

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"encoding/base64"

	"github.com/devopsext/utils"
)

type GraylogOptions struct {
	URL       string
	Timeout   int
	Insecure  bool
	User      string
	Password  string
	Streams   string
	Query     string
	RangeType string
	Sort      string
	Limit     int
	From      string
	To        string
	Range     string
}

type Graylog struct {
	client  *http.Client
	options GraylogOptions
}

func (g *Graylog) get(URL string) ([]byte, error) {

	auth := ""
	if !utils.IsEmpty(g.options.User) {
		basic := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", g.options.User, g.options.Password)))
		auth = fmt.Sprintf("Basic %s", basic)
	}

	return utils.HttpGetRaw(g.client, URL, "application/json", auth)
}

// https://graylog.some.host/api/search/universal/relative?query=*&range=3600&limit=100&sort=timestamp:desc&pretty=true
func (g *Graylog) searchRelative(URL, streams, query, sort string, ranges string, limit int) ([]byte, error) {

	params := make(url.Values)
	if !utils.IsEmpty(streams) {
		params.Add("streams", streams)
	}
	if !utils.IsEmpty(query) {
		params.Add("query", query)
	}
	if !utils.IsEmpty(sort) {
		params.Add("sort", sort)
	}
	if !utils.IsEmpty(ranges) {
		params.Add("range", ranges)
	}
	if limit > 0 {
		params.Add("limit", strconv.Itoa(limit))
	}

	u, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "/api/search/universal/relative")
	u.RawQuery = params.Encode()

	return g.get(u.String())
}

// https://graylog.some.host/api/search/universal/absolute?query=*&limit=100&to=2022-04-05T10:54:15.354Z&from=2022-04-05T00:00:00.000Z
func (g *Graylog) searchAbsolute(URL, streams, query, sort string, from, to string, limit int) ([]byte, error) {

	params := make(url.Values)
	if !utils.IsEmpty(streams) {
		params.Add("streams", streams)
	}
	if !utils.IsEmpty(query) {
		params.Add("query", query)
	}
	if !utils.IsEmpty(sort) {
		params.Add("sort", sort)
	}
	if !utils.IsEmpty(from) {
		params.Add("from", from)
	}
	if !utils.IsEmpty(to) {
		params.Add("to", to)
	}
	if limit > 0 {
		params.Add("limit", strconv.Itoa(limit))
	}

	u, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "/api/search/universal/absolute")
	u.RawQuery = params.Encode()
	return g.get(u.String())
}

func (g *Graylog) GetLogs() ([]byte, error) {

	switch g.options.RangeType {
	case "relative":
		return g.searchRelative(g.options.URL, g.options.Streams, g.options.Query, g.options.Sort, g.options.Range, g.options.Limit)
	case "absolute":
		return g.searchAbsolute(g.options.URL, g.options.Streams, g.options.Query, g.options.Sort, g.options.From, g.options.To, g.options.Limit)
	default:
		return nil, fmt.Errorf("no range type")
	}
}

func NewGraylog(options GraylogOptions) *Graylog {

	graylog := &Graylog{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
	return graylog
}
