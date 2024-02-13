package common

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/devopsext/utils"
)

type OutputOptions struct {
	Output string
	Query  string
}

func FormatBasicAuth(user, pass string) string {
	auth := user + ":" + pass
	return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth))
}

// we need custom json marshal due to no html escaption
func JsonMarshal(t interface{}) ([]byte, error) {
	buffer := &bytes.Buffer{}
	encoder := json.NewEncoder(buffer)
	encoder.SetEscapeHTML(false)
	err := encoder.Encode(t)
	return buffer.Bytes(), err
}

func InterfaceToMap(prefix string, i interface{}) (map[string]interface{}, error) {
	data, err := JsonMarshal(i)
	if err != nil {
		return nil, err
	}
	var m map[string]interface{}
	err = json.Unmarshal(data, &m)
	if err != nil {
		return nil, err
	}
	r := make(map[string]interface{})
	for k, v := range m {
		k = fmt.Sprintf("%s%s", prefix, k)
		r[k] = v
	}
	return r, nil
}

func Output(query, to string, prefix string, opts []interface{}, bytes []byte, stdout *Stdout) {

	stdout.Debug("Raw output => %s", string(bytes))

	b, err := utils.Content(query)
	if err != nil {
		stdout.Panic(err)
	}
	query = string(b)

	output := string(bytes)
	if !utils.IsEmpty(query) {

		jnata := NewJsonata(JsonataOptions{})

		for _, v := range opts {
			vars, err := InterfaceToMap(prefix, v)
			if err == nil {
				jnata.RegisterVars(vars)
			}
		}

		var v interface{}
		err = json.Unmarshal(bytes, &v)
		if err != nil {
			stdout.Panic(err)
		}

		v1, err := jnata.Eval(v, query)
		if err != nil {
			stdout.Panic(err)
		}

		// v1 is json object
		_, ok := v1.(map[string]interface{})
		if ok {
			b, err = JsonMarshal(v1)
			if err != nil {
				output = fmt.Sprintf("%v", v1)
			} else {
				output = strings.TrimSpace(string(b))
			}
		} else {
			// v1 is json array
			_, ok = v1.([]interface{})
			if ok {
				b, err = JsonMarshal(v1)
				if err != nil {
					output = fmt.Sprintf("%v", v1)
				} else {
					output = strings.TrimSpace(string(b))
				}
			}
		}
		if !ok {
			output = fmt.Sprintf("%v", v1)
		}
	}

	if utils.IsEmpty(to) {
		stdout.Info(output)
	} else {
		stdout.Debug("Writing output to %s...", to)
		err := os.WriteFile(to, []byte(output), 0600)
		if err != nil {
			stdout.Error(err)
		}
	}
}

func OutputJson(outputOpts OutputOptions, prefix string, opts []interface{}, bytes []byte, stdout *Stdout) {
	Output(outputOpts.Query, outputOpts.Output, prefix, opts, bytes, stdout)
}

func OutputRaw(output string, bytes []byte, stdout *Stdout) {

	stdout.Debug("Raw output => %s", string(bytes))

	out := string(bytes)
	if utils.IsEmpty(output) {
		stdout.Info(out)
	} else {
		stdout.Debug("Writing output to %s...", output)
		err := os.WriteFile(output, bytes, 0600)
		if err != nil {
			stdout.Error(err)
		}
	}
}

func Debug(prefix string, obj interface{}, stdout *Stdout) {
	vars, err := InterfaceToMap(prefix, obj)
	if err != nil {
		stdout.Panic(err)
	}
	for k, v := range vars {
		if !utils.IsEmpty(v) {
			stdout.Debug("%s: %v", k, v)
		}
	}
}

func TruncateString(str string, length int) string {
	if length <= 0 {
		return ""
	}
	truncated := ""
	count := 0
	for _, char := range str {
		truncated += string(char)
		count++
		if count >= length {
			break
		}
	}
	return truncated
}

func ReadAndMarshal(input string) (map[string]interface{}, error) {

	var result map[string]interface{}

	dat, err := os.ReadFile(input)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {

			err = json.Unmarshal([]byte(input), &result)

			if err != nil {
				return nil, err
			}
		} else {
			return nil, err
		}
	} else {

		err = json.Unmarshal(dat, &result)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}

func RemoveEmptyStrings(items []string) []string {

	r := []string{}

	for _, v := range items {
		if utils.IsEmpty(v) {
			continue
		}
		r = append(r, strings.TrimSpace(v))
	}

	return r
}
