package vendors

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	iSO8601BasicFormat          = "20060102T150405Z"
	iSO8601BasicFormatShort     = "20060102"
	defaultAWSEC2RegionsURL     = "https://ec2.us-east-1.amazonaws.com/?Action=DescribeRegions&Version=2016-11-15"
	defaultAWSAccountDetailsURL = "https://sts.us-east-1.amazonaws.com/?Action=GetCallerIdentity&Version=2011-06-15"
)

var lf = []byte{'\n'}

type AWSKeys struct {
	AccessKey string
	SecretKey string
}

type AWSClient struct {
	AccountID  string
	Region     string
	Keys       *AWSKeys
	HttpClient *http.Client
	Url        string
}

type AWSService struct {
	Name   string
	Region string
}

type AWSRegion struct {
	RegionName     string `xml:"regionName"`
	RegionEndpoint string `xml:"regionEndpoint"`
}

type awsDescribeRegionsResponse struct {
	XMLName    xml.Name `xml:"http://ec2.amazonaws.com/doc/2016-11-15/ DescribeRegionsResponse"`
	RequestId  string   `xml:"requestId"`
	RegionInfo struct {
		Items []AWSRegion `xml:"item"`
	} `xml:"regionInfo"`
}

type awsEC2describeInstancesResponse struct {
	XMLName       xml.Name                 `xml:"http://ec2.amazonaws.com/doc/2016-11-15/ DescribeInstancesResponse"`
	InstanceItems []awsEC2InstanceMetadata `xml:"reservationSet>item>instancesSet>item"`
}

type awsEC2InstanceMetadata struct {
	InstanceId string `xml:"instanceId"`
	KeyName    string `xml:"keyName"`
	IpAddress  string `xml:"ipAddress"`
	Platform   string `xml:"platformDetails"`
	Tags       []struct {
		Key   string `xml:"key"`
		Value string `xml:"value"`
	} `xml:"tagSet>item"`
}

type awsAccountMetadata struct {
	XMLName   xml.Name `xml:"https://sts.amazonaws.com/doc/2011-06-15/ GetCallerIdentityResponse"`
	AccountID string   `xml:"GetCallerIdentityResult>Account"`
}

type awsBase struct {
	clients []*AWSClient // Per AWS design, one client is needed per region
	keys    AWSKeys
}

type AWSEC2 struct {
	awsBase
}

type AWSEC2Instance struct {
	AccountID string
	Host      string
	IP        string
	Region    string
	OS        string
	Server    string
	Vendor    string
	Cluster   string
}

// NewAWSEC2 creates a new instance of AWSEC2 with the given AWS keys.
// It retrieves available regions and generates clients for each region.
func NewAWSEC2(keys AWSKeys) (*AWSEC2, error) {
	if keys.AccessKey == "" || keys.SecretKey == "" {
		return nil, fmt.Errorf("cannot create ec2 object, aws keys not present")
	}
	ec2 := AWSEC2{
		awsBase{
			keys: keys,
		},
	}
	regions, err := ec2.getAvailableAWSRegions()
	if err != nil {
		return nil, err
	}
	err = ec2.generateAWSClients(regions, keys)
	if err != nil {
		return nil, err
	}

	return &ec2, nil
}

