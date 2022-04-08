package vendors

import (
	"net/http"

	"github.com/devopsext/utils"
)

type GraylogOptions struct {
	URL      string
	Timeout  int
	Insecure bool

	Output string
	Query  string
}

type Graylog struct {
	client  *http.Client
	options GraylogOptions
}

func (g *Graylog) Logs() ([]byte, error) {
	return nil, nil
}

func NewGraylog(options GraylogOptions) *Graylog {

	return &Graylog{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
