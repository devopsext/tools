// Package lite provides a dependency-light template funcmap that can be used
// on its own (for example by external consumers that only need a small set
// of Sprig-style helper functions) or embedded by a richer template
// implementation, such as github.com/devopsext/tools/render.
//
// Unlike render.Template, Template in this package never imports
// github.com/devopsext/tools/vendors or any of the cloud/identity SDKs
// (AWS, Kubernetes, Teleport, LDAP, gRPC) that vendors pulls in, so
// importing this package keeps a consumer's dependency graph small.
package lite

import (
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"net"
	"net/http"
	"net/url"
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
	utils "github.com/devopsext/utils"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

// Default timeout and synthetic HTTP status codes used by the http* template
// functions when a real status code cannot be determined (for example when
// the request never reaches the server).
const (
	DefaultTimeout = 5 * time.Second
	ErrorCodeParam = -1
	ErrorCodeTLS   = -2
	ErrorCodeHTTP  = -3
)

// TemplateOptions configures a Template. It mirrors the subset of
// render.TemplateOptions that the lite function set needs.
type TemplateOptions struct {
	// Name is the name of the template to execute. When empty, the root
	// template is executed.
	Name string
	// Content is the template source text.
	Content string
	// TimeFormat is the layout used to parse strings passed to timeFormat.
	TimeFormat string
	// Funcs overrides or extends the default funcmap.
	Funcs map[string]any
	// FilterFuncs, when true, restricts the funcmap to only the functions
	// whose name appears as a substring of Content.
	FilterFuncs bool
}

// Template holds the state shared by the lite template functions: the
// options it was configured with, the logger used by the log* functions,
// the funcmap that was built for it, and the parsed template node used for
// diagnostics.
//
// Template is safe to embed by higher-level template implementations. Since
// its fields are unexported, embedders must use Configure, SetNode, and
// SetFuncs to initialize its state.
type Template struct {
	options TemplateOptions
	logger  common.Logger
	funcs   template.FuncMap
	tpl     interface{}
}

// TextTemplate is a minimal, dependency-light text/template wrapper built on
// top of Template's funcmap.
type TextTemplate struct {
	Template
	template *txtTemplate.Template
}

// HTTPResult is the result of an http* template function call.
type HTTPResult struct {
	Body       []byte
	Error      string
	StatusCode int
}

// ReadFileResult is the result of a readFile template function call.
type ReadFileResult struct {
	Content []byte
	Error   string
}

// Configure sets the options and logger used by Template's functions (for
// example timeFormat, env, and templateRender). It is exported so that
// higher-level template implementations that embed Template, such as
// render.Template, can keep the embedded Template's state in sync with
// their own, richer option set.
func (tpl *Template) Configure(options TemplateOptions, logger common.Logger) {
	tpl.options = options
	tpl.logger = logger
}

// SetNode stores the parsed template node used by ParserLine for
// diagnostics. It is exported for use by higher-level template
// implementations that parse their own text/template or html/template tree
// on top of an embedded Template.
func (tpl *Template) SetNode(node interface{}) {
	tpl.tpl = node
}

// FuncMap returns the map of lite template functions provided by Template,
// merged over sprig's text funcmap. Consumers that only need the funcmap
// (for example, to build their own text/template.Template) can call this
// without constructing a full TextTemplate.
func (tpl *Template) FuncMap() template.FuncMap {
	funcs := sprig.TxtFuncMap()
	tpl.SetFuncs(funcs)
	return funcs
}

// SetFuncs registers Template's lite function set into funcs. Higher-level
// template implementations, such as render.Template, call this to obtain
// the lite funcmap before layering additional functions on top.
func (tpl *Template) SetFuncs(funcs map[string]any) {

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
	funcs["tryFromJson"] = tpl.TryFromJson
	funcs["toYaml"] = tpl.ToYaml
	funcs["toYml"] = tpl.ToYaml
	funcs["fromYaml"] = tpl.FromYaml
	funcs["fromYml"] = tpl.FromYaml
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

	funcs["error"] = tpl.Error
	funcs["ifError"] = tpl.IfError

	funcs["content"] = tpl.Content
	funcs["tagExists"] = tpl.TagExists
	funcs["tagValue"] = tpl.TagValue
	funcs["dateParse"] = tpl.DateParse
	funcs["durationBetween"] = tpl.DurationBetween
	funcs["duration"] = tpl.Duration
	funcs["durationString"] = tpl.DurationString
	funcs["nowFmt"] = tpl.NowFmt
	funcs["uuid"] = tpl.UUID

	funcs["stringList"] = tpl.StringList
	funcs["strings"] = tpl.StringList
	funcs["appendString"] = tpl.AppendString

	funcs["intList"] = tpl.IntList
	funcs["floatList"] = tpl.FloatList

	funcs["httpGetHeader"] = tpl.HttpGetHeader
	funcs["httpGet"] = tpl.HttpGet
	funcs["httpGetExt"] = tpl.HttpGetExt
	funcs["httpGetSilent"] = tpl.HttpGetSilent
	funcs["httpPost"] = tpl.HttpPost
	funcs["httpPostExt"] = tpl.HttpPostExt
	funcs["tryHttpPost"] = tpl.TryHttpPost
	funcs["httpPut"] = tpl.HttpPut
	funcs["httpPatch"] = tpl.HttpPatch
	funcs["httpForm"] = tpl.HttpForm

	funcs["readFile"] = tpl.ReadFile

	funcs["templateRender"] = tpl.TemplateRender
	funcs["templateRenderFile"] = tpl.TemplateRenderFile

	funcs["dirCreate"] = tpl.DirCreate
	funcs["fileCreate"] = tpl.FileCreate

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

// ParserLine returns the position, in the parsed template source, of the
// current template's root node. It is mainly useful for diagnostics; it
// requires SetNode to have been called with a *text/template.Template or
// html/template.Template value.
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

	return line, nil
}

// LogError writes obj and args to the configured logger's Error method. It
// is a no-op, returning no error, when no logger is configured.
func (tpl *Template) LogError(obj interface{}, args ...interface{}) (string, error) {
	if tpl.logger == nil {
		return "", nil
	}
	tpl.logger.Error(obj, args...)
	return "", nil
}

// LogWarn writes obj and args to the configured logger's Warn method. It is
// a no-op, returning no error, when no logger is configured.
func (tpl *Template) LogWarn(obj interface{}, args ...interface{}) (string, error) {
	if tpl.logger == nil {
		return "", nil
	}
	tpl.logger.Warn(obj, args...)
	return "", nil
}

// LogDebug writes obj and args to the configured logger's Debug method. It
// is a no-op, returning no error, when no logger is configured.
func (tpl *Template) LogDebug(obj interface{}, args ...interface{}) (string, error) {
	if tpl.logger == nil {
		return "", nil
	}
	tpl.logger.Debug(obj, args...)
	return "", nil
}

// LogInfo writes obj and args to the configured logger's Info method. It is
// a no-op, returning no error, when no logger is configured.
func (tpl *Template) LogInfo(obj interface{}, args ...interface{}) (string, error) {
	if tpl.logger == nil {
		return "", nil
	}
	tpl.logger.Info(obj, args...)
	return "", nil
}

// ReplaceAll replaces all occurrences of f in s with t.
func (tpl *Template) ReplaceAll(f, t, s string) (string, error) {
	return strings.Replace(s, f, t, -1), nil
}

// RegexReplaceAll replaces all substrings of s matching the regular
// expression re with pl.
func (tpl *Template) RegexReplaceAll(re, pl, s string) (string, error) {
	compiled, err := regexp.Compile(re)
	if err != nil {
		return "", err
	}
	return compiled.ReplaceAllString(s, pl), nil
}

// RegexMatch reports whether s matches the regular expression re.
func (tpl *Template) RegexMatch(re, s string) (bool, error) {
	compiled, err := regexp.Compile(re)
	if err != nil {
		return false, err
	}
	return compiled.MatchString(s), nil
}

// RegexFindSubmatch returns the leftmost match of regex in s, along with its
// submatches.
func (tpl *Template) RegexFindSubmatch(regex string, s string) []string {
	r := regexp.MustCompile(regex)
	return r.FindStringSubmatch(s)
}

// regexMatchFindKey reports whether v is a map whose field entry matches
// value as a regular expression anchored at the start of value.
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

// RegexMatchFindKeys returns the keys (or indices) of obj whose field entry
// matches value as a regular expression.
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

// RegexMatchFindKey returns the first key found by RegexMatchFindKeys, or
// value if no key matched.
func (tpl *Template) RegexMatchFindKey(obj interface{}, field, value string) interface{} {

	keys := tpl.RegexMatchFindKeys(obj, field, value)
	if len(keys) == 0 {
		return value
	}
	return keys[0]
}

// RegexMatchObjectByField returns the element of obj (a slice or a map)
// whose field entry matches value as a regular expression, or nil when no
// element matches.
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

// Compare reports whether v1 and v2 are equal, applying loose numeric and
// slice-membership comparisons where useful for template data.
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

// findKey reports whether v is a map whose field entry compares equal to
// value.
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

// FindKeys returns the keys (or indices) of obj whose field entry compares
// equal to value.
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

// FindKey returns the first key found by FindKeys, or value if no key
// matched.
func (tpl *Template) FindKey(obj interface{}, field string, value interface{}) interface{} {

	keys := tpl.FindKeys(obj, field, value)
	if len(keys) == 0 {
		return value
	}
	return keys[0]
}

// FindObject returns the element of obj (a slice or a map) whose field
// entry compares equal to value, or nil when no element matches.
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

// FindObjects returns every element of obj (a slice or a map) whose field
// entry compares equal to value.
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

// CountOccurrences counts how many times each string element of list
// occurs.
func (tpl *Template) CountOccurrences(list []interface{}) map[string]int {
	occurrences := make(map[string]int)
	for _, item := range list {
		occurrences[item.(string)]++
	}

	return occurrences
}

// SortOccurrences sorts occurrences by descending count and returns, at
// most, count "key<sep>count" strings.
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

// ToLower converts s (usually piped in) to lowercase.
func (tpl *Template) ToLower(s string) (string, error) {
	return strings.ToLower(s), nil
}

// ToTitle converts s (usually piped in) to titlecase.
func (tpl *Template) ToTitle(s string) (string, error) {
	return strings.ToTitle(s), nil
}

// ToUpper converts s (usually piped in) to uppercase.
func (tpl *Template) ToUpper(s string) (string, error) {
	return strings.ToUpper(s), nil
}

// ToJson converts i into a deeply nested JSON string.
func (tpl *Template) ToJson(i interface{}) (string, error) {

	result, err := json.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(result)), err
}

