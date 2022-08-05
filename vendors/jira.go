package vendors

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
)

type JiraIssueCreateOptions struct {
	ProjectKey string
	Type       string
	Priority   string
	Assignee   string
	Reporter   string
}

type JiraIssueOptions struct {
	IdOrKey      string
	Summary      string
	Description  string
	CustomFields string
	Labels       []string
}

type JiraIssueAddCommentOptions struct {
	Body string
}

type JiraIssueAddAttachmentOptions struct {
	File string
	Name string
}

type JiraOptions struct {
	URL         string
	Timeout     int
	Insecure    bool
	User        string
	Password    string
	AccessToken string
}

type JiraIssueProject struct {
	Key string `json:"key"`
}

type JiraIssueType struct {
	Name string `json:"name"`
}

type JiraIssuePriority struct {
	Name string `json:"name"`
}

type JiraIssueAssignee struct {
	Name string `json:"name"`
}

type JiraIssueReporter struct {
	Name string `json:"name"`
}
type JiraIssueFields struct {
	Project     *JiraIssueProject  `json:"project,omitempty"`
	IssueType   *JiraIssueType     `json:"issuetype,omitempty"`
	Summary     string             `json:"summary,omitempty"`
	Description string             `json:"description,omitempty"`
	Labels      []string           `json:"labels,omitempty"`
	Priority    *JiraIssuePriority `json:"priority,omitempty"`
	Assignee    *JiraIssueAssignee `json:"assignee,omitempty"`
	Reporter    *JiraIssueReporter `json:"reporter,omitempty"`
}

type JiraIssueCreate struct {
	Fields *JiraIssueFields `json:"fields"`
}

type JiraIssueUpdate struct {
	Fields *JiraIssueFields `json:"fields"`
}

type JiraIssueAddComment struct {
	Body string `json:"body"`
}

type Jira struct {
	client  *http.Client
	options JiraOptions
}

// we need custom json marshal for Jira due to possible using of custom fields
func jsonJiraMarshal(issue interface{}, cf map[string]interface{}) ([]byte, error) {
	m, err := common.InterfaceToMap("", issue)
	if err != nil {
		return nil, err
	}
	if len(cf) > 0 {
		value, err := common.InterfaceToMap("", m["fields"])
		if err != nil {
			return nil, err
		}
		for k, v := range cf {
			value[k] = v
		}
		m["fields"] = value
	}
	return json.Marshal(m)
}

func (j *Jira) getAuth(opts JiraOptions) string {

	auth := ""
	if !utils.IsEmpty(opts.User) {
		userPass := fmt.Sprintf("%s:%s", opts.User, opts.Password)
		auth = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(userPass)))
		return auth
	}
	if !utils.IsEmpty(opts.AccessToken) {
		auth = fmt.Sprintf("Bearer %s", opts.AccessToken)
		return auth
	}
	return auth
}

func (j *Jira) CustomIssueCreate(jiraOptions JiraOptions, issueOptions JiraIssueOptions, issueCreateOptions JiraIssueCreateOptions) ([]byte, error) {

	issue := &JiraIssueCreate{
		Fields: &JiraIssueFields{
			Project: &JiraIssueProject{
				Key: issueCreateOptions.ProjectKey,
			},
			IssueType: &JiraIssueType{
				Name: issueCreateOptions.Type,
			},
			Summary:     issueOptions.Summary,
			Description: issueOptions.Description,
			Labels:      issueOptions.Labels,
		},
	}

	if !utils.IsEmpty(issueCreateOptions.Priority) {
		issue.Fields.Priority = &JiraIssuePriority{
			Name: issueCreateOptions.Priority,
		}
	}

	if !utils.IsEmpty(issueCreateOptions.Assignee) {
		issue.Fields.Assignee = &JiraIssueAssignee{
			Name: issueCreateOptions.Assignee,
		}
	}

	if !utils.IsEmpty(issueCreateOptions.Reporter) {
		issue.Fields.Reporter = &JiraIssueReporter{
			Name: issueCreateOptions.Reporter,
		}
	}

	cf := make(map[string]interface{})

	if !utils.IsEmpty(issueOptions.CustomFields) {
		var err error
		cf, err = common.ReadAndMarshal(issueOptions.CustomFields)
		if err != nil {
			return nil, err
		}
	}

	req, err := jsonJiraMarshal(&issue, cf)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(jiraOptions.URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "/rest/api/2/issue")
	return common.HttpPostRaw(j.client, u.String(), "application/json", j.getAuth(jiraOptions), req)
}

