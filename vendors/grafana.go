package vendors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"regexp"
	"strconv"
	"strings"
	"time"

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
	URL         string
	Timeout     int
	Insecure    bool
	APIKey      string
	OrgID       string
	UID         string
	FolderUID   string
	FolderID    int
	Annotations []string
	PanelIDs    []string
	PanelTitles []string
	PanelSeries []string
	LegendRight bool
	Arrange     bool
	Count       int
	Width       int
	Height      int
}

type GrafanaDahboardOptions struct {
	Title     string
	UID       string
	Slug      string
	Timezone  string
	FolderUID string
	FolderID  int
	Tags      []string
	From      string
	To        string
	SaveUID   bool
	Overwrite bool
	Cloned    GrafanaClonedDahboardOptions
}

type GrafanaLibraryElementOptions struct {
	Name     string
	UID      string
	FolderID int
	Kind     string
	SaveUID  bool
	Cloned   GrafanaClonedLibraryElementOptions
}

type GrafanaClonedLibraryElementOptions struct {
	URL      string
	Timeout  int
	Insecure bool
	APIKey   string
	OrgID    string
	Name     string
	UID      string
	FolderID int
	Kind     string
}

type GrafanaOptions struct {
	URL      string
	Timeout  int
	Insecure bool
	APIKey   string
	OrgID    string
}

type GrafanaDashboardTime struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type GrafanaDashboardAnnotations struct {
	List         []interface{} `json:"list"`
	GraphTooltip int           `json:"graphTooltip"`
}

type GrafanaDashboardTemplating struct {
	List []interface{} `json:"list"`
}

type GrafanaDashboard struct {
	ID            *int                        `json:"id,omitempty"`
	UID           string                      `json:"uid,omitempty"`
	Title         string                      `json:"title,omitempty"`
	Tags          []string                    `json:"tags,omitempty"`
	Timezone      string                      `json:"timezone,omitempty"`
	SchemaVersion int                         `json:"schemaVersion,omitempty"`
	Version       int                         `json:"version,omitempty"`
	GraphTooltip  int                         `json:"graphTooltip,omitempty"`
	Time          GrafanaDashboardTime        `json:"time,omitempty"`
	Annotations   GrafanaDashboardAnnotations `json:"annotations,omitempty"`
	Templating    GrafanaDashboardTemplating  `json:"templating,omitempty"`
	Panels        []interface{}               `json:"panels,omitempty"`
}

type GrafanaLibraryElementSearchResult struct {
	Result struct {
		TotalCount int                     `json:"totalCount,omitempty"`
		Elements   []GrafanaLibraryElement `json:"elements,omitempty"`
		Page       int                     `json:"page,omitempty"`
		PerPage    int                     `json:"perPage,omitempty"`
	} `json:"result,omitempty"`
}

type GrafanaLibraryElementResult struct {
	Result GrafanaLibraryElement `json:"result,omitempty"`
}

type GrafanaLibraryElement struct {
	ID          int                       `json:"id,omitempty"`
	OrgID       int                       `json:"orgId,omitempty"`
	FolderID    int                       `json:"folderId,omitempty"`
	UID         string                    `json:"uid,omitempty"`
	Name        string                    `json:"name,omitempty"`
	Kind        int                       `json:"kind,omitempty"`
	Type        string                    `json:"type,omitempty"`
	Description string                    `json:"description,omitempty"`
	Model       interface{}               `json:"model,omitempty"`
	Version     int                       `json:"version,omitempty"`
	Meta        GrafanaLibraryElementMeta `json:"meta,omitempty"`
}

type GrafanaLibraryElementMeta struct {
	FolderName          string                         `json:"folderName,,omitempty"`
	FolderUID           string                         `json:"folderUid,omitempty"`
	ConnectedDashboards int                            `json:"connectedDashboards,omitempty"`
	Created             time.Time                      `json:"created,omitempty"`
	Updated             time.Time                      `json:"updated,omitempty"`
	CreatedBy           GrafanaLibraryElementCreatedBy `json:"createdBy,omitempty"`
	UpdatedBy           GrafanaLibraryElementUpdatedBy `json:"updatedBy,omitempty"`
}

type GrafanaLibraryElementCreatedBy struct {
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatarUrl,omitempty"`
}

