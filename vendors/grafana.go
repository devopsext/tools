package vendors

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
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
	MatchAny    bool
}

type GrafanaCreateAnnotationOptions struct {
	Time    string
	TimeEnd string
	Tags    string
	Text    string
}

type GrafanaClonedDahboardOptions struct {
	UID         string
	Annotations []string
	PanelIDs    []string
	PanelSeries []string
	LegendRight bool
	Arrange     bool
	Count       int
	Width       int
	Height      int
}

type GrafanaCreateDahboardOptions struct {
	Title     string
	FolderUID string
	Tags      []string
	From      string
	To        string
	Cloned    GrafanaClonedDahboardOptions
}

type GrafanaOptions struct {
	URL               string
	Timeout           int
	Insecure          bool
	APIKey            string
	OrgID             string
	DashboardUID      string
	DashboardSlug     string
	DashboardTimezone string
}

type GrafanaDashboardTime struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type GrafanaDashboardAnnotations struct {
	List         []interface{} `json:"list"`
	GraphTooltip int           `json:"graphTooltip"`
}

type GrafanaDashboard struct {
	ID            int                         `json:"id"`
	UID           string                      `json:"uid"`
	Title         string                      `json:"title"`
	Tags          []string                    `json:"tags"`
	Timezone      string                      `json:"timezone"`
	SchemaVersion int                         `json:"schemaVersion"`
	Version       int                         `json:"version"`
	GraphTooltip  int                         `json:"graphTooltip"`
	Time          GrafanaDashboardTime        `json:"time"`
	Annotations   GrafanaDashboardAnnotations `json:"annotations"`
	Panels        []interface{}               `json:"panels"`
}

