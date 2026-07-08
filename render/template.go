package render

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	htmlTemplate "html/template"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"
	txtTemplate "text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/render/lite"
	"github.com/devopsext/tools/vendors"
	utils "github.com/devopsext/utils"
)

// Default timeout and synthetic HTTP status codes used by the http* template
// functions. These are re-exported from render/lite for backward
// compatibility.
const (
	DefaultTimeout = lite.DefaultTimeout
	ErrorCodeParam = lite.ErrorCodeParam
	ErrorCodeTLS   = lite.ErrorCodeTLS
	ErrorCodeHTTP  = lite.ErrorCodeHTTP
)

// TemplateOptions configures a Template, TextTemplate, or HtmlTemplate. Its
// Name, Content, TimeFormat, Funcs, and FilterFuncs fields are also used to
// configure the embedded, dependency-light render/lite function set; see
// lite.TemplateOptions.
//
// TemplateOptions is a flat struct, rather than one embedding
// lite.TemplateOptions, so that existing keyed struct literals such as
// TemplateOptions{Content: "..."} keep compiling.
type TemplateOptions struct {
	// Name is the name of the template to execute. When empty, the root
	// template is executed.
	Name string
	// Object is a JSON-encoded data object used by Render when no object is
	// passed explicitly.
	Object string
	// Content is the template source text.
	Content string
	// Files is an additional list of template files to parse.
	Files []string
	// TimeFormat is the layout used to parse strings passed to timeFormat.
	TimeFormat string
	// Pattern is a glob pattern of additional template files to parse.
	Pattern string
	// Funcs overrides or extends the default funcmap.
	Funcs map[string]any
	// FilterFuncs, when true, restricts the funcmap to only the functions
	// whose name appears as a substring of Content.
	FilterFuncs bool
}

// liteOptions returns the subset of options that configures the embedded
// render/lite function set.
func (options TemplateOptions) liteOptions() lite.TemplateOptions {
	return lite.TemplateOptions{
		Name:        options.Name,
		Content:     options.Content,
		TimeFormat:  options.TimeFormat,
		Funcs:       options.Funcs,
		FilterFuncs: options.FilterFuncs,
	}
}

// Template holds the state shared by TextTemplate and HtmlTemplate: the
// heavy, vendor-backed template functions (Jira, Grafana, AWS, Kubernetes,
// and so on) defined in this package, plus the lite, dependency-light
// function set embedded from render/lite.
type Template struct {
	lite.Template
	options TemplateOptions
	logger  common.Logger
	funcs   template.FuncMap
}

// TextTemplate is a text/template.Template built with render's full
// funcmap.
type TextTemplate struct {
	Template
	template *txtTemplate.Template
}

// HtmlTemplate is an html/template.Template built with render's full
// funcmap.
type HtmlTemplate struct {
	Template
	template *htmlTemplate.Template
}

func (tpl *Template) tryToWaitUntil(t time.Time, timeout time.Duration) {

	t2 := time.Now()
	diff := t2.Sub(t)
	if diff < timeout {
		time.Sleep(timeout - diff)
	}
}

func (tpl *Template) URLWait(url string, timeout, retry int, size int64) []byte {

	if utils.IsEmpty(url) {
		return nil
	}

	if retry <= 0 {
		retry = 1
	}

	tpl.LogDebug("URLWait url => %s [%d, %d, %d]", url, timeout, retry, size)

	var transport = &http.Transport{
		Dial:                (&net.Dialer{Timeout: time.Duration(timeout) * time.Second}).Dial,
		TLSHandshakeTimeout: time.Duration(timeout) * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: true},
	}

	client := http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}

	for i := 0; i < retry; i++ {

		t1 := time.Now()

		tpl.LogDebug("URLWait(%d) get %s...", i, url)

		data, err := utils.HttpGetRaw(&client, url, "", "")
		if err != nil {
			tpl.LogDebug("URLWait(%d) get %s err => %s", i, url, err.Error())
			tpl.tryToWaitUntil(t1, client.Timeout)
			continue
		}

		l := int64(len(data))
		tpl.LogDebug("URLWait(%d) %s len(data) = %d", i, url, l)

		if l < size {
			tpl.tryToWaitUntil(t1, client.Timeout)
			continue
		} else if l >= size {
			return data
		}
	}
	return nil
}

func (tpl *Template) GitlabPipelineVars(URL string, token string, projectID int, query string, limit int) string {

	gitlabOptions := vendors.GitlabOptions{
		Timeout:  30,
		Insecure: false,
		URL:      URL,
		Token:    token,
	}

	gitlab := vendors.NewGitlab(gitlabOptions)

	if limit <= 0 {
		limit = 100
	}

	pipelineOptions := vendors.GitlabPipelineOptions{
		ProjectID: projectID,
		Scope:     "finished",
		OrderBy:   "updated_at",
		Sort:      "desc",
		Limit:     limit,
	}

	pipelineGetVariablesOptions := vendors.GitlabGetPipelineVariablesOptions{
		Query: strings.Split(query, ","),
	}

	b, err := gitlab.GetPipelineVariables(pipelineOptions, pipelineGetVariablesOptions)
	if err != nil {
		tpl.LogInfo("GitlabPipelineVars err => %s", err.Error())
		return ""
	}
	return string(b)
}

func (tpl *Template) Sleep(ms int) string {
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return ""
}

func (tpl *Template) JiraSearchAssets(params map[string]interface{}) ([]byte, error) {

	if len(params) == 0 {
		return nil, fmt.Errorf("JiraSearchAssets err => %s", "no params allowed")
	}

	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 10
	}
	insecure, _ := params["insecure"].(bool)
	user, _ := params["user"].(string)
	password, _ := params["password"].(string)
	token, _ := params["token"].(string)

	jiraOptions := vendors.JiraOptions{
		URL:         url,
		Timeout:     timeout,
		Insecure:    insecure,
		User:        user,
		Password:    password,
		AccessToken: token,
	}

	jira := vendors.NewJira(jiraOptions)

	query, _ := params["query"].(string)
	limit, _ := params["limit"].(int)
	if limit == 0 {
		limit = 50
	}

	assetsOptions := vendors.JiraSearchAssetOptions{
		SearchPattern: query,
		ResultPerPage: limit,
	}

	return jira.SearchAssets(assetsOptions)
}

