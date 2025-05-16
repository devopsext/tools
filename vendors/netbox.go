package vendors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/devopsext/utils"
)

type NetboxOptions struct {
	Timeout  int
	Insecure bool
	URL      string
	Token    string
	Limit    string
	Brief    bool
	Filter   map[string]string
}

type NetboxDeviceOptions struct {
	DeviceID string
}

type Netbox struct {
	client  *http.Client
	options NetboxOptions
}

type NetxboxAPIResponse struct {
	Count    int             `json:"count"`
	Next     string          `json:"next"`
	Previous string          `json:"previous"`
	Results  json.RawMessage `json:"results"`
}

func (n *Netbox) getAuth(options NetboxOptions) string {

	auth := ""
	if !utils.IsEmpty(options.Token) {
		auth = fmt.Sprintf("Token %s", options.Token)
		return auth
	}
	return auth
}

func (n *Netbox) setParams(options NetboxOptions) url.Values {

	var params = make(url.Values)
	params.Add("limit", options.Limit)

	if options.Brief {
		params.Add("brief", "1")
	}

	if !utils.IsEmpty(options.Filter) {
		for param, val := range options.Filter {
			params.Add(param, val)
		}
	}

	return params
}

func (n *Netbox) CustomGetDevices(options NetboxOptions, netboxDeviceOptions NetboxDeviceOptions) ([]byte, error) {

	u, err := url.Parse(options.URL)
	if err != nil {
		return nil, err
	}

	u.RawQuery = n.setParams(options).Encode()

	u.Path = path.Join(u.Path, "/api/dcim/devices/")

	if !utils.IsEmpty(netboxDeviceOptions.DeviceID) {
		u.Path = path.Join(u.Path, fmt.Sprintf("%s/", netboxDeviceOptions.DeviceID))

		return utils.HttpGetRaw(n.client, u.String(), "application/json", n.getAuth(options))
	}

	var devices json.RawMessage

	for {
		resp, err := utils.HttpGetRaw(n.client, u.String(), "application/json", n.getAuth(options))

		if err != nil {
			return nil, err
		}

		var apiResp NetxboxAPIResponse

		err = json.Unmarshal(resp, &apiResp)
		if err != nil {
			return nil, err
		}

		devices = append(devices, apiResp.Results...)

		if apiResp.Next == "" {
			break
		}

		u, err = url.Parse(apiResp.Next)
		if err != nil {
			return nil, err
		}
	}

	return devices, nil
}

func (n *Netbox) GetDevices(deviceOptions NetboxDeviceOptions) ([]byte, error) {
	return n.CustomGetDevices(n.options, deviceOptions)
}

func NewNetbox(options NetboxOptions) *Netbox {

	return &Netbox{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
