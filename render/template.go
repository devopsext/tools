package render

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	htmlTemplate "html/template"
	"io/ioutil"
	"net"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	txtTemplate "text/template"
	"time"

	"github.com/tidwall/gjson"

	"github.com/Masterminds/sprig/v3"
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	utils "github.com/devopsext/utils"
)

type TemplateOptions struct {
	Name       string
	Object     string
	Content    string
	Files      []string
	TimeFormat string
	Pattern    string
	Funcs      map[string]any
}

type Template struct {
	options TemplateOptions
	logger  common.Logger
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
func (tpl *Template) fLogError(obj interface{}, args ...interface{}) (string, error) {
	if tpl.logger == nil {
		return "", nil
	}
	tpl.logger.Error(obj, args...)
	return "", nil
}

// put warnings to logger
func (tpl *Template) fLogWarn(obj interface{}, args ...interface{}) (string, error) {
	if tpl.logger == nil {
		return "", nil
	}
	tpl.logger.Warn(obj, args...)
	return "", nil
}

// put warnings to logger
func (tpl *Template) fLogDebug(obj interface{}, args ...interface{}) (string, error) {
	if tpl.logger == nil {
		return "", nil
	}
	tpl.logger.Debug(obj, args...)
	return "", nil
}

// put information to logger
func (tpl *Template) fLogInfo(obj interface{}, args ...interface{}) (string, error) {
	if tpl.logger == nil {
		return "", nil
	}
	tpl.logger.Info(obj, args...)
	return "", nil
}

// replaceAll replaces all occurrences of a value in a string with the given
// replacement value.
func (tpl *Template) fReplaceAll(f, t, s string) (string, error) {
	return strings.Replace(s, f, t, -1), nil
}

// regexReplaceAll replaces all occurrences of a regular expression with
// the given replacement value.
func (tpl *Template) fRegexReplaceAll(re, pl, s string) (string, error) {
	compiled, err := regexp.Compile(re)
	if err != nil {
		return "", err
	}
	return compiled.ReplaceAllString(s, pl), nil
}

// regexMatch returns true or false if the string matches
// the given regular expression
func (tpl *Template) fRegexMatch(re, s string) (bool, error) {
	compiled, err := regexp.Compile(re)
	if err != nil {
		return false, err
	}
	return compiled.MatchString(s), nil
}

func (tpl *Template) fRegexFindSubmatch(regex string, s string) []string {
	r := regexp.MustCompile(regex)
	return r.FindStringSubmatch(s)
}

func (tpl *Template) RegexMatchObjectNamesByField(obj map[string]interface{}, field, value string) []interface{} {

	var r []interface{}
	if obj == nil || utils.IsEmpty(field) {
		return r
	}

	for k, v := range obj {

		m, ok := v.(map[string]interface{})
		if !ok {
			continue
		}
		if m[field] == nil {
			continue
		}
		s, ok := m[field].(string)
		if !ok {
			continue
		}
		match, _ := regexp.MatchString(fmt.Sprintf("^%s", s), value)
		if match {
			r = append(r, k)
		}
	}
	return r
}

func (tpl *Template) fRegexMatchObjectNameByField(obj map[string]interface{}, field, value string) interface{} {

	keys := tpl.RegexMatchObjectNamesByField(obj, field, value)
	if len(keys) == 0 {
		return value
	}
	return keys[0]
}

func (tpl *Template) fRegexMatchObjectByField(obj map[string]interface{}, field, value string) interface{} {

	key := tpl.fRegexMatchObjectNameByField(obj, field, value)
	if obj == nil {
		return nil
	}
	s, ok := key.(string)
	if !ok {
		return nil
	}
	return obj[s]
}

// toLower converts the given string (usually by a pipe) to lowercase.
func (tpl *Template) fToLower(s string) (string, error) {
	return strings.ToLower(s), nil
}

// toTitle converts the given string (usually by a pipe) to titlecase.
func (tpl *Template) fToTitle(s string) (string, error) {
	return strings.Title(s), nil
}

// toUpper converts the given string (usually by a pipe) to uppercase.
func (tpl *Template) fToUpper(s string) (string, error) {
	return strings.ToUpper(s), nil
}

// toJSON converts the given structure into a deeply nested JSON string.
func (tpl *Template) fToJSON(i interface{}) (string, error) {
	result, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(result)), err
}