func (tpl *Template) JiraCreateAsset(params map[string]interface{}) ([]byte, error) {

	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 10
	}
	insecure, _ := params["insecure"].(bool)
	user, _ := params["user"].(string)
	password, _ := params["password"].(string)
	token, _ := params["token"].(string)

	objectTypeId, _ := params["objectTypeId"].(int)
	objectSchemeId, _ := params["objectSchemeId"].(string)
	nameId, _ := params["nameId"].(int)
	name, _ := params["name"].(string)
	descriptionId, _ := params["descriptionId"].(int)
	description, _ := params["description"].(string)
	repositoryId, _ := params["repositoryId"].(int)
	repository, _ := params["repository"].(string)
	titleId, _ := params["titleId"].(int)
	title, _ := params["title"].(string)
	tierId, _ := params["tierId"].(int)
	tier, _ := params["tier"].(string)
	businessProcessId, _ := params["businessProcessId"].(int)
	businessProcessesKeysRaw, _ := params["businessProcessesKeys"].([]interface{})
	dependenciesId, _ := params["dependenciesId"].(int)
	dependenciesKeysRaw, _ := params["dependenciesKeys"].([]interface{})
	teamId, _ := params["teamId"].(int)
	teamKey, _ := params["teamKey"].(string)
	groupId, _ := params["groupId"].(int)
	groupKey, _ := params["groupKey"].(string)
	thirdPartyId, _ := params["thirdPartyId"].(int)
	thirdPartyKey, _ := params["thirdPartyKey"].(string)
	decommissionedId, _ := params["decommissionedId"].(int)
	decommissionedKey, _ := params["decommissionedKey"].(string)

	businessProcessesKeys := make([]string, len(businessProcessesKeysRaw))
	for i, key := range businessProcessesKeysRaw {
		businessProcessesKeys[i] = fmt.Sprint(key)
	}
	businessProcessesAttributesValues := make([]vendors.JiraAssetAttributeValue, 0)
	for _, businessProcessKey := range businessProcessesKeys {
		businessProcessesAttributesValues = append(businessProcessesAttributesValues, vendors.JiraAssetAttributeValue{
			Value: businessProcessKey,
		})
	}
	businessProcesses := vendors.JiraAssetAttribute{
		ObjectTypeAttributeId: businessProcessId,
		ObjectAttributeValues: businessProcessesAttributesValues,
	}

	dependenciesKeys := make([]string, len(dependenciesKeysRaw))
	for i, key := range dependenciesKeysRaw {
		dependenciesKeys[i] = fmt.Sprint(key)
	}

	dependenciesAttributesValues := make([]vendors.JiraAssetAttributeValue, 0)
	for _, dependencyKey := range dependenciesKeys {
		dependenciesAttributesValues = append(dependenciesAttributesValues, vendors.JiraAssetAttributeValue{
			Value: dependencyKey,
		})
	}
	dependencies := vendors.JiraAssetAttribute{
		ObjectTypeAttributeId: dependenciesId,
		ObjectAttributeValues: dependenciesAttributesValues,
	}

	team := vendors.JiraAssetAttribute{
		ObjectTypeAttributeId: teamId,
		ObjectAttributeValues: []vendors.JiraAssetAttributeValue{
			{Value: teamKey},
		},
	}

	group := vendors.JiraAssetAttribute{
		ObjectTypeAttributeId: groupId,
		ObjectAttributeValues: []vendors.JiraAssetAttributeValue{
			{Value: groupKey},
		},
	}

	isThirdParty := vendors.JiraAssetAttribute{
		ObjectTypeAttributeId: thirdPartyId,
		ObjectAttributeValues: []vendors.JiraAssetAttributeValue{
			{Value: thirdPartyKey},
		},
	}

	isDecommissioned := vendors.JiraAssetAttribute{
		ObjectTypeAttributeId: decommissionedId,
		ObjectAttributeValues: []vendors.JiraAssetAttributeValue{
			{Value: decommissionedKey},
		},
	}

	jiraOptions := vendors.JiraOptions{
		URL:         url,
		Timeout:     timeout,
		Insecure:    insecure,
		User:        user,
		Password:    password,
		AccessToken: token,
	}
	jiraIssueOptions := vendors.JiraCreateAssetOptions{
		Name:              name,
		ObjectSchemeId:    objectSchemeId,
		ObjectTypeId:      objectTypeId,
		RepositoryId:      repositoryId,
		NameId:            nameId,
		DescriptionId:     descriptionId,
		Description:       description,
		Repository:        repository,
		TitleId:           titleId,
		Title:             title,
		TierId:            tierId,
		Tier:              tier,
		BusinessProcesses: &businessProcesses,
		Team:              &team,
		Dependencies:      &dependencies,
		Group:             &group,
		IsThirdParty:      &isThirdParty,
		IsDecommissioned:  &isDecommissioned,
	}

	jira := vendors.NewJira(jiraOptions)

	response, err := jira.CreateAsset(jiraIssueOptions)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (tpl *Template) JiraUpdateAsset(params map[string]interface{}) ([]byte, error) {
	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 10
	}
	insecure, _ := params["insecure"].(bool)
	user, _ := params["user"].(string)
	password, _ := params["password"].(string)
	token, _ := params["token"].(string)

	objectId, _ := params["objectId"].(string)
	jsonData, _ := params["json"].(string)

	jiraOptions := vendors.JiraOptions{
		URL:         url,
		Timeout:     timeout,
		Insecure:    insecure,
		User:        user,
		Password:    password,
		AccessToken: token,
	}

	jiraUpdateOptions := vendors.JiraUpdateAssetOptions{
		ObjectId: objectId,
		Json:     jsonData,
	}

	jira := vendors.NewJira(jiraOptions)

	response, err := jira.UpdateAsset(jiraUpdateOptions)
	if err != nil {
		fmt.Printf("\n\n\nJira API error response: %s\n\n\n", string(response))
		return nil, err
	}

	return response, nil
}

func (tpl *Template) JiraMoveIssue(params map[string]interface{}) ([]byte, error) {

	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 10
	}
	insecure, _ := params["insecure"].(bool)
	user, _ := params["user"].(string)
	password, _ := params["password"].(string)
	token, _ := params["token"].(string)

	key, _ := params["key"].(string)
	issueType, _ := params["issueType"].(string)

	jiraOptions := vendors.JiraOptions{
		URL:         url,
		Timeout:     timeout,
		Insecure:    insecure,
		User:        user,
		Password:    password,
		AccessToken: token,
	}

	jiraIssueOptions := vendors.JiraIssueOptions{
		IdOrKey: key,
		Type:    issueType,
	}

	jira := vendors.NewJira(jiraOptions)

	response, err := jira.MoveIssue(jiraIssueOptions)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (tpl *Template) JiraAddComment(params map[string]interface{}) ([]byte, error) {

	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 10
	}
	insecure, _ := params["insecure"].(bool)
	user, _ := params["user"].(string)
	password, _ := params["password"].(string)
	token, _ := params["token"].(string)
	body, _ := params["body"].(string)
	key, _ := params["key"].(string)

	jiraOptions := vendors.JiraOptions{
		URL:         url,
		Timeout:     timeout,
		Insecure:    insecure,
		User:        user,
		Password:    password,
		AccessToken: token,
	}
	jiraCommentOptions := vendors.JiraAddIssueCommentOptions{

		Body: body,
	}
	jiraIssueOptions := vendors.JiraIssueOptions{
		IdOrKey: key,
	}
	jira := vendors.NewJira(jiraOptions)

	response, err := jira.IssueAddComment(jiraIssueOptions, jiraCommentOptions)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (tpl *Template) JiraGetIssueTransition(params map[string]interface{}) ([]byte, error) {

	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 10
	}
	insecure, _ := params["insecure"].(bool)
	user, _ := params["user"].(string)
	password, _ := params["password"].(string)
	token, _ := params["token"].(string)

	key, _ := params["key"].(string)

	jiraOptions := vendors.JiraOptions{
		URL:         url,
		Timeout:     timeout,
		Insecure:    insecure,
		User:        user,
		Password:    password,
		AccessToken: token,
	}
	jiraIssueOptions := vendors.JiraIssueOptions{
		IdOrKey: key,
	}

	jira := vendors.NewJira(jiraOptions)

	response, err := jira.GetIssueTransitions(jiraOptions, jiraIssueOptions)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (tpl *Template) JiraIssueTransition(params map[string]interface{}) ([]byte, error) {

	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 10
	}
	insecure, _ := params["insecure"].(bool)
	user, _ := params["user"].(string)
	password, _ := params["password"].(string)
	token, _ := params["token"].(string)

	transitionId, _ := params["id"].(string)
	key, _ := params["key"].(string)

	comment, _ := params["comment"].(string)

	jiraOptions := vendors.JiraOptions{
		URL:         url,
		Timeout:     timeout,
		Insecure:    insecure,
		User:        user,
		Password:    password,
		AccessToken: token,
	}

	jiraIssueOptions := vendors.JiraIssueOptions{
		TransitionID:     transitionId,
		IdOrKey:          key,
		UpdateAddComment: comment,
	}

	jira := vendors.NewJira(jiraOptions)

	response, err := jira.CustomChangeIssueTransitions(jiraOptions, jiraIssueOptions)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (tpl *Template) JiraUpdateIssue(params map[string]interface{}) ([]byte, error) {

	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 10
	}
	insecure, _ := params["insecure"].(bool)
	user, _ := params["user"].(string)
	password, _ := params["password"].(string)
	token, _ := params["token"].(string)

	var updateAddLabels []string
	if addLabels, ok := params["addLabels"].(string); ok {
		updateAddLabels = strings.Split(addLabels, ",")
	}
	var updateRemoveLabels []string
	if removeLabels, ok := params["removeLabels"].(string); ok {
		updateRemoveLabels = strings.Split(removeLabels, ",")
	}

	key, _ := params["key"].(string)
	summary, _ := params["summary"].(string)
	description, _ := params["description"].(string)
	customFields, _ := params["customFields"].(string)

	jiraOptions := vendors.JiraOptions{
		URL:         url,
		Timeout:     timeout,
		Insecure:    insecure,
		User:        user,
		Password:    password,
		AccessToken: token,
	}
	jiraIssueOptions := vendors.JiraIssueOptions{
		IdOrKey:            key,
		Summary:            summary,
		Description:        description,
		CustomFields:       customFields,
		UpdateAddLabels:    updateAddLabels,
		UpdateRemoveLabels: updateRemoveLabels,
	}

	jira := vendors.NewJira(jiraOptions)

	response, err := jira.UpdateIssue(jiraIssueOptions)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (tpl *Template) JiraSearchIssue(params map[string]interface{}) ([]byte, error) {

	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 10
	}
	insecure, _ := params["insecure"].(bool)
	user, _ := params["user"].(string)
	password, _ := params["password"].(string)
	token, _ := params["token"].(string)

	jql, _ := params["jql"].(string)
	fields := strings.Split(params["fields"].(string), ",")
	maxResults, _ := params["maxResults"].(int)

	jiraOptions := vendors.JiraOptions{
		URL:         url,
		Timeout:     timeout,
		Insecure:    insecure,
		User:        user,
		Password:    password,
		AccessToken: token,
	}
	jiraSearchOptions := vendors.JiraSearchIssueOptions{
		SearchPattern: jql,
		MaxResults:    maxResults,
		Fields:        fields,
	}

	jira := vendors.NewJira(jiraOptions)

	response, err := jira.SearchIssue(jiraSearchOptions)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (tpl *Template) JiraCreateIssue(params map[string]interface{}) ([]byte, error) {

	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 10
	}
	insecure, _ := params["insecure"].(bool)
	user, _ := params["user"].(string)
	password, _ := params["password"].(string)
	token, _ := params["token"].(string)

	key, _ := params["projectKey"].(string)
	summary, _ := params["summary"].(string)
	description, _ := params["description"].(string)
	assignee, _ := params["assignee"].(string)
	reporter, _ := params["reporter"].(string)
	labels := strings.Split(params["labels"].(string), ",")
	priority, _ := params["priority"].(string)
	components, _ := params["components"].(string)
	issueType, _ := params["issueType"].(string)
	customFields, _ := params["customFields"].(string)

	jiraOptions := vendors.JiraOptions{
		URL:         url,
		Timeout:     timeout,
		Insecure:    insecure,
		User:        user,
		Password:    password,
		AccessToken: token,
	}
	jiraIssueOptions := vendors.JiraIssueOptions{
		ProjectKey:   key,
		Summary:      summary,
		Description:  description,
		Type:         issueType,
		Priority:     priority,
		Labels:       labels,
		Components:   components,
		Assignee:     assignee,
		Reporter:     reporter,
		CustomFields: customFields,
	}

	jira := vendors.NewJira(jiraOptions)

	response, err := jira.CreateIssue(jiraIssueOptions)
	if err != nil {
		tpl.logger.Error("Failed to create Jira issue: %v", response)
		return nil, err
	}

	return response, nil
}

