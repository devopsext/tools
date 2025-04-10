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
	"strconv"
	"strings"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
)

type Jira struct {
	client  *http.Client
	options JiraOptions
}
type JiraOptions struct {
	URL         string
	Timeout     int
	Insecure    bool
	User        string
	Password    string
	AccessToken string
}

type JiraIssueOptions struct {
	IdOrKey      string
	ProjectKey   string
	Type         string
	Priority     string
	Assignee     string
	Reporter     string
	Summary      string
	Description  string
	CustomFields string
	TransitionID string
	Components   string
	Labels       []string
}

type JiraAddIssueCommentOptions struct {
	Body string
}

type JiraAddIssueAttachmentOptions struct {
	File string
	Name string
}

type JiraSearchIssueOptions struct {
	SearchPattern string
	MaxResults    int
	Fields        []string
}

type JiraSearchAssetOptions struct {
	SearchPattern string
	ResultPerPage int
}

type JiraCreateAssetOptions struct {
	Description      string
	Name             string
	Repository       string
	DescriptionId    int
	NameId           int
	RepositoryId     int
	ObjectSchemeId   string
	ObjectTypeId     int
	TitleId          int
	Title            string
	TierId           int
	Tier             string
	BusinessProcess  *JiraAssetAttribute
	Team             *JiraAssetAttribute
	Dependencies     *JiraAssetAttribute
	Group            *JiraAssetAttribute
	IsThirdParty     *JiraAssetAttribute
	IsDecommissioned *JiraAssetAttribute
}

type JiraUpdateAssetOptions struct {
	ObjectId string
	Json     string
}

type JiraIssueCreate struct {
	Fields *JiraIssueFields `json:"fields"`
}

type JiraIssueUpdate struct {
	Fields *JiraIssueFields `json:"fields"`
}
type JiraIssueFields struct {
	Project     *JiraIssueProject      `json:"project,omitempty"`
	IssueType   *JiraIssueType         `json:"issuetype,omitempty"`
	Summary     string                 `json:"summary,omitempty"`
	Description string                 `json:"description,omitempty"`
	Labels      []string               `json:"labels,omitempty"`
	Priority    *JiraIssuePriority     `json:"priority,omitempty"`
	Components  *[]JiraIssueComponents `json:"components,omitempty"`
	Assignee    *JiraIssueAssignee     `json:"assignee,omitempty"`
	Reporter    *JiraIssueReporter     `json:"reporter,omitempty"`
}

type JiraIssueAddComment struct {
	Body string `json:"body"`
}

type JiraIssueType struct {
	Name string `json:"name"`
}

type JiraIssueProject struct {
	Key string `json:"key"`
}
type JiraIssuePriority struct {
	Name string `json:"name"`
}

type JiraIssueComponents struct {
	Name string `json:"name"`
}

type JiraIssueAssignee struct {
	Name string `json:"name"`
}

type JiraIssueReporter struct {
	Name string `json:"name"`
}

type JiraTransition struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type JiraTransitions struct {
	Transitions []JiraTransition `json:"transitions"`
}
type JiraIssueTransition struct {
	Transition *JiraTransition `json:"transition"`
}

type OutputCode struct {
	Code int `json:"code"`
}

type JiraAssetAttributeValue struct {
	Value string `json:"value"`
}

type JiraAssetAttribute struct {
	ObjectTypeAttributeId int                       `json:"objectTypeAttributeId"`
	ObjectAttributeValues []JiraAssetAttributeValue `json:"objectAttributeValues"`
}

