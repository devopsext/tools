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
	IdOrKey     string
	Summary     string
	Description string
	Labels      []string
}

type JiraIssueAddCommentOptions struct {
	Body string
}

type JiraIssueAddAttachmentOptions struct {
	File string
	Name string
}

type JiraIssueUpdateOptions struct {
}

type JiraOptions struct {
	URL                       string
	Timeout                   int
	Insecure                  bool
	User                      string
	Password                  string
	IssueCreateOptions        *JiraIssueCreateOptions
	IssueOptions              *JiraIssueOptions
	IssueAddCommentOptions    *JiraIssueAddCommentOptions
	IssueAddAttachmentOptions *JiraIssueAddAttachmentOptions
	IssueUpdateOptions        *JiraIssueUpdateOptions
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

func (j *Jira) getAuth(opts JiraOptions) string {

	auth := ""
	if !utils.IsEmpty(opts.User) {
		userPass := fmt.Sprintf("%s:%s", opts.User, opts.Password)
		auth = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(userPass)))
	}
	return auth
}

func (j *Jira) IssueCreateCustom(opts JiraOptions) ([]byte, error) {

	if opts.IssueOptions == nil || opts.IssueCreateOptions == nil {
		return nil, fmt.Errorf("options are not enough")
	}

	issue := &JiraIssueCreate{
		Fields: &JiraIssueFields{
			Project: &JiraIssueProject{
				Key: opts.IssueCreateOptions.ProjectKey,
			},
			IssueType: &JiraIssueType{
				Name: opts.IssueCreateOptions.Type,
			},
			Summary:     opts.IssueOptions.Summary,
			Description: opts.IssueOptions.Description,
			Labels:      opts.IssueOptions.Labels,
		},
	}

	if !utils.IsEmpty(opts.IssueCreateOptions.Priority) {
		issue.Fields.Priority = &JiraIssuePriority{
			Name: opts.IssueCreateOptions.Priority,
		}
	}

	if !utils.IsEmpty(opts.IssueCreateOptions.Assignee) {
		issue.Fields.Assignee = &JiraIssueAssignee{
			Name: opts.IssueCreateOptions.Assignee,
		}
	}

	if !utils.IsEmpty(opts.IssueCreateOptions.Reporter) {
		issue.Fields.Reporter = &JiraIssueReporter{
			Name: opts.IssueCreateOptions.Reporter,
		}
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
	return common.HttpPostRaw(j.client, u.String(), "application/json", j.getAuth(opts), req)
}

func (j *Jira) IssueCreate() ([]byte, error) {
	return j.IssueCreateCustom(j.options)
}

func (j *Jira) IssueAddCustomComment(opts JiraOptions) ([]byte, error) {

	if opts.IssueOptions == nil || opts.IssueAddCommentOptions == nil {
		return nil, fmt.Errorf("options are not enough")
	}

	comment := &JiraIssueAddComment{
		Body: opts.IssueAddCommentOptions.Body,
	}

	req, err := json.Marshal(&comment)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("/rest/api/2/issue/%s/comment", opts.IssueOptions.IdOrKey))
	return common.HttpPostRaw(j.client, u.String(), "application/json", j.getAuth(opts), req)
}

func (j *Jira) IssueAddComment() ([]byte, error) {
	return j.IssueAddCustomComment(j.options)
}

func (j *Jira) IssueAddCustomAttachment(opts JiraOptions) ([]byte, error) {

	if opts.IssueOptions == nil || opts.IssueAddAttachmentOptions == nil {
		return nil, fmt.Errorf("options are not enough")
	}

	u, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("/rest/api/2/issue/%s/attachments", opts.IssueOptions.IdOrKey))

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	fw, err := w.CreateFormFile("file", opts.IssueAddAttachmentOptions.Name)
	if err != nil {
		return nil, err
	}

	if _, err := fw.Write([]byte(opts.IssueAddAttachmentOptions.File)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	headers := make(map[string]string)
	headers["Content-type"] = w.FormDataContentType()
	headers["Authorization"] = j.getAuth(opts)
	headers["X-Atlassian-Token"] = "no-check"
	return common.HttpPostRawWithHeaders(j.client, u.String(), headers, body.Bytes())
}

func (j *Jira) IssueAddAttachment() ([]byte, error) {
	return j.IssueAddCustomAttachment(j.options)
}

func (j *Jira) IssueUpdateCustom(opts JiraOptions) ([]byte, error) {

	if opts.IssueOptions == nil || opts.IssueUpdateOptions == nil {
		return nil, fmt.Errorf("options are not enough")
	}

	issue := &JiraIssueUpdate{
		Fields: &JiraIssueFields{
			Summary:     opts.IssueOptions.Summary,
			Description: opts.IssueOptions.Description,
		},
	}

	if len(opts.IssueOptions.Labels) > 0 {
		for _, v := range opts.IssueOptions.Labels {
			if !utils.IsEmpty(v) {
				issue.Fields.Labels = opts.IssueOptions.Labels
				break
			}
		}
	}

	req, err := json.Marshal(&issue)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("/rest/api/2/issue/%s", opts.IssueOptions.IdOrKey))
	return common.HttpPutRaw(j.client, u.String(), "application/json", j.getAuth(opts), req)
}

func (j *Jira) IssueUpdate() ([]byte, error) {
	return j.IssueUpdateCustom(j.options)
}

func NewJira(options JiraOptions) *Jira {

	return &Jira{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