func (tpl *Template) JiraGetUserByEmail(params map[string]interface{}) ([]byte, error) {

	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 10
	}
	insecure, _ := params["insecure"].(bool)
	user, _ := params["user"].(string)
	password, _ := params["password"].(string)
	token, _ := params["token"].(string)
	email, _ := params["email"].(string)

	jiraOptions := vendors.JiraOptions{
		URL:         url,
		Timeout:     timeout,
		Insecure:    insecure,
		User:        user,
		Password:    password,
		AccessToken: token,
	}

	jira := vendors.NewJira(jiraOptions)

	jiraUser, err := jira.GetUserByEmail(jiraOptions, email)
	if err != nil {
		return nil, err
	}

	return json.Marshal(jiraUser)
}

func (tpl *Template) awsOptionsFromParams(params map[string]interface{}) vendors.AWSOptions {
	accessKey, _ := params["accessKey"].(string)
	secretKey, _ := params["secretKey"].(string)
	account, _ := params["account"].(string)
	role, _ := params["role"].(string)
	roleTimeout, _ := params["roleTimeout"].(int)
	roleSessionName, _ := params["roleSessionName"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 30
	}
	insecure, _ := params["insecure"].(bool)
	return vendors.AWSOptions{
		AWSKeys:         vendors.AWSKeys{AccessKey: accessKey, SecretKey: secretKey},
		Accounts:        account,
		Role:            role,
		RoleTimeout:     roleTimeout,
		RoleSessionName: roleSessionName,
		Timeout:         timeout,
		Insecure:        insecure,
	}
}

func (tpl *Template) AWSS3ListObjects(params map[string]interface{}) ([]byte, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("AWSS3ListObjects err => no params provided")
	}
	region, _ := params["region"].(string)
	bucket, _ := params["bucket"].(string)
	prefix, _ := params["prefix"].(string)

	s3, err := vendors.NewAWSS3(tpl.awsOptionsFromParams(params))
	if err != nil {
		return nil, fmt.Errorf("AWSS3ListObjects err => %w", err)
	}
	return s3.ListObjects(region, bucket, prefix)
}

func (tpl *Template) AWSS3GetObject(params map[string]interface{}) ([]byte, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("AWSS3GetObject err => no params provided")
	}
	region, _ := params["region"].(string)
	bucket, _ := params["bucket"].(string)
	key, _ := params["key"].(string)
	noerror, _ := params["noerror"].(bool)

	s3, err := vendors.NewAWSS3(tpl.awsOptionsFromParams(params))
	if err != nil {
		if noerror {
			return nil, nil
		}
		return nil, fmt.Errorf("AWSS3GetObject err => %w", err)
	}
	data, err := s3.GetObject(region, bucket, key)
	if err != nil && noerror {
		fmt.Printf("AWSS3GetObject warning => %v\n", err)
		return nil, nil

	}
	return data, err
}

func (tpl *Template) AWSS3PutObject(params map[string]interface{}) ([]byte, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("AWSS3PutObject err => no params provided")
	}
	region, _ := params["region"].(string)
	bucket, _ := params["bucket"].(string)
	key, _ := params["key"].(string)
	contentType, _ := params["contentType"].(string)
	if contentType == "" {
		contentType = "application/json"
	}
	var body []byte
	switch v := params["body"].(type) {
	case string:
		body = []byte(v)
	case []byte:
		body = v
	}

	s3, err := vendors.NewAWSS3(tpl.awsOptionsFromParams(params))
	if err != nil {
		return nil, fmt.Errorf("AWSS3PutObject err => %w", err)
	}
	return s3.PutObject(region, bucket, key, contentType, body)
}

func (tpl *Template) LdapGetGroupMembers(params map[string]interface{}) ([]byte, error) {

	ldapOptions := vendors.LdapOptions{
		URL:      params["url"].(string),
		User:     params["user"].(string),
		Password: params["password"].(string),
		Timeout:  params["timeout"].(int),
		Insecure: params["insecure"].(bool),
		BaseDN:   params["baseDN"].(string),
	}

	filterObjectValue := params["filterObjectValue"].(string)
	filterCNValue := params["filterCNValue"].(string)

	ldap, err := vendors.NewLdapClient(ldapOptions, tpl.logger)
	if err != nil {
		tpl.logger.Error("Failed to create LDAP client: %v", err)
		return nil, err
	}
	defer ldap.Close()

	filter := fmt.Sprintf("(&(objectClass=%s*)(cn=%s))", filterObjectValue, filterCNValue)

	attributes := []string{"distinguishedName", "cn", "memberUid"} // check

	membersJson, err := ldap.GetGroupMembers(filter, attributes)
	if err != nil {
		tpl.logger.Error("Failed to get group members: %v", err)
		return nil, err
	}
	return membersJson, nil
}

