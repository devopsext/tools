package vendors

import (
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
)

type ZabbixHostGetOptions struct {
	Fields     []string
	Inventory  []string
	Interfaces []string
}

type ZabbixOptions struct {
	Timeout  int
	Insecure bool
	URL      string
	User     string
	Password string
	Auth     string
}

type Zabbix struct {
	client  *http.Client
	options ZabbixOptions
}

type ZabbixUserLoginResponse struct {
	Result string `json:"result"`
}

type ZabbixUserLoginParams struct {
	User     string `json:"user"`
	Password string `json:"password"`
}

type ZabbixUserLogin struct {
	JsonRPC string                 `json:"jsonrpc"`
	Method  string                 `json:"method"`
	Params  *ZabbixUserLoginParams `json:"params"`
	ID      int                    `json:"id"`
}

type ZabbixHostGetParams struct {
	Output           []string `json:"output"`
	SelectInventory  []string `json:"selectInventory"`
	SelectInterfaces []string `json:"selectInterfaces"`
}

type ZabbixHostGet struct {
	JsonRPC string               `json:"jsonrpc"`
	Method  string               `json:"method"`
	Params  *ZabbixHostGetParams `json:"params"`
	Auth    string               `json:"auth"`
	ID      int                  `json:"id"`
}

const (
	zabbixContentType    = "application/json-rpc"
	zabbixJsonRpcURL     = "/api_jsonrpc.php"
	zabbixJsonRpcVersion = "2.0"
)

func (o *Zabbix) getZabbixAuth(opts ZabbixOptions) (*ZabbixUserLoginResponse, error) {

	u, err := url.Parse(opts.URL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, zabbixJsonRpcURL)

	r := &ZabbixUserLogin{

		JsonRPC: zabbixJsonRpcVersion,
		Method:  "user.login",
		ID:      1,
		Params: &ZabbixUserLoginParams{
			User:     opts.User,
			Password: opts.Password,
		},
	}

	req, err := json.Marshal(&r)
	if err != nil {
		return nil, err
	}

	res, err := common.HttpPostRaw(o.client, u.String(), zabbixContentType, "", req)
	if err != nil {
		return nil, err
	}

	var zr ZabbixUserLoginResponse
	err = json.Unmarshal(res, &zr)
	if err != nil {
		return nil, err
	}
	return &zr, nil
}

func (o *Zabbix) CustomGetHosts(options ZabbixOptions, hostGetOptions ZabbixHostGetOptions) ([]byte, error) {

	auth := options.Auth
	if utils.IsEmpty(auth) {
		za, err := o.getZabbixAuth(options)
		if err != nil {
			return nil, err
		}
		auth = za.Result
	}

	u, err := url.Parse(options.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, zabbixJsonRpcURL)

	r := &ZabbixHostGet{

		JsonRPC: zabbixJsonRpcVersion,
		Method:  "host.get",
		Auth:    auth,
		ID:      1,
		Params: &ZabbixHostGetParams{
			Output:           common.RemoveEmptyStrings(hostGetOptions.Fields),
			SelectInventory:  common.RemoveEmptyStrings(hostGetOptions.Inventory),
			SelectInterfaces: common.RemoveEmptyStrings(hostGetOptions.Interfaces),
		},
	}

	req, err := json.Marshal(&r)
	if err != nil {
		return nil, err
	}

	return common.HttpPostRaw(o.client, u.String(), zabbixContentType, "", req)
}

func (o *Zabbix) GetHosts(options ZabbixHostGetOptions) ([]byte, error) {
	return o.CustomGetHosts(o.options, options)
}

func NewZabbix(options ZabbixOptions) *Zabbix {

	return &Zabbix{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