// FromJson parses i (a []byte or something convertible to a string) as
// JSON.
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

// TryFromJson is like FromJson but returns nil instead of an error.
func (tpl *Template) TryFromJson(i interface{}) interface{} {

	r, err := tpl.FromJson(i)
	if err != nil {
		return nil
	}
	return r
}

// ToYaml converts i into a deeply nested YAML string.
func (tpl *Template) ToYaml(i interface{}) (string, error) {

	result, err := yaml.Marshal(i)
	if err != nil {
		return "", err
	}
	return string(bytes.TrimSpace(result)), err
}

// FromYaml parses i (a []byte or something convertible to a string) as
// YAML.
func (tpl *Template) FromYaml(i interface{}) (interface{}, error) {

	var d []byte
	var r interface{}
	ds, ok := i.([]byte)
	if ok {
		d = ds
	} else {
		d = []byte(fmt.Sprintf("%v", i))
	}
	err := yaml.Unmarshal(d, &r)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// Split is a version of strings.Split that can be piped.
func (tpl *Template) Split(sep, s string) ([]string, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return []string{}, nil
	}
	return strings.Split(s, sep), nil
}

// Join is a version of strings.Join that can be piped.
func (tpl *Template) Join(sep string, a []string) (string, error) {
	return strings.Join(a, sep), nil
}