func (tpl *Template) grafanaGetAlerts(params map[string]interface{}) ([]byte, error) {
	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 15
	}
	insecure, _ := params["insecure"].(bool)
	token, _ := params["token"].(string)
	orgID, _ := params["orgid"].(string)
	suppressed, _ := params["suppressed"].(bool)
	groupBy, _ := params["groupby"].(string)
	if groupBy == "" {
		groupBy = "alertname" // default
	}
	filter, _ := params["filter"].(string)

	grafanaOptions := vendors.GrafanaOptions{
		URL:      url,
		Timeout:  timeout,
		Insecure: insecure,
		APIKey:   token,
		OrgID:    orgID,
	}

	grafanaGetAlertsOptions := vendors.GrafanaGetAlertsOptions{
		Suppressed: suppressed,
		GroupBy:    groupBy,
		Filter:     filter,
	}

	grafana := vendors.NewGrafana(grafanaOptions)

	response, err := grafana.GetAlerts(grafanaGetAlertsOptions)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (tpl *Template) GrafanaCreateDashboard(params map[string]interface{}) ([]byte, error) {

	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 15
	}
	insecure, _ := params["insecure"].(bool)
	token, _ := params["token"].(string)
	orgID, _ := params["orgid"].(string)
	dUID, _ := params["uid"].(string)
	dSlug, _ := params["slug"].(string)
	dTimeZone, _ := params["timezone"].(string)

	title, _ := params["title"].(string)
	fUID, _ := params["fuid"].(string)
	tag, _ := params["tags"].(string)
	tags := []string{}
	if tag != "" {
		tags = strings.Split(tag, ",")
	}
	from, _ := params["from"].(string)
	to, _ := params["to"].(string)

	clonedUID, _ := params["cloneduid"].(string)

	panelIDS := params["panelids"].(string)
	var cpanelIDs []string
	if panelIDS != "" {
		cpanelIDs = strings.Split(panelIDS, ",")
	}
	titles := params["ptitles"].(string)
	var ptitles []string
	if titles != "" {
		ptitles = strings.Split(titles, ",")
	}

	grafanaOptions := vendors.GrafanaOptions{
		URL:      url,
		Timeout:  timeout,
		Insecure: insecure,
		APIKey:   token,
		OrgID:    orgID,
	}

	grafanaCreateDashboardOptions := vendors.GrafanaDashboardOptions{
		Title:     title,
		UID:       dUID,
		Slug:      dSlug,
		Timezone:  dTimeZone,
		FolderUID: fUID,
		Tags:      tags,
		From:      from,
		To:        to,
		Cloned: vendors.GrafanaClonedDashboardOptions{
			UID:         clonedUID,
			PanelIDs:    cpanelIDs,
			PanelTitles: ptitles,
			Count:       3,
			Width:       7,
			Height:      7,
		},
	}

	grafana := vendors.NewGrafana(grafanaOptions)

	response, err := grafana.CreateDashboard(grafanaCreateDashboardOptions)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (tpl *Template) GrafanaCopyDashboard(params map[string]interface{}) ([]byte, error) {

	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 15
	}
	insecure, _ := params["insecure"].(bool)
	token, _ := params["token"].(string)
	orgID, _ := params["orgid"].(string)
	dUID, _ := params["uid"].(string)
	dSlug, _ := params["slug"].(string)
	dTimeZone, _ := params["timezone"].(string)

	title, _ := params["title"].(string)
	fUID, _ := params["fuid"].(string)
	tag, _ := params["tags"].(string)
	tags := []string{}
	if tag != "" {
		tags = strings.Split(tag, ",")
	}
	from, _ := params["from"].(string)
	to, _ := params["to"].(string)

	clonedUID, _ := params["cloneduid"].(string)

	grafanaOptions := vendors.GrafanaOptions{
		URL:      url,
		Timeout:  timeout,
		Insecure: insecure,
		APIKey:   token,
		OrgID:    orgID,
	}

	grafanaCopyDashboardOptions := vendors.GrafanaDashboardOptions{
		Title:     title,
		UID:       dUID,
		Slug:      dSlug,
		Timezone:  dTimeZone,
		FolderUID: fUID,
		Tags:      tags,
		From:      from,
		To:        to,
		Cloned: vendors.GrafanaClonedDashboardOptions{
			UID: clonedUID,
		},
	}

	grafana := vendors.NewGrafana(grafanaOptions)

	response, err := grafana.CopyDashboard(grafanaCopyDashboardOptions)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func (tpl *Template) PagerDutyCreateIncident(params map[string]interface{}) ([]byte, error) {

	if len(params) == 0 {
		return nil, fmt.Errorf("PagerDutyEscalate err => %s", "no params passed")
	}

	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 10
	}
	insecure, _ := params["insecure"].(bool)
	token, _ := params["token"].(string)

	pagerDutyOptions := vendors.PagerDutyOptions{
		URL:      url,
		Timeout:  timeout,
		Insecure: insecure,
		Token:    token,
	}

	pagerDuty := vendors.NewPagerDuty(pagerDutyOptions, tpl.logger)

	title, _ := params["title"].(string)
	body, _ := params["body"].(string)
	urgency, _ := params["urgency"].(string)
	serviceID, _ := params["serviceID"].(string)
	priorityID, _ := params["priorityID"].(string)
	incidentType, _ := params["incidentType"].(string)

	incidentOptions := vendors.PagerDutyIncidentOptions{
		Title:        title,
		Body:         body,
		Urgency:      urgency,
		ServiceID:    serviceID,
		PriorityID:   priorityID,
		IncidentType: incidentType,
	}

	from, _ := params["from"].(string)

	createOptions := vendors.PagerDutyCreateIncidentOptions{
		From: from,
	}

	return pagerDuty.CreateIncident(incidentOptions, createOptions)
}

func (tpl *Template) PagerDutySendNoteToIncident(params map[string]interface{}) ([]byte, error) {

	if len(params) == 0 {
		return nil, fmt.Errorf("PagerDutyEscalate err => %s", "no params passed")
	}

	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 10
	}
	insecure, _ := params["insecure"].(bool)
	token, _ := params["token"].(string)

	pagerDutyOptions := vendors.PagerDutyOptions{
		URL:      url,
		Timeout:  timeout,
		Insecure: insecure,
		Token:    token,
	}

	pagerDuty := vendors.NewPagerDuty(pagerDutyOptions, tpl.logger)

	incidentId, _ := params["incidentid"].(string)
	noteContent, _ := params["notecontent"].(string)

	noteOptions := vendors.PagerDutyIncidentNoteOptions{
		IncidentID:  incidentId,
		NoteContent: noteContent,
	}

	from, _ := params["from"].(string)

	createOptions := vendors.PagerDutyCreateIncidentOptions{
		From: from,
	}

	return pagerDuty.CreateIncidentNote(noteOptions, createOptions)
}

func (tpl *Template) PrometheusGet(params map[string]interface{}) ([]byte, error) {

	if len(params) == 0 {
		return nil, fmt.Errorf("PrometheusGet err => %s", "no params allowed")
	}

	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 5
	}

	insecure, _ := params["insecure"].(bool)
	user, _ := params["user"].(string)
	password, _ := params["password"].(string)

	query, _ := params["query"].(string)
	if utils.IsEmpty(query) {
		return nil, fmt.Errorf("PrometheusGet err => %s", "query is empty")
	}

	from, _ := params["from"].(string)
	to, _ := params["to"].(string)
	step, _ := params["step"].(string)
	prms, _ := params["params"].(string)

	noerror, _ := params["noerror"].(bool)

	prometheusOptions := vendors.PrometheusOptions{
		URL:      url,
		User:     user,
		Password: password,
		Timeout:  timeout,
		Insecure: insecure,
		Query:    query,
		From:     from,
		To:       to,
		Step:     step,
		Params:   prms,
	}

	prometheus := vendors.NewPrometheus(prometheusOptions)

	d, err := prometheus.Get()
	if noerror {
		err = nil
	}
	return d, err
}

