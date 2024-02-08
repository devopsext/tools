package render

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	htmlTemplate "html/template"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"text/template"
	txtTemplate "text/template"
	"time"

	"github.com/Masterminds/sprig/v3"
	"github.com/araddon/dateparse"
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	utils "github.com/devopsext/utils"

	"github.com/tidwall/gjson"
)

type TemplateOptions struct {
	Name        string
	Object      string
	Content     string
	Files       []string
	TimeFormat  string
	Pattern     string
	Funcs       map[string]any
	FilterFuncs bool
}

type Template struct {
	options TemplateOptions
	logger  common.Logger
	funcs   template.FuncMap
}

type TextTemplate struct {
	Template
	template *txtTemplate.Template
}

type HtmlTemplate struct {
	Template
	template *htmlTemplate.Template
}

// put errors to logger
func (tpl *Template) LogError(obj interface{}, args ...interface{}) (string, error) {
	if tpl.logger == nil {
		return "", nil
	}
	tpl.logger.Error(obj, args...)
	return "", nil
}

// put warnings to logger
func (tpl *Template) LogWarn(obj interface{}, args ...interface{}) (string, error) {
	if tpl.logger == nil {
		return "", nil
	}
	tpl.logger.Warn(obj, args...)
	return "", nil
}

// put warnings to logger
func (tpl *Template) LogDebug(obj interface{}, args ...interface{}) (string, error) {
	if tpl.logger == nil {
		return "", nil
	}
	tpl.logger.Debug(obj, args...)
	return "", nil
}

// put information to logger
func (tpl *Template) LogInfo(obj interface{}, args ...interface{}) (string, error) {
	if tpl.logger == nil {
		return "", nil
	}
	tpl.logger.Info(obj, args...)
	return "", nil
}

// replaceAll replaces all occurrences of a value in a string with the given
// replacement value.
func (tpl *Template) ReplaceAll(f, t, s string) (string, error) {
	return strings.Replace(s, f, t, -1), nil
}

// regexReplaceAll replaces all occurrences of a regular expression with
// the given replacement value.
func (tpl *Template) RegexReplaceAll(re, pl, s string) (string, error) {
	compiled, err := regexp.Compile(re)
	if err != nil {
		return "", err
	}
	return compiled.ReplaceAllString(s, pl), nil
}

// regexMatch returns true or alse if the string matches
// the given regular expression
func (tpl *Template) RegexMatch(re, s string) (bool, error) {
	compiled, err := regexp.Compile(re)
	if err != nil {
		return false, err
	}
	return compiled.MatchString(s), nil
}

func (tpl *Template) RegexFindSubmatch(regex string, s string) []string {
	r := regexp.MustCompile(regex)
	return r.FindStringSubmatch(s)
}

func (tpl *Template) regexMatchFindKey(v interface{}, field, value string) bool {

	m, ok := v.(map[string]interface{})
	if !ok {
		return false
	}
	if m[field] == nil {
		return false
	}
	s := fmt.Sprintf("%v", m[field])
	match, _ := regexp.MatchString(fmt.Sprintf("^%s", s), value)
	return match
}

func (tpl *Template) RegexMatchFindKeys(obj interface{}, field, value string) []interface{} {

	var r []interface{}
	if obj == nil || utils.IsEmpty(field) || utils.IsEmpty(value) {
		return r
	}

	a, ok := obj.([]interface{})
	if ok {
		for k, v := range a {
			if !tpl.regexMatchFindKey(v, field, value) {
				continue
			}
			r = append(r, k)
		}
		return r
	}

	m, ok := obj.(map[string]interface{})
	if ok {
		for k, v := range m {
			if !tpl.regexMatchFindKey(v, field, value) {
				continue
			}
			r = append(r, k)
		}
	}

	return r
}

func (tpl *Template) RegexMatchFindKey(obj interface{}, field, value string) interface{} {

	keys := tpl.RegexMatchFindKeys(obj, field, value)
	if len(keys) == 0 {
		return value
	}
	return keys[0]
}

func (tpl *Template) RegexMatchObjectByField(obj interface{}, field, value string) interface{} {

	if obj == nil {
		return nil
	}
	key := tpl.RegexMatchFindKey(obj, field, value)
	if key == value {
		return nil
	}

	a, ok := obj.([]interface{})
	ka, _ := key.(int)
	if ok {
		return a[ka]
	}

	m, ok := obj.(map[string]interface{})
	km, _ := key.(string)
	if ok {
		return m[km]
	}

	return nil
}