// IsEmpty reports whether v is considered empty by utils.IsEmpty.
func (tpl *Template) IsEmpty(v interface{}) (bool, error) {
	return utils.IsEmpty(v), nil
}

// IsNotEmpty is the negation of IsEmpty.
func (tpl *Template) IsNotEmpty(v interface{}) (bool, error) {
	return !utils.IsEmpty(v), nil
}

// Env returns the value of the environment variable key, or "" if unset.
func (tpl *Template) Env(key string) (string, error) {
	return utils.EnvGet(key, "").(string), nil
}

// TimeFormat parses s using the options' TimeFormat layout and reformats it
// using format.
func (tpl *Template) TimeFormat(s string, format string) (string, error) {

	t, err := time.Parse(tpl.options.TimeFormat, s)
	if err != nil {
		return s, err
	}
	return t.Format(format), nil
}

// TimeNano parses s as an RFC3339 timestamp with nanoseconds and returns its
// Unix nanosecond count as a string.
func (tpl *Template) TimeNano(s string) (string, error) {

	t1, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(t1.UnixNano(), 10), nil
}

// JsonEscape returns s JSON-encoded as a quoted string.
func (tpl *Template) JsonEscape(s string) (string, error) {

	b, err := json.Marshal(s)
	if err != nil {
		return "", err
	}

	return string(b), nil
}

