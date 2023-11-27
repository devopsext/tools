package vendors

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
)

type ObserviumOptions struct {
	Timeout  int
	Insecure bool
	URL      string
	User     string
	Password string
	Token    string
}

type Observium struct {
	client  *http.Client
	options ObserviumOptions
}

func (o *Observium) getAuth(opts ObserviumOptions) string {

	auth := ""
	if !utils.IsEmpty(opts.User) {
		userPass := fmt.Sprintf("%s:%s", opts.User, opts.Password)
		auth = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(userPass)))
		return auth
	}
	if !utils.IsEmpty(opts.Token) {
		auth = fmt.Sprintf("Bearer %s", opts.Token)
		return auth
	}
	return auth
}

func (o *Observium) CustomGetDevices(options ObserviumOptions) ([]byte, error) {

	u, err := url.Parse(options.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "/api/v0/devices/")

	return common.HttpGetRaw(o.client, u.String(), "application/json", o.getAuth(options))
}

func (o *Observium) GetDevices() ([]byte, error) {
	return o.CustomGetDevices(o.options)
}

func NewObservium(options ObserviumOptions) *Observium {

	return &Observium{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
