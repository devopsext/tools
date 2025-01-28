package vendors

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
)

const catchpointAPIURL = "https://io.catchpoint.com/api/"
const catchpointAPIVersion = "v3.2"

const (
	catchpointAPIInstantTest = "instanttests"
	catchpointAPINodesGroups = "nodes/groups/"
	catchpointAPINodesAll    = "nodes/all"
)

const catchpointRetryHeader = "Retry-After"

type Catchpoint struct {
	client  *http.Client
	options CatchpointOptions
}

type CatchpointOptions struct {
	APIToken string
	Timeout  int
	Insecure bool
	Retries  int
}

type CatchpointSearchNodesWithOptions struct {
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
	ID               int                           `json:"id"`
	InstantTestNodes *[]CatchpointInstantTestNodes `json:"instantTestNodes"`
}

type CatchpointIstantTestResponse struct {
	Data *CatchpointIstantTestData `json:"data,omitempty"`
	*CatchpointReponse
}

type CatchpointMessage struct {
	Message string `json:"message"`
}

type CatchpointReponse struct {
	Errors    *[]CatchpointMessage `json:"errors"`
	Messages  *[]CatchpointMessage `json:"messages"`
	Completed bool                 `json:"completed"`
	TraceId   string               `json:"traceId"`
}

type CatchpointSearchNodesWithOptionsData struct {
	Nodes *[]Node `json:"nodes"`
}

type CatchpointSearchNodesWithOptionsResponse struct {
	Data      *CatchpointSearchNodesWithOptionsData `json:"data"`
	Messages  *[]CatchpointMessage                  `json:"messages"`
	Completed bool                                  `json:"completed"`
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

type CatchpointHTTPMethodType struct {
	ID int `json:"id"`
}

type CatchpointInstantTestType struct {
	ID int `json:"id"`
}

type CatchpointMonitorType struct {
	ID int `json:"id"`
}

type CatchpointNetworkType struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type CatchpointCountry struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type Node struct {
	ID          int                    `json:"id"`
	Name        string                 `json:"name"`
	NetworkType *CatchpointNetworkType `json:"networkType"`
	Country     *CatchpointCountry     `json:"country"`
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
	Data *NodeGroupData `json:"data"`
	*CatchpointReponse
}

type CatchpointInstantTest struct {
	URL             string                     `json:"url"`
	NodesIds        *[]ID                      `json:"nodesIds"`
	HTTPMethodType  *CatchpointHTTPMethodType  `json:"httpMethodType"`
	InstantTestType *CatchpointInstantTestType `json:"instantTestType"`
	MonitorType     *CatchpointMonitorType     `json:"monitorType"`
}

type CatchpointInstantTestWithNodeGroup struct {
	URL             string                     `json:"url"`
	NodeGroupID     int                        `json:"nodesIds"`
	HTTPMethodType  *CatchpointHTTPMethodType  `json:"httpMethodType"`
	InstantTestType *CatchpointInstantTestType `json:"instantTestType"`
	MonitorType     *CatchpointMonitorType     `json:"monitorType"`
}

type CatchpointInstantTestResultHostsFields struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
}

type CatchpointInstantTestResultHostsMetrics struct {
	HostName string    `json:"hostName"`
	Items    []float64 `json:"items"`
}

type CatchpointInstantTestResultHosts struct {
	Fields  *[]CatchpointInstantTestResultHostsFields  `json:"fields"`
	Metrics *[]CatchpointInstantTestResultHostsMetrics `json:"metrics"`
}

type CatchpointInstantTestResultWebRecordItemsNavigationUrl struct {
	Scheme       string `json:"scheme"`
	Host         string `json:"host"`
	PathAndQuery string `json:"pathAndQuery"`
	AbsoluteUri  string `json:"absoluteUri"`
}

type CatchpointInstantTestResultWebRecordItems struct {
	IPAddess      string                                                  `json:"ipAddress"`
	NavigationUrl *CatchpointInstantTestResultWebRecordItemsNavigationUrl `json:"navigationUrl"`
	ResponseCode  int                                                     `json:"responseCode"`
}