type GrafanaBoard struct {
	Dashboard GrafanaDashboard `json:"dashboard"`
	FolderID  int              `json:"folderId"`
	FolderUID string           `json:"folderUid"`
	Message   string           `json:"message"`
	Overwrite bool             `json:"overwrite"`
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

func (g *Grafana) getAuth(options GrafanaOptions) string {
	auth := ""
	if !utils.IsEmpty(options.APIKey) {
		auth = fmt.Sprintf("Bearer %s", options.APIKey)
	}
	return auth
}

func (g *Grafana) CustomRenderImage(grafanaOptions GrafanaOptions, renderImageOptions GrafanaRenderImageOptions) ([]byte, error) {

	u, err := url.Parse(grafanaOptions.URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("/render/d-solo/%s/%s", grafanaOptions.DashboardUID, grafanaOptions.DashboardSlug))

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
	params.Add("tz", grafanaOptions.DashboardTimezone)

	u.RawQuery = params.Encode()
	return common.HttpGetRaw(g.client, u.String(), "", g.getAuth(grafanaOptions))
}

func (g *Grafana) RenderImage(options GrafanaRenderImageOptions) ([]byte, error) {
	return g.CustomRenderImage(g.options, options)
}

func (g *Grafana) CustomGetDashboards(grafanaOptions GrafanaOptions) ([]byte, error) {
	u, err := url.Parse(grafanaOptions.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("/api/dashboards/uid/%s", grafanaOptions.DashboardUID))
	return common.HttpGetRaw(g.client, u.String(), "", g.getAuth(grafanaOptions))
}

func (g *Grafana) GetDashboards() ([]byte, error) {
	return g.CustomGetDashboards(g.options)
}

func (g Grafana) CustomCreateAnnotation(grafanaOptions GrafanaOptions, createAnnotationOptions GrafanaCreateAnnotationOptions) ([]byte, error) {
	u, err := url.Parse(grafanaOptions.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "/api/annotations")

	b, err := json.Marshal(g.createAnnotation(&createAnnotationOptions))
	if err != nil {
		return nil, err
	}
	return common.HttpPostRaw(g.client, u.String(), "application/json", g.getAuth(grafanaOptions), b)
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

	u.Path = path.Join(u.Path, "/api/annotations")

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
	if getAnnotationsOptions.MatchAny {
		params.Add("matchAny", "true")
	}
	params.Add("tz", grafanaOptions.DashboardTimezone)

	u.RawQuery = params.Encode()
	return common.HttpGetRaw(g.client, u.String(), "", g.getAuth(grafanaOptions))
}

func (g *Grafana) GetAnnotations(options GrafanaGetAnnotationsOptions) ([]byte, error) {
	return g.CustomGetAnnotations(g.options, options)
}

/*
- make dashboard by name if it's not exists yet
- clone the panel to new dashboard (or existed one)
- filter series in the pannel by name (regexp)
- save dashboard with cloned panel
- render panel to chart
*/

func (g Grafana) matchFilter(filter []string, value string) bool {

	for _, v := range filter {
		match, _ := regexp.MatchString(v, value)
		if match {
			return true
		}
	}
	return false
}

func (g Grafana) copyAnnotations(source, dest *GrafanaDashboardAnnotations, names []string) {

	if len(source.List) <= 0 {
		return
	}

	for _, v := range source.List {
		m, ok := v.(map[string]interface{})
		if ok {
			name, ok := m["name"].(string)
			if ok && g.matchFilter(names, name) {
				dest.List = append(dest.List, v)
			}
		}
	}
}

func (g Grafana) panelIsType(pm map[string]interface{}, typ string) bool {
	t, ok := pm["type"].(string)
	if !ok {
		return false
	}
	return t == typ
}

func (g Grafana) setLegend(pm map[string]interface{}, right bool) {

	legend, okLegend := pm["legend"].(map[string]interface{})
	if !okLegend {
		legend = make(map[string]interface{})
	}
	legend["rightSide"] = right
}

func (g Grafana) deleteAlerts(pm map[string]interface{}) {

	_, okAlert := pm["alert"].(map[string]interface{})
	if !okAlert {
		return
	}
	delete(pm, "alert")
}

func (g Grafana) setTransformations(pm map[string]interface{}, pattern string) {

	transformations, okTransformations := pm["transformations"].([]interface{})
	if !okTransformations {
		transformations = []interface{}{}
	}

	transformation := make(map[string]interface{})
	transformation["id"] = "filterFieldsByName"

	options := make(map[string]interface{})
	include := make(map[string]interface{})
	names := []interface{}{"Time"}
	include["names"] = names
	include["pattern"] = pattern
	options["include"] = include
	transformation["options"] = options
	transformations = append(transformations, transformation)

	pm["transformations"] = transformations
}

func (g Grafana) findPanel(source *[]interface{}, ID string) map[string]interface{} {

	for _, p := range *source {
		pm, ok := p.(map[string]interface{})
		if ok {

			id, okID := pm["id"].(float64)
			if okID {
				if fmt.Sprintf("%.f", id) == ID {
					return pm
				} else if g.panelIsType(pm, "row") {
					pnls, okPnls := pm["panels"].([]interface{})
					if okPnls {
						rp := g.findPanel(&pnls, ID)
						if rp != nil {
							return rp
						}
					}
				}
			}
		}
	}
	return nil
}

func (g Grafana) copyPanels(source, dest *[]interface{}, clonedDashboardOptions GrafanaClonedDahboardOptions) {

	if len(*source) <= 0 {
		return
	}

	IDs := clonedDashboardOptions.PanelIDs
	series := clonedDashboardOptions.PanelSeries

	if !utils.IsEmpty(IDs) {
		for idx, id := range IDs {

			pm := g.findPanel(source, id)
			if pm != nil {
				if !g.panelIsType(pm, "row") {
					g.setLegend(pm, clonedDashboardOptions.LegendRight)
					g.deleteAlerts(pm)
					if (len(series) > idx) && !utils.IsEmpty(series[idx]) {
						g.setTransformations(pm, series[idx])
					}
				}
				*dest = append(*dest, pm)
			}
		}
	} else {
		for _, p := range *source {
			pm, ok := p.(map[string]interface{})
			if ok {
				if !g.panelIsType(pm, "row") {
					g.setLegend(pm, clonedDashboardOptions.LegendRight)
					g.deleteAlerts(pm)
				}
				*dest = append(*dest, pm)
			}
		}
	}
}

func (g Grafana) arrangePanels(panels *[]interface{}, clonedDashboardOptions GrafanaClonedDahboardOptions) {

	if len(*panels) <= 0 {
		return
	}

	var xn float64 = 0
	var y float64 = 0
	cnt := 0

	mod := clonedDashboardOptions.Count
	if mod <= 0 {
		mod = 3
	}

	height := clonedDashboardOptions.Height
	if height <= 0 {
		height = 7
	}

	width := clonedDashboardOptions.Width
	if width <= 0 {
		width = 6
	}

	for _, p := range *panels {
		pm, ok := p.(map[string]interface{})
		if ok && !g.panelIsType(pm, "row") {
			gp, okGP := pm["gridPos"].(map[string]interface{})
			if okGP {
				m := cnt % clonedDashboardOptions.Count
				if m == 0 {
					xn = 0
				}
				var w float64 = float64(width)
				gp["h"] = float64(height)
				gp["w"] = w
				gp["x"] = xn
				gp["y"] = y * float64(m)
				pm["gridPos"] = gp
				xn = xn + w
				cnt++
			}
		}
	}
}

func (g Grafana) CustomCreateDashboard(grafanaOptions GrafanaOptions, createDashboardOptions GrafanaCreateDahboardOptions) ([]byte, error) {

	u, err := url.Parse(grafanaOptions.URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "/api/dashboards/db")

	cloned := &GrafanaBoard{}
	if !utils.IsEmpty(createDashboardOptions.Cloned.UID) {
		clonedOpts := GrafanaOptions{
			URL:          grafanaOptions.URL,
			Timeout:      grafanaOptions.Timeout,
			Insecure:     grafanaOptions.Insecure,
			APIKey:       grafanaOptions.APIKey,
			OrgID:        grafanaOptions.OrgID,
			DashboardUID: createDashboardOptions.Cloned.UID,
		}
		b, err := g.CustomGetDashboards(clonedOpts)
		if err != nil {
			return nil, err
		}
		cloned = &GrafanaBoard{}
		err = json.Unmarshal(b, cloned)
		if err != nil {
			return nil, err
		}
	}

	req := &GrafanaBoard{
		FolderID:  0,
		FolderUID: createDashboardOptions.FolderUID,
		Overwrite: false,
	}
	req.Dashboard.Title = createDashboardOptions.Title
	req.Dashboard.Tags = createDashboardOptions.Tags
	req.Dashboard.Timezone = grafanaOptions.DashboardTimezone
	req.Dashboard.GraphTooltip = 1
	req.Dashboard.Time.From = createDashboardOptions.From
	req.Dashboard.Time.To = createDashboardOptions.To

	g.copyAnnotations(&cloned.Dashboard.Annotations, &req.Dashboard.Annotations, createDashboardOptions.Cloned.Annotations)
	g.copyPanels(&cloned.Dashboard.Panels, &req.Dashboard.Panels, createDashboardOptions.Cloned)

	if createDashboardOptions.Cloned.Arrange {
		g.arrangePanels(&req.Dashboard.Panels, createDashboardOptions.Cloned)
	}

	b, err := json.Marshal(&req)
	if err != nil {
		return nil, err
	}
	return common.HttpPostRaw(g.client, u.String(), "application/json", g.getAuth(grafanaOptions), b)
}

func (g *Grafana) CreateDashboard(options GrafanaCreateDahboardOptions) ([]byte, error) {
	return g.CustomCreateDashboard(g.options, options)
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