// TemplateRender renders the named template from the current template's
// content, using obj as the data object, and returns the result as a
// string. Unlike lite.Template.TemplateRender, this override gives the
// nested render access to this Template's full funcmap (including its
// heavy, vendor-backed functions), matching this method's behavior before
// the lite function set was extracted into render/lite.
func (tpl *Template) TemplateRender(name string, obj interface{}) (string, error) {

	opts := TemplateOptions{
		Content:     tpl.options.Content,
		Funcs:       tpl.funcs,
		FilterFuncs: tpl.options.FilterFuncs,
	}
	t, err := NewTextTemplate(opts, tpl.logger)
	if err != nil {
		return "", err
	}

	b, err := t.customRender(name, obj)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// TemplateRenderFile renders the template loaded from path, using obj as
// the data object, and returns the result as a string. See TemplateRender
// for why this overrides lite.Template.TemplateRenderFile.
func (tpl *Template) TemplateRenderFile(path string, obj interface{}) (string, error) {

	content, err := utils.Content(path)
	if err != nil {
		return "", err
	}

	opts := TemplateOptions{
		Content:     string(content),
		Funcs:       tpl.funcs,
		FilterFuncs: tpl.options.FilterFuncs,
	}
	t, err := NewTextTemplate(opts, tpl.logger)
	if err != nil {
		return "", err
	}

	b, err := t.RenderObject(obj)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (tpl *Template) GoogleCalendarGetEvents(params map[string]interface{}) ([]byte, error) {

	if len(params) == 0 {
		return nil, fmt.Errorf("GoogleCalendarGetEvents err => %s", "no params allowed")
	}

	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 10
	}
	insecure, _ := params["insecure"].(bool)
	clientID, _ := params["clientID"].(string)
	clientSecret, _ := params["clientSecret"].(string)
	token, _ := params["token"].(string)

	googleOptions := vendors.GoogleOptions{
		Timeout:           timeout,
		Insecure:          insecure,
		OAuthClientID:     clientID,
		OAuthClientSecret: clientSecret,
		RefreshToken:      token,
	}

	google := vendors.NewGoogle(googleOptions, tpl.logger)

	id, _ := params["ID"].(string)
	calendarOptions := vendors.GoogleCalendarOptions{
		ID: id,
	}

	timeMin, _ := params["timeMin"].(string)
	timeMax, _ := params["timeMax"].(string)
	timeZone, _ := params["timeZone"].(string)

	calendarGetEventsOptions := vendors.GoogleCalendarGetEventsOptions{
		TimeMin:      timeMin,
		TimeMax:      timeMax,
		TimeZone:     timeZone,
		SingleEvents: true,
		OrderBy:      "startTime",
	}

	return google.CalendarGetEvents(calendarOptions, calendarGetEventsOptions)
}

func (tpl *Template) GoogleCalendarInsertEvent(params map[string]interface{}) ([]byte, error) {

	if len(params) == 0 {
		return nil, fmt.Errorf("GoogleCalendarInsertEvent err => %s", "no params allowed")
	}

	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 10
	}
	insecure, _ := params["insecure"].(bool)
	clientID, _ := params["clientID"].(string)
	clientSecret, _ := params["clientSecret"].(string)
	token, _ := params["token"].(string)

	googleOptions := vendors.GoogleOptions{
		Timeout:           timeout,
		Insecure:          insecure,
		OAuthClientID:     clientID,
		OAuthClientSecret: clientSecret,
		RefreshToken:      token,
	}

	google := vendors.NewGoogle(googleOptions, tpl.logger)

	id, _ := params["ID"].(string)
	calendarOptions := vendors.GoogleCalendarOptions{
		ID: id,
	}

	summary, _ := params["summary"].(string)
	description, _ := params["description"].(string)
	start, _ := params["start"].(string)
	end, _ := params["end"].(string)
	timeZone, _ := params["timeZone"].(string)
	visibility, _ := params["visibility"].(string)
	conferenceID, _ := params["conferenceID"].(string)

	calendarInsertEventOptions := vendors.GoogleCalendarInsertEventOptions{
		Summary:      summary,
		Description:  description,
		Start:        start,
		End:          end,
		TimeZone:     timeZone,
		Visibility:   visibility,
		ConferenceID: conferenceID,
	}

	return google.CalendarInsertEvent(calendarOptions, calendarInsertEventOptions)
}

func (tpl *Template) GoogleCalendarDeleteEvents(params map[string]interface{}) ([]byte, error) {

	if len(params) == 0 {
		return nil, fmt.Errorf("GoogleCalendarDeleteEvents err => %s", "no params allowed")
	}

	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 10
	}
	insecure, _ := params["insecure"].(bool)
	clientID, _ := params["clientID"].(string)
	clientSecret, _ := params["clientSecret"].(string)
	token, _ := params["token"].(string)

	googleOptions := vendors.GoogleOptions{
		Timeout:           timeout,
		Insecure:          insecure,
		OAuthClientID:     clientID,
		OAuthClientSecret: clientSecret,
		RefreshToken:      token,
	}

	google := vendors.NewGoogle(googleOptions, tpl.logger)

	id, _ := params["ID"].(string)
	calendarOptions := vendors.GoogleCalendarOptions{
		ID: id,
	}

	timeMin, _ := params["timeMin"].(string)
	timeMax, _ := params["timeMax"].(string)
	timeZone, _ := params["timeZone"].(string)

	calendarGetEventsOptions := vendors.GoogleCalendarGetEventsOptions{
		TimeMin:  timeMin,
		TimeMax:  timeMax,
		TimeZone: timeZone,
	}

	return google.CalendarDeleteEvents(calendarOptions, calendarGetEventsOptions)
}

func (tpl *Template) GoogleMeetCreateSpace(params map[string]interface{}) ([]byte, error) {

	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 30
	}
	insecure, _ := params["insecure"].(bool)

	// Read service account credentials from environment variables
	serviceAccountKey := utils.EnvGet("GOOGLE_SERVICE_ACCOUNT_KEY", "").(string)
	impersonateEmail := utils.EnvGet("GOOGLE_IMPERSONATE_EMAIL", "").(string)

	if utils.IsEmpty(serviceAccountKey) {
		return nil, fmt.Errorf("GOOGLE_SERVICE_ACCOUNT_KEY environment variable not set")
	}

	googleOptions := vendors.GoogleOptions{
		Timeout:           timeout,
		Insecure:          insecure,
		ServiceAccountKey: serviceAccountKey,
		ImpersonateEmail:  impersonateEmail,
	}

	google := vendors.NewGoogle(googleOptions, tpl.logger)

	accessType, _ := params["accessType"].(string)
	if accessType == "" {
		accessType = "TRUSTED" // Default to organization-only
	}

	meetOptions := vendors.GoogleMeetOptions{
		AccessType: accessType,
	}

	meetResponse, err := google.CreateMeetSpace(meetOptions)
	if err != nil {
		return nil, err
	}

	// Debug: Log the actual response
	tpl.logger.Debug("Google Meet API response: %+v", meetResponse)
	tpl.logger.Debug("Meeting URI: %s", meetResponse.MeetingUri)
	tpl.logger.Debug("Meeting Code: %s", meetResponse.MeetingCode)

	// Convert response to JSON bytes
	return json.Marshal(meetResponse)
}

func (tpl *Template) GoogleDocsCopyDocument(params map[string]interface{}) ([]byte, error) {

	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 10
	}
	insecure, _ := params["insecure"].(bool)
	clientID, _ := params["clientID"].(string)
	clientSecret, _ := params["clientSecret"].(string)
	token, _ := params["token"].(string)
	domain, _ := params["domain"].(string)

	googleOptions := vendors.GoogleOptions{
		Timeout:           timeout,
		Insecure:          insecure,
		OAuthClientID:     clientID,
		OAuthClientSecret: clientSecret,
		RefreshToken:      token,
	}

	google := vendors.NewGoogle(googleOptions, tpl.logger)

	documentId, _ := params["documentId"].(string)
	docsOptions := vendors.GoogleDocsOptions{
		ID:     documentId,
		Domain: domain,
	}

	return google.DocsCopyDocument(docsOptions)
}

func (tpl *Template) SSHRun(params map[string]interface{}) ([]byte, error) {

	user, _ := params["user"].(string)
	host, _ := params["host"].(string)
	command, _ := params["command"].(string)
	key, _ := params["key"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 40
	}

	privateKey, err := utils.Content(key)
	if err != nil {
		return nil, err
	}

	sshOptions := vendors.SSHOptions{
		User:       user,
		Address:    host,
		PrivateKey: privateKey,
		Command:    command,
		Timeout:    timeout,
	}

	ssh := vendors.NewSSH(sshOptions)
	response, err := ssh.Run(sshOptions)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (tpl *Template) ListFilesWithModTime(rootDir string) (map[string]string, error) {
	filesMap := make(map[string]string)

	// Walk through the directory tree
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			// Handle the error (e.g., skip the file or directory)
			fmt.Println("Error accessing", path, ":", err)
			return nil
		}

		if !info.IsDir() {
			// Extract and collect filename and last modified date
			filename := filepath.Base(path)
			filesMap[filename] = info.ModTime().Format(time.RFC3339)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking the directory: %w", err)
	}

	return filesMap, nil
}

