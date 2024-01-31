package common

import (
	"encoding/json"
	"net/http"
	"os"

	"github.com/blues/jsonata-go"
	"github.com/blues/jsonata-go/jtypes"
	"github.com/devopsext/utils"
)

type JsonataOptions struct {
}

type Jsonata struct {
	options JsonataOptions
}

func (j *Jsonata) RegisterVars(vars map[string]interface{}) error {
	return jsonata.RegisterVars(vars)
}

func (j *Jsonata) fEnv(key string) string {
	return os.Getenv(key)
}

func (j *Jsonata) fHttpGet(URL, contentType string) interface{} {

	var r interface{}

	bytes, err := utils.HttpGetRaw(http.DefaultClient, URL, contentType, "")
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &r)
	if err != nil {
		return err
	}
	return r
}

/*func (j *Jsonata) fRegexMatchObjectByFields(obj map[string]interface{}, field, value string) []interface{} {

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

func (j *Jsonata) fRegexMatchObjectByField(obj map[string]interface{}, field, value string) interface{} {

	keys := j.fRegexMatchObjectByFields(obj, field, value)
	if len(keys) == 0 {
		return value
	}
	return keys[0]
}*/

func (j *Jsonata) Eval(data interface{}, query string) (interface{}, error) {

	exts := make(map[string]jsonata.Extension)
	exts["env"] = jsonata.Extension{
		Func:               j.fEnv,
		UndefinedHandler:   jtypes.ArgUndefined(0),
		EvalContextHandler: jtypes.ArgCountEquals(0),
	}
	exts["httpGet"] = jsonata.Extension{
		Func:               j.fHttpGet,
		UndefinedHandler:   jtypes.ArgUndefined(0),
		EvalContextHandler: jtypes.ArgCountEquals(0),
	}
	/*exts["regexMatchObjectByField"] = jsonata.Extension{
		Func:               tpl,
		UndefinedHandler:   jtypes.ArgUndefined(0),
		EvalContextHandler: jtypes.ArgCountEquals(0),
	}*/
	jsonata.RegisterExts(exts)

	expr := jsonata.MustCompile(query)
	v, err := expr.Eval(data)
	if err != nil {
		return nil, err
	}
	return v, nil
}

func NewJsonata(options JsonataOptions) *Jsonata {

	jsonata := &Jsonata{
		options: options,
	}
	return jsonata
}