// ToString converts i to its string representation.
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

// EscapeString escapes special HTML characters in s.
func (tpl *Template) EscapeString(s string) (string, error) {
	return html.EscapeString(s), nil
}

// UnescapeString unescapes HTML entities in s.
func (tpl *Template) UnescapeString(s string) (string, error) {
	return html.UnescapeString(s), nil
}

// Jsonata evaluates the JSONata expression query against data. query and
// data may each be a filesystem path to load the expression or JSON data
// from.
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

// Gjson evaluates the gjson path expression path against obj, which may be a
// JSON string, a filesystem path to a JSON file, a []byte of JSON, or any
// value marshalable to JSON.
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
			b, err := os.ReadFile(v)
			if err != nil {
				return "", err
			}
			value = gjson.GetBytes(b, path)
		} else {
			value = gjson.Get(v, path)
		}
	} else {

		data, ok := obj.([]byte)
		if ok {
			value = gjson.GetBytes(data, path)
		} else {
			b, err := common.JsonMarshal(obj)
			if err != nil {
				return "", err
			}
			value = gjson.GetBytes(b, path)
		}
	}
	return value.String(), nil
}

// IfDef returns i converted to a string, or def when i is empty.
func (tpl *Template) IfDef(i interface{}, def string) (string, error) {

	if utils.IsEmpty(i) {
		return def, nil
	}
	return tpl.ToString(i)
}

// IfElse walks vars as (match, result) pairs and returns the result paired
// with the first match equal to o, or o itself when nothing matches.
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

// IfIP reports whether obj is a valid IP address.
func (tpl *Template) IfIP(obj interface{}) bool {

	if obj == nil {
		return false
	}

	a := net.ParseIP(fmt.Sprintf("%v", obj))
	return a != nil
}

// IfIPAndPort reports whether obj is a "host:port" string whose host part
// is a valid IP address.
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

// Error formats format and a as an error, following fmt.Errorf semantics.
func (tpl *Template) Error(format string, a ...any) error {

	err := fmt.Errorf(format, a...)
	return err
}

// IfError returns an error built from format and a when flag is not empty,
// and no error otherwise.
func (tpl *Template) IfError(flag interface{}, format string, a ...any) (string, error) {

	if !utils.IsEmpty(flag) {
		return "", tpl.Error(format, a...)
	}
	return "", nil
}

