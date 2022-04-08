package common

import (
	"encoding/json"
	"fmt"
	"io/ioutil"

	"github.com/blues/jsonata-go"
	"github.com/devopsext/utils"
)

func Output(query, to string, bytes []byte, stdout *Stdout) {

	b, err := utils.Content(query)
	if err != nil {
		stdout.Panic(err)
	}
	query = string(b)

	output := string(bytes)
	if !utils.IsEmpty(query) {
		expr := jsonata.MustCompile(query)

		var v interface{}
		err := json.Unmarshal(bytes, &v)
		if err != nil {
			stdout.Panic(err)
		}
		v, err = expr.Eval(v)
		if err != nil {
			stdout.Panic(err)
		}
		output = fmt.Sprintf("%v", v)
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