func (tpl *Template) VMReset(params map[string]interface{}) ([]byte, error) {

	user, _ := params["user"].(string)
	url, _ := params["url"].(string)
	password, _ := params["password"].(string)
	vms := strings.Split(params["vms"].(string), ",")
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 20
	}
	insecure, _ := params["insecure"].(bool)

	vcenterOptions := vendors.VCenterOptions{
		URL:      url,
		User:     user,
		Password: password,
		Timeout:  timeout,
		Insecure: insecure,
	}

	vmNames := vendors.VCenterVMNameOptions{
		Names: vms,
	}

	vcenterOptions, err := vendors.InitializeVCenterSession(vcenterOptions)
	if err != nil {
		return nil, err
	}

	vcenter := vendors.NewVCenter(vcenterOptions)

	var vi vendors.VMsResponse

	vmInfo, err := vcenter.GetVMsByName(vmNames)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(vmInfo, &vi)
	if err != nil {
		return nil, err
	}

	if len(vi.Value) == 0 {
		return nil, fmt.Errorf("no VMs found")
	}

	var response []byte

	if len(vi.Value) > 0 {
		for _, vm := range vi.Value {
			response, err = vcenter.ResetVM(vm.VM)
			if err != nil {
				return nil, err
			}
		}
	}

	return response, nil
}

func (tpl *Template) VMStop(params map[string]interface{}) ([]byte, error) {

	user, _ := params["user"].(string)
	url, _ := params["url"].(string)
	password, _ := params["password"].(string)
	vms := strings.Split(params["vms"].(string), ",")
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 20
	}
	insecure, _ := params["insecure"].(bool)

	vcenterOptions := vendors.VCenterOptions{
		URL:      url,
		User:     user,
		Password: password,
		Timeout:  timeout,
		Insecure: insecure,
	}

	vmNames := vendors.VCenterVMNameOptions{
		Names: vms,
	}

	vcenterOptions, err := vendors.InitializeVCenterSession(vcenterOptions)
	if err != nil {
		return nil, err
	}

	vcenter := vendors.NewVCenter(vcenterOptions)

	var vi vendors.VMsResponse

	vmInfo, err := vcenter.GetVMsByName(vmNames)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(vmInfo, &vi)
	if err != nil {
		return nil, err
	}

	if len(vi.Value) == 0 {
		return nil, fmt.Errorf("no VMs found")
	}

	var response []byte

	if len(vi.Value) > 0 {
		for _, vm := range vi.Value {
			response, err = vcenter.StopVM(vm.VM)
			if err != nil {
				return nil, err
			}
		}
	}

	return response, nil
}

func (tpl *Template) VMStart(params map[string]interface{}) ([]byte, error) {

	user, _ := params["user"].(string)
	url, _ := params["url"].(string)
	password, _ := params["password"].(string)
	vms := strings.Split(params["vms"].(string), ",")
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 20
	}
	insecure, _ := params["insecure"].(bool)

	vcenterOptions := vendors.VCenterOptions{
		URL:      url,
		User:     user,
		Password: password,
		Timeout:  timeout,
		Insecure: insecure,
	}

	vmNames := vendors.VCenterVMNameOptions{
		Names: vms,
	}

	vcenterOptions, err := vendors.InitializeVCenterSession(vcenterOptions)
	if err != nil {
		return nil, err
	}

	vcenter := vendors.NewVCenter(vcenterOptions)

	var vi vendors.VMsResponse

	vmInfo, err := vcenter.GetVMsByName(vmNames)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(vmInfo, &vi)
	if err != nil {
		return nil, err
	}

	if len(vi.Value) == 0 {
		return nil, fmt.Errorf("no VMs found")
	}

	var response []byte

	if len(vi.Value) > 0 {
		for _, vm := range vi.Value {
			response, err = vcenter.StartVM(vm.VM)
			if err != nil {
				return nil, err
			}
		}
	}

	return response, nil

}

func (tpl *Template) VMStatus(params map[string]interface{}) ([]byte, error) {

	user, _ := params["user"].(string)
	url, _ := params["url"].(string)
	password, _ := params["password"].(string)
	vms := strings.Split(params["vms"].(string), ",")
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 20
	}
	insecure, _ := params["insecure"].(bool)

	vcenterOptions := vendors.VCenterOptions{
		URL:      url,
		User:     user,
		Password: password,
		Timeout:  timeout,
		Insecure: insecure,
	}

	vmNames := vendors.VCenterVMNameOptions{
		Names: vms,
	}

	vcenterOptions, err := vendors.InitializeVCenterSession(vcenterOptions)
	if err != nil {
		return nil, err
	}

	vcenter := vendors.NewVCenter(vcenterOptions)

	var vi vendors.VMsResponse

	vmInfo, err := vcenter.GetVMsByName(vmNames)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(vmInfo, &vi)
	if err != nil {
		return nil, err
	}

	var response []byte

	if len(vi.Value) > 0 {
		for _, vm := range vi.Value {
			response, err = vcenter.GetVM(vm.VM)
			if err != nil {
				return nil, err
			}
		}
	}

	return response, nil

}

func (tpl *Template) VMReboot(params map[string]interface{}) ([]byte, error) {

	user, _ := params["user"].(string)
	url, _ := params["url"].(string)
	password, _ := params["password"].(string)
	vms := strings.Split(params["vms"].(string), ",")
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 20
	}
	insecure, _ := params["insecure"].(bool)

	vcenterOptions := vendors.VCenterOptions{
		URL:      url,
		User:     user,
		Password: password,
		Timeout:  timeout,
		Insecure: insecure,
	}

	vmNames := vendors.VCenterVMNameOptions{
		Names: vms,
	}

	vcenterOptions, err := vendors.InitializeVCenterSession(vcenterOptions)
	if err != nil {
		return nil, err
	}

	vcenter := vendors.NewVCenter(vcenterOptions)

	var vi vendors.VMsResponse

	vmInfo, err := vcenter.GetVMsByName(vmNames)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(vmInfo, &vi)
	if err != nil {
		return nil, err
	}

	if len(vi.Value) == 0 {
		return nil, fmt.Errorf("no VMs found")
	}

	var response []byte

	if len(vi.Value) > 0 {
		for _, vm := range vi.Value {
			response, err = vcenter.RebootVM(vm.VM)
			if err != nil {
				return nil, err
			}
		}
	}

	return response, nil
}

func (tpl *Template) VMShutdown(params map[string]interface{}) ([]byte, error) {

	user, _ := params["user"].(string)
	url, _ := params["url"].(string)
	password, _ := params["password"].(string)
	vms := strings.Split(params["vms"].(string), ",")
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 20
	}
	insecure, _ := params["insecure"].(bool)

	vcenterOptions := vendors.VCenterOptions{
		URL:      url,
		User:     user,
		Password: password,
		Timeout:  timeout,
		Insecure: insecure,
	}

	vmNames := vendors.VCenterVMNameOptions{
		Names: vms,
	}

	vcenterOptions, err := vendors.InitializeVCenterSession(vcenterOptions)
	if err != nil {
		return nil, err
	}

	vcenter := vendors.NewVCenter(vcenterOptions)

	var vi vendors.VMsResponse

	vmInfo, err := vcenter.GetVMsByName(vmNames)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(vmInfo, &vi)
	if err != nil {
		return nil, err
	}

	if len(vi.Value) == 0 {
		return nil, fmt.Errorf("no VMs found")
	}

	var response []byte

	if len(vi.Value) > 0 {
		for _, vm := range vi.Value {
			response, err = vcenter.ShutdownVM(vm.VM)
			if err != nil {
				return nil, err
			}
		}
	}

	return response, nil
}