func (tpl *Template) Compare(v1, v2 interface{}) bool {

	if v1 == v2 {
		return true
	}
	if utils.IsEmpty(v1) && utils.IsEmpty(v2) {
		return true
	}
	if utils.IsEmpty(v1) {
		return false
	}
	if utils.IsEmpty(v2) {
		return false
	}

	switch v1.(type) {
	case int, int16, int32, int64:
		v2s := fmt.Sprintf("%v", v2)
		v21, err := strconv.ParseInt(v2s, 10, 64)
		if err == nil {
			v11, _ := v1.(int64)
			return v11 == v21
		}
	case float32, float64:
		v2s := fmt.Sprintf("%v", v2)
		v21, err := strconv.ParseFloat(v2s, 64)
		if err == nil {
			v11, _ := v1.(float64)
			return v11 == v21
		}
	default:
		v21 := fmt.Sprintf("%v", v2)
		return v1.(string) == v21
	}
	return false
}

func (tpl *Template) findKey(v interface{}, field string, value interface{}) bool {

	m, ok := v.(map[string]interface{})
	if !ok {
		return false
	}
	if m[field] == nil {
		return false
	}

	return tpl.Compare(m[field], value)
}

func (tpl *Template) FindKeys(obj interface{}, field string, value interface{}) []interface{} {

	var r []interface{}
	if obj == nil || utils.IsEmpty(field) || utils.IsEmpty(value) {
		return r
	}

	a, ok := obj.([]interface{})
	if ok {
		for k, v := range a {
			if !tpl.findKey(v, field, value) {
				continue
			}
			r = append(r, k)
		}
		return r
	}

	m, ok := obj.(map[string]interface{})
	if ok {
		for k, v := range m {
			if !tpl.findKey(v, field, value) {
				continue
			}
			r = append(r, k)
		}
	}

	return r
}

func (tpl *Template) FindKey(obj interface{}, field string, value interface{}) interface{} {

	keys := tpl.FindKeys(obj, field, value)
	if len(keys) == 0 {
		return value
	}
	return keys[0]
}

func (tpl *Template) FindObjectByField(obj interface{}, field string, value interface{}) interface{} {

	if obj == nil {
		return nil
	}
	key := tpl.FindKey(obj, field, value)
	if key == value {
		return nil
	}

	a, ok := obj.([]interface{})
	ka, _ := key.(int)
	if ok {
		return a[ka]
	}

	m, ok := obj.(map[string]interface{})
	km, _ := key.(string)
	if ok {
		return m[km]
	}

	return nil
}

// toLower converts the given string (usually by a pipe) to lowercase.
func (tpl *Template) ToLower(s string) (string, error) {
	return strings.ToLower(s), nil
}

// toTitle converts the given string (usually by a pipe) to titlecase.
func (tpl *Template) ToTitle(s string) (string, error) {
	return strings.Title(s), nil
}

// toUpper converts the given string (usually by a pipe) to uppercase.
func (tpl *Template) ToUpper(s string) (string, error) {
	return strings.ToUpper(s), nil
}

// toJSON converts the given structure into a deeply nested JSON string.
func (tpl *Template) ToJson(i interface{}) (string, error) {

	result, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(result)), err
}

func (tpl *Template) FromJson(i interface{}) (interface{}, error) {

	var d []byte
	var r interface{}
	ds, ok := i.([]byte)
	if ok {
		d = ds
	} else {
		d = []byte(fmt.Sprintf("%v", i))
	}
	err := json.Unmarshal(d, &r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// split is a version of strings.Split that can be piped
func (tpl *Template) Split(sep, s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}, nil
	}
	return strings.Split(s, sep), nil
}

// join is a version of strings.Join that can be piped
func (tpl *Template) Join(sep string, a []string) (string, error) {
	return strings.Join(a, sep), nil
}

func (tpl *Template) IsEmpty(s string) (bool, error) {
	s1 := strings.TrimSpace(s)
	return len(s1) == 0, nil
}

func (tpl *Template) Env(key string) (string, error) {
	return utils.EnvGet(key, "").(string), nil
}

func (tpl *Template) TimeFormat(s string, format string) (string, error) {

	t, err := time.Parse(tpl.options.TimeFormat, s)
	if err != nil {

		return s, err
	}
	return t.Format(format), nil
}

func (tpl *Template) TimeNano(s string) (string, error) {

	t1, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(t1.UnixNano(), 10), nil
}