// split is a version of strings.Split that can be piped
func (tpl *Template) fSplit(sep, s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}, nil
	}
	return strings.Split(s, sep), nil
}

// join is a version of strings.Join that can be piped
func (tpl *Template) fJoin(sep string, a []string) (string, error) {
	return strings.Join(a, sep), nil
}

func (tpl *Template) fIsEmpty(s string) (bool, error) {
	s1 := strings.TrimSpace(s)
	return len(s1) == 0, nil
}

func (tpl *Template) fEnv(key string) (string, error) {
	return utils.EnvGet(key, "").(string), nil
}

func (tpl *Template) fTimeFormat(s string, format string) (string, error) {

	t, err := time.Parse(tpl.options.TimeFormat, s)
	if err != nil {

		return s, err
	}
	return t.Format(format), nil
}

func (tpl *Template) fTimeNano(s string) (string, error) {

	t1, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(t1.UnixNano(), 10), nil
}

func (tpl *Template) fJsonEscape(s string) (string, error) {

	bytes, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	return string(bytes), nil
}

// toString converts the given value to string
func (tpl *Template) fToString(i interface{}) (string, error) {

	if i != nil {
		return fmt.Sprintf("%v", i), nil
	}
	return "", nil
}

func (tpl *Template) fEscapeString(s string) (string, error) {
	return html.EscapeString(s), nil
}

func (tpl *Template) fUnescapeString(s string) (string, error) {
	return html.UnescapeString(s), nil
}