// getAvailableRegions retrieves available AWS regions using the given AWS keys.
func (e *awsBase) getAvailableAWSRegions() ([]AWSRegion, error) {
	client := &AWSClient{
		Keys:       &e.keys,
		HttpClient: http.DefaultClient,
		Url:        defaultAWSEC2RegionsURL,
	}
	r, err := http.NewRequest("GET", client.Url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := client.Do(r)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	response := awsDescribeRegionsResponse{}
	err = xml.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response.RegionInfo.Items, nil
}

// generateClients generates AWS clients for the given regions and AWS keys.
func (e *awsBase) generateAWSClients(regions []AWSRegion, keys AWSKeys) error {
	clients := make([]*AWSClient, 0)

	for _, region := range regions {
		client := AWSClient{
			Region:     region.RegionName,
			Keys:       &keys,
			HttpClient: http.DefaultClient,
			Url:        "https://" + region.RegionEndpoint + "/?Action=DescribeInstances&Version=2016-11-15",
		}
		client.getAWSAccountID()
		clients = append(clients, &client)
	}
	e.clients = clients
	return nil
}

// GetAllEC2Instances retrieves all EC2 instances associated with the AWSEC2 service.
func (e *AWSEC2) GetAllAWSEC2Instances() ([]AWSEC2Instance, error) {
	var wg sync.WaitGroup
	var errs []error
	instances := make([]AWSEC2Instance, 0)

	for _, c := range e.clients {
		wg.Add(1)
		go func(c *AWSClient) {
			defer wg.Done()
			r, err := http.NewRequest("GET", c.Url, nil)
			if err != nil {
				errs = append(errs, err)
				return
			}

			resp, err := c.Do(r)
			if err != nil {
				errs = append(errs, err)
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				errs = append(errs, err)
				return
			}

			response := awsEC2describeInstancesResponse{}
			err = xml.Unmarshal(body, &response)
			if err != nil {
				errs = append(errs, err)
				return
			}
			for _, instance := range response.InstanceItems {
				host := instance.InstanceId
				for _, tag := range instance.Tags {
					if tag.Key == "Name" {
						host = strings.ReplaceAll(strings.TrimSpace(tag.Value), " ", "_")
					}
				}

				instances = append(instances, AWSEC2Instance{
					AccountID: c.AccountID,
					Host:      host,
					IP:        instance.IpAddress,
					Region:    c.Region,
					OS:        instance.Platform,
					Vendor:    "aws",
					Server:    instance.InstanceId,
					Cluster:   "aws",
				})
			}
		}(c)
	}
	wg.Wait()

	if len(errs) > 0 {
		return instances, fmt.Errorf("encountered errors while retrieving EC2 instances: %v", errs)
	}

	return instances, nil
}

// getAWSAccountID retrieves the AWS account ID associated with the AWS client.
func (c *AWSClient) getAWSAccountID() error {
	r, err := http.NewRequest("GET", defaultAWSAccountDetailsURL, nil)
	if err != nil {
		return err
	}

	resp, err := c.Do(r)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	response := awsAccountMetadata{}
	err = xml.Unmarshal(body, &response)
	if err != nil {
		return err
	}

	c.AccountID = response.AccountID
	return nil
}

// Do executes the HTTP request with AWS authentication using the standard AWS4 format.
func (c *AWSClient) Do(req *http.Request) (*http.Response, error) {
	c.awsSignRequest(c.Keys, req)
	return c.HttpClient.Do(req)
}

// Sign signs the HTTP request with AWS authentication.
func (c *AWSClient) awsSignRequest(keys *AWSKeys, r *http.Request) error {
	parts := strings.Split(r.Host, ".")
	if len(parts) < 4 {
		return fmt.Errorf("invalid AWS Endpoint: %s", r.Host)
	}
	sv := new(AWSService)
	sv.Name = parts[0]
	sv.Region = parts[1]
	sv.awsSignService(keys, r)
	return nil
}

// sign signs the HTTP request with AWS authentication for a specific service.
func (s *AWSService) awsSignService(keys *AWSKeys, r *http.Request) error {
	date := r.Header.Get("Date")
	t := time.Now().UTC()
	if date != "" {
		var err error
		t, err = time.Parse(http.TimeFormat, date)
		if err != nil {
			return err
		}
	}
	r.Header.Set("Date", t.Format(iSO8601BasicFormat))

	k := keys.sign(s, t)
	h := hmac.New(sha256.New, k)
	s.writeStringToSign(h, t, r)

	auth := bytes.NewBufferString("AWS4-HMAC-SHA256 ")
	auth.Write([]byte("Credential=" + keys.AccessKey + "/" + s.creds(t)))
	auth.Write([]byte{',', ' '})
	auth.Write([]byte("SignedHeaders="))
	s.writeHeaderList(auth, r)
	auth.Write([]byte{',', ' '})
	auth.Write([]byte("Signature=" + fmt.Sprintf("%x", h.Sum(nil))))

	r.Header.Set("Authorization", auth.String())

	return nil
}

func (s *AWSService) creds(t time.Time) string {
	return t.Format(iSO8601BasicFormatShort) + "/" + s.Region + "/" + s.Name + "/aws4_request"
}

func (s *AWSService) writeQuery(w io.Writer, r *http.Request) {
	var a []string
	for k, vs := range r.URL.Query() {
		k = url.QueryEscape(k)
		for _, v := range vs {
			if v == "" {
				a = append(a, k)
			} else {
				v = url.QueryEscape(v)
				a = append(a, k+"="+v)
			}
		}
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			w.Write([]byte{'&'})
		}
		w.Write([]byte(s))
	}
}

