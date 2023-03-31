package common

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"

	"github.com/blues/jsonata-go"
	"github.com/blues/jsonata-go/jtypes"
	"github.com/devopsext/utils"
)

type OutputOptions struct {
	Output string
	Query  string
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

func jsonataEnv(key string) string {
	return os.Getenv(key)
}

func jsonantaHttpGet(URL, contentType string) interface{} {

	var result interface{}

	bytes, err := HttpGetRaw(http.DefaultClient, URL, contentType, "")
	if err != nil {
		return err
	}
	err = json.Unmarshal(bytes, &result)
	if err != nil {
		return err
	}
	return result
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
		for _, v := range opts {
			vars, err := InterfaceToMap(prefix, v)
			if err == nil {
				jsonata.RegisterVars(vars)
			}
		}

		exts := make(map[string]jsonata.Extension)
		exts["env"] = jsonata.Extension{
			Func:               jsonataEnv,
			UndefinedHandler:   jtypes.ArgUndefined(0),
			EvalContextHandler: jtypes.ArgCountEquals(0),
		}
		exts["httpGet"] = jsonata.Extension{
			Func:               jsonantaHttpGet,
			UndefinedHandler:   jtypes.ArgUndefined(0),
			EvalContextHandler: jtypes.ArgCountEquals(0),
		}
		jsonata.RegisterExts(exts)

		expr := jsonata.MustCompile(query)

		var v interface{}
		err = json.Unmarshal(bytes, &v)
		if err != nil {
			stdout.Panic(err)
		}
		v1, err := expr.Eval(v)
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
		err := ioutil.WriteFile(to, []byte(output), 0600)
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
		err := ioutil.WriteFile(output, bytes, 0600)
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

func HttpRequestRawWithHeaders(client *http.Client, method, URL string, headers map[string]string, raw []byte) ([]byte, error) {

	reader := bytes.NewReader(raw)

	req, err := http.NewRequest(method, URL, reader)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		if utils.IsEmpty(v) {
			continue
		}
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("response status code: %s", resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func HttpPostRawWithHeaders(client *http.Client, URL string, headers map[string]string, raw []byte) ([]byte, error) {
	return HttpRequestRawWithHeaders(client, "POST", URL, headers, raw)
}

func HttpPostRaw(client *http.Client, URL, contentType string, authorization string, raw []byte) ([]byte, error) {

	headers := make(map[string]string)
	if !utils.IsEmpty(contentType) {
		headers["Content-Type"] = contentType
	}
	if !utils.IsEmpty(authorization) {
		headers["Authorization"] = authorization
	}
	return HttpPostRawWithHeaders(client, URL, headers, raw)
}

func HttpRequestRawWithHeadersOutCode(client *http.Client, method, URL string, headers map[string]string, raw []byte) (body []byte, code int, err error) {

	reader := bytes.NewReader(raw)

	req, err := http.NewRequest(method, URL, reader)
	if err != nil {
		return nil, 0, err
	}

	for k, v := range headers {
		if utils.IsEmpty(v) {
			continue
		}
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, 0, fmt.Errorf(resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, 0, err
	}

	return b, resp.StatusCode, nil
}

func HttpPostRawWithHeadersOutCode(client *http.Client, URL string, headers map[string]string, raw []byte) (body []byte, code int, err error) {
	return HttpRequestRawWithHeadersOutCode(client, "POST", URL, headers, raw)
}

func HttpPostRawOutCode(client *http.Client, URL, contentType string, authorization string, raw []byte) (body []byte, code int, err error) {

	headers := make(map[string]string)
	if !utils.IsEmpty(contentType) {
		headers["Content-Type"] = contentType
	}
	if !utils.IsEmpty(authorization) {
		headers["Authorization"] = authorization
	}
	return HttpPostRawWithHeadersOutCode(client, URL, headers, raw)
}

func HttpPutRawWithHeaders(client *http.Client, URL string, headers map[string]string, raw []byte) ([]byte, error) {
	return HttpRequestRawWithHeaders(client, "PUT", URL, headers, raw)
}

func HttpPutRaw(client *http.Client, URL, contentType string, authorization string, raw []byte) ([]byte, error) {

	headers := make(map[string]string)
	if !utils.IsEmpty(contentType) {
		headers["Content-Type"] = contentType
	}
	if !utils.IsEmpty(authorization) {
		headers["Authorization"] = authorization
	}
	return HttpPutRawWithHeaders(client, URL, headers, raw)
}

func HttpGetRawWithHeaders(client *http.Client, URL string, headers map[string]string) ([]byte, error) {

	req, err := http.NewRequest("GET", URL, nil)
	if err != nil {
		return nil, err
	}

	for k, v := range headers {
		if utils.IsEmpty(v) {
			continue
		}
		req.Header.Set(k, v)
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(resp.Status)
	}

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func HttpGetRaw(client *http.Client, URL, contentType string, authorization string) ([]byte, error) {

	headers := make(map[string]string)
	if !utils.IsEmpty(contentType) {
		headers["Content-Type"] = contentType
	}
	if !utils.IsEmpty(authorization) {
		headers["Authorization"] = authorization
	}

	return HttpGetRawWithHeaders(client, URL, headers)
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

func ReadAndMarshal(path string) (map[string]interface{}, error) {
	var result map[string]interface{}

	dat, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(dat, &result)
	if err != nil {
		return nil, err
	}

	return result, nil
}