func (tpl *Template) fJsonata(data interface{}, query string) (string, error) {

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

func (tpl *Template) fGjson(obj interface{}, path string) (string, error) {

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
			bytes, err := ioutil.ReadFile(v)
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

func (tpl *Template) fIfDef(i interface{}, def string) (string, error) {

	if utils.IsEmpty(i) {
		return def, nil
	}
	return tpl.fToString(i)
}

func (tpl *Template) fIfElse(o interface{}, vars []interface{}) interface{} {

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

func (tpl *Template) fIfIP(s string) bool {

	a := net.ParseIP(s)
	return a != nil
}

func (tpl *Template) fIfIPAndPort(s string) bool {

	arr := strings.Split(s, ":")
	if len(arr) > 0 {
		s = strings.TrimSpace(arr[0])
	} else {
		return false
	}
	return tpl.fIfIP(s)
}

func (tpl *Template) fContent(s string) (string, error) {

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

func (tpl *Template) fURLWait(url string, timeout, retry int, size int64) []byte {

	if utils.IsEmpty(url) {
		return nil
	}

	if retry <= 0 {
		retry = 1
	}

	tpl.fLogInfo("fURLWait url => %s [%d, %d, %d]", url, timeout, retry, size)

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

		tpl.fLogInfo("fURLWait(%d) get %s...", i, url)

		data, err := common.HttpGetRaw(&client, url, "", "")
		if err != nil {
			tpl.fLogInfo("fURLWait(%d) get %s err => %s", i, url, err.Error())
			tpl.tryToWaitUntil(t1, client.Timeout)
			continue
		}

		l := int64(len(data))
		tpl.fLogInfo("fURLWait(%d) %s len(data) = %d", i, url, l)

		if l < size {
			tpl.tryToWaitUntil(t1, client.Timeout)
			continue
		} else if l >= size {
			return data
		}
	}
	return nil
}

func (tpl *Template) fGitlabPipelineVars(URL string, token string, projectID int, query string, limit int) string {

	gitlabOptions := vendors.GitlabOptions{
		Timeout:  30,
		Insecure: false,
		URL:      URL,
		Token:    token,
	}

	gitlab, err := vendors.NewGitlab(gitlabOptions)
	if err != nil {
		tpl.fLogInfo("fGitlabPipelineVars err => %s", err.Error())
		return ""
	}

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

	pipelineGetVariablesOptions := vendors.GitlabPipelineGetVariablesOptions{
		Query: strings.Split(query, ","),
	}

	b, err := gitlab.PipelineGetVariables(pipelineOptions, pipelineGetVariablesOptions)
	if err != nil {
		tpl.fLogInfo("fGitlabPipelineVars err => %s", err.Error())
		return ""
	}
	return string(b)
}

func (tpl *Template) fTagExists(s, key string) (bool, error) {

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

func (tpl *Template) fTagValue(s, key string) (string, error) {

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

func (tpl *Template) setTemplateFuncs(funcs map[string]any) {

	funcs["logError"] = tpl.fLogError
	funcs["logWarn"] = tpl.fLogWarn
	funcs["logDebug"] = tpl.fLogDebug
	funcs["logInfo"] = tpl.fLogInfo

	funcs["regexReplaceAll"] = tpl.fRegexReplaceAll
	funcs["regexMatch"] = tpl.fRegexMatch
	funcs["regexFindSubmatch"] = tpl.fRegexFindSubmatch
	funcs["regexMatchObjectNamesByField"] = tpl.RegexMatchObjectNamesByField
	funcs["regexMatchObjectNameByField"] = tpl.fRegexMatchObjectNameByField
	funcs["regexMatchObjectByField"] = tpl.fRegexMatchObjectByField

	funcs["replaceAll"] = tpl.fReplaceAll
	funcs["toLower"] = tpl.fToLower
	funcs["toTitle"] = tpl.fToTitle
	funcs["toUpper"] = tpl.fToUpper
	funcs["toJSON"] = tpl.fToJSON
	funcs["split"] = tpl.fSplit
	funcs["join"] = tpl.fJoin
	funcs["isEmpty"] = tpl.fIsEmpty
	funcs["env"] = tpl.fEnv
	funcs["getEnv"] = tpl.fEnv
	funcs["timeFormat"] = tpl.fTimeFormat
	funcs["timeNano"] = tpl.fTimeNano
	funcs["jsonEscape"] = tpl.fJsonEscape
	funcs["toString"] = tpl.fToString
	funcs["escapeString"] = tpl.fEscapeString
	funcs["unescapeString"] = tpl.fUnescapeString
	funcs["jsonata"] = tpl.fJsonata
	funcs["gjson"] = tpl.fGjson
	funcs["ifDef"] = tpl.fIfDef
	funcs["ifElse"] = tpl.fIfElse
	funcs["ifIP"] = tpl.fIfIP
	funcs["ifIPAndPort"] = tpl.fIfIPAndPort
	funcs["content"] = tpl.fContent
	funcs["urlWait"] = tpl.fURLWait
	funcs["gitlabPipelineVars"] = tpl.fGitlabPipelineVars
	funcs["tagExists"] = tpl.fTagExists
	funcs["tagValue"] = tpl.fTagValue
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

	if empty, _ := tpl.fIsEmpty(name); empty {
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

	if utils.IsEmpty(options.Files) && utils.IsEmpty(options.Pattern) {
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
	tpl.options = options
	tpl.logger = logger
	return &tpl, nil
}

func (tpl *HtmlTemplate) customRender(name string, obj interface{}) ([]byte, error) {

	var b bytes.Buffer
	var err error

	if empty, _ := tpl.fIsEmpty(name); empty {
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

	if utils.IsEmpty(options.Files) && utils.IsEmpty(options.Pattern) {
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
	tpl.options = options
	tpl.logger = logger
	return &tpl, nil
}