type CatchpointInstantTestResultWebRecord struct {
	Items *[]CatchpointInstantTestResultWebRecordItems `json:"items"`
}
type CatchpointInstantTestResult struct {
	Hosts      *CatchpointInstantTestResultHosts     `json:"hosts"`
	WebRecords *CatchpointInstantTestResultWebRecord `json:"webRecords"`
}

type CatchpointInstantTestResultRecord struct {
	TestResult  *CatchpointInstantTestResult `json:"testResult"`
	ID          int                          `json:"id"`
	Node        *Node                        `json:"node"`
	MonitorType *CatchpointMonitorType       `json:"monitorType"`
	PublicLink  string                       `json:"publicLink"`
}

type CatchpointInstantTestResultData struct {
	InstantTestStatus string                             `json:"instantTestStatus"`
	InstantTestRecord *CatchpointInstantTestResultRecord `json:"instantTestRecord"`
}

type CatchpointInstantTestResultReponse struct {
	*CatchpointReponse
	Data *CatchpointInstantTestResultData `json:"data,omitempty"`
}

type TestSummary struct {
	TestID       int    `json:"test_id"`
	NodeName     string `json:"node_name"`
	Country      string `json:"country"`
	ResponseCode int    `json:"response_code,omitempty"`
	ResponseMs   int    `json:"response_ms,omitempty"`
	PublicLink   string `json:"public_link"`
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

func (c *Catchpoint) CheckError(data []byte, e error) error {
	if e != nil && strings.Contains(e.Error(), "max retries exceeded") {
		return fmt.Errorf("max retries exceeded")
	}

	r := &CatchpointReponse{}
	err := json.Unmarshal(data, &r)
	if err != nil {
		return fmt.Errorf("error parsing response: %s", err)
	}

	if r.Errors != nil && len(*r.Errors) != 0 || !r.Completed {
		return fmt.Errorf("%s", r.Messages)
	}
	return e
}

func (c *Catchpoint) GetNodesFromGroup(options CatchpointNodeGroup) ([]byte, error) {
	return c.CustomGetNodesFromGroup(c.options, options)
}

func (c *Catchpoint) CustomGetNodesFromGroup(catchpointOptions CatchpointOptions, options CatchpointNodeGroup) ([]byte, error) {
	return utils.HttpGetRawRetry(c.client, c.apiURL(catchpointAPINodesGroups+fmt.Sprintf("%d", options.ID)), "application/json", c.getAuth(catchpointOptions), catchpointOptions.Retries, catchpointRetryHeader)
}

func (c *Catchpoint) InstantTest(options CatchpointInstantTestOptions) ([]byte, error) {
	return c.CustomInstantTest(c.options, options)
}

func (c *Catchpoint) InstantTestWithNodeGroup(options CatchpointInstantTestWithNodeGroupOptions) ([]byte, error) {
	return c.CustomInstantTestWithNodeGroup(c.options, options)
}

func (c *Catchpoint) SearchNodesWithOptions(options CatchpointSearchNodesWithOptions) ([]byte, error) {
	return c.CustomSearchNodesWithOptions(c.options, options)
}

func (c *Catchpoint) GetInstantTestResult(testID string, nodeID int) ([]byte, error) {
	return c.CustomGetInstantTestResult(c.options, testID, nodeID)
}

func filterNodesByID(allNodes []*Node, ids []int) []*Node {
	idSet := make(map[int]bool)
	for _, id := range ids {
		idSet[id] = true
	}

	var filtered []*Node
	for _, node := range allNodes {
		if idSet[node.ID] {
			filtered = append(filtered, node)
		}
	}
	return filtered
}

func (c *Catchpoint) WaitPollSuccessOrCancelDetailed(ctx context.Context, pollDelay int, testID int, nodes []*Node, catchpointOptions CatchpointOptions) (bool, []*Node) {

	// the method will return true if all nodes are ready, otherwise false + list of success nodes

	var wg sync.WaitGroup
	results := make(chan struct {
		nodeID  int
		success bool
	}, len(nodes))

	successNodes := []int{}
	var mu sync.Mutex

	strTestId := strconv.Itoa(testID)

	wg.Add(len(nodes))

	for _, node := range nodes {
		go func(ctx context.Context, nodeID int) {
			defer wg.Done()

			ticker := time.NewTicker(time.Duration(pollDelay) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					results <- struct {
						nodeID  int
						success bool
					}{nodeID, false}
					return
				case <-ticker.C:
					d, err := c.CustomGetInstantTestResult(catchpointOptions, strTestId, nodeID)
					err = c.CheckError(d, err)
					if err != nil {
						continue
					}

					r := CatchpointInstantTestResultReponse{}
					err = json.Unmarshal(d, &r)
					if err != nil {
						continue
					}

					if r.Data.InstantTestRecord.TestResult != nil {
						results <- struct {
							nodeID  int
							success bool
						}{nodeID, true}
						return
					}
				}
			}
		}(ctx, node.ID)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		mu.Lock()
		if result.success {
			successNodes = append(successNodes, result.nodeID)
		}
		mu.Unlock()
	}

	allReady := len(successNodes) == len(nodes)

	var successNodesDetails []*Node
	if allReady {
		successNodesDetails = nodes
	} else {
		successNodesDetails = filterNodesByID(nodes, successNodes)
	}

	return allReady, successNodesDetails
}

