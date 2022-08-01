package vendors

import (
	"encoding/json"
	"errors"
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

type GrafanaCreateAnnotationOptions struct {
	Time    string
	TimeEnd string
	Tags    string
	Text    string
}

type GrafanaOptions struct {
	URL      string
	Timeout  int
	Insecure bool
	APIKey   string
	OrgID    string
	UID      string
	Slug     string
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

func (g *Grafana) CustomRenderImage(grafanaOptions GrafanaOptions, renderImageOptions GrafanaRenderImageOptions) ([]byte, error) {

	var params = make(url.Values)
	if !utils.IsEmpty(grafanaOptions.OrgID) {
		params.Add("orgId", grafanaOptions.OrgID)
	}
	if !utils.IsEmpty(renderImageOptions.PanelID) {
		params.Add("panelId", renderImageOptions.PanelID)
	}
	if renderImageOptions.Width > 0 {
		params.Add("width", strconv.Itoa(renderImageOptions.Width))
	}
	if renderImageOptions.Height > 0 {
		params.Add("height", strconv.Itoa(renderImageOptions.Height))
	}
	if !utils.IsEmpty(renderImageOptions.From) {
		params.Add("from", g.toRFC3339NanoStr(renderImageOptions.From))
	}
	if !utils.IsEmpty(renderImageOptions.To) {
		params.Add("to", g.toRFC3339NanoStr(renderImageOptions.To))
	}
	params.Add("tz", "UTC")

	u, err := url.Parse(grafanaOptions.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("/render/d-solo/%s/%s", grafanaOptions.UID, grafanaOptions.Slug))
	u.RawQuery = params.Encode()

	auth := ""
	if !utils.IsEmpty(grafanaOptions.APIKey) {
		auth = fmt.Sprintf("Bearer %s", grafanaOptions.APIKey)
	}
	return common.HttpGetRaw(g.client, u.String(), "", auth)
}

func (g *Grafana) RenderImage(options GrafanaRenderImageOptions) ([]byte, error) {
	return g.CustomRenderImage(g.options, options)
}

func (g *Grafana) CustomGetDashboards(grafanaOptions GrafanaOptions) ([]byte, error) {
	u, err := url.Parse(grafanaOptions.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("api/dashboards/uid/%s", grafanaOptions.UID))

	auth := ""
	if !utils.IsEmpty(grafanaOptions.APIKey) {
		auth = fmt.Sprintf("Bearer %s", grafanaOptions.APIKey)
	}
	return common.HttpGetRaw(g.client, u.String(), "", auth)
}

func (g *Grafana) GetDashboards() ([]byte, error) {
	return g.CustomGetDashboards(g.options)
}

func (g Grafana) CustomCreateAnnotation(grafanaOptions GrafanaOptions, createAnnotationOptions GrafanaCreateAnnotationOptions) ([]byte, error) {
	u, err := url.Parse(grafanaOptions.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "api/annotations")

	auth := ""
	if !utils.IsEmpty(grafanaOptions.APIKey) {
		auth = fmt.Sprintf("Bearer %s", g.options.APIKey)
	}
	b, err := json.Marshal(g.createAnnotation(&createAnnotationOptions))
	if err != nil {
		return nil, err
	}
	return common.HttpPostRaw(g.client, u.String(), "application/json", auth, b)
}

func (g *Grafana) createAnnotation(o *GrafanaCreateAnnotationOptions) *GrafanaAnnotation {
	t := g.toRFC3339Nano(o.Time)
	tEnd := g.toRFC3339Nano(o.TimeEnd)
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

func (g *Grafana) toRFC3339NanoStr(ts string) string {
	res := ts
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err == nil {
		res = strconv.Itoa(int(t.UTC().UnixMilli()))
	}
	return res
}

func (g *Grafana) toRFC3339Nano(ts string) int64 {
	t, err := time.Parse(time.RFC3339Nano, ts)
	if err == nil {
		return t.UTC().UnixMilli()
	}
	return time.Now().UnixMilli()
}

func (g *Grafana) CreateAnnotation(options GrafanaCreateAnnotationOptions) ([]byte, error) {
	return g.CustomCreateAnnotation(g.options, options)
}

func (g *Grafana) CustomGetAnnotations(grafanaOptions GrafanaOptions, getAnnotationsOptions GrafanaGetAnnotationsOptions) ([]byte, error) {
	u, err := url.Parse(grafanaOptions.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "api/annotations")

	var params = make(url.Values)
	for _, tag := range strings.Split(getAnnotationsOptions.Tags, ",") {
		if !utils.IsEmpty(tag) {
			params.Add("tags", tag)
		}
	}
	if !utils.IsEmpty(grafanaOptions.OrgID) {
		params.Add("orgId", grafanaOptions.OrgID)
	}
	if !utils.IsEmpty(getAnnotationsOptions.From) {
		params.Add("from", g.toRFC3339NanoStr(getAnnotationsOptions.From))
	}
	if !utils.IsEmpty(getAnnotationsOptions.To) {
		params.Add("to", g.toRFC3339NanoStr(getAnnotationsOptions.To))
	}
	if getAnnotationsOptions.Type == "alert" || getAnnotationsOptions.Type == "annotation" {
		params.Add("type", getAnnotationsOptions.Type)
	}
	if getAnnotationsOptions.Limit > 0 {
		params.Add("limit", strconv.Itoa(getAnnotationsOptions.Limit))
	}
	if getAnnotationsOptions.AlertID > 0 {
		params.Add("alertId", strconv.Itoa(getAnnotationsOptions.AlertID))
	}
	if getAnnotationsOptions.DashboardID > 0 {
		params.Add("dashboardId", strconv.Itoa(getAnnotationsOptions.DashboardID))
	}
	if getAnnotationsOptions.PanelID > 0 {
		params.Add("panelId", strconv.Itoa(getAnnotationsOptions.PanelID))
	}
	params.Add("tz", "UTC")

	u.RawQuery = params.Encode()

	auth := ""
	if !utils.IsEmpty(grafanaOptions.APIKey) {
		auth = fmt.Sprintf("Bearer %s", grafanaOptions.APIKey)
	}
	return common.HttpGetRaw(g.client, u.String(), "", auth)
}

func (g *Grafana) GetAnnotations(options GrafanaGetAnnotationsOptions) ([]byte, error) {
	return g.CustomGetAnnotations(g.options, options)
}

func NewGrafana(options GrafanaOptions) (*Grafana, error) {

	client := utils.NewHttpClient(options.Timeout, options.Insecure)
	if client == nil {
		return nil, errors.New("no http client")
	}

	grafana := &Grafana{
		client:  client,
		options: options,
	}
	return grafana, nil
}
