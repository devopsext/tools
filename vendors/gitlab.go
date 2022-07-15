package vendors

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
)

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
	ProjectID int
	Scope     string
	Status    string
	Ref       string
	OrderBy   string
	Sort      string
	Limit     int
}

type GitlabPipelineGetVariablesOptions struct {
	Query []string
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

func (g Gitlab) getLastPipeline(project int, ref string) (*GitlabPipelinesResp, error) {
	u, err := url.Parse(g.options.URL)
	if err != nil {
		return nil, err
	}

	u.Path = fmt.Sprintf("/api/v4/projects/%d/pipelines", project)

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

func (g Gitlab) GetLastPipeline(project int, ref string) ([]byte, error) {
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

func (g Gitlab) getPipelineVariables(project int, pipeline int) (map[string]interface{}, error) {
	u, err := url.Parse(g.options.URL)
	if err != nil {
		return nil, err
	}

	u.Path = fmt.Sprintf("/api/v4/projects/%d/pipelines/%d/variables", project, pipeline)

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

func (g Gitlab) GetLastPipelineVariables(project int, ref string) ([]byte, error) {
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

func (g *Gitlab) getPipelines(gitlabOptions GitlabOptions, pipelineOptions GitlabPipelineOptions) ([]GitlabPipelinesResp, error) {

	var params = make(url.Values)
	if !utils.IsEmpty(pipelineOptions.Scope) {
		params.Add("scope", pipelineOptions.Scope)
	}
	if !utils.IsEmpty(pipelineOptions.Status) {
		params.Add("status", pipelineOptions.Status)
	}
	if !utils.IsEmpty(pipelineOptions.Ref) {
		params.Add("ref", pipelineOptions.Ref)
	}
	if !utils.IsEmpty(pipelineOptions.OrderBy) {
		params.Add("order_by", pipelineOptions.OrderBy)
	}
	if !utils.IsEmpty(pipelineOptions.Sort) {
		params.Add("sort", pipelineOptions.Sort)
	}

	u, err := url.Parse(gitlabOptions.URL)
	if err != nil {
		return nil, err
	}
	u.Path = fmt.Sprintf("/api/v4/projects/%d/pipelines", pipelineOptions.ProjectID)
	u.RawQuery = params.Encode()

	headers := make(map[string]string)
	headers["PRIVATE-TOKEN"] = gitlabOptions.Token

	b, err := common.HttpGetRawWithHeaders(g.client, u.String(), headers)
	if err != nil {
		return nil, err
	}

	var pipelines []GitlabPipelinesResp
	err = json.Unmarshal(b, &pipelines)
	if err != nil {
		return nil, err
	}
	return pipelines, err
}

func (g *Gitlab) getPipelineVariablesEx(gitlabOptions GitlabOptions, projectID, pipelineID int) ([]GitlabPipelineVariableResp, error) {

	u, err := url.Parse(gitlabOptions.URL)
	if err != nil {
		return nil, err
	}
	u.Path = fmt.Sprintf("/api/v4/projects/%d/pipelines/%d/variables", projectID, pipelineID)

	headers := make(map[string]string)
	headers["PRIVATE-TOKEN"] = gitlabOptions.Token

	b, err := common.HttpGetRawWithHeaders(g.client, u.String(), headers)
	if err != nil {
		return nil, err
	}

	var variables []GitlabPipelineVariableResp
	err = json.Unmarshal(b, &variables)
	if err != nil {
		return nil, err
	}
	return variables, err
}

func (g *Gitlab) variablesExist(variables []GitlabPipelineVariableResp, query string) bool {

	for _, variable := range variables {

		key := strings.TrimSpace(variable.Key)
		value := strings.TrimSpace(variable.Value)

		arr := strings.Split(query, "=")
		if len(arr) == 0 {
			return false
		}

		if len(arr) < 2 {
			keyMatched, _ := regexp.MatchString(strings.TrimSpace(arr[0]), key)
			if keyMatched {
				return true
			}
		}

		if len(arr) == 2 {
			keyMatched, _ := regexp.MatchString(strings.TrimSpace(arr[0]), key)
			valueMatched, _ := regexp.MatchString(strings.TrimSpace(arr[1]), value)
			if keyMatched && valueMatched {
				return true
			}
		}
	}
	return false
}

func (g *Gitlab) CustomPipelineGetVariables(gitlabOptions GitlabOptions, pipelineOptions GitlabPipelineOptions,
	pipelineGetVariablesOptions GitlabPipelineGetVariablesOptions) ([]byte, error) {

	// 1. get pipeline list by pipeline variable key=value
	// 2. reverse pipeline list and get first success pipeline
	// 3. get variables and values from

	pipelines, err := g.getPipelines(gitlabOptions, pipelineOptions)
	if err != nil {
		return nil, err
	}

	counter := 0

	for _, pipeline := range pipelines {

		if pipelineOptions.Limit > 0 && counter >= pipelineOptions.Limit {
			break
		}
		counter++

		variables, err := g.getPipelineVariablesEx(gitlabOptions, pipeline.ProjectID, pipeline.ID)
		if err != nil {
			continue
		}

		r := false
		if len(pipelineGetVariablesOptions.Query) > 0 {
			r = true
		}

		for _, q := range pipelineGetVariablesOptions.Query {
			r = r && g.variablesExist(variables, q)
		}

		if r {
			b, err := json.Marshal(variables)
			if err != nil {
				return nil, err
			}
			return b, nil
		}
	}
	return nil, errors.New("no pipeline or variables found")
}

func (g *Gitlab) PipelineGetVariables(pipelineOptions GitlabPipelineOptions, pipelineGetVariablesOptions GitlabPipelineGetVariablesOptions) ([]byte, error) {
	return g.CustomPipelineGetVariables(g.options, pipelineOptions, pipelineGetVariablesOptions)
}

func NewGitlab(options GitlabOptions) (*Gitlab, error) {

	client := utils.NewHttpClient(options.Timeout, options.Insecure)
	if client == nil {
		return nil, errors.New("no http client")
	}

	gitlab := &Gitlab{
		client:  client,
		options: options,
	}
	return gitlab, nil
}