func (c *Catchpoint) WaitPollSuccessOrCancel(ctx context.Context, pollDelay int, testID int, Nodes []*Node, catchpointOptions CatchpointOptions) bool {

	// the method will return true if all nodes are ready, otherwise false

	var wg sync.WaitGroup
	results := make(chan bool, len(Nodes))
	successCount := 0
	var mu sync.Mutex

	strTestId := strconv.Itoa(testID)

	wg.Add(len(Nodes))

	for _, node := range Nodes {
		go func(ctx context.Context, nodeID int) {
			defer wg.Done()

			ticker := time.NewTicker(time.Duration(pollDelay) * time.Second)
			defer ticker.Stop()

			for {
				select {
				case <-ctx.Done():
					results <- false
					return
				case <-ticker.C:

					d, err := c.CustomGetInstantTestResult(catchpointOptions, strTestId, nodeID)
					err = c.CheckError(d, err)
					if err != nil {
						continue
					}

					r := CatchpointInstantTestResultReponse{}
					err = json.Unmarshal(d, &r)
					if err != nil {
						continue
					}

					if r.Data.InstantTestRecord.TestResult != nil {
						mu.Lock()
						successCount++
						mu.Unlock()
						results <- true
						return
					}
				}
			}
		}(ctx, node.ID)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	for result := range results {
		if !result {
			return false
		}
	}

	mu.Lock()
	allReady := successCount == len(Nodes)
	mu.Unlock()

	return allReady
}

func (c *Catchpoint) GetLogReport(catchpointOptions CatchpointOptions, testID int, nodes []*Node) (*[]CatchpointInstantTestResultReponse, error) {

	var reportOpts []CatchpointInstantTestResultReponse
	strTestId := strconv.Itoa(testID)

	for _, node := range nodes {

		d, err := c.CustomGetInstantTestResult(catchpointOptions, strTestId, node.ID)
		if err != nil {
			return nil, c.CheckError(d, err)
		}
		r := CatchpointInstantTestResultReponse{}
		err = json.Unmarshal(d, &r)
		if err != nil {
			return nil, err
		}
		r.Data.InstantTestRecord.Node.ID = node.ID
		r.Data.InstantTestRecord.Node.Country = node.Country
		r.Data.InstantTestRecord.Node.Name = node.Name
		reportOpts = append(reportOpts, r)

	}
	return &reportOpts, nil
}

func (c *Catchpoint) GenerateSummary(results *[]CatchpointInstantTestResultReponse) ([]TestSummary, error) {
	var summaries []TestSummary

	for _, result := range *results {
		record := result.Data.InstantTestRecord
		if record == nil || record.TestResult == nil || record.TestResult.WebRecords == nil {
			continue
		}

		var responseCode int
		if record.TestResult.WebRecords.Items != nil && len(*record.TestResult.WebRecords.Items) > 0 {
			responseCode = (*record.TestResult.WebRecords.Items)[0].ResponseCode
		}

		var responseMs int
		if record.TestResult.Hosts.Metrics != nil && len(*record.TestResult.Hosts.Metrics) > 0 {
			responseMs = int((*record.TestResult.Hosts.Metrics)[0].Items[2])
		}

		summary := TestSummary{
			TestID:       record.ID,
			NodeName:     record.Node.Name,
			Country:      record.Node.Country.Name,
			ResponseCode: responseCode,
			ResponseMs:   responseMs,
			PublicLink:   record.PublicLink,
		}
		summaries = append(summaries, summary)
	}

	if len(summaries) == 0 {
		return nil, fmt.Errorf("no valid summaries generated")
	}
	return summaries, nil
}

func (c *Catchpoint) CustomGetInstantTestResult(catchpointOptions CatchpointOptions, testID string, nodeID int) ([]byte, error) {

	u, err := url.Parse(catchpointAPIURL + catchpointAPIVersion)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, catchpointAPIInstantTest, testID)

	params := make(url.Values)
	params.Add("nodeId", strconv.Itoa(nodeID))
	u.RawQuery = params.Encode()

	return utils.HttpGetRawRetry(c.client, u.String(), "application/json", c.getAuth(catchpointOptions), catchpointOptions.Retries, catchpointRetryHeader)
}