func (tpl *Template) JsonEscape(s string) (string, error) {

	bytes, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// toString converts the given value to string
func (tpl *Template) ToString(i interface{}) (string, error) {

	if !utils.IsEmpty(i) {

		ds, ok := i.([]byte)
		if ok {
			return string(ds), nil
		}
		return fmt.Sprintf("%v", i), nil
	}
	return "", nil
}

func (tpl *Template) EscapeString(s string) (string, error) {
	return html.EscapeString(s), nil
}

func (tpl *Template) UnescapeString(s string) (string, error) {
	return html.UnescapeString(s), nil
}

func (tpl *Template) Jsonata(data interface{}, query string) (string, error) {

	if utils.IsEmpty(query) {
		return "", errors.New("query is empty")
	}

	if _, err := os.Stat(query); err == nil {
		content, err := os.ReadFile(query)
		if err != nil {
			return "", err
		}
		query = string(content)
	}

	s, ok := data.(string) // possibly json as string
	if ok {

		if _, err := os.Stat(s); err == nil {
			content, err := os.ReadFile(s)
			if err != nil {
				return "", err
			}
			s = string(content)
		}

		var v interface{}
		err := json.Unmarshal([]byte(s), &v)
		if err != nil {
			return "", err
		}
		data = v
	}

	jnata := common.NewJsonata(common.JsonataOptions{})
	m, err := jnata.Eval(data, query)
	if err != nil {
		return "", err
	}

	ret := ""

	_, ok = m.(map[string]interface{}) // could be as object
	if ok {
		b, err := common.JsonMarshal(m)
		if err != nil {
			return "", err
		}
		ret = strings.TrimSpace(string(b)) // issue with adding new line
	} else {
		ret = fmt.Sprintf("%v", m)
	}

	return ret, nil
}

func (tpl *Template) Gjson(obj interface{}, path string) (string, error) {

	if utils.IsEmpty(path) {
		err := errors.New("path is empty")
		return "", err
	}

	if obj == nil {
		err := errors.New("object is not defined")
		return "", err
	}

	json := ""
	v, ok := obj.(string)
	if ok {
		if _, err := os.Stat(v); err == nil {
			bytes, err := os.ReadFile(v)
			if err != nil {
				return "", err
			}
			json = string(bytes)
		}
	}

	if utils.IsEmpty(json) {
		bytes, err := common.JsonMarshal(obj)
		if err != nil {
			return "", err
		}
		json = string(bytes)
	}

	if utils.IsEmpty(json) {
		err := errors.New("json is empty")
		return "", err
	}

	value := gjson.Get(json, path)
	return value.String(), nil
}

func (tpl *Template) IfDef(i interface{}, def string) (string, error) {

	if utils.IsEmpty(i) {
		return def, nil
	}
	return tpl.ToString(i)
}

func (tpl *Template) IfElse(o interface{}, vars []interface{}) interface{} {

	if len(vars) == 0 {
		return o
	}
	for k, v := range vars {
		if k%2 == 0 {
			if o == v && len(vars) > k+1 {
				return vars[k+1]
			}
		}
	}
	return o
}

func (tpl *Template) IfIP(obj interface{}) bool {

	if obj == nil {
		return false
	}

	a := net.ParseIP(fmt.Sprintf("%v", obj))
	return a != nil
}

func (tpl *Template) IfIPAndPort(obj interface{}) bool {

	if obj == nil {
		return false
	}
	s := fmt.Sprintf("%v", obj)
	arr := strings.Split(s, ":")
	if len(arr) > 0 {
		s = strings.TrimSpace(arr[0])
	} else {
		return false
	}
	return tpl.IfIP(s)
}

func (tpl *Template) Content(s string) (string, error) {

	if utils.IsEmpty(s) {
		return "", nil
	}

	bytes, err := utils.Content(s)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
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

func (tpl *Template) TagExists(s, key string) (bool, error) {

	// DataDog tags
	tags := strings.Split(s, ",")
	if len(tags) > 0 {
		for _, tag := range tags {
			kv := strings.Split(tag, ":")
			k := ""
			if len(kv) > 0 {
				k = kv[0]
			}
			if strings.TrimSpace(k) == strings.TrimSpace(key) {
				return true, nil
			}
		}
	}
	return false, nil
}

func (tpl *Template) TagValue(s, key string) (string, error) {

	// DataDog tags
	tags := strings.Split(s, ",")
	if len(tags) > 0 {
		for _, tag := range tags {
			kv := strings.Split(tag, ":")
			k := ""
			v := ""
			if len(kv) > 0 {
				k = kv[0]
			}
			if len(kv) > 1 {
				v = kv[1]
			}
			if strings.TrimSpace(k) == strings.TrimSpace(key) {
				return v, nil
			}
		}
	}
	return s, nil
}

func (tpl *Template) DateParse(d string) (time.Time, error) {
	t, err := dateparse.ParseAny(d)
	if err != nil {
		return time.Now(), err
	}
	return t, nil
}

// url, contentType, authorization string, timeout int
func (tpl *Template) HttpGet(params map[string]interface{}) ([]byte, error) {

	if len(params) == 0 {
		return nil, fmt.Errorf("HttpGet err => %s", "no params allowed")
	}

	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 5
	}

	insecure, _ := params["insecure"].(bool)
	contentType, _ := params["contentType"].(string)
	authorization, _ := params["authorization"].(string)

	var transport = &http.Transport{
		Dial:                (&net.Dialer{Timeout: time.Duration(timeout) * time.Second}).Dial,
		TLSHandshakeTimeout: time.Duration(timeout) * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: insecure},
	}

	client := http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}
	return utils.HttpGetRaw(&client, url, contentType, authorization)
}