func (s *AWSService) writeHeader(w io.Writer, r *http.Request) {
	i, a := 0, make([]string, len(r.Header))
	for k, v := range r.Header {
		sort.Strings(v)
		a[i] = strings.ToLower(k) + ":" + strings.Join(v, ",")
		i++
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			w.Write(lf)
		}
		io.WriteString(w, s)
	}
}

func (s *AWSService) writeHeaderList(w io.Writer, r *http.Request) {
	i, a := 0, make([]string, len(r.Header))
	for k := range r.Header {
		a[i] = strings.ToLower(k)
		i++
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			w.Write([]byte{';'})
		}
		w.Write([]byte(s))
	}
}

func (s *AWSService) writeBody(w io.Writer, r *http.Request) {
	var b []byte
	if r.Body == nil {
		b = []byte("")
	} else {
		var err error
		b, err = io.ReadAll(r.Body)
		if err != nil {
			panic(err)
		}
		r.Body = io.NopCloser(bytes.NewBuffer(b))
	}

	h := sha256.New()
	h.Write(b)
	fmt.Fprintf(w, "%x", h.Sum(nil))
}

func (s *AWSService) writeURI(w io.Writer, r *http.Request) {
	path := r.URL.RequestURI()
	if r.URL.RawQuery != "" {
		path = path[:len(path)-len(r.URL.RawQuery)-1]
	}
	slash := strings.HasSuffix(path, "/")
	path = filepath.Clean(path)
	if path != "/" && slash {
		path += "/"
	}
	w.Write([]byte(path))
}

func (s *AWSService) writeRequest(w io.Writer, r *http.Request) {
	r.Header.Set("host", r.Host)

	w.Write([]byte(r.Method))
	w.Write(lf)
	s.writeURI(w, r)
	w.Write(lf)
	s.writeQuery(w, r)
	w.Write(lf)
	s.writeHeader(w, r)
	w.Write(lf)
	w.Write(lf)
	s.writeHeaderList(w, r)
	w.Write(lf)
	s.writeBody(w, r)
}

func (s *AWSService) writeStringToSign(w io.Writer, t time.Time, r *http.Request) {
	w.Write([]byte("AWS4-HMAC-SHA256"))
	w.Write(lf)
	w.Write([]byte(t.Format(iSO8601BasicFormat)))
	w.Write(lf)

	w.Write([]byte(s.creds(t)))
	w.Write(lf)

	h := sha256.New()
	s.writeRequest(h, r)
	fmt.Fprintf(w, "%x", h.Sum(nil))
}

func (k *AWSKeys) sign(s *AWSService, t time.Time) []byte {
	h := ghmac([]byte("AWS4"+k.SecretKey), []byte(t.Format(iSO8601BasicFormatShort)))
	h = ghmac(h, []byte(s.Region))
	h = ghmac(h, []byte(s.Name))
	h = ghmac(h, []byte("aws4_request"))
	return h
}

func ghmac(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