func (c *Catchpoint) CustomSearchNodesWithOptions(catchpointOptions CatchpointOptions, catchpointNodesGetAllOptions CatchpointSearchNodesWithOptions) ([]byte, error) {

	params := make(url.Values)
	if !utils.IsEmpty(catchpointNodesGetAllOptions.Name) {
		params.Add("name", catchpointNodesGetAllOptions.Name)
	}
	if !utils.IsEmpty(strconv.FormatBool(catchpointNodesGetAllOptions.Targeted)) {
		params.Add("targeted", strconv.FormatBool(catchpointNodesGetAllOptions.Targeted))
	}
	if !utils.IsEmpty(strconv.FormatBool(catchpointNodesGetAllOptions.Active)) {
		params.Add("active", strconv.FormatBool(catchpointNodesGetAllOptions.Active))
	}
	if !utils.IsEmpty(strconv.FormatBool(catchpointNodesGetAllOptions.Paused)) {
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
	if !utils.IsEmpty(strconv.FormatBool(catchpointNodesGetAllOptions.IPv6)) {
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

	return utils.HttpGetRawRetry(c.client, u.String(), "application/json", c.getAuth(catchpointOptions), catchpointOptions.Retries, catchpointRetryHeader)
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
		InstantTestType: &CatchpointInstantTestType{ID: catchpointInstantTestWithNodeGroupOptions.InstantTestType},
		HTTPMethodType:  &CatchpointHTTPMethodType{ID: catchpointInstantTestWithNodeGroupOptions.HTTPMethodType},
		MonitorType:     &CatchpointMonitorType{ID: catchpointInstantTestWithNodeGroupOptions.MonitorType},
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

	return utils.HttpPostRawRetry(c.client, u.String(), "application/json", c.getAuth(catchpointOptions), req, catchpointOptions.Retries, catchpointRetryHeader)
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
		InstantTestType: &CatchpointInstantTestType{ID: catchpointInstantTestOptions.InstantTestType},
		HTTPMethodType:  &CatchpointHTTPMethodType{ID: catchpointInstantTestOptions.HTTPMethodType},
		MonitorType:     &CatchpointMonitorType{ID: catchpointInstantTestOptions.MonitorType},
	}

	req, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}

	return utils.HttpPostRawRetry(c.client, u.String(), "application/json", c.getAuth(catchpointOptions), req, catchpointOptions.Retries, catchpointRetryHeader)
}

func NewCatchpoint(options CatchpointOptions, logger common.Logger) *Catchpoint {

	catchpoint := &Catchpoint{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
	return catchpoint
}
