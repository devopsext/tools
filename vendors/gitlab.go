package vendors

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"time"

	"github.com/devopsext/utils"
)

type GitlabPipelineGetVariablesOptions struct {
}

type GitlabOptions struct {
	Timeout  int
	Insecure bool
	URL      string
	Token    string
}

type Gitlab struct {
	client  *http.Client
	options GitlabOptions
}

type GitlabPipelineOptions struct {
	Project string
	Ref     string
}

type GitlabPipelinesResp struct {
	ID        int       `json:"id"`
	Iid       int       `json:"iid"`
	ProjectID int       `json:"project_id"`
	Sha       string    `json:"sha"`
	Ref       string    `json:"ref"`
	Status    string    `json:"status"`
	Source    string    `json:"source"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	WebURL    string    `json:"web_url"`
}

type GitlabPipelineVariableResp struct {
	VariableType string `json:"variable_type"`
	Key          string `json:"key"`
	Value        string `json:"value"`
}

func (g *Gitlab) get(url string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("PRIVATE-TOKEN", g.options.Token)

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

func (g Gitlab) getLastPipeline(project string, ref string) (*GitlabPipelinesResp, error) {
	u, err := url.Parse(g.options.URL)
	if err != nil {
		return nil, err
	}

	u.Path = fmt.Sprintf("/api/v4/projects/%s/pipelines", project)

	q := u.Query()
	if ref != "" {
		q.Add("ref", ref)
	}
	q.Add("order_by", "updated_at")
	q.Add("sort", "desc")

	b, err := g.get(u.String())
	if err != nil {
		return nil, err
	}

	var pipelines []GitlabPipelinesResp
	err = json.Unmarshal(b, &pipelines)
	if err != nil {
		return nil, err
	}

	if len(pipelines) == 0 {
		return nil, fmt.Errorf("no pipelines found")
	}

	return &pipelines[0], nil
}

func (g Gitlab) GetLastPipeline(project string, ref string) ([]byte, error) {
	p, err := g.getLastPipeline(project, ref)
	if err != nil {
		return nil, err
	}
	b, err := json.Marshal(p)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (g Gitlab) getPipelineVariables(project string, pipeline int) (map[string]interface{}, error) {
	u, err := url.Parse(g.options.URL)
	if err != nil {
		return nil, err
	}

	u.Path = fmt.Sprintf("/api/v4/projects/%s/pipelines/%d/variables", project, pipeline)

	b, err := g.get(u.String())
	if err != nil {
		return nil, err
	}

	var variables []GitlabPipelineVariableResp
	err = json.Unmarshal(b, &variables)
	if err != nil {
		return nil, err
	}

	var p string
	for _, v := range variables {
		if v.Key == "TRIGGER_PAYLOAD" {
			p = v.Value
			break
		}
	}

	if p == "" {
		return nil, fmt.Errorf("no trigger payload found")
	}

	var data map[string]interface{}
	err = json.Unmarshal([]byte(p), &data)
	if err != nil {
		return nil, err
	}

	return data["variables"].(map[string]interface{}), nil
}

func (g Gitlab) GetLastPipelineVariables(project string, ref string) ([]byte, error) {
	pipeline, err := g.getLastPipeline(project, ref)
	if err != nil {
		return nil, err
	}

	v, err := g.getPipelineVariables(project, pipeline.ID)
	if err != nil {
		return nil, err
	}

	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (g *Gitlab) CustomPipelineGetVariables(gitlabOpts GitlabOptions, pipelineGetVariablesOptions GitlabPipelineGetVariablesOptions) ([]byte, error) {
	return nil, nil
}

func (g *Gitlab) PipelineGetVariables(options GitlabPipelineGetVariablesOptions) ([]byte, error) {
	return g.CustomPipelineGetVariables(g.options, options)
}

func NewGitlab(options GitlabOptions) *Gitlab {
	return &Gitlab{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
