package vendors

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
	"net/http"
	"net/url"
	"path"
)

type CmdbOptions struct {
	Timeout  int
	Insecure bool
	APIURL   string
}
type CmdbOutputOptions struct {
	Output      string // path to output if empty to stdout
	OutputQuery string
}

type Cmdb struct {
	client  *http.Client
	options CmdbOptions
}

func (c Cmdb) GetComponentManifest(s string) ([]byte, error) {
	u, err := url.Parse(c.options.APIURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "components", s)
	return common.HttpGetRaw(c.client, u.String(), "", "")
}

func NewCmdb(options CmdbOptions) *Cmdb {
	return &Cmdb{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