// Content returns the content referenced by s (a literal string, a
// filesystem path, or a URL, per utils.Content).
func (tpl *Template) Content(s string) (string, error) {

	if utils.IsEmpty(s) {
		return "", nil
	}

	b, err := utils.Content(s)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// TagExists reports whether key is present among the DataDog-style,
// comma-separated "key:value" tags in s.
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

// TagValue returns the value of key among the DataDog-style,
// comma-separated "key:value" tags in s, or s itself when key is not found.
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

// DateParse parses d using a best-effort format detection.
func (tpl *Template) DateParse(d string) (time.Time, error) {
	t, err := dateparse.ParseAny(d)
	if err != nil {
		return time.Now(), err
	}
	return t, nil
}

// DurationBetween returns the whole days, remaining hours, and remaining
// minutes between start and end.
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

// Duration returns the elapsed time from start to now, formatted like
// "1.234s" or "500ms".
func (tpl *Template) Duration(start time.Time) string {
	duration := time.Since(start)
	return duration.String()
}

// DurationString returns the duration from start to end, formatted like
// "1.234s" or "500ms".
func (tpl *Template) DurationString(start, end time.Time) string {
	duration := end.Sub(start)
	return duration.String()
}

// NowFmt returns the current time formatted using the layout f.
func (tpl *Template) NowFmt(f string) string {

	t := time.Now()
	s := t.Format(f)

	return s
}

// UUID returns a new random (version 4) UUID string.
func (tpl *Template) UUID() string {

	id := uuid.New()
	return id.String()
}

// StringList returns items as a []string, allowing variadic string
// arguments to be collected into a list.
func (tpl *Template) StringList(items ...string) []string {
	return items
}

// AppendString appends items to arr and returns the result.
func (tpl *Template) AppendString(arr []string, items ...string) []string {
	return append(arr, items...)
}

// IntList returns items as a []int64, allowing variadic int64 arguments to
// be collected into a list.
func (tpl *Template) IntList(items ...int64) []int64 {
	return items
}

// FloatList returns items as a []float64, allowing variadic float64
// arguments to be collected into a list.
func (tpl *Template) FloatList(items ...float64) []float64 {
	return items
}

func newHTTPTransport(timeout int, insecure bool, clientCrt, clientKey, clientCA string) (*http.Transport, error) {

	var certs []tls.Certificate
	if !utils.IsEmpty(clientCrt) && !utils.IsEmpty(clientKey) {
		pair, err := tls.X509KeyPair([]byte(clientCrt), []byte(clientKey))
		if err != nil {
			return nil, err
		}
		certs = append(certs, pair)
	}

	var rootCAs *x509.CertPool
	if !utils.IsEmpty(clientCA) {
		rootCAs = x509.NewCertPool()
		rootCAs.AppendCertsFromPEM([]byte(clientCA))
	}

	return &http.Transport{
		Dial:                (&net.Dialer{Timeout: time.Duration(timeout) * time.Second}).Dial,
		DialContext:         (&net.Dialer{Timeout: time.Duration(timeout) * time.Second}).DialContext,
		TLSHandshakeTimeout: time.Duration(timeout) * time.Second,
		TLSClientConfig: &tls.Config{
			RootCAs:            rootCAs,
			Certificates:       certs,
			InsecureSkipVerify: insecure,
		},
	}, nil
}

// HttpGetHeader performs an HTTP GET against params["url"] and returns the
// response headers as a JSON-encoded []byte. Recognized params: url
// (required), timeout, insecure, clientCrt, clientKey, clientCA.
func (tpl *Template) HttpGetHeader(params map[string]interface{}) ([]byte, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("HttpGetHeader err => %s", "no params allowed")
	}

	u, ok := params["url"].(string)
	if !ok || u == "" {
		return nil, fmt.Errorf("HttpGetHeader err => %s", "invalid or missing URL")
	}
	timeout, ok := params["timeout"].(int)
	if !ok || timeout <= 0 {
		timeout = 5
	}

	insecure, _ := params["insecure"].(bool)
	clientCrt, _ := params["clientCrt"].(string)
	clientKey, _ := params["clientKey"].(string)
	clientCA, _ := params["clientCA"].(string)

	transport, err := newHTTPTransport(timeout, insecure, clientCrt, clientKey, clientCA)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}

	headers, err := utils.HttpGetHeader(&client, u)
	if err != nil {
		return nil, fmt.Errorf("HttpGetHeader err => %w", err)
	}

	headersBytes, err := json.Marshal(headers)
	if err != nil {
		return nil, fmt.Errorf("HttpGetHeader err => %w", err)
	}

	return headersBytes, nil
}

// HttpGet performs an HTTP GET and returns the response body. It is a
// convenience wrapper over HttpGetExt that turns a non-empty
// HTTPResult.Error into a Go error.
func (tpl *Template) HttpGet(params map[string]interface{}) ([]byte, error) {
	result := tpl.HttpGetExt(params)
	if result.Error != "" {
		return nil, fmt.Errorf("%s", result.Error)
	}
	return result.Body, nil
}

