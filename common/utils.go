package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/blues/jsonata-go"
	"github.com/devopsext/utils"
)

// we need custom json marshal due to no html escaption
func jsonMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

func jsonataVars(p string, t interface{}) (map[string]interface{}, error) {

	bytes, err := jsonMarshal(t)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(bytes, &m)
	if err != nil {
		return nil, err
	}
	r := make(map[string]interface{})
	for k, v := range m {
		k = fmt.Sprintf("%s%s", p, k)
		r[k] = v
	}
	return r, nil
}

func Output(query, to string, prefix string, opts interface{}, bytes []byte, stdout *Stdout) {

	b, err := utils.Content(query)
	if err != nil {
		stdout.Panic(err)
	}
	query = string(b)

	output := string(bytes)
	if !utils.IsEmpty(query) {

		vars, err := jsonataVars(prefix, opts)
		if err == nil {
			jsonata.RegisterVars(vars)
		}

		expr := jsonata.MustCompile(query)

		var v interface{}
		err = json.Unmarshal(bytes, &v)
		if err != nil {
			stdout.Panic(err)
		}
		v, err = expr.Eval(v)
		if err != nil {
			stdout.Panic(err)
		}
		b, err = jsonMarshal(v)
		if err != nil {
			output = fmt.Sprintf("%v", v)
		} else {
			output = string(b)
		}
	}

	if utils.IsEmpty(to) {
		stdout.Info(output)
	} else {
		stdout.Debug("Writing output to %s...", to)
		err := ioutil.WriteFile(to, []byte(output), 0644)
		if err != nil {
			stdout.Error(err)
		}
	}
}
