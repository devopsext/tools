package vendors

import (
	"net/http"

	"github.com/devopsext/utils"
)

type JiraOptions struct {
	URL      string
	Timeout  int
	Insecure bool
	User     string
	Password string
}

type Jira struct {
	client  *http.Client
	options JiraOptions
}

func (j *Jira) CreateTask() ([]byte, error) {
	return nil, nil
}

func NewJira(options JiraOptions) *Jira {

	return &Jira{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
