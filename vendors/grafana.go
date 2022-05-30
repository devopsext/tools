package vendors

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
)

type GrafanaRenderImageOptions struct {
	Width  int
	Height int
}

type GrafanaGetAnnotationsOptions struct {
	Tags string
}

type GrafanaOptions struct {
	PanelID               string
	URL                   string
	Timeout               int
	Insecure              bool
	APIKey                string
	OrgID                 string
	From                  string
	To                    string
	UID                   string
	Slug                  string
	RenderImageOptions    *GrafanaRenderImageOptions
	GetAnnotationsOptions *GrafanaGetAnnotationsOptions
}

type Grafana struct {
	client  *http.Client
	options GrafanaOptions
}

func (g *Grafana) RenderCustomImage(opts GrafanaOptions) ([]byte, error) {
	if opts.RenderImageOptions == nil {
		return nil, fmt.Errorf("options are not enough")
	}

	var params = make(url.Values)
	if !utils.IsEmpty(opts.OrgID) {
		params.Add("orgId", opts.OrgID)
	}
	if !utils.IsEmpty(opts.PanelID) {
		params.Add("panelId", opts.PanelID)
	}
	if opts.RenderImageOptions.Width > 0 {
		params.Add("width", strconv.Itoa(opts.RenderImageOptions.Width))
	}
	if opts.RenderImageOptions.Height > 0 {
		params.Add("height", strconv.Itoa(opts.RenderImageOptions.Height))
	}
	if !utils.IsEmpty(opts.From) {
		params.Add("from", toRFC3339Nano(opts.From))
	}
	if !utils.IsEmpty(opts.To) {
		params.Add("to", toRFC3339Nano(opts.To))
	}
	params.Add("tz", "UTC")

	u, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("/render/d-solo/%s/%s", opts.UID, opts.Slug))
	u.RawQuery = params.Encode()

	auth := ""
	if !utils.IsEmpty(opts.APIKey) {
		auth = fmt.Sprintf("Bearer %s", opts.APIKey)
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
	if !utils.IsEmpty(opts.APIKey) {
		auth = fmt.Sprintf("Bearer %s", opts.APIKey)
	}
	return common.HttpGetRaw(g.client, u.String(), "", auth)
}

func (g *Grafana) GetDashboards() ([]byte, error) {
	return g.GetCustomDashboards(g.options)
}

func (g *Grafana) GetAnnotations() ([]byte, error) {
	return g.GetCustomAnnotations(g.options)
}

func toRFC3339Nano(ts string) string {
	res := ts
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err == nil {
		res = strconv.Itoa(int(t.UTC().UnixMilli()))
	}
	return res
}

func (g *Grafana) GetCustomAnnotations(options GrafanaOptions) ([]byte, error) {
	u, err := url.Parse(options.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "api/annotations")

	var params = make(url.Values)
	for _, tag := range strings.Split(options.GetAnnotationsOptions.Tags, ",") {
		if !utils.IsEmpty(tag) {
			params.Add("tags", tag)
		}
	}
	if !utils.IsEmpty(options.OrgID) {
		params.Add("orgId", options.OrgID)
	}
	if !utils.IsEmpty(options.From) {
		params.Add("from", toRFC3339Nano(options.From))
	}
	if !utils.IsEmpty(options.To) {
		params.Add("to", toRFC3339Nano(options.To))
	}
	params.Add("tz", "UTC")

	u.RawQuery = params.Encode()

	auth := ""
	if !utils.IsEmpty(options.APIKey) {
		auth = fmt.Sprintf("Bearer %s", options.APIKey)
	}
	return common.HttpGetRaw(g.client, u.String(), "", auth)
}

func NewGrafana(options GrafanaOptions) *Grafana {
	return &Grafana{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