// HttpGetSilent performs an HTTP GET, treating every param other than url,
// timeout, insecure, clientCrt, clientKey, and clientCA as an extra request
// header.
func (tpl *Template) HttpGetSilent(params map[string]interface{}) ([]byte, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("HttpGetSilent err => %s", "no params allowed")
	}

	u, _ := params["url"].(string)
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

	clientCrt, _ := params["clientCrt"].(string)
	clientKey, _ := params["clientKey"].(string)
	clientCA, _ := params["clientCA"].(string)

	transport, err := newHTTPTransport(timeout, insecure, clientCrt, clientKey, clientCA)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}

	body, code, err := utils.HttpRequestRawWithHeadersOutCodeSilent(&client, "GET", u, headers, nil)
	if err != nil {
		return nil, fmt.Errorf("HttpGetSilent err => HTTP status %d, error: %v", code, err)
	}

	return body, nil
}

// HttpGetExt performs an HTTP GET and returns the full HTTPResult, including
// its status code and any error, instead of returning a Go error.
// Recognized params: url (required), timeout, insecure, contentType,
// authorization, clientCrt, clientKey, clientCA.
func (tpl *Template) HttpGetExt(params map[string]interface{}) HTTPResult {

	result := HTTPResult{}

	if len(params) == 0 {
		result.Error = "no parameters provided"
		result.StatusCode = ErrorCodeParam
		return result
	}

	u, _ := params["url"].(string)
	if u == "" {
		result.Error = "URL parameter is required"
		result.StatusCode = ErrorCodeParam
		return result
	}

	var timeout int
	switch tp := params["timeout"].(type) {
	case int:
		timeout = tp
	case float64:
		timeout = int(tp)
	default:
		timeout = 5
	}

	insecure, _ := params["insecure"].(bool)
	contentType, _ := params["contentType"].(string)
	authorization, _ := params["authorization"].(string)
	clientCrt, _ := params["clientCrt"].(string)
	clientKey, _ := params["clientKey"].(string)
	clientCA, _ := params["clientCA"].(string)

	transport, err := newHTTPTransport(timeout, insecure, clientCrt, clientKey, clientCA)
	if err != nil {
		result.Error = fmt.Errorf("failed to load client key pair: %w", err).Error()
		result.StatusCode = ErrorCodeTLS
		return result
	}

	client := http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}

	start := time.Now()
	defer func() {
		if tpl.logger != nil {
			tpl.logger.Debug("HTTP request completed",
				"url", u,
				"duration", time.Since(start),
				"status", result.StatusCode)
		}
	}()

	body, err := utils.HttpGetRaw(&client, u, contentType, authorization)
	if err != nil {
		result.Error = fmt.Errorf("HTTP request failed: %w", err).Error()
		result.StatusCode = ErrorCodeHTTP

		if respErr, ok := err.(interface{ StatusCode() int }); ok {
			result.StatusCode = respErr.StatusCode()
		} else {
			e := err.Error()
			if strings.Contains(e, "404") {
				result.StatusCode = http.StatusNotFound
			} else if strings.Contains(e, "403") {
				result.StatusCode = http.StatusForbidden
			} else if strings.Contains(e, "500") {
				result.StatusCode = http.StatusInternalServerError
			}
		}
		result.Body = body
		return result
	}

	result.Body = body
	result.StatusCode = http.StatusOK

	return result
}

// HttpPost performs an HTTP POST and returns the response body. It is a
// convenience wrapper over HttpPostExt that turns a non-empty
// HTTPResult.Error into a Go error.
func (tpl *Template) HttpPost(params map[string]interface{}) ([]byte, error) {
	result := tpl.HttpPostExt(params)
	if result.Error != "" {
		return nil, fmt.Errorf("%s", result.Error)
	}
	return result.Body, nil
}

// TryHttpPost is like HttpPost but returns nil instead of an error.
func (tpl *Template) TryHttpPost(params map[string]interface{}) []byte {

	r, err := tpl.HttpPost(params)
	if err != nil {
		return nil
	}
	return r
}

func bodyFromParam(b interface{}) []byte {
	if utils.IsEmpty(b) {
		return nil
	}
	switch v := b.(type) {
	case string:
		return []byte(v)
	case []byte:
		return v
	default:
		return []byte(fmt.Sprintf("%s", v))
	}
}

