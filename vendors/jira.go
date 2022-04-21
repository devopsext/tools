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

type JiraCreateIssueOptions struct {
	ProjectKey  string
	Type        string
	Summary     string
	Description string
	Priority    string
	Labels      []string
	Assignee    string
	Reporter    string
}

type JiraIssueOptions struct {
	IdOrKey string
}

type JiraIssueAddCommentOptions struct {
	Body string
}

type JiraIssueAddAttachmentOptions struct {
	File string
	Name string
}

type JiraOptions struct {
	URL                       string
	Timeout                   int
	Insecure                  bool
	User                      string
	Password                  string
	CreateIssueOptions        *JiraCreateIssueOptions
	IssueOptions              *JiraIssueOptions
	IssueAddCommentOptions    *JiraIssueAddCommentOptions
	IssueAddAttachmentOptions *JiraIssueAddAttachmentOptions
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
	Project     *JiraIssueProject  `json:"project"`
	IssueType   *JiraIssueType     `json:"issuetype"`
	Summary     string             `json:"summary"`
	Description string             `json:"description"`
	Labels      []string           `json:"labels"`
	Priority    *JiraIssuePriority `json:"priority,omitempty"`
	Assignee    *JiraIssueAssignee `json:"assignee,omitempty"`
	Reporter    *JiraIssueReporter `json:"reporter,omitempty"`
}

type JiraCreateIssue struct {
	Fields *JiraIssueFields `json:"fields"`
}

type JiraIssueAddComment struct {
	Body string `json:"body"`
}

type Jira struct {
	client  *http.Client
	options JiraOptions
}

func (j *Jira) CreateCustomIssue(opts JiraOptions) ([]byte, error) {

	if opts.CreateIssueOptions == nil {
		return nil, fmt.Errorf("options are not enough")
	}

	issue := &JiraCreateIssue{
		Fields: &JiraIssueFields{
			Project: &JiraIssueProject{
				Key: opts.CreateIssueOptions.ProjectKey,
			},
			IssueType: &JiraIssueType{
				Name: opts.CreateIssueOptions.Type,
			},
			Summary:     opts.CreateIssueOptions.Summary,
			Description: opts.CreateIssueOptions.Description,
			Labels:      opts.CreateIssueOptions.Labels,
		},
	}

	if !utils.IsEmpty(opts.CreateIssueOptions.Priority) {
		issue.Fields.Priority = &JiraIssuePriority{
			Name: opts.CreateIssueOptions.Priority,
		}
	}

	if !utils.IsEmpty(opts.CreateIssueOptions.Assignee) {
		issue.Fields.Assignee = &JiraIssueAssignee{
			Name: opts.CreateIssueOptions.Assignee,
		}
	}

	if !utils.IsEmpty(opts.CreateIssueOptions.Reporter) {
		issue.Fields.Reporter = &JiraIssueReporter{
			Name: opts.CreateIssueOptions.Reporter,
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

	auth := ""
	if !utils.IsEmpty(opts.User) {
		userPass := fmt.Sprintf("%s:%s", opts.User, opts.Password)
		auth = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(userPass)))
	}

	return common.HttpPostRaw(j.client, u.String(), "application/json", auth, req)
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
	if !utils.IsEmpty(opts.User) {
		userPass := fmt.Sprintf("%s:%s", opts.User, opts.Password)
		headers["Authorization"] = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(userPass)))
	}
	headers["X-Atlassian-Token"] = "no-check"

	return common.HttpPostRawWithHeaders(j.client, u.String(), headers, body.Bytes())
}

func (j *Jira) IssueAddAttachment() ([]byte, error) {
	return j.IssueAddCustomAttachment(j.options)
}

func NewJira(options JiraOptions) *Jira {

	return &Jira{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
