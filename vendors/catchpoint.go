package vendors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"

	"github.com/devopsext/utils"
)

const catchpointAPIURL = "https://io.catchpoint.com/api/"
const catchpointAPIVersion = "v3.2"

const (
	catchpointAPIInstantTest = "instanttests"
	catchpointAPINodesGroups = "nodes/groups/"
	catchpointAPINodesAll    = "nodes/all"
)

type Catchpoint struct {
	client  *http.Client
	options CatchpointOptions
}

type CatchpointOptions struct {
	APIToken string
	Timeout  int
	Insecure bool
}

type CatchpointNodesGetAllOptions struct {
	Name        string
	Targeted    bool
	Active      bool
	Paused      bool
	NetworkType int
	City        string
	Country     string
	IPv6        bool
	ASN         string
	AsNumber    int
	PageNumber  int
	PageSize    int
}

type CatchpointInstantTestNodes struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	InstantTestStatus string `json:"instantTestStatus"`
}

type CatchpointIstantTestData struct {
	ID               int                         `json:"id"`
	InstantTestNodes *CatchpointInstantTestNodes `json:"instantTestNodes"`
}

type CatchpointIstantTestResponse struct {
	Data      *CatchpointIstantTestData `json:"data,omitempty"`
	Completed bool                      `json:"completed"`
}

type CatchpointInstantTestOptions struct {
	URL             string
	NodesIds        string
	HTTPMethodType  int
	InstantTestType int
	MonitorType     int
	OnDemand        bool
}

type CatchpointInstantTestWithNodeGroupOptions struct {
	URL             string
	NodeGroupID     int
	HTTPMethodType  int
	InstantTestType int
	MonitorType     int
	OnDemand        bool
}

type ID struct {
	ID int `json:"id"`
}

type CatchpointNodeGroup struct {
	ID int `json:"id"`
}

type HTTPMethodType struct {
	ID int `json:"id"`
}

type InstantTestType struct {
	ID int `json:"id"`
}

type MonitorType struct {
	ID int `json:"id"`
}

type NetworkType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Node struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	NetworkType *NetworkType `json:"networkType"`
}

type NodeGroupItem struct {
	ID            int           `json:"id"`
	Name          string        `json:"name"`
	DivisionID    int           `json:"divisionId"`
	Nodes         *[]Node       `json:"nodes"`
	NodeLocations []interface{} `json:"nodeLocations"`
}

type NodeGroupData struct {
	NodeGroups *[]NodeGroupItem `json:"nodeGroups"`
}

type NodeGroup struct {
	Data      *NodeGroupData `json:"data"`
	Messages  []interface{}  `json:"messages"`
	Errors    []interface{}  `json:"errors"`
	Completed bool           `json:"completed"`
	TraceID   string         `json:"traceId"`
}

type CatchpointInstantTest struct {
	URL             string           `json:"url"`
	NodesIds        *[]ID            `json:"nodesIds"`
	HTTPMethodType  *HTTPMethodType  `json:"httpMethodType"`
	InstantTestType *InstantTestType `json:"instantTestType"`
	MonitorType     *MonitorType     `json:"monitorType"`
}

type CatchpointInstantTestWithNodeGroup struct {
	URL             string           `json:"url"`
	NodeGroupID     int              `json:"nodesIds"`
	HTTPMethodType  *HTTPMethodType  `json:"httpMethodType"`
	InstantTestType *InstantTestType `json:"instantTestType"`
	MonitorType     *MonitorType     `json:"monitorType"`
}

type CatchpointResponsePayload struct {
	Tests []struct {
		ID          int    `json:"id"`
		Status      string `json:"status"`
		ResultURL   string `json:"resultUrl"`
		Description string `json:"description"`
	} `json:"tests"`
}

func (c *Catchpoint) apiURL(cmd string) string {
	return catchpointAPIURL + catchpointAPIVersion + "/" + cmd
}

func (c *Catchpoint) getAuth(opts CatchpointOptions) string {

	auth := ""
	if !utils.IsEmpty(opts.APIToken) {
		auth = fmt.Sprintf("Bearer %s", opts.APIToken)
		return auth
	}
	return auth
}

func (c *Catchpoint) GetNodesFromGroup(options CatchpointNodeGroup) ([]byte, error) {
	return c.CustomGetNodesFromGroup(c.options, options)
}

func (c *Catchpoint) CustomGetNodesFromGroup(catchpointOptions CatchpointOptions, options CatchpointNodeGroup) ([]byte, error) {
	return utils.HttpGetRaw(c.client, c.apiURL(catchpointAPINodesGroups+fmt.Sprintf("%d", options.ID)), "application/json", c.getAuth(catchpointOptions))
}

func (c *Catchpoint) InstantTest(options CatchpointInstantTestOptions) ([]byte, error) {
	return c.CustomInstantTest(c.options, options)
}

func (c *Catchpoint) InstantTestWithNodeGroup(options CatchpointInstantTestWithNodeGroupOptions) ([]byte, error) {
	return c.CustomInstantTestWithNodeGroup(c.options, options)
}

func (c *Catchpoint) SearchNodesWithOptions(options CatchpointNodesGetAllOptions) ([]byte, error) {
	return c.CustomSearchNodesWithOptions(c.options, options)
}

