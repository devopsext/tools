package vendors

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
)

type GrafanaRenderImageOptions struct {
	PanelID string
	From    string
	To      string
	Width   int
	Height  int
}

type GrafanaGetDashboardsOptions struct {
	PanelID string
	From    string
	To      string
}

type GrafanaOptions struct {
	URL                  string
	Timeout              int
	Insecure             bool
	ApiKey               string
	OrgID                string
	UID                  string
	Slug                 string
	RenderImageOptions   *GrafanaRenderImageOptions
	GetDashboardsOptions *GrafanaGetDashboardsOptions
}

type Grafana struct {
	client  *http.Client
	options GrafanaOptions
}

func (g *Grafana) RenderCustomImage(opts GrafanaOptions) ([]byte, error) {

	if opts.RenderImageOptions == nil {
		return nil, fmt.Errorf("options are not enough")
	}

	params := make(url.Values)
	if !utils.IsEmpty(opts.OrgID) {
		params.Add("orgId", opts.OrgID)
	}
	if !utils.IsEmpty(opts.RenderImageOptions.PanelID) {
		params.Add("panelId", opts.RenderImageOptions.PanelID)
	}
	if opts.RenderImageOptions.Width > 0 {
		params.Add("width", strconv.Itoa(opts.RenderImageOptions.Width))
	}
	if opts.RenderImageOptions.Height > 0 {
		params.Add("height", strconv.Itoa(opts.RenderImageOptions.Height))
	}
	if !utils.IsEmpty(opts.RenderImageOptions.From) {
		from := opts.RenderImageOptions.From
		t, err := time.Parse(time.RFC3339Nano, from)
		if err == nil {
			from = strconv.Itoa(int(t.UTC().UnixMilli()))
		}
		params.Add("from", from)
	}
	if !utils.IsEmpty(opts.RenderImageOptions.To) {
		to := opts.RenderImageOptions.To
		t, err := time.Parse(time.RFC3339Nano, to)
		if err == nil {
			to = strconv.Itoa(int(t.UTC().UnixMilli()))
		}
		params.Add("to", to)
	}
	params.Add("tz", "UTC")

	u, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("/render/d-solo/%s/%s", opts.UID, opts.Slug))
	if params != nil {
		u.RawQuery = params.Encode()
	}

	auth := ""
	if !utils.IsEmpty(opts.ApiKey) {
		auth = fmt.Sprintf("Bearer %s", opts.ApiKey)
	}
	return common.HttpGetRaw(g.client, u.String(), "", auth)
}

func (g *Grafana) RenderImage() ([]byte, error) {
	return g.RenderCustomImage(g.options)
}

func (g *Grafana) GetCustomDashboards(opts GrafanaOptions) ([]byte, error) {

	u, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("api/dashboards/uid/%s", opts.UID))

	auth := ""
	if !utils.IsEmpty(opts.ApiKey) {
		auth = fmt.Sprintf("Bearer %s", opts.ApiKey)
	}
	return common.HttpGetRaw(g.client, u.String(), "", auth)
}

func (g *Grafana) GetDashboards() ([]byte, error) {
	return g.GetCustomDashboards(g.options)
}

func NewGrafana(options GrafanaOptions) *Grafana {

	return &Grafana{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