func (tpl *Template) CatchpointInstantTest(params map[string]interface{}) ([]byte, error) {

	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 10
	}
	token, _ := params["token"].(string)
	pollTime, _ := params["pollTime"].(int)
	pollDelay, _ := params["pollDelay"].(int)
	pageSize, _ := params["pageSize"].(int)
	pageNumber, _ := params["pageNumber"].(int)

	countries := strings.Split(params["countries"].(string), ",")

	catchpointOptions := vendors.CatchpointOptions{
		APIToken: token,
		Timeout:  timeout,
		Insecure: true,
		Retries:  3,
	}

	catchpoint := vendors.NewCatchpoint(catchpointOptions, tpl.logger)

	nodesOpts := vendors.CatchpointSearchNodesWithOptions{
		Targeted:    false,
		Active:      true,
		Paused:      false,
		NetworkType: 0,
		IPv6:        false,
		PageNumber:  pageNumber,
		PageSize:    pageSize,
	}
	var d vendors.CatchpointSearchNodesWithOptionsResponse

	var nodes []*vendors.Node
	for _, country := range countries {

		ctr := common.CountryByShort(country)
		nodesOpts.Country = ctr
		r, _ := catchpoint.CustomSearchNodesWithOptions(catchpointOptions, nodesOpts)
		json.Unmarshal(r, &d)

		for _, n := range *d.Data.Nodes {
			nodes = append(nodes, &vendors.Node{
				ID:   n.ID,
				Name: n.Name,
				Country: &vendors.CatchpointCountry{
					Name: n.Country.Name,
				},
			})

		}
	}
	if len(nodes) == 0 {
		return nil, fmt.Errorf("no nodes found with the given parameters")
	}

	var nodesStr string
	for _, n := range nodes {
		nodesStr = nodesStr + strconv.Itoa(n.ID) + ","
	}
	nodesStr = nodesStr[:len(nodesStr)-1]
	instantTestOptions := vendors.CatchpointInstantTestOptions{
		URL:             url,
		NodesIds:        nodesStr,
		HTTPMethodType:  0,
		InstantTestType: 0,
		MonitorType:     2,
		OnDemand:        true,
	}

	wmrByte, err := catchpoint.CustomInstantTest(catchpointOptions, instantTestOptions)
	if err != nil {
		return nil, catchpoint.CheckError(wmrByte, err)
	}

	wmr := vendors.CatchpointIstantTestResponse{}
	json.Unmarshal(wmrByte, &wmr)

	var cancel context.CancelFunc
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(pollTime)*time.Second)
	defer cancel()

	var testResults *[]vendors.CatchpointInstantTestResultReponse

	allReady, successNodes := catchpoint.WaitPollSuccessOrCancelDetailed(ctx, pollDelay, wmr.Data.ID, nodes, catchpointOptions)

	if allReady {
		testResults, err = catchpoint.GetLogReport(catchpointOptions, wmr.Data.ID, nodes)
		if err != nil {
			return nil, fmt.Errorf("unable to get the test results: %w", err)
		}
	} else {
		if len(successNodes) > 0 {
			testResults, err = catchpoint.GetLogReport(catchpointOptions, wmr.Data.ID, successNodes)
			if err != nil {
				return nil, fmt.Errorf("unable to get the test results: %w", err)
			}
		} else {
			return nil, fmt.Errorf("no nodes succeeded in polling;")
		}
	}

	summary, err := catchpoint.GenerateSummary(testResults)
	if err != nil {
		return nil, fmt.Errorf("unable to generate the summary")
	}
	summaryBytes, _ := json.Marshal(summary)

	return summaryBytes, nil
}

