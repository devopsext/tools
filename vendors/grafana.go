package vendors

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"time"

	"github.com/devopsext/utils"
)

type GrafanaOptions struct {
	URL         string
	Timeout     int
	Insecure    bool
	ApiKey      string
	OrgID       string
	UID         string
	Slug        string
	PanelID     string
	From        string
	To          string
	ImageWidth  int
	ImageHeight int
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

	if !utils.IsEmpty(g.options.ApiKey) {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", g.options.ApiKey))
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

func (g *Grafana) renderImage(URL, uid, slug, orgId, panelId, width, height, from, to string) ([]byte, error) {

	params := make(url.Values)
	if !utils.IsEmpty(orgId) {
		params.Add("orgId", orgId)
	}
	if !utils.IsEmpty(panelId) {
		params.Add("panelId", panelId)
	}
	if !utils.IsEmpty(width) {
		params.Add("width", width)
	}
	if !utils.IsEmpty(height) {
		params.Add("height", height)
	}
	if !utils.IsEmpty(from) {
		t, err := time.Parse(time.RFC3339Nano, from)
		if err == nil {
			from = strconv.Itoa(int(t.UTC().UnixMilli()))
		}
		params.Add("from", from)
	}
	if !utils.IsEmpty(to) {
		t, err := time.Parse(time.RFC3339Nano, to)
		if err == nil {
			to = strconv.Itoa(int(t.UTC().UnixMilli()))
		}
		params.Add("to", to)
	}
	params.Add("tz", "UTC")

	u, err := url.Parse(URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("/render/d-solo/%s/%s", uid, slug))
	if params != nil {
		u.RawQuery = params.Encode()
	}
	return g.get(u.String())
}

func (g *Grafana) GetImage() ([]byte, error) {

	width := strconv.Itoa(g.options.ImageWidth)
	height := strconv.Itoa(g.options.ImageHeight)

	return g.renderImage(g.options.URL, g.options.UID, g.options.Slug, g.options.OrgID, g.options.PanelID, width, height, g.options.From, g.options.To)
}

func NewGrafana(options GrafanaOptions) *Grafana {

	return &Grafana{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