// HttpPostExt performs an HTTP POST and returns the full HTTPResult,
// including its status code and any error, instead of returning a Go
// error. Recognized params: url (required), body, timeout, insecure,
// contentType, authorization, clientCrt, clientKey, clientCA.
func (tpl *Template) HttpPostExt(params map[string]interface{}) HTTPResult {

	result := HTTPResult{}

	if len(params) == 0 {
		result.Error = "no params allowed"
		result.StatusCode = ErrorCodeParam
		return result
	}

	u, _ := params["url"].(string)
	if u == "" {
		result.Error = "URL parameter is required"
		result.StatusCode = ErrorCodeParam
		return result
	}

	var timeout int
	switch tp := params["timeout"].(type) {
	case int:
		timeout = tp
	case float64:
		timeout = int(tp)
	default:
		timeout = 5
	}

	insecure, _ := params["insecure"].(bool)
	contentType, _ := params["contentType"].(string)
	authorization, _ := params["authorization"].(string)
	body := bodyFromParam(params["body"])
	clientCrt, _ := params["clientCrt"].(string)
	clientKey, _ := params["clientKey"].(string)
	clientCA, _ := params["clientCA"].(string)

	transport, err := newHTTPTransport(timeout, insecure, clientCrt, clientKey, clientCA)
	if err != nil {
		result.Error = fmt.Errorf("failed to load client key pair: %w", err).Error()
		result.StatusCode = ErrorCodeTLS
		return result
	}

	client := http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}

	start := time.Now()
	defer func() {
		if tpl.logger != nil {
			tpl.logger.Debug("HTTP POST request completed",
				"url", u,
				"duration", time.Since(start),
				"status", result.StatusCode)
		}
	}()

	respBody, err := utils.HttpPostRaw(&client, u, contentType, authorization, body)
	if err != nil {
		result.Error = fmt.Errorf("HTTP request failed: %w", err).Error()
		result.StatusCode = ErrorCodeHTTP
		if respErr, ok := err.(interface{ StatusCode() int }); ok {
			result.StatusCode = respErr.StatusCode()
		} else {
			e := err.Error()
			if strings.Contains(e, "404") {
				result.StatusCode = http.StatusNotFound
			} else if strings.Contains(e, "403") {
				result.StatusCode = http.StatusForbidden
			} else if strings.Contains(e, "500") {
				result.StatusCode = http.StatusInternalServerError
			}
		}
		result.Body = respBody
		return result
	}

	result.Body = respBody
	result.StatusCode = http.StatusOK
	return result
}

// HttpPut performs an HTTP PUT and returns the response body. Recognized
// params: url, body, timeout, insecure, contentType, authorization,
// clientCrt, clientKey, clientCA.
func (tpl *Template) HttpPut(params map[string]interface{}) ([]byte, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("HttpPut err => %s", "no params allowed")
	}

	u, _ := params["url"].(string)
	var timeout int
	switch tp := params["timeout"].(type) {
	case int:
		timeout = tp
	case float64:
		timeout = int(tp)
	default:
		timeout = 5
	}

	insecure, _ := params["insecure"].(bool)
	contentType, _ := params["contentType"].(string)
	authorization, _ := params["authorization"].(string)
	body := bodyFromParam(params["body"])
	clientCrt, _ := params["clientCrt"].(string)
	clientKey, _ := params["clientKey"].(string)
	clientCA, _ := params["clientCA"].(string)

	transport, err := newHTTPTransport(timeout, insecure, clientCrt, clientKey, clientCA)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}
	return utils.HttpPutRaw(&client, u, contentType, authorization, body)
}

// HttpPatch performs an HTTP PATCH and returns the response body.
// Recognized params: url, body, timeout, insecure, contentType,
// authorization, clientCrt, clientKey, clientCA.
func (tpl *Template) HttpPatch(params map[string]interface{}) ([]byte, error) {
	if len(params) == 0 {
		return nil, fmt.Errorf("HttpPut err => %s", "no params allowed")
	}

	u, _ := params["url"].(string)
	timeout, _ := params["timeout"].(int)
	if timeout == 0 {
		timeout = 5
	}

	insecure, _ := params["insecure"].(bool)
	contentType, _ := params["contentType"].(string)
	authorization, _ := params["authorization"].(string)
	body := bodyFromParam(params["body"])
	clientCrt, _ := params["clientCrt"].(string)
	clientKey, _ := params["clientKey"].(string)
	clientCA, _ := params["clientCA"].(string)

	transport, err := newHTTPTransport(timeout, insecure, clientCrt, clientKey, clientCA)
	if err != nil {
		return nil, err
	}

	client := http.Client{
		Timeout:   time.Duration(timeout) * time.Second,
		Transport: transport,
	}
	return utils.HttpPatchRaw(&client, u, contentType, authorization, body)
}