type GrafanaLibraryElementUpdatedBy struct {
	ID        int    `json:"id,omitempty"`
	Name      string `json:"name,omitempty"`
	AvatarURL string `json:"avatarUrl,omitempty"`
}

type GrafanaBoard struct {
	Dashboard GrafanaDashboard `json:"dashboard,omitempty"`
	FolderID  int              `json:"folderId,omitempty"`
	FolderUID string           `json:"folderUid,omitempty"`
	Message   string           `json:"message,omitempty"`
	Overwrite bool             `json:"overwrite,omitempty"`
	Meta      DashboardMeta    `json:"meta,omitempty"`
}

type DashboardMeta struct {
	IsStarred bool   `json:"isStarred"`
	Slug      string `json:"slug"`
	Folder    *int64 `json:"folderId"`
	FolderUID string `json:"folderUid"`
	URL       string `json:"url"`
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

func (g *Grafana) CustomRenderImage(grafanaOptions GrafanaOptions, grafanaDashboardOptions GrafanaDahboardOptions, renderImageOptions GrafanaRenderImageOptions) ([]byte, error) {

	u, err := url.Parse(grafanaOptions.URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("/render/d-solo/%s/%s", grafanaDashboardOptions.UID, grafanaDashboardOptions.Slug))

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
	params.Add("tz", grafanaDashboardOptions.Timezone)

	u.RawQuery = params.Encode()
	return utils.HttpGetRaw(g.client, u.String(), "", g.getAuth(grafanaOptions))
}

func (g *Grafana) RenderImage(dashboardOptions GrafanaDahboardOptions, renderOptions GrafanaRenderImageOptions) ([]byte, error) {
	return g.CustomRenderImage(g.options, dashboardOptions, renderOptions)
}

func (g *Grafana) CustomGetLibraryElement(grafanaOptions GrafanaOptions, grafanaLibraryElementOptions GrafanaLibraryElementOptions) ([]byte, error) {
	u, err := url.Parse(grafanaOptions.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("/api/library-elements/%s", grafanaLibraryElementOptions.UID))
	result, err := utils.HttpGetRaw(g.client, u.String(), "", g.getAuth(grafanaOptions))
	if err != nil {
		return nil, err
	}

	libraryElementResult := &GrafanaLibraryElementResult{}
	err = json.Unmarshal(result, libraryElementResult)
	if err != nil {
		return nil, err
	}

	libraryElement := libraryElementResult.Result
	return json.Marshal(libraryElement)
}

func (g *Grafana) GetLibraryElement(libraryElementOptions GrafanaLibraryElementOptions) ([]byte, error) {
	return g.CustomGetLibraryElement(g.options, libraryElementOptions)
}

func (g *Grafana) CustomGetDashboards(grafanaOptions GrafanaOptions, grafanaDashboardOptions GrafanaDahboardOptions) ([]byte, error) {
	u, err := url.Parse(grafanaOptions.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("/api/dashboards/uid/%s", grafanaDashboardOptions.UID))
	return utils.HttpGetRaw(g.client, u.String(), "", g.getAuth(grafanaOptions))
}

func (g *Grafana) GetDashboards(dashboardOptions GrafanaDahboardOptions) ([]byte, error) {
	return g.CustomGetDashboards(g.options, dashboardOptions)
}

func (g *Grafana) CustomDeleteDashboards(grafanaOptions GrafanaOptions, grafanaDashboardOptions GrafanaDahboardOptions) ([]byte, error) {
	u, err := url.Parse(grafanaOptions.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("/api/dashboards/uid/%s", grafanaDashboardOptions.UID))
	return utils.HttpDeleteRaw(g.client, u.String(), "application/json", g.getAuth(grafanaOptions), []byte{})
}

func (g *Grafana) DeleteDashboards(dashboardOptions GrafanaDahboardOptions) ([]byte, error) {
	return g.CustomDeleteDashboards(g.options, dashboardOptions)
}

func (g *Grafana) CustomSearchDashboards(grafanaOptions GrafanaOptions, grafanaDashboardOptions GrafanaDahboardOptions) ([]byte, error) {
	u, err := url.Parse(grafanaOptions.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "/api/search")

	var params = make(url.Values)

	switch {
	case !utils.IsEmpty(grafanaDashboardOptions.FolderUID):
		params.Add("folderUIDs", grafanaDashboardOptions.FolderUID)
	case grafanaDashboardOptions.FolderID != 0:
		params.Add("folderIds", strconv.Itoa(grafanaDashboardOptions.FolderID))
	case !utils.IsEmpty(grafanaDashboardOptions.UID):
		params.Add("dashboardUIDs", grafanaDashboardOptions.FolderUID)
	}

	u.RawQuery = params.Encode()

	return utils.HttpGetRaw(g.client, u.String(), "", g.getAuth(grafanaOptions))
}

func (g *Grafana) SearchDashboards(dashboardOptions GrafanaDahboardOptions) ([]byte, error) {
	return g.CustomSearchDashboards(g.options, dashboardOptions)
}

func (g *Grafana) CustomSearchLibraryElements(grafanaOptions GrafanaOptions, grafanaLibraryElementOptions GrafanaLibraryElementOptions) ([]byte, error) {
	u, err := url.Parse(grafanaOptions.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "/api/library-elements")

	var params = make(url.Values)

	if !utils.IsEmpty(grafanaLibraryElementOptions.FolderID) {
		params.Add("folderFilter", strconv.Itoa(grafanaLibraryElementOptions.FolderID))
	}

	u.RawQuery = params.Encode()

	result, err := utils.HttpGetRaw(g.client, u.String(), "", g.getAuth(grafanaOptions))
	if err != nil {
		return nil, err
	}

	libraryElementsResult := &GrafanaLibraryElementSearchResult{}
	err = json.Unmarshal(result, libraryElementsResult)
	if err != nil {
		return nil, err
	}

	libraryElement := libraryElementsResult.Result.Elements
	return json.Marshal(libraryElement)
}

func (g *Grafana) SearchLibraryElements(libraryElementOptions GrafanaLibraryElementOptions) ([]byte, error) {
	return g.CustomSearchLibraryElements(g.options, libraryElementOptions)
}

func (g Grafana) CustomCopyDashboard(grafanaOptions GrafanaOptions, grafanaDashboardOptions GrafanaDahboardOptions) ([]byte, error) {
	u, err := url.Parse(grafanaOptions.URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "/api/dashboards/db")

	copyDashboard := &GrafanaBoard{}
	if !utils.IsEmpty(grafanaDashboardOptions.Cloned.UID) {

		copyOpts := grafanaOptions
		copyDashboardOpts := grafanaDashboardOptions
		copyDashboardOpts.UID = grafanaDashboardOptions.Cloned.UID

		if !utils.IsEmpty(grafanaDashboardOptions.Cloned.URL) {
			copyOpts.URL = grafanaDashboardOptions.Cloned.URL
			copyOpts.Timeout = grafanaDashboardOptions.Cloned.Timeout
			copyOpts.Insecure = grafanaDashboardOptions.Cloned.Insecure
			copyOpts.APIKey = grafanaDashboardOptions.Cloned.APIKey
			copyOpts.OrgID = grafanaDashboardOptions.Cloned.OrgID
		}

		b, err := g.CustomGetDashboards(copyOpts, copyDashboardOpts)
		if err != nil {
			return nil, err
		}
		copyDashboard = &GrafanaBoard{}
		err = json.Unmarshal(b, copyDashboard)
		if err != nil {
			return nil, err
		}
	}

	copyDashboard.FolderID = 0
	copyDashboard.Overwrite = grafanaDashboardOptions.Overwrite
	copyDashboard.Meta.Folder = nil
	copyDashboard.Meta.FolderUID = ""
	copyDashboard.Dashboard.ID = nil

	if !grafanaDashboardOptions.SaveUID {
		copyDashboard.Dashboard.UID = ""
	}
	if !utils.IsEmpty(grafanaDashboardOptions.Title) {
		copyDashboard.Dashboard.Title = grafanaDashboardOptions.Title
	}
	if !utils.IsEmpty(grafanaDashboardOptions.FolderUID) {
		copyDashboard.FolderUID = grafanaDashboardOptions.FolderUID
	}
	if !utils.IsEmpty(grafanaDashboardOptions.FolderID) {
		copyDashboard.FolderID = grafanaDashboardOptions.FolderID
		copyDashboard.FolderUID = ""
	}
	if !utils.IsEmpty(grafanaDashboardOptions.Tags) {
		copyDashboard.Dashboard.Tags = grafanaDashboardOptions.Tags
	}

	b, err := json.Marshal(copyDashboard)
	if err != nil {
		return nil, err
	}

	return utils.HttpPostRaw(g.client, u.String(), "application/json", g.getAuth(grafanaOptions), b)
}

func (g *Grafana) CopyDashboard(grafanaCreateOptions GrafanaDahboardOptions) ([]byte, error) {
	return g.CustomCopyDashboard(g.options, grafanaCreateOptions)
}

func (g Grafana) CustomCopyLibraryElement(grafanaOptions GrafanaOptions, grafanaLibraryElementOptions GrafanaLibraryElementOptions) ([]byte, error) {
	u, err := url.Parse(grafanaOptions.URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "/api/library-elements")

	copyLibraryElement := &GrafanaLibraryElement{}
	if !utils.IsEmpty(grafanaLibraryElementOptions.Cloned.UID) {

		copyOpts := grafanaOptions
		copyLibraryElementsOpts := grafanaLibraryElementOptions
		copyLibraryElementsOpts.UID = grafanaLibraryElementOptions.Cloned.UID

		if !utils.IsEmpty(grafanaLibraryElementOptions.Cloned.URL) {
			copyOpts.URL = grafanaLibraryElementOptions.Cloned.URL
			copyOpts.Timeout = grafanaLibraryElementOptions.Cloned.Timeout
			copyOpts.Insecure = grafanaLibraryElementOptions.Cloned.Insecure
			copyOpts.APIKey = grafanaLibraryElementOptions.Cloned.APIKey
			copyOpts.OrgID = grafanaLibraryElementOptions.Cloned.OrgID
		}

		b, err := g.CustomGetLibraryElement(copyOpts, copyLibraryElementsOpts)
		if err != nil {
			return nil, err
		}
		copyLibraryElement = &GrafanaLibraryElement{}
		err = json.Unmarshal(b, copyLibraryElement)
		if err != nil {
			return nil, err
		}
	}

	newlibElement := &GrafanaLibraryElement{}
	newlibElement.Name = copyLibraryElement.Name
	newlibElement.Kind = copyLibraryElement.Kind
	newlibElement.FolderID = copyLibraryElement.FolderID
	newlibElement.Model = copyLibraryElement.Model

	if grafanaLibraryElementOptions.SaveUID {
		newlibElement.UID = copyLibraryElement.UID
	}

	if !utils.IsEmpty(grafanaLibraryElementOptions.Name) {
		newlibElement.Name = grafanaLibraryElementOptions.Name
	}
	if !utils.IsEmpty(grafanaLibraryElementOptions.FolderID) {
		newlibElement.FolderID = grafanaLibraryElementOptions.FolderID
	}
	if !utils.IsEmpty(grafanaLibraryElementOptions.UID) {
		newlibElement.UID = grafanaLibraryElementOptions.UID
	}

	l, err := json.Marshal(newlibElement)
	if err != nil {
		return nil, err
	}

	result, err := utils.HttpPostRaw(g.client, u.String(), "application/json", g.getAuth(grafanaOptions), l)
	fmt.Println(string(result))
	return result, err
}

func (g *Grafana) CopyLibraryElement(grafanaLibraryElementOptions GrafanaLibraryElementOptions) ([]byte, error) {
	return g.CustomCopyLibraryElement(g.options, grafanaLibraryElementOptions)
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
	return utils.HttpPostRaw(g.client, u.String(), "application/json", g.getAuth(grafanaOptions), b)
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

func (g *Grafana) CustomGetAnnotations(grafanaOptions GrafanaOptions, grafanaDashboardOptions GrafanaDahboardOptions, getAnnotationsOptions GrafanaGetAnnotationsOptions) ([]byte, error) {
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
	params.Add("tz", grafanaDashboardOptions.Timezone)

	u.RawQuery = params.Encode()
	return utils.HttpGetRaw(g.client, u.String(), "", g.getAuth(grafanaOptions))
}

func (g *Grafana) GetAnnotations(dashboardOptions GrafanaDahboardOptions, annotationsOptions GrafanaGetAnnotationsOptions) ([]byte, error) {
	return g.CustomGetAnnotations(g.options, dashboardOptions, annotationsOptions)
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

func (g Grafana) findPanelByID(source *[]interface{}, ID string) map[string]interface{} {

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
						rp := g.findPanelByID(&pnls, ID)
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

func (g Grafana) findPanelsByTitle(source *[]interface{}, title string, pms *[]map[string]interface{}) {

	for _, p := range *source {
		pm, ok := p.(map[string]interface{})
		if ok {

			t, okID := pm["title"].(string)
			if okID {
				match, _ := regexp.MatchString(title, t)
				if match {
					*pms = append(*pms, pm)
				} else if g.panelIsType(pm, "row") {
					pnls, okPnls := pm["panels"].([]interface{})
					if okPnls {
						g.findPanelsByTitle(&pnls, title, pms)
					}
				}
			}
		}
	}
}

func (g Grafana) copyPanels(source, dest *[]interface{}, clonedDashboardOptions GrafanaClonedDahboardOptions) {

	if len(*source) <= 0 {
		return
	}

	IDs := clonedDashboardOptions.PanelIDs
	titles := clonedDashboardOptions.PanelTitles
	series := clonedDashboardOptions.PanelSeries

	if !utils.IsEmpty(IDs) {
		for idx, id := range IDs {

			pm := g.findPanelByID(source, id)
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
	}

	if !utils.IsEmpty(titles) {
		for idx, title := range titles {

			pms := [](map[string]interface{}){}
			g.findPanelsByTitle(source, title, &pms)
			for _, pm := range pms {
				if !g.panelIsType(pm, "row") {
					g.setLegend(pm, clonedDashboardOptions.LegendRight)
					g.deleteAlerts(pm)
					if (len(series) > idx) && !utils.IsEmpty(series[idx]) {
						g.setTransformations(pm, series[idx])
					}
				} else {
					pnls, okPnls := pm["panels"].([]interface{})
					if okPnls {
						for _, pm1 := range pnls {
							pm2, ok := pm1.(map[string]interface{})
							if ok {
								g.setLegend(pm2, clonedDashboardOptions.LegendRight)
								g.deleteAlerts(pm2)
								if (len(series) > idx) && !utils.IsEmpty(series[idx]) {
									g.setTransformations(pm2, series[idx])
								}
							}
						}
					}
				}
				*dest = append(*dest, pm)
			}
		}
	}

	if utils.IsEmpty(IDs) && utils.IsEmpty(titles) {
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

func (g Grafana) CustomCreateDashboard(grafanaOptions GrafanaOptions, createDashboardOptions GrafanaDahboardOptions) ([]byte, error) {

	u, err := url.Parse(grafanaOptions.URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "/api/dashboards/db")

	cloned := &GrafanaBoard{}

	if !utils.IsEmpty(createDashboardOptions.Cloned.UID) {

		clonedOpts := grafanaOptions
		clonedDashboardOpts := createDashboardOptions
		clonedDashboardOpts.UID = createDashboardOptions.Cloned.UID

		if !utils.IsEmpty(createDashboardOptions.Cloned.URL) {
			clonedOpts.URL = createDashboardOptions.Cloned.URL
			clonedOpts.Timeout = createDashboardOptions.Cloned.Timeout
			clonedOpts.Insecure = createDashboardOptions.Cloned.Insecure
			clonedOpts.APIKey = createDashboardOptions.Cloned.APIKey
			clonedOpts.OrgID = createDashboardOptions.Cloned.OrgID
		}
		b, err := g.CustomGetDashboards(clonedOpts, clonedDashboardOpts)
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
	req.Dashboard.Timezone = createDashboardOptions.Timezone
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
	return utils.HttpPostRaw(g.client, u.String(), "application/json", g.getAuth(grafanaOptions), b)
}

func (g *Grafana) CreateDashboard(options GrafanaDahboardOptions) ([]byte, error) {
	return g.CustomCreateDashboard(g.options, options)
}

func NewGrafana(options GrafanaOptions) *Grafana {

	grafana := &Grafana{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
	return grafana
}