func (tpl *Template) HttpPost(params map[string]interface{}) ([]byte, error) {

	if len(params) == 0 {
		return nil, fmt.Errorf("HttpPost err => %s", "no params allowed")
	}

	url, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 5
	}

	insecure, _ := params["insecure"].(bool)
	contentType, _ := params["contentType"].(string)
	authorization, _ := params["authorization"].(string)

	var body []byte
	b := params["body"]
	if !utils.IsEmpty(b) {
		switch b.(type) {
		case string:
			bs, _ := b.(string)
			body = []byte(bs)
		case []byte:
			body, _ = b.([]byte)
		default:
			bs := fmt.Sprintf("%s", b)
			body = []byte(bs)
		}
	}

	var transport = &http.Transport{
		Dial:                (&net.Dialer{Timeout: time.Duration(timeout) * time.Second}).Dial,
		TLSHandshakeTimeout: time.Duration(timeout) * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: insecure},
	}

	client := http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}
	return utils.HttpPostRaw(&client, url, contentType, authorization, body)
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

	assetsOptions := vendors.JiraSearchAssetsOptions{
		SearchPattern: query,
		ResultPerPage: limit,
	}

	return jira.SearchAssets(assetsOptions)
}

func (tpl *Template) JiraCreateIssue(params map[string]interface{}) ([]byte, error) {

	if len(params) == 0 {
		return nil, fmt.Errorf("JiraSearchAssets err => %s", "no params allowed")
	}

	return []byte("asdadasdad"), nil
}

