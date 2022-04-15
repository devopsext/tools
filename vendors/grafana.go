package vendors

import (
	"fmt"
	"io/ioutil"
	"net/http"

	"encoding/base64"

	"github.com/devopsext/utils"
)

type GrafanaOptions struct {
	URL         string
	Timeout     int
	Insecure    bool
	User        string
	Password    string
	Output      string
	OutputQuery string
}

type Grafana struct {
	client  *http.Client
	options GrafanaOptions
}

func (g *Grafana) get(URL string) ([]byte, error) {

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	if !utils.IsEmpty(g.options.User) {
		basic := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s:%s", g.options.User, g.options.Password)))
		req.Header.Set("Authorization", fmt.Sprintf("Basic %s", basic))
	}

	resp, err := g.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

/*func (g *Graylog) searchAbsolute(URL, streams, query, sort string, from, to string, limit int) ([]byte, error) {

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
	if params != nil {
		u.RawQuery = params.Encode()
	}
	return g.get(u.String())
}*/

func (g *Grafana) GetImage() ([]byte, error) {

	return nil, fmt.Errorf("not implemented")
}

func NewGrafana(options GrafanaOptions) *Grafana {

	return &Grafana{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