func (tpl *Template) K8sResourceDescribe(params map[string]interface{}) ([]byte, error) {

	config, _ := params["config"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout <= 0 {
		timeout = 30
	}

	options := vendors.K8sOptions{
		Config:  config,
		Timeout: timeout,
	}

	k8s := vendors.NewK8s(options, tpl.logger)

	kind, _ := params["kind"].(string)
	namespace, _ := params["namespace"].(string)
	name, _ := params["name"].(string)

	describeOptions := vendors.K8sResourceDescribeOptions{
		K8sResourceOptions: vendors.K8sResourceOptions{
			Kind:      kind,
			Namespace: namespace,
			Name:      name,
		},
	}

	return k8s.CustomResourceDescribe(options, describeOptions)
}

func (tpl *Template) K8sResourceDelete(params map[string]interface{}) ([]byte, error) {

	config, _ := params["config"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout <= 0 {
		timeout = 30
	}

	options := vendors.K8sOptions{
		Config:  config,
		Timeout: timeout,
	}

	k8s := vendors.NewK8s(options, tpl.logger)

	kind, _ := params["kind"].(string)
	namespace, _ := params["namespace"].(string)
	name, _ := params["name"].(string)

	deleteOptions := vendors.K8sResourceDeleteOptions{
		K8sResourceOptions: vendors.K8sResourceOptions{
			Kind:      kind,
			Namespace: namespace,
			Name:      name,
		},
	}

	return k8s.CustomResourceDelete(options, deleteOptions)
}

func (tpl *Template) K8sResourceScale(params map[string]interface{}) ([]byte, error) {

	config, _ := params["config"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout <= 0 {
		timeout = 30
	}

	options := vendors.K8sOptions{
		Config:  config,
		Timeout: timeout,
	}

	k8s := vendors.NewK8s(options, tpl.logger)

	kind, _ := params["kind"].(string)
	namespace, _ := params["namespace"].(string)
	name, _ := params["name"].(string)

	replicas := tpl.paramAsInt(params["replicas"], 0)
	waitTimeout := tpl.paramAsInt(params["waitTimeout"], 0)
	pollTimeout := tpl.paramAsInt(params["pollTimeout"], 1)

	scaleOptions := vendors.K8sResourceScaleOptions{
		K8sResourceOptions: vendors.K8sResourceOptions{
			Kind:      kind,
			Namespace: namespace,
			Name:      name,
		},
		Replicas:    replicas,
		WaitTimeout: waitTimeout,
		PollTimeout: pollTimeout,
	}

	return k8s.CustomResourceScale(options, scaleOptions)
}

func (tpl *Template) paramAsInt(param interface{}, def int) int {

	if utils.IsEmpty(param) {
		return def
	}
	value, ok := param.(int)
	if !ok {
		valueStr := fmt.Sprintf("%v", param)
		value, _ = strconv.Atoi(valueStr)
	}
	if utils.IsEmpty(param) {
		return def
	}
	return value
}

func (tpl *Template) K8sResourceRestart(params map[string]interface{}) ([]byte, error) {

	config, _ := params["config"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout <= 0 {
		timeout = 30
	}

	options := vendors.K8sOptions{
		Config:  config,
		Timeout: timeout,
	}

	k8s := vendors.NewK8s(options, tpl.logger)

	kind, _ := params["kind"].(string)
	namespace, _ := params["namespace"].(string)
	name, _ := params["name"].(string)

	waitTimeout := tpl.paramAsInt(params["waitTimeout"], 5)
	pollTimeout := tpl.paramAsInt(params["pollTimeout"], 1)

	restartOptions := vendors.K8sResourceRestartOptions{
		K8sResourceOptions: vendors.K8sResourceOptions{
			Kind:      kind,
			Namespace: namespace,
			Name:      name,
		},
		WaitTimeout: waitTimeout,
		PollTimeout: pollTimeout,
	}

	return k8s.CustomResourceRestart(options, restartOptions)
}

func (tpl *Template) DirRemove(path string, recursive bool) error {

	var err error
	if recursive {
		err = os.RemoveAll(path)
	} else {
		err = os.Remove(path)
	}
	if err != nil {
		return err
	}
	return nil
}

func (tpl *Template) Exec(path string, timeout int, params []string) ([]byte, error) {

	name := filepath.Base(path)
	dir := filepath.Dir(path)

	ctx := context.Background()
	if timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
		defer cancel()
	}

	cmd := exec.CommandContext(ctx, name, params...)
	cmd.Dir = dir

	r, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("%s: %s", err, r)
	}
	return r, nil
}

// setTemplateFuncs registers render's full funcmap into funcs: first the
// lite, dependency-light function set from render/lite, then the heavy,
// vendor-backed functions defined in this package. templateRender and
// templateRenderFile are re-registered on top of the lite versions so that
// nested renders keep access to the full funcmap, matching this method's
// pre-render/lite behavior.
func (tpl *Template) setTemplateFuncs(funcs map[string]any) {

	tpl.Template.SetFuncs(funcs)

	funcs["urlWait"] = tpl.URLWait
	funcs["gitlabPipelineVars"] = tpl.GitlabPipelineVars
	funcs["sleep"] = tpl.Sleep

	funcs["catchpointInstantTest"] = tpl.CatchpointInstantTest

	funcs["jiraSearchAssets"] = tpl.JiraSearchAssets
	funcs["jiraCreateIssue"] = tpl.JiraCreateIssue
	funcs["jiraSearchIssue"] = tpl.JiraSearchIssue
	funcs["jiraCreateAsset"] = tpl.JiraCreateAsset
	funcs["jiraMoveIssue"] = tpl.JiraMoveIssue
	funcs["jiraAddComment"] = tpl.JiraAddComment
	funcs["jiraUpdateIssue"] = tpl.JiraUpdateIssue
	funcs["jiraIssueTransition"] = tpl.JiraIssueTransition
	funcs["jiraGetIssueTransition"] = tpl.JiraGetIssueTransition
	funcs["jiraUpdateAsset"] = tpl.JiraUpdateAsset
	funcs["jiraGetUserByEmail"] = tpl.JiraGetUserByEmail

	funcs["grafanaGetAlerts"] = tpl.grafanaGetAlerts
	funcs["grafanaCreateDashboard"] = tpl.GrafanaCreateDashboard
	funcs["grafanaCopyDashboard"] = tpl.GrafanaCopyDashboard

	funcs["pagerDutyCreateIncident"] = tpl.PagerDutyCreateIncident
	funcs["pagerDutySendNoteToIncident"] = tpl.PagerDutySendNoteToIncident

	// Override the lite versions so nested renders see this template's
	// full funcmap, including the heavy functions registered above.
	funcs["templateRender"] = tpl.TemplateRender
	funcs["templateRenderFile"] = tpl.TemplateRenderFile

	funcs["googleCalendarGetEvents"] = tpl.GoogleCalendarGetEvents
	funcs["googleCalendarInsertEvent"] = tpl.GoogleCalendarInsertEvent
	funcs["googleCalendarDeleteEvents"] = tpl.GoogleCalendarDeleteEvents
	funcs["googleMeetCreateSpace"] = tpl.GoogleMeetCreateSpace
	funcs["googleDocsCopyDocument"] = tpl.GoogleDocsCopyDocument

	funcs["sshRun"] = tpl.SSHRun
	funcs["listFilesWithModTime"] = tpl.ListFilesWithModTime

	funcs["vmReset"] = tpl.VMReset
	funcs["vmStart"] = tpl.VMStart
	funcs["vmStop"] = tpl.VMStop
	funcs["vmStatus"] = tpl.VMStatus
	funcs["vmReboot"] = tpl.VMReboot
	funcs["vmShutdown"] = tpl.VMShutdown

	funcs["awsS3ListObjects"] = tpl.AWSS3ListObjects
	funcs["awsS3GetObject"] = tpl.AWSS3GetObject
	funcs["awsS3PutObject"] = tpl.AWSS3PutObject

	funcs["ldapGetGroupMember"] = tpl.LdapGetGroupMembers

	funcs["prometheusGet"] = tpl.PrometheusGet

	funcs["k8sResourceDescribe"] = tpl.K8sResourceDescribe
	funcs["k8sResourceDelete"] = tpl.K8sResourceDelete
	funcs["k8sResourceScale"] = tpl.K8sResourceScale
	funcs["k8sResourceRestart"] = tpl.K8sResourceRestart

	funcs["dirCreate"] = tpl.DirCreate
	funcs["dirRemove"] = tpl.DirRemove
	funcs["fileCreate"] = tpl.FileCreate
	funcs["exec"] = tpl.Exec

	tpl.funcs = funcs
}

// filterFuncsByContent restricts funcs to the entries whose name appears as
// a substring of content.
func filterFuncsByContent(funcs map[string]any, content string) map[string]any {

	m := make(map[string]any)
	for k, v := range funcs {
		if strings.Contains(content, k) {
			m[k] = v
		}
	}
	return m
}

func (tpl *TextTemplate) customRender(name string, obj interface{}) ([]byte, error) {

	var b bytes.Buffer
	var err error

	if empty, _ := tpl.IsEmpty(name); empty {
		err = tpl.template.Execute(&b, obj)
	} else {
		err = tpl.template.ExecuteTemplate(&b, name, obj)
	}
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// CustomRenderWithOptions renders the template using opts.Name and, when
// opts.Object is set, opts.Object unmarshaled as JSON as the data object.
func (tpl *TextTemplate) CustomRenderWithOptions(opts TemplateOptions) ([]byte, error) {

	var obj interface{}
	if !utils.IsEmpty(opts.Object) {
		err := json.Unmarshal([]byte(opts.Object), &obj)
		if err != nil {
			return nil, err
		}
	}
	return tpl.customRender(tpl.options.Name, obj)
}

// Render renders the template using its configured options.
func (tpl *TextTemplate) Render() ([]byte, error) {
	return tpl.CustomRenderWithOptions(tpl.options)
}

// RenderObject renders the template using obj as the data object.
func (tpl *TextTemplate) RenderObject(obj interface{}) ([]byte, error) {
	return tpl.customRender(tpl.options.Name, obj)
}

// NewTextTemplate parses options.Content as a text/template.Template, with
// render's full funcmap (sprig, the lite function set, and the heavy,
// vendor-backed functions defined in this package, all overridable by
// options.Funcs) available to it, and returns the resulting TextTemplate.
func NewTextTemplate(options TemplateOptions, logger common.Logger) (*TextTemplate, error) {

	if utils.IsEmpty(options.Content) {
		return nil, errors.New("no content")
	}

	tpl := &TextTemplate{}
	tpl.Template.Configure(options.liteOptions(), logger)

	funcs := sprig.TxtFuncMap()
	tpl.setTemplateFuncs(funcs)
	for k, v := range options.Funcs {
		funcs[k] = v
	}

	if options.FilterFuncs {
		funcs = filterFuncsByContent(funcs, options.Content)
	}

	t, err := txtTemplate.New(options.Name).Funcs(funcs).Parse(options.Content)
	if err != nil {
		return nil, err
	}

	if !utils.IsEmpty(options.Files) {
		t, err = t.ParseFiles(options.Files...)
		if err != nil {
			return nil, err
		}
	}

	if !utils.IsEmpty(options.Pattern) {
		t, err = t.ParseGlob(options.Pattern)
		if err != nil {
			return nil, err
		}
	}

	tpl.Template.SetNode(t)
	tpl.template = t
	tpl.funcs = funcs
	tpl.options = options
	tpl.logger = logger
	return tpl, nil
}

func (tpl *HtmlTemplate) customRender(name string, obj interface{}) ([]byte, error) {

	var b bytes.Buffer
	var err error

	if empty, _ := tpl.IsEmpty(name); empty {
		err = tpl.template.Execute(&b, obj)
	} else {
		err = tpl.template.ExecuteTemplate(&b, name, obj)
	}
	if err != nil {
		return nil, err
	}
	return b.Bytes(), nil
}

// CustomRenderWithOptions renders the template using opts.Name and, when
// opts.Object is set, opts.Object unmarshaled as JSON as the data object.
func (tpl *HtmlTemplate) CustomRenderWithOptions(opts TemplateOptions) ([]byte, error) {

	var obj interface{}
	if !utils.IsEmpty(opts.Object) {
		err := json.Unmarshal([]byte(opts.Object), &obj)
		if err != nil {
			return nil, err
		}
	}
	return tpl.customRender(tpl.options.Name, obj)
}

// Render renders the template using its configured options.
func (tpl *HtmlTemplate) Render() ([]byte, error) {
	return tpl.CustomRenderWithOptions(tpl.options)
}

// RenderObject renders the template using obj as the data object.
func (tpl *HtmlTemplate) RenderObject(obj interface{}) ([]byte, error) {
	return tpl.customRender(tpl.options.Name, obj)
}

// NewHtmlTemplate parses options.Content as an html/template.Template, with
// render's full funcmap (sprig, the lite function set, and the heavy,
// vendor-backed functions defined in this package, all overridable by
// options.Funcs) available to it, and returns the resulting HtmlTemplate.
func NewHtmlTemplate(options TemplateOptions, logger common.Logger) (*HtmlTemplate, error) {

	if utils.IsEmpty(options.Content) {
		return nil, errors.New("no content")
	}

	tpl := &HtmlTemplate{}
	tpl.Template.Configure(options.liteOptions(), logger)

	funcs := sprig.HtmlFuncMap()
	tpl.setTemplateFuncs(funcs)
	for k, v := range options.Funcs {
		funcs[k] = v
	}

	if options.FilterFuncs {
		funcs = filterFuncsByContent(funcs, options.Content)
	}

	t, err := htmlTemplate.New(options.Name).Funcs(funcs).Parse(options.Content)
	if err != nil {
		return nil, err
	}

	if !utils.IsEmpty(options.Files) {
		t, err = t.ParseFiles(options.Files...)
		if err != nil {
			return nil, err
		}
	}

	if !utils.IsEmpty(options.Pattern) {
		t, err = t.ParseGlob(options.Pattern)
		if err != nil {
			return nil, err
		}
	}

	tpl.Template.SetNode(t)
	tpl.template = t
	tpl.funcs = funcs
	tpl.options = options
	tpl.logger = logger
	return tpl, nil
}