func (tpl *Template) PagerDutyCreateIncident(params map[string]interface{}) ([]byte, error) {

	if len(params) == 0 {
		return nil, fmt.Errorf("PagerDutyEscalate err => %s", "no params allowed")
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

	incidentOptions := vendors.PagerDutyIncidentOptions{
		Title:      title,
		Body:       body,
		Urgency:    urgency,
		ServiceID:  serviceID,
		PriorityID: priorityID,
	}

	from, _ := params["from"].(string)

	createOptions := vendors.PagerDutyCreateIncidentOptions{
		From: from,
	}

	return pagerDuty.CreateIncident(incidentOptions, createOptions)
}

func (tpl *Template) TemplateRenderFile(path string, obj interface{}) (string, error) {

	content, err := utils.Content(path)
	if err != nil {
		return "", err
	}

	opts := TemplateOptions{
		Content:     string(content),
		Funcs:       tpl.funcs,
		FilterFuncs: false,
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

	calendarGetEventsOptions := vendors.GoogleCalendarGetEventsOptions{
		TimeMin:            timeMin,
		TimeMax:            timeMax,
		AlwaysIncludeEmail: true,
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

func (tpl *Template) setTemplateFuncs(funcs map[string]any) {

	funcs["logError"] = tpl.LogError
	funcs["logWarn"] = tpl.LogWarn
	funcs["logDebug"] = tpl.LogDebug
	funcs["logInfo"] = tpl.LogInfo

	funcs["regexReplaceAll"] = tpl.RegexReplaceAll
	funcs["regexMatch"] = tpl.RegexMatch
	funcs["regexFindSubmatch"] = tpl.RegexFindSubmatch

	funcs["regexMatchFindKeys"] = tpl.RegexMatchFindKeys
	funcs["regexMatchFindKey"] = tpl.RegexMatchFindKey
	funcs["regexMatchObjectByField"] = tpl.RegexMatchObjectByField

	funcs["findKeys"] = tpl.FindKeys
	funcs["findKey"] = tpl.FindKey
	funcs["findObjectByField"] = tpl.FindObjectByField

	funcs["replaceAll"] = tpl.ReplaceAll
	funcs["toLower"] = tpl.ToLower
	funcs["toTitle"] = tpl.ToTitle
	funcs["toUpper"] = tpl.ToUpper
	funcs["toJSON"] = tpl.ToJson // deprecated
	funcs["toJson"] = tpl.ToJson
	funcs["fromJson"] = tpl.FromJson
	funcs["split"] = tpl.Split
	funcs["join"] = tpl.Join
	funcs["isEmpty"] = tpl.IsEmpty
	funcs["env"] = tpl.Env
	funcs["getEnv"] = tpl.Env // deprecated
	funcs["timeFormat"] = tpl.TimeFormat
	funcs["timeNano"] = tpl.TimeNano
	funcs["jsonEscape"] = tpl.JsonEscape
	funcs["toString"] = tpl.ToString
	funcs["escapeString"] = tpl.EscapeString
	funcs["unescapeString"] = tpl.UnescapeString
	funcs["jsonata"] = tpl.Jsonata
	funcs["gjson"] = tpl.Gjson
	funcs["ifDef"] = tpl.IfDef
	funcs["ifElse"] = tpl.IfElse
	funcs["ifIP"] = tpl.IfIP
	funcs["ifIPAndPort"] = tpl.IfIPAndPort
	funcs["content"] = tpl.Content
	funcs["urlWait"] = tpl.URLWait
	funcs["gitlabPipelineVars"] = tpl.GitlabPipelineVars
	funcs["tagExists"] = tpl.TagExists
	funcs["tagValue"] = tpl.TagValue
	funcs["dateParse"] = tpl.DateParse

	funcs["httpGet"] = tpl.HttpGet
	funcs["httpPost"] = tpl.HttpPost
	funcs["jiraSearchAssets"] = tpl.JiraSearchAssets
	funcs["jiraCreateIssue"] = tpl.JiraCreateIssue
	funcs["pagerDutyCreateIncident"] = tpl.PagerDutyCreateIncident
	funcs["templateRenderFile"] = tpl.TemplateRenderFile
	funcs["googleCalendarGetEvents"] = tpl.GoogleCalendarGetEvents
	funcs["googleCalendarInsertEvent"] = tpl.GoogleCalendarInsertEvent
}

func (tpl *Template) filterFuncsByContent(funcs map[string]any, content string) map[string]any {

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

func (tpl *TextTemplate) Render() ([]byte, error) {
	return tpl.CustomRenderWithOptions(tpl.options)
}

func (tpl *TextTemplate) RenderObject(obj interface{}) ([]byte, error) {
	return tpl.customRender(tpl.options.Name, obj)
}

func NewTextTemplate(options TemplateOptions, logger common.Logger) (*TextTemplate, error) {

	if utils.IsEmpty(options.Content) {
		return nil, errors.New("no content")
	}

	var tpl = TextTemplate{}
	var t *txtTemplate.Template

	funcs := sprig.TxtFuncMap()
	tpl.setTemplateFuncs(funcs)
	for k, v := range options.Funcs {
		funcs[k] = v
	}

	if options.FilterFuncs {
		funcs = tpl.filterFuncsByContent(funcs, options.Content)
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

	tpl.template = t
	tpl.funcs = funcs
	tpl.options = options
	tpl.logger = logger
	return &tpl, nil
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

func (tpl *HtmlTemplate) Render() ([]byte, error) {
	return tpl.CustomRenderWithOptions(tpl.options)
}

func (tpl *HtmlTemplate) RenderObject(obj interface{}) ([]byte, error) {
	return tpl.customRender(tpl.options.Name, obj)
}

func NewHtmlTemplate(options TemplateOptions, logger common.Logger) (*HtmlTemplate, error) {

	if utils.IsEmpty(options.Content) {
		return nil, errors.New("no content")
	}

	var tpl = HtmlTemplate{}
	var t *htmlTemplate.Template

	funcs := sprig.HtmlFuncMap()
	tpl.setTemplateFuncs(funcs)
	for k, v := range options.Funcs {
		funcs[k] = v
	}

	if options.FilterFuncs {
		funcs = tpl.filterFuncsByContent(funcs, options.Content)
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

	tpl.template = t
	tpl.funcs = funcs
	tpl.options = options
	tpl.logger = logger
	return &tpl, nil
}