// ReadFile reads the file at params["filePath"] and returns its content.
func (tpl *Template) ReadFile(params map[string]any) ReadFileResult {
	if len(params) == 0 {
		return ReadFileResult{Error: "no params allowed"}
	}

	filePath, _ := params["filePath"].(string)
	if filePath == "" {
		return ReadFileResult{Error: "filePath is required"}
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return ReadFileResult{Error: err.Error()}
	}

	return ReadFileResult{Content: content}
}

// HttpForm URL-encodes params as an application/x-www-form-urlencoded body,
// JSON-encoding any map or slice values.
func (tpl *Template) HttpForm(params map[string]interface{}) ([]byte, error) {

	if len(params) == 0 {
		return nil, fmt.Errorf("HttpForm err => %s", "no params allowed")
	}

	form := url.Values{}
	for k, v := range params {

		// encode map to json
		m, ok := v.(map[string]interface{})
		if ok {

			b, err := json.Marshal(m)
			if err != nil {
				return nil, err
			}
			form.Add(k, string(b))
			continue
		}

		// encode interface array to json
		ia, ok := v.([]interface{})
		if ok {
			b, err := json.Marshal(ia)
			if err != nil {
				return nil, err
			}
			form.Add(k, string(b))
			continue
		}

		// encode string array to json
		sa, ok := v.([]string)
		if ok {
			b, err := json.Marshal(sa)
			if err != nil {
				return nil, err
			}
			form.Add(k, string(b))
			continue
		}

		form.Add(k, fmt.Sprintf("%v", v))
	}
	return []byte(form.Encode()), nil
}

// TemplateRender renders the named template from the current template's
// content, using obj as the data object, and returns the result as a
// string. The rendered template only has access to Template's lite
// funcmap.
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
// the data object, and returns the result as a string. The rendered
// template only has access to Template's lite funcmap.
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

// DirCreate creates path, and any missing parents, using the given
// permission mode. A mode of 0 defaults to 0755.
func (tpl *Template) DirCreate(path string, mode int) error {

	m := mode
	if m <= 0 {
		m = 0755
	}

	return os.MkdirAll(path, os.FileMode(m))
}

// FileCreate writes content to path, creating any missing parent
// directories, using the given permission mode. A mode of 0 defaults to
// 0644.
func (tpl *Template) FileCreate(path, content string, mode int) error {

	m := mode
	if m <= 0 {
		m = 0644
	}
	err := os.MkdirAll(filepath.Dir(path), 0755)
	if err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), os.FileMode(m))
}

// customRender renders name against obj using tpl's parsed template.
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

// Render executes the template using tpl's configured options and returns
// the result.
func (tpl *TextTemplate) Render() ([]byte, error) {
	return tpl.customRender(tpl.options.Name, nil)
}

// RenderObject executes the template using obj as the data object and
// returns the result.
func (tpl *TextTemplate) RenderObject(obj interface{}) ([]byte, error) {
	return tpl.customRender(tpl.options.Name, obj)
}

// NewTextTemplate parses options.Content as a text/template.Template, with
// Template's lite funcmap (overridden by options.Funcs) available to it,
// and returns the resulting TextTemplate.
func NewTextTemplate(options TemplateOptions, logger common.Logger) (*TextTemplate, error) {

	if utils.IsEmpty(options.Content) {
		return nil, errors.New("no content")
	}

	tt := &TextTemplate{}
	tt.Configure(options, logger)

	funcs := sprig.TxtFuncMap()
	tt.SetFuncs(funcs)
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

	tt.SetNode(t)
	tt.template = t
	tt.funcs = funcs
	return tt, nil
}