func (j *Jira) IssueCreate(issueOptions JiraIssueOptions, issueCreateOptions JiraIssueCreateOptions) ([]byte, error) {
	return j.CustomIssueCreate(j.options, issueOptions, issueCreateOptions)
}

func (j *Jira) CustomIssueAddComment(jiraOptions JiraOptions, issueOptions JiraIssueOptions, addCommentOptions JiraIssueAddCommentOptions) ([]byte, error) {

	comment := &JiraIssueAddComment{
		Body: addCommentOptions.Body,
	}

	req, err := json.Marshal(&comment)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(jiraOptions.URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("/rest/api/2/issue/%s/comment", issueOptions.IdOrKey))
	return common.HttpPostRaw(j.client, u.String(), "application/json", j.getAuth(jiraOptions), req)
}

func (j *Jira) IssueAddComment(issueOptions JiraIssueOptions, addCommentOptions JiraIssueAddCommentOptions) ([]byte, error) {
	return j.CustomIssueAddComment(j.options, issueOptions, addCommentOptions)
}

func (j *Jira) CustomIssueAddAttachment(jiraOptions JiraOptions, issueOptions JiraIssueOptions, addAttachmentOptions JiraIssueAddAttachmentOptions) ([]byte, error) {

	u, err := url.Parse(jiraOptions.URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("/rest/api/2/issue/%s/attachments", issueOptions.IdOrKey))

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	fw, err := w.CreateFormFile("file", addAttachmentOptions.Name)
	if err != nil {
		return nil, err
	}

	if _, err := fw.Write([]byte(addAttachmentOptions.File)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	headers := make(map[string]string)
	headers["Content-type"] = w.FormDataContentType()
	headers["Authorization"] = j.getAuth(jiraOptions)
	headers["X-Atlassian-Token"] = "no-check"
	return common.HttpPostRawWithHeaders(j.client, u.String(), headers, body.Bytes())
}

func (j *Jira) IssueAddAttachment(issueOptions JiraIssueOptions, addAttachmentOptions JiraIssueAddAttachmentOptions) ([]byte, error) {
	return j.CustomIssueAddAttachment(j.options, issueOptions, addAttachmentOptions)
}

func (j *Jira) CustomIssueUpdate(jiraOptions JiraOptions, issueOptions JiraIssueOptions) ([]byte, error) {

	issue := &JiraIssueUpdate{
		Fields: &JiraIssueFields{
			Summary:     issueOptions.Summary,
			Description: issueOptions.Description,
		},
	}

	if len(issueOptions.Labels) > 0 {
		for _, v := range issueOptions.Labels {
			if !utils.IsEmpty(v) {
				issue.Fields.Labels = issueOptions.Labels
				break
			}
		}
	}

	cf := make(map[string]interface{})

	if !utils.IsEmpty(issueOptions.CustomFields) {
		var err error
		cf, err = common.ReadAndMarshal(issueOptions.CustomFields)
		if err != nil {
			return nil, err
		}
	}

	req, err := jsonJiraMarshal(&issue, cf)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(jiraOptions.URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("/rest/api/2/issue/%s", issueOptions.IdOrKey))
	return common.HttpPutRaw(j.client, u.String(), "application/json", j.getAuth(jiraOptions), req)
}

func (j *Jira) IssueUpdate(options JiraIssueOptions) ([]byte, error) {
	return j.CustomIssueUpdate(j.options, options)
}

func NewJira(options JiraOptions) (*Jira, error) {

	jira := &Jira{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
	return jira, nil
}