type JiraAsset struct {
	ObjectTypeId int                  `json:"objectTypeId"`
	Attributes   []JiraAssetAttribute `json:"attributes"`
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

// we need custom json unmarshal for Jira Assets to support pagination
func jsonJiraAssetsUnmarshal(a []byte) (map[string]interface{}, error) {
	var assets interface{}
	err := json.Unmarshal(a, &assets)
	if err != nil {
		return nil, err
	}
	m := assets.(map[string]interface{})
	return m, nil
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

func (j *Jira) CustomCreateIssue(jiraOptions JiraOptions, createOptions JiraIssueOptions) ([]byte, error) {

	issue := &JiraIssueCreate{
		Fields: &JiraIssueFields{
			Project: &JiraIssueProject{
				Key: createOptions.ProjectKey,
			},
			IssueType: &JiraIssueType{
				Name: createOptions.Type,
			},
			Summary:     createOptions.Summary,
			Description: createOptions.Description,
		},
	}

	if !utils.IsEmpty(createOptions.Priority) {
		issue.Fields.Priority = &JiraIssuePriority{
			Name: createOptions.Priority,
		}
	}

	if (!utils.IsEmpty(createOptions.Labels)) && (len(createOptions.Labels) > 0) {
		issue.Fields.Labels = createOptions.Labels
	}

	if !utils.IsEmpty(createOptions.Assignee) {
		issue.Fields.Assignee = &JiraIssueAssignee{
			Name: createOptions.Assignee,
		}
	}

	if !utils.IsEmpty(createOptions.Reporter) {
		issue.Fields.Reporter = &JiraIssueReporter{
			Name: createOptions.Reporter,
		}
	}

	cf := make(map[string]interface{})

	if !utils.IsEmpty(createOptions.CustomFields) {
		var err error
		err = json.Unmarshal([]byte(createOptions.CustomFields), &cf)
		if err != nil {
			return nil, err
		}
	}

	if !utils.IsEmpty(createOptions.Components) {
		components := make([]JiraIssueComponents, 0)
		for _, v := range strings.Split(createOptions.Components, ",") {
			components = append(components, JiraIssueComponents{
				Name: v,
			})
		}
		issue.Fields.Components = &components
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
	return utils.HttpPostRaw(j.client, u.String(), "application/json", j.getAuth(jiraOptions), req)
}

func (j *Jira) CreateIssue(issueCreateOptions JiraIssueOptions) ([]byte, error) {
	return j.CustomCreateIssue(j.options, issueCreateOptions)
}

func (j *Jira) CustomAddIssueComment(jiraOptions JiraOptions, issueOptions JiraIssueOptions, addCommentOptions JiraAddIssueCommentOptions) ([]byte, error) {

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
	return utils.HttpPostRaw(j.client, u.String(), "application/json", j.getAuth(jiraOptions), req)
}

func (j *Jira) IssueAddComment(issueOptions JiraIssueOptions, addCommentOptions JiraAddIssueCommentOptions) ([]byte, error) {
	return j.CustomAddIssueComment(j.options, issueOptions, addCommentOptions)
}

func (j *Jira) CustomAddIssueAttachment(jiraOptions JiraOptions, issueOptions JiraIssueOptions, addAttachmentOptions JiraAddIssueAttachmentOptions) ([]byte, error) {

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
	return utils.HttpPostRawWithHeaders(j.client, u.String(), headers, body.Bytes())
}

func (j *Jira) AddIssueAttachment(issueOptions JiraIssueOptions, addAttachmentOptions JiraAddIssueAttachmentOptions) ([]byte, error) {
	return j.CustomAddIssueAttachment(j.options, issueOptions, addAttachmentOptions)
}

func (j *Jira) CustomUpdateIssue(jiraOptions JiraOptions, issueOptions JiraIssueOptions) ([]byte, error) {

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
		err = json.Unmarshal([]byte(issueOptions.CustomFields), &cf)
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
	return utils.HttpPutRaw(j.client, u.String(), "application/json", j.getAuth(jiraOptions), req)
}

func (j *Jira) CustomMoveIssue(jiraOptions JiraOptions, moveOptions JiraIssueOptions) ([]byte, error) {

	issue := &JiraIssueUpdate{
		Fields: &JiraIssueFields{
			IssueType: &JiraIssueType{
				Name: moveOptions.Type,
			},
		},
	}

	cf := make(map[string]interface{})

	req, err := jsonJiraMarshal(&issue, cf)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(jiraOptions.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("/rest/api/2/issue/%s", moveOptions.IdOrKey))
	return utils.HttpPutRaw(j.client, u.String(), "application/json", j.getAuth(jiraOptions), req)
}

func (j *Jira) MoveIssue(options JiraIssueOptions) ([]byte, error) {
	return j.CustomMoveIssue(j.options, options)
}

func (j *Jira) UpdateIssue(options JiraIssueOptions) ([]byte, error) {
	return j.CustomUpdateIssue(j.options, options)
}

func (j *Jira) GetIssueTransitions(jiraOptions JiraOptions, issueOptions JiraIssueOptions) ([]byte, error) {
	u, err := url.Parse(jiraOptions.URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("/rest/api/2/issue/%s/transitions", issueOptions.IdOrKey))
	q := u.Query()
	q.Set("expand", "transitions.fields")
	u.RawQuery = q.Encode()

	t, err := utils.HttpGetRaw(j.client, u.String(), "application/json", j.getAuth(jiraOptions))
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (j *Jira) CustomChangeIssueTransitions(jiraOptions JiraOptions, issueOptions JiraIssueOptions) ([]byte, error) {

	transition := &JiraIssueTransition{
		Transition: &JiraTransition{ID: issueOptions.TransitionID},
	}

	req, err := json.Marshal(transition)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(jiraOptions.URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, fmt.Sprintf("/rest/api/2/issue/%s/transitions", issueOptions.IdOrKey))

	_, c, err := utils.HttpPostRawOutCode(j.client, u.String(), "application/json", j.getAuth(jiraOptions), req)
	if err != nil {
		return nil, err
	}

	code, err := common.JsonMarshal(&OutputCode{
		Code: c,
	})
	if err != nil {
		return nil, err
	}

	return code, nil
}

func (j *Jira) ChangeIssueTransitions(options JiraIssueOptions) ([]byte, error) {
	return j.CustomChangeIssueTransitions(j.options, options)
}

func (j *Jira) CustomSearchIssue(jiraOptions JiraOptions, search JiraSearchIssueOptions) ([]byte, error) {

	params := make(url.Values)
	params.Add("jql", search.SearchPattern)
	params.Add("maxResults", strconv.Itoa(search.MaxResults))
	params.Add("validateQuery", "strict")
	params.Add("fields", strings.Join(search.Fields, ","))

	u, err := url.Parse(jiraOptions.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "/rest/api/2/search")
	u.RawQuery = params.Encode()

	return utils.HttpGetRaw(j.client, u.String(), "application/json", j.getAuth(jiraOptions))
}

func (j *Jira) SearchIssue(options JiraSearchIssueOptions) ([]byte, error) {
	return j.CustomSearchIssue(j.options, options)
}

func (j *Jira) CustomSearchAssets(jiraOptions JiraOptions, search JiraSearchAssetOptions) ([]byte, error) {

	params := make(url.Values)
	params.Add("qlQuery", search.SearchPattern)
	params.Add("resultPerPage", strconv.Itoa(search.ResultPerPage))

	u, err := url.Parse(jiraOptions.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "/rest/insight/1.0/aql/objects")
	u.RawQuery = params.Encode()
	a, err := utils.HttpGetRaw(j.client, u.String(), "application/json", j.getAuth(jiraOptions))
	if err != nil {
		return nil, err
	}

	// We need to check if there is a pagination in the answer, if so we need to get all results
	m, err := jsonJiraAssetsUnmarshal(a)
	if err != nil {
		return nil, err
	}
	assetsObj := m["objectEntries"].([]interface{})
	objAttr := m["objectTypeAttributes"].([]interface{})
	pageSize := m["pageSize"].(float64)
	if pageSize > 1 {
		for i := 2; i <= int(pageSize); i++ {
			params.Set("page", strconv.Itoa(i))
			u.RawQuery = params.Encode()
			a, err := utils.HttpGetRaw(j.client, u.String(), "application/json", j.getAuth(jiraOptions))
			if err != nil {
				return nil, err
			}
			m, err := jsonJiraAssetsUnmarshal(a)
			if err != nil {
				return nil, err
			}
			assetsObjPage := m["objectEntries"].([]interface{})
			assetsObj = append(assetsObj, assetsObjPage...)
		}

	}
	result := map[string]interface{}{
		"objects":    assetsObj,
		"attributes": objAttr,
	}
	return json.Marshal(result)
}

func (j *Jira) SearchAssets(options JiraSearchAssetOptions) ([]byte, error) {
	return j.CustomSearchAssets(j.options, options)
}

func (j *Jira) CustomCreateAsset(jiraOptions JiraOptions, createOptions JiraCreateAssetOptions) ([]byte, error) {
	attributes := []JiraAssetAttribute{
		{
			ObjectTypeAttributeId: createOptions.NameId,
			ObjectAttributeValues: []JiraAssetAttributeValue{
				{
					Value: createOptions.Name,
				},
			},
		},
		{
			ObjectTypeAttributeId: createOptions.DescriptionId,
			ObjectAttributeValues: []JiraAssetAttributeValue{
				{
					Value: createOptions.Description,
				},
			},
		},
		{
			ObjectTypeAttributeId: createOptions.TierId,
			ObjectAttributeValues: []JiraAssetAttributeValue{
				{
					Value: createOptions.Tier,
				},
			},
		},
		{
			ObjectTypeAttributeId: createOptions.RepositoryId,
			ObjectAttributeValues: []JiraAssetAttributeValue{
				{
					Value: createOptions.Repository,
				},
			},
		},
		{
			ObjectTypeAttributeId: createOptions.TitleId,
			ObjectAttributeValues: []JiraAssetAttributeValue{
				{
					Value: createOptions.Title,
				},
			},
		},
	}

	if createOptions.BusinessProcess != nil {
		attributes = append(attributes, *createOptions.BusinessProcess)
	}

	if createOptions.Team != nil {
		attributes = append(attributes, *createOptions.Team)
	}

	if createOptions.Dependencies != nil {
		attributes = append(attributes, *createOptions.Dependencies)
	}

	if createOptions.Group != nil {
		attributes = append(attributes, *createOptions.Group)
	}

	if createOptions.IsThirdParty != nil {
		attributes = append(attributes, *createOptions.IsThirdParty)
	}

	if createOptions.IsDecommissioned != nil {
		attributes = append(attributes, *createOptions.IsDecommissioned)
	}

	object := &JiraAsset{
		ObjectTypeId: createOptions.ObjectTypeId,
		Attributes:   attributes,
	}

	req, err := json.Marshal(object)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(jiraOptions.URL)
	if err != nil {
		return nil, err
	}

	params := make(url.Values)
	params.Add("objectSchemaId", createOptions.ObjectSchemeId)
	u.Path = path.Join(u.Path, "rest/assets/1.0/object/create")
	u.RawQuery = params.Encode()
	return utils.HttpPostRaw(j.client, u.String(), "application/json", j.getAuth(jiraOptions), req)
}

func (j *Jira) CreateAsset(createOptions JiraCreateAssetOptions) ([]byte, error) {
	return j.CustomCreateAsset(j.options, createOptions)
}

func (j *Jira) CustomUpdateAsset(jiraOptions JiraOptions, updateOptions JiraUpdateAssetOptions) ([]byte, error) {

	u, err := url.Parse(jiraOptions.URL)
	if err != nil {
		return nil, err
	}

	params := make(url.Values)
	u.Path = path.Join(u.Path, fmt.Sprintf("rest/assets/1.0/object/%s", updateOptions.ObjectId))
	u.RawQuery = params.Encode()

	return utils.HttpPutRaw(j.client, u.String(), "application/json", j.getAuth(jiraOptions), []byte(updateOptions.Json))
}

func (j *Jira) UpdateAsset(updateOptions JiraUpdateAssetOptions) ([]byte, error) {
	return j.CustomUpdateAsset(j.options, updateOptions)
}

func NewJira(options JiraOptions) *Jira {

	jira := &Jira{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
	return jira
}
