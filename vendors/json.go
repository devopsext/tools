package vendors

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
	"net/http"
)

type JSONOptions struct {
	Timeout  int
	Insecure bool
	URL      string
}
type JSONOutputOptions struct {
	Output      string // path to output if empty to stdout
	OutputQuery string
}

type JSON struct {
	client  *http.Client
	options JSONOptions
}

func (c JSON) Get() ([]byte, error) {
	return common.HttpGetRaw(c.client, c.options.URL, "", "")
}

func NewJSON(options JSONOptions) *JSON {
	return &JSON{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
