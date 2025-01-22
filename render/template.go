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
	"path/filepath"
	"regexp"
	"sort"
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
	tpl     interface{}
}

type TextTemplate struct {
	Template
	template *txtTemplate.Template
}

type HtmlTemplate struct {
	Template
	template *htmlTemplate.Template
}

func (tpl *Template) ParserLine() (int, error) {

	var i interface{} = tpl.tpl

	line := 0

	txt, ok := i.(*txtTemplate.Template)
	if ok {
		line = int(txt.Tree.Root.Pos)

		for _, v := range txt.Root.Nodes {

			l, c := txt.ErrorContext(v)
			tpl.logger.Debug("%s | %d | %s = %s", v.String(), v.Position(), l, c)
		}
	}

	html, ok := i.(htmlTemplate.Template)
	if ok {
		line = int(html.Tree.Root.Pos)
	}

	return line, nil
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
	if key == nil {
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
	case []interface{}, []string:
		return utils.Contains(v1, v2)
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

func (tpl *Template) FindObject(obj interface{}, field string, value interface{}) interface{} {

	if obj == nil {
		return nil
	}
	key := tpl.FindKey(obj, field, value)
	if key == nil {
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

func (tpl *Template) FindObjects(obj interface{}, field string, value interface{}) []interface{} {

	r := []interface{}{}
	if obj == nil {
		return r
	}
	keys := tpl.FindKeys(obj, field, value)
	if len(keys) == 0 {
		return r
	}

	a, ok := obj.([]interface{})
	if ok {
		for _, v := range keys {
			ka, _ := v.(int)
			r = append(r, a[ka])
		}
		return r
	}

	m, ok := obj.(map[string]interface{})
	if ok {
		for _, v := range keys {
			km, _ := v.(string)
			r = append(r, m[km])
		}
		return r
	}
	return r
}

func (tpl *Template) CountOccurrences(list []interface{}) map[string]int {
	occurrences := make(map[string]int)
	for _, item := range list {
		occurrences[item.(string)]++
	}

	return occurrences
}

func (tpl *Template) SortOccurrences(occurrences map[string]int, sep string, count int) []string {
	keys := make([]string, 0, len(occurrences))
	for k := range occurrences {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return occurrences[keys[i]] > occurrences[keys[j]]
	})

	sortedKeyValues := make([]string, 0, len(keys))
	for _, k := range keys {
		sortedKeyValues = append(sortedKeyValues, k+sep+strconv.Itoa(occurrences[k]))
	}

	if count > len(sortedKeyValues) {
		return sortedKeyValues
	}
	return sortedKeyValues[:count]
}

// toLower converts the given string (usually by a pipe) to lowercase.
func (tpl *Template) ToLower(s string) (string, error) {
	return strings.ToLower(s), nil
}

// toTitle converts the given string (usually by a pipe) to titlecase.
func (tpl *Template) ToTitle(s string) (string, error) {
	return strings.ToTitle(s), nil
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

func (tpl *Template) IsEmpty(v interface{}) (bool, error) {
	return utils.IsEmpty(v), nil
}

func (tpl *Template) IsNotEmpty(v interface{}) (bool, error) {
	return !utils.IsEmpty(v), nil
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

	if obj == nil {
		err := errors.New("object is not defined")
		return "", err
	}

	if utils.IsEmpty(path) {
		err := errors.New("path is empty")
		return "", err
	}

	var value gjson.Result
	v, ok := obj.(string)
	if ok {

		if _, err := os.Stat(v); err == nil {
			bytes, err := os.ReadFile(v)
			if err != nil {
				return "", err
			}
			value = gjson.GetBytes(bytes, path)
		} else {
			value = gjson.Get(v, path)
		}
	} else {

		data, ok := obj.([]byte)
		if ok {
			value = gjson.GetBytes(data, path)
		} else {
			bytes, err := common.JsonMarshal(obj)
			if err != nil {
				return "", err
			}
			value = gjson.GetBytes(bytes, path)
		}
	}
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

func (tpl *Template) DurationBetween(start, end time.Time) map[string]int {
	duration := end.Sub(start)

	days := int(duration.Hours()) / 24
	hours := int(duration.Hours()) % 24
	minutes := int(duration.Minutes()) % 60

	return map[string]int{
		"Days":    days,
		"Hours":   hours,
		"Minutes": minutes,
	}
}

func (tpl *Template) NowFmt(f string) string {

	t := time.Now()
	s := t.Format(f)

	return s
}

func (tpl *Template) Sleep(ms int) string {
	time.Sleep(time.Duration(ms) * time.Millisecond)
	return ""
}

func (tpl *Template) Error(format string, a ...any) (string, error) {

	err := fmt.Errorf(format, a...)
	return err.Error(), err
}

// url, contentType, authorization string, timeout int
func (tpl *Template) HttpGetHeader(params map[string]interface{}) ([]byte, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("HttpGetHeader err => %s", "no params allowed")
	}

	url, ok := params["url"].(string)
	if !ok || url == "" {
		return nil, fmt.Errorf("HttpGetHeader err => %s", "invalid or missing URL")
	}
	timeout, ok := params["timeout"].(int)
	if !ok || timeout <= 0 {
		timeout = 5
	}

	insecure, _ := params["insecure"].(bool)

	var transport = &http.Transport{
		Dial:                (&net.Dialer{Timeout: time.Duration(timeout) * time.Second}).Dial,
		TLSHandshakeTimeout: time.Duration(timeout) * time.Second,
		TLSClientConfig:     &tls.Config{InsecureSkipVerify: insecure},
	}

	client := http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}

	// Call the GetHeaders function
	headers, err := utils.HttpGetHeader(&client, url)
	if err != nil {
		return nil, fmt.Errorf("HttpGetHeader err => %w", err)
	}

	headersBytes, err := json.Marshal(headers)
	if err != nil {
		return nil, fmt.Errorf("HttpGetHeader err => %w", err)
	}

	return headersBytes, nil
}

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
	repositoryId, _ := params["repositoryId"].(int)
	nameId, _ := params["nameId"].(int)
	descriptionId, _ := params["descriptionId"].(int)
	description, _ := params["description"].(string)
	name, _ := params["name"].(string)
	repository, _ := params["repository"].(string)
	titleId, _ := params["titleId"].(int)
	title, _ := params["title"].(string)
	tierId, _ := params["tierId"].(int)
	tier, _ := params["tier"].(string)

	jiraOptions := vendors.JiraOptions{
		URL:         url,
		Timeout:     timeout,
		Insecure:    insecure,
		User:        user,
		Password:    password,
		AccessToken: token,
	}
	jiraIssueOptions := vendors.JiraCreateAssetOptions{
		Name:           name,
		ObjectSchemeId: objectSchemeId,
		ObjectTypeId:   objectTypeId,
		RepositoryId:   repositoryId,
		NameId:         nameId,
		DescriptionId:  descriptionId,
		Description:    description,
		Repository:     repository,
		TitleId:        titleId,
		Title:          title,
		TierId:         tierId,
		Tier:           tier,
	}

	jira := vendors.NewJira(jiraOptions)

	response, err := jira.CreateAsset(jiraIssueOptions)
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

	jiraOptions := vendors.JiraOptions{
		URL:         url,
		Timeout:     timeout,
		Insecure:    insecure,
		User:        user,
		Password:    password,
		AccessToken: token,
	}

	jiraIssueOptions := vendors.JiraIssueOptions{
		TransitionID: transitionId,
		IdOrKey:      key,
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
		IdOrKey:      key,
		Summary:      summary,
		Description:  description,
		CustomFields: customFields,
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
		URL:               url,
		Timeout:           timeout,
		Insecure:          insecure,
		APIKey:            token,
		OrgID:             orgID,
		DashboardUID:      dUID,
		DashboardSlug:     dSlug,
		DashboardTimezone: dTimeZone,
	}

	grafanaCreateDashboardOptions := vendors.GrafanaCreateDahboardOptions{
		Title:     title,
		FolderUID: fUID,
		Tags:      tags,
		From:      from,
		To:        to,
		Cloned: vendors.GrafanaClonedDahboardOptions{
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
		URL:               url,
		Timeout:           timeout,
		Insecure:          insecure,
		APIKey:            token,
		OrgID:             orgID,
		DashboardUID:      dUID,
		DashboardSlug:     dSlug,
		DashboardTimezone: dTimeZone,
	}

	grafanaCreateDashboardOptions := vendors.GrafanaCreateDahboardOptions{
		Title:     title,
		FolderUID: fUID,
		Tags:      tags,
		From:      from,
		To:        to,
		Cloned: vendors.GrafanaClonedDahboardOptions{
			UID: clonedUID,
		},
	}

	grafana := vendors.NewGrafana(grafanaOptions)

	response, err := grafana.CopyDashboard(grafanaCreateDashboardOptions)
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

func (tpl *Template) VMRestart(params map[string]interface{}) ([]byte, error) {

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
			response, err = vcenter.RestartVM(vm.VM)
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

func (tpl *Template) setTemplateFuncs(funcs map[string]any) {

	funcs["parserLine"] = tpl.ParserLine

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
	funcs["findObject"] = tpl.FindObject
	funcs["findObjects"] = tpl.FindObjects
	funcs["findObjectByField"] = tpl.FindObject
	funcs["countOccurrences"] = tpl.CountOccurrences
	funcs["sortOccurrences"] = tpl.SortOccurrences

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
	funcs["isNotEmpty"] = tpl.IsNotEmpty
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
	funcs["durationBetween"] = tpl.DurationBetween
	funcs["nowFmt"] = tpl.NowFmt
	funcs["sleep"] = tpl.Sleep
	funcs["error"] = tpl.Error

	funcs["httpGetHeader"] = tpl.HttpGetHeader
	funcs["httpGet"] = tpl.HttpGet
	funcs["httpGetSilent"] = tpl.HttpGetSilent
	funcs["httpPost"] = tpl.HttpPost
	funcs["jiraSearchAssets"] = tpl.JiraSearchAssets
	funcs["jiraCreateIssue"] = tpl.JiraCreateIssue
	funcs["jiraSearchIssue"] = tpl.JiraSearchIssue
	funcs["jiraCreateAsset"] = tpl.JiraCreateAsset
	funcs["jiraAddComment"] = tpl.JiraAddComment
	funcs["jiraUpdateIssue"] = tpl.JiraUpdateIssue
	funcs["jiraIssueTransition"] = tpl.JiraIssueTransition
	funcs["jiraGetIssueTransition"] = tpl.JiraGetIssueTransition
	funcs["grafanaCreateDashboard"] = tpl.GrafanaCreateDashboard
	funcs["grafanaCopyDashboard"] = tpl.GrafanaCopyDashboard
	funcs["pagerDutyCreateIncident"] = tpl.PagerDutyCreateIncident
	funcs["pagerDutySendNoteToIncident"] = tpl.PagerDutySendNoteToIncident
	funcs["templateRender"] = tpl.TemplateRender
	funcs["templateRenderFile"] = tpl.TemplateRenderFile
	funcs["googleCalendarGetEvents"] = tpl.GoogleCalendarGetEvents
	funcs["googleCalendarInsertEvent"] = tpl.GoogleCalendarInsertEvent
	funcs["googleCalendarDeleteEvents"] = tpl.GoogleCalendarDeleteEvents
	funcs["sshRun"] = tpl.SSHRun
	funcs["listFilesWithModTime"] = tpl.ListFilesWithModTime
	funcs["vmRestart"] = tpl.VMRestart
	funcs["vmStart"] = tpl.VMStart
	funcs["vmStop"] = tpl.VMStop
	funcs["vmStatus"] = tpl.VMStatus

	funcs["prometheusGet"] = tpl.PrometheusGet
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

	tpl.tpl = t
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

	tpl.tpl = t
	tpl.template = t
	tpl.funcs = funcs
	tpl.options = options
	tpl.logger = logger
	return &tpl, nil
}

func (tpl *Template) HttpGetSilent(params map[string]interface{}) ([]byte, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("HttpGetSilent err => %s", "no params allowed")
	}

	url, _ := params["url"].(string)
	insecure, _ := params["insecure"].(bool)
	timeout, ok := params["timeout"].(int)
	if !ok || timeout <= 0 {
		timeout = 5
	}

	headers := map[string]string{}
	for key, value := range params {
		if key == "url" || key == "timeout" || key == "insecure" {
			continue
		}
		if strValue, ok := value.(string); ok {
			headers[key] = strValue
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

	body, code, err := utils.HttpRequestRawWithHeadersOutCodeSilent(&client, "GET", url, headers, nil)

	if err != nil {
		return nil, fmt.Errorf("HttpGetSilent err => HTTP status %d, error: %v", code, err)
	}

	return body, nil
}
