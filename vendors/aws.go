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
	iSO8601BasicFormat      = "20060102T150405Z"
	iSO8601BasicFormatShort = "20060102"
	defaultQueryURL         = "https://ec2.us-east-1.amazonaws.com/?Action=DescribeRegions&Version=2016-11-15"
)

var lf = []byte{'\n'}

type AwsKeys struct {
	AccessKey string
	SecretKey string
}

type Client struct {
	Region string
	Keys   *AwsKeys
	Client *http.Client
	Url    string
}

type Service struct {
	Name   string
	Region string
}

type DescribeRegionsResponse struct {
	XMLName    xml.Name `xml:"http://ec2.amazonaws.com/doc/2016-11-15/ DescribeRegionsResponse"`
	RequestId  string   `xml:"requestId"`
	RegionInfo struct {
		Items []Region `xml:"item"`
	} `xml:"regionInfo"`
}

type Region struct {
	RegionName     string `xml:"regionName"`
	RegionEndpoint string `xml:"regionEndpoint"`
}

type DescribeInstancesResponse struct {
	XMLName       xml.Name           `xml:"http://ec2.amazonaws.com/doc/2016-11-15/ DescribeInstancesResponse"`
	InstanceItems []InstanceMetadata `xml:"reservationSet>item>instancesSet>item"`
}

type InstanceMetadata struct {
	InstanceId string `xml:"instanceId"`
	KeyName    string `xml:"keyName"`
	IpAddress  string `xml:"ipAddress"`
	Platform   string `xml:"platformDetails"`
	Tags       []struct {
		Key   string `xml:"key"`
		Value string `xml:"value"`
	} `xml:"tagSet>item"`
}

type EC2 struct {
	clients []*Client // Per AWS design, one client is needed per region
	keys    AwsKeys
}

type EC2Instance struct {
	Host    string
	IP      string
	Region  string
	OS      string
	Server  string
	Vendor  string
	Cluster string
}

func NewEC2(keys AwsKeys) (*EC2, error) {
	if keys.AccessKey == "" || keys.SecretKey == "" {
		return nil, fmt.Errorf("cannot create ec2 object, aws keys not present")
	}
	regions, err := getAvailableRegions(keys)
	if err != nil {
		return nil, err
	}
	clients := generateClients(regions, keys)

	return &EC2{
		clients: clients,
		keys:    keys,
	}, nil
}

func getAvailableRegions(keys AwsKeys) ([]Region, error) {
	client := &Client{
		Keys:   &keys,
		Client: http.DefaultClient,
		Url:    defaultQueryURL,
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
	response := DescribeRegionsResponse{}
	err = xml.Unmarshal(body, &response)
	if err != nil {
		return nil, err
	}

	return response.RegionInfo.Items, nil
}

func generateClients(regions []Region, keys AwsKeys) []*Client {
	clients := make([]*Client, 0)
	for _, region := range regions {
		clients = append(clients, &Client{
			Region: region.RegionName,
			Keys:   &keys,
			Client: http.DefaultClient,
			Url:    "https://" + region.RegionEndpoint + "/?Action=DescribeInstances&Version=2016-11-15",
		})
	}
	return clients
}

func (e *EC2) GetAllEC2Instances() ([]EC2Instance, error) {
	var wg sync.WaitGroup
	var reqErr error
	instances := make([]EC2Instance, 0)

	for _, c := range e.clients {
		wg.Add(1)
		go func(c *Client) { // Process web requests concurrently, it's much faster
			defer wg.Done()
			r, err := http.NewRequest("GET", c.Url, nil)
			if err != nil {
				reqErr = err
				return
			}

			resp, err := c.Do(r)
			if err != nil {
				reqErr = err
				return
			}
			defer resp.Body.Close()

			body, err := io.ReadAll(resp.Body)
			if err != nil {
				reqErr = err
				return
			}

			response := DescribeInstancesResponse{}
			err = xml.Unmarshal(body, &response)
			if err != nil {
				reqErr = err
				return
			}
			for _, instance := range response.InstanceItems {
				host := instance.InstanceId
				for _, tag := range instance.Tags {
					if tag.Key == "Name" {
						host = strings.ReplaceAll(strings.TrimSpace(tag.Value), " ", "_")
					}
				}

				instances = append(instances, EC2Instance{
					Host:    host,
					IP:      instance.IpAddress,
					Region:  c.Region,
					OS:      instance.Platform,
					Vendor:  "aws",
					Server:  instance.InstanceId,
					Cluster: "aws",
				})
			}
		}(c)
	}
	wg.Wait()

	if reqErr != nil {
		return nil, reqErr
	}

	return instances, nil
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	Sign(c.Keys, req)
	return c.Client.Do(req)
}

func Sign(keys *AwsKeys, r *http.Request) error {
	parts := strings.Split(r.Host, ".")
	if len(parts) < 4 {
		return fmt.Errorf("invalid AWS Endpoint: %s", r.Host)
	}
	sv := new(Service)
	sv.Name = parts[0]
	sv.Region = parts[1]
	sv.Sign(keys, r)
	return nil
}

func (s *Service) Sign(keys *AwsKeys, r *http.Request) error {
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

func (s *Service) creds(t time.Time) string {
	return t.Format(iSO8601BasicFormatShort) + "/" + s.Region + "/" + s.Name + "/aws4_request"
}

func (s *Service) writeQuery(w io.Writer, r *http.Request) {
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

func (s *Service) writeHeader(w io.Writer, r *http.Request) {
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

func (s *Service) writeHeaderList(w io.Writer, r *http.Request) {
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

func (s *Service) writeBody(w io.Writer, r *http.Request) {
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

func (s *Service) writeURI(w io.Writer, r *http.Request) {
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

func (s *Service) writeRequest(w io.Writer, r *http.Request) {
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

func (s *Service) writeStringToSign(w io.Writer, t time.Time, r *http.Request) {
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

func (k *AwsKeys) sign(s *Service, t time.Time) []byte {
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