func (c *Catchpoint) CustomSearchNodesWithOptions(catchpointOptions CatchpointOptions, catchpointNodesGetAllOptions CatchpointNodesGetAllOptions) ([]byte, error) {

	params := make(url.Values)
	if !utils.IsEmpty(catchpointNodesGetAllOptions.Name) {
		params.Add("name", catchpointNodesGetAllOptions.Name)
	}
	if !utils.IsEmpty(catchpointNodesGetAllOptions.Targeted) {
		params.Add("targeted", strconv.FormatBool(catchpointNodesGetAllOptions.Targeted))
	}
	if !utils.IsEmpty(catchpointNodesGetAllOptions.Active) {
		params.Add("active", strconv.FormatBool(catchpointNodesGetAllOptions.Active))
	}
	if !utils.IsEmpty(catchpointNodesGetAllOptions.Paused) {
		params.Add("paused", strconv.FormatBool(catchpointNodesGetAllOptions.Paused))
	}
	if !utils.IsEmpty(catchpointNodesGetAllOptions.NetworkType) {
		params.Add("networkType", strconv.Itoa(catchpointNodesGetAllOptions.NetworkType))
	}
	if !utils.IsEmpty(catchpointNodesGetAllOptions.City) {
		params.Add("city", catchpointNodesGetAllOptions.City)
	}
	if !utils.IsEmpty(catchpointNodesGetAllOptions.Country) {
		params.Add("country", catchpointNodesGetAllOptions.Country)
	}
	if !utils.IsEmpty(catchpointNodesGetAllOptions.IPv6) {
		params.Add("ipv6", strconv.FormatBool(catchpointNodesGetAllOptions.IPv6))
	}
	if !utils.IsEmpty(catchpointNodesGetAllOptions.ASN) {
		params.Add("asn", catchpointNodesGetAllOptions.ASN)
	}
	if !utils.IsEmpty(catchpointNodesGetAllOptions.AsNumber) {
		params.Add("asNumber", strconv.Itoa(catchpointNodesGetAllOptions.AsNumber))
	}
	if !utils.IsEmpty(catchpointNodesGetAllOptions.PageNumber) {
		params.Add("pageNumber", strconv.Itoa(catchpointNodesGetAllOptions.PageNumber))
	}
	if !utils.IsEmpty(catchpointNodesGetAllOptions.PageSize) {
		params.Add("pageSize", strconv.Itoa(catchpointNodesGetAllOptions.PageSize))
	}

	u, err := url.Parse(catchpointAPIURL + catchpointAPIVersion)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, catchpointAPINodesAll)
	u.RawQuery = params.Encode()

	return utils.HttpGetRaw(c.client, u.String(), "application/json", c.getAuth(catchpointOptions))
}

func (c *Catchpoint) CustomInstantTestWithNodeGroup(catchpointOptions CatchpointOptions, catchpointInstantTestWithNodeGroupOptions CatchpointInstantTestWithNodeGroupOptions) ([]byte, error) {

	nodeIDsBytes, err := c.GetNodesFromGroup(CatchpointNodeGroup{ID: catchpointInstantTestWithNodeGroupOptions.NodeGroupID})
	if err != nil {
		return nil, err
	}

	var nodeIDs *NodeGroup
	err = json.Unmarshal(nodeIDsBytes, &nodeIDs)
	if err != nil {
		return nil, err
	}

	var ids []ID
	for _, group := range *nodeIDs.Data.NodeGroups {
		for _, node := range *group.Nodes {
			ids = append(ids, ID{ID: node.ID})
		}
	}

	body := &CatchpointInstantTest{
		URL:             catchpointInstantTestWithNodeGroupOptions.URL,
		NodesIds:        &ids,
		InstantTestType: &InstantTestType{ID: catchpointInstantTestWithNodeGroupOptions.InstantTestType},
		HTTPMethodType:  &HTTPMethodType{ID: catchpointInstantTestWithNodeGroupOptions.HTTPMethodType},
		MonitorType:     &MonitorType{ID: catchpointInstantTestWithNodeGroupOptions.MonitorType},
	}

	req, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	params := make(url.Values)
	params.Add("onDemand", strconv.FormatBool(catchpointInstantTestWithNodeGroupOptions.OnDemand))

	u, err := url.Parse(catchpointAPIURL + catchpointAPIVersion)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, catchpointAPIInstantTest)
	u.RawQuery = params.Encode()

	return utils.HttpPostRaw(c.client, u.String(), "application/json", c.getAuth(catchpointOptions), req)
}

func (c *Catchpoint) CustomInstantTest(catchpointOptions CatchpointOptions, catchpointInstantTestOptions CatchpointInstantTestOptions) ([]byte, error) {

	params := make(url.Values)
	params.Add("onDemand", strconv.FormatBool(catchpointInstantTestOptions.OnDemand))

	u, err := url.Parse(catchpointAPIURL + catchpointAPIVersion)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, catchpointAPIInstantTest)
	u.RawQuery = params.Encode()

	var ids []ID
	for _, id := range strings.Split(catchpointInstantTestOptions.NodesIds, ",") {
		idInt, _ := strconv.Atoi(id)
		ids = append(ids, ID{ID: idInt})
	}

	body := &CatchpointInstantTest{
		URL:             catchpointInstantTestOptions.URL,
		NodesIds:        &ids,
		InstantTestType: &InstantTestType{ID: catchpointInstantTestOptions.InstantTestType},
		HTTPMethodType:  &HTTPMethodType{ID: catchpointInstantTestOptions.HTTPMethodType},
		MonitorType:     &MonitorType{ID: catchpointInstantTestOptions.MonitorType},
	}

	req, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return utils.HttpPostRaw(c.client, u.String(), "application/json", c.getAuth(catchpointOptions), req)
}

func NewCatchpoint(options CatchpointOptions) *Catchpoint {

	catchpoint := &Catchpoint{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
	return catchpoint
}
