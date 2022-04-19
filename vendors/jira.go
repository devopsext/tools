package vendors

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
)

type JiraOptions struct {
	URL         string
	Timeout     int
	Insecure    bool
	User        string
	Password    string
	ProjectKey  string
	IssueType   string
	Summary     string
	Description string
}

type JiraIssueProject struct {
	Key string `json:"key"`
}

type JiraIssueType struct {
	Name string `json:"name"`
}

type JiraIssueFields struct {
	Project     *JiraIssueProject `json:"project"`
	IssueType   *JiraIssueType    `json:"issuetype"`
	Summary     string            `json:"summary"`
	Description string            `json:"description"`
}

type JiraCreateIssue struct {
	Fields *JiraIssueFields `json:"fields"`
}

type Jira struct {
	client  *http.Client
	options JiraOptions
}

func (j *Jira) CreateCustomIssue(opts JiraOptions) ([]byte, error) {

	issue := &JiraCreateIssue{
		Fields: &JiraIssueFields{
			Project: &JiraIssueProject{
				Key: opts.ProjectKey,
			},
			IssueType: &JiraIssueType{
				Name: opts.IssueType,
			},
			Summary:     opts.Summary,
			Description: opts.Description,
		},
	}

	req, err := json.Marshal(&issue)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "/rest/api/2/issue")

	auth := ""
	if !utils.IsEmpty(opts.User) {
		userPass := fmt.Sprintf("%s:%s", opts.User, opts.Password)
		auth = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(userPass)))
	}

	return common.HttpPostRaw(j.client, u.String(), "application/json", auth, req)
}

func (j *Jira) CreateIssue() ([]byte, error) {
	return j.CreateCustomIssue(j.options)
}

func NewJira(options JiraOptions) *Jira {

	return &Jira{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
