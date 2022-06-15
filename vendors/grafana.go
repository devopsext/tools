package vendors

import (
	"encoding/json"
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

type GrafanaGetAnnotationsOptions struct {
	From        string
	To          string
	Tags        string
	Type        string
	Limit       int
	AlertID     int
	DashboardID int
	PanelID     int
}

type GrafanaOptions struct {
	URL                     string
	Timeout                 int
	Insecure                bool
	APIKey                  string
	OrgID                   string
	UID                     string
	Slug                    string
	RenderImageOptions      *GrafanaRenderImageOptions
	GetDashboardsOptions    *GrafanaGetDashboardsOptions
	GetAnnotationsOptions   *GrafanaGetAnnotationsOptions
	CreateAnnotationOptions *GrafanaCreateAnnotationOptions
}

type GrafanaCreateAnnotationOptions struct {
	Time    string
	TimeEnd string
	Tags    string
	Text    string
}

type GrafanaAnnotation struct {
	Time    int64    `json:"time"`
	TimeEnd int64    `json:"timeEnd"`
	Tags    []string `json:"tags"`
	Text    string   `json:"text"`
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
		params.Add("from", toRFC3339NanoStr(opts.RenderImageOptions.From))
	}
	if !utils.IsEmpty(opts.RenderImageOptions.To) {
		params.Add("to", toRFC3339NanoStr(opts.RenderImageOptions.To))
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

func (g Grafana) CreateAnnotation() ([]byte, error) {
	u, err := url.Parse(g.options.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "api/annotations")

	auth := ""
	if !utils.IsEmpty(g.options.APIKey) {
		auth = fmt.Sprintf("Bearer %s", g.options.APIKey)
	}
	b, err := json.Marshal(createAnnotation(g.options.CreateAnnotationOptions))
	if err != nil {
		return nil, err
	}
	return common.HttpPostRaw(g.client, u.String(), "application/json", auth, b)
}

func createAnnotation(o *GrafanaCreateAnnotationOptions) *GrafanaAnnotation {
	t := toRFC3339Nano(o.Time)
	tEnd := toRFC3339Nano(o.TimeEnd)
	if o.TimeEnd == "" {
		tEnd = t
	}

	return &GrafanaAnnotation{
		Time:    t,
		TimeEnd: tEnd,
		Tags:    strings.Split(o.Tags, ","),
		Text:    o.Text,
	}
}

func toRFC3339NanoStr(ts string) string {
	res := ts
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err == nil {
		res = strconv.Itoa(int(t.UTC().UnixMilli()))
	}
	return res
}

func toRFC3339Nano(ts string) int64 {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err == nil {
		return t.UTC().UnixMilli()
	}
	return time.Now().UnixMilli()
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
	if !utils.IsEmpty(options.GetAnnotationsOptions.From) {
		params.Add("from", toRFC3339NanoStr(options.GetAnnotationsOptions.From))
	}
	if !utils.IsEmpty(options.GetAnnotationsOptions.To) {
		params.Add("to", toRFC3339NanoStr(options.GetAnnotationsOptions.To))
	}
	if options.GetAnnotationsOptions.Type == "alert" || options.GetAnnotationsOptions.Type == "annotation" {
		params.Add("type", options.GetAnnotationsOptions.Type)
	}
	if options.GetAnnotationsOptions.Limit > 0 {
		params.Add("limit", strconv.Itoa(options.GetAnnotationsOptions.Limit))
	}
	if options.GetAnnotationsOptions.AlertID > 0 {
		params.Add("alertId", strconv.Itoa(options.GetAnnotationsOptions.AlertID))
	}
	if options.GetAnnotationsOptions.DashboardID > 0 {
		params.Add("dashboardId", strconv.Itoa(options.GetAnnotationsOptions.DashboardID))
	}
	if options.GetAnnotationsOptions.PanelID > 0 {
		params.Add("panelId", strconv.Itoa(options.GetAnnotationsOptions.PanelID))
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
