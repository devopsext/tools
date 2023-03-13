package vendors

import (
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
)

type PrometheusOptions struct {
	URL      string
	Timeout  int
	Insecure bool
	Query    string
	From     string
	To       string
	Step     string
	Params   string
}
type PrometheusOutputOptions struct {
	Output      string
	OutputQuery string
}

type Prometheus struct {
	client  *http.Client
	options PrometheusOptions
}

func (p *Prometheus) toPrometheusTimestamp(ts string) string {
	res := ts
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err == nil {
		res = strconv.Itoa(int(t.UTC().Unix()))
	}
	return res
}

func (p *Prometheus) CustomGet(options PrometheusOptions) ([]byte, error) {

	params := make(url.Values)
	params.Add("query", options.Query)

	if !utils.IsEmpty(options.From) {
		params.Add("start", p.toPrometheusTimestamp(options.From))
	}
	if !utils.IsEmpty(options.To) {
		params.Add("end", p.toPrometheusTimestamp(options.To))
	}
	if !utils.IsEmpty(options.Step) {
		params.Add("step", options.Step)
	}

	vls, err := url.ParseQuery(options.Params)
	if err == nil {
		for k, arr := range vls {
			for _, v := range arr {
				params.Add(k, v)
			}
		}
	}

	u, err := url.Parse(options.URL)
	if err != nil {
		return nil, err
	}

	apiURL := "/api/v1/query_range"
	if utils.IsEmpty(options.From) && utils.IsEmpty(options.To) {
		apiURL = "/api/v1/query"
	}

	u.Path = path.Join(u.Path, apiURL)
	u.RawQuery = params.Encode()

	return common.HttpGetRaw(p.client, u.String(), "application/json", "")
}

func (p *Prometheus) Get() ([]byte, error) {
	return p.CustomGet(p.options)
}

func NewPrometheus(options PrometheusOptions) *Prometheus {
	return &Prometheus{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
