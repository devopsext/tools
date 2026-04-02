package vendors

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
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

	utils "github.com/devopsext/utils"
)

const (
	iSO8601BasicFormat      = "20060102T150405Z"
	iSO8601BasicFormatShort = "20060102"
	awsSTSURL               = "https://sts.us-east-1.amazonaws.com/"
	awsEC2RegionsURL        = "https://ec2.us-east-1.amazonaws.com/?Action=DescribeRegions&Version=2016-11-15"
	awsRoleRefreshGrace     = 5 * time.Minute
)

var lf = []byte{'\n'}

type AWSOptions struct {
	Accounts        string
	Role            string
	RoleTimeout     int
	RoleSessionName string
	Timeout         int
	Insecure        bool
	AWSKeys
}

type AWSKeys struct {
	AccessKey    string
	SecretKey    string
	SessionToken string
}

// awsBase manages per-account credentials with lazy, auto-refreshing role assumption.
type awsBase struct {
	account    string
	staticKeys AWSKeys // original caller credentials, never rotated
	roleKeys   AWSKeys // temporary assumed-role credentials
	roleExpiry time.Time
	opts       AWSOptions
	client     *http.Client
	mu         sync.Mutex
}

func newAWSBase(account string, opts AWSOptions) *awsBase {
	return &awsBase{
		account:    account,
		staticKeys: AWSKeys{AccessKey: opts.AccessKey, SecretKey: opts.SecretKey},
		opts:       opts,
		client:     utils.NewHttpClient(opts.Timeout, opts.Insecure),
	}
}

// keys returns valid credentials. If a role is configured it assumes/refreshes as needed.
func (b *awsBase) keys() (*AWSKeys, error) {
	if b.opts.Role == "" {
		return &b.staticKeys, nil
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	if !b.roleExpiry.IsZero() && time.Now().Before(b.roleExpiry.Add(-awsRoleRefreshGrace)) {
		return &b.roleKeys, nil
	}
	if err := b.assumeRole(); err != nil {
		return nil, err
	}
	return &b.roleKeys, nil
}

// assumeRole calls STS AssumeRole and stores the resulting temporary credentials.
// Must be called with b.mu held.
func (b *awsBase) assumeRole() error {
	sessionName := b.opts.RoleSessionName
	if sessionName == "" {
		sessionName = "tools_session"
	}
	duration := b.opts.RoleTimeout
	if duration == 0 {
		duration = 3600
	}
	rawURL := fmt.Sprintf(
		"%s?Action=AssumeRole&Version=2011-06-15&RoleSessionName=%s&RoleArn=arn:aws:iam::%s:role/%s&DurationSeconds=%d",
		awsSTSURL, sessionName, b.account, b.opts.Role, duration,
	)
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return err
	}
	resp, err := b.do(req, &b.staticKeys)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	var result awsAssumeRoleResponse
	if err := xml.Unmarshal(body, &result); err != nil {
		return err
	}
	b.roleKeys = AWSKeys{
		AccessKey:    result.AccessKey,
		SecretKey:    result.SecretKey,
		SessionToken: result.SessionToken,
	}
	b.roleExpiry = time.Now().Add(time.Duration(duration) * time.Second)
	return nil
}

// accountID fetches the caller's AWS account ID via STS GetCallerIdentity.
func (b *awsBase) accountID() (string, error) {
	keys, err := b.keys()
	if err != nil {
		return "", err
	}
	req, err := http.NewRequest("GET", awsSTSURL+"?Action=GetCallerIdentity&Version=2011-06-15", nil)
	if err != nil {
		return "", err
	}
	resp, err := b.do(req, keys)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	var result awsAccountMetadata
	if err := xml.Unmarshal(body, &result); err != nil {
		return "", err
	}
	return result.AccountID, nil
}

// regions returns the list of available EC2 regions for this account.
func (b *awsBase) regions() ([]AWSRegion, error) {
	keys, err := b.keys()
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest("GET", awsEC2RegionsURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := b.do(req, keys)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result awsDescribeRegionsResponse
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	return result.RegionInfo.Items, nil
}

// do signs and executes an HTTP request with the given credentials.
func (b *awsBase) do(req *http.Request, keys *AWSKeys) (*http.Response, error) {
	if keys.SessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", keys.SessionToken)
	}
	if err := awsSign(keys, req); err != nil {
		return nil, err
	}
	return b.client.Do(req)
}

// ---- EC2 ---------------------------------------------------------------

type AWSEC2 struct {
	bases []*awsBase
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

func NewAWSEC2(opts AWSOptions) (*AWSEC2, error) {
	if opts.AccessKey == "" || opts.SecretKey == "" {
		return nil, fmt.Errorf("AWS access key and secret key are required")
	}
	accounts := strings.Split(opts.Accounts, ",")
	bases := make([]*awsBase, 0, len(accounts))
	for _, account := range accounts {
		bases = append(bases, newAWSBase(strings.TrimSpace(account), opts))
	}
	return &AWSEC2{bases: bases}, nil
}

func (e *AWSEC2) GetAllAWSEC2Instances() ([]AWSEC2Instance, error) {
	var (
		mu        sync.Mutex
		wg        sync.WaitGroup
		instances []AWSEC2Instance
		errs      []error
	)

	for _, base := range e.bases {
		regions, err := base.regions()
		if err != nil {
			return nil, fmt.Errorf("account %s: failed to list regions: %w", base.account, err)
		}
		accountID, err := base.accountID()
		if err != nil {
			return nil, fmt.Errorf("account %s: failed to get account ID: %w", base.account, err)
		}
		for _, region := range regions {
			wg.Add(1)
			go func(b *awsBase, region AWSRegion, accountID string) {
				defer wg.Done()
				got, err := fetchEC2Instances(b, region, accountID)
				mu.Lock()
				defer mu.Unlock()
				if err != nil {
					errs = append(errs, err)
					return
				}
				instances = append(instances, got...)
			}(base, region, accountID)
		}
	}
	wg.Wait()

	if len(errs) > 0 {
		return instances, fmt.Errorf("errors retrieving EC2 instances: %v", errs)
	}
	return instances, nil
}

func fetchEC2Instances(b *awsBase, region AWSRegion, accountID string) ([]AWSEC2Instance, error) {
	keys, err := b.keys()
	if err != nil {
		return nil, err
	}
	rawURL := "https://" + region.RegionEndpoint + "/?Action=DescribeInstances&Version=2016-11-15"
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := b.do(req, keys)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	var result awsEC2describeInstancesResponse
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, err
	}
	out := make([]AWSEC2Instance, 0, len(result.InstanceItems))
	for _, inst := range result.InstanceItems {
		host := inst.InstanceId
		for _, tag := range inst.Tags {
			if tag.Key == "Name" {
				host = strings.ReplaceAll(strings.TrimSpace(tag.Value), " ", "_")
			}
		}
		ip := inst.IpAddress
		if ip == "" {
			ip = inst.PrivateIpAddress
		}
		out = append(out, AWSEC2Instance{
			AccountID: accountID,
			Host:      host,
			IP:        ip,
			Region:    region.RegionName,
			OS:        inst.Platform,
			Vendor:    "aws",
			Server:    inst.InstanceId,
			Cluster:   "aws",
		})
	}
	return out, nil
}

// ---- S3 ----------------------------------------------------------------

type AWSS3 struct {
	base *awsBase
}

func NewAWSS3(opts AWSOptions) (*AWSS3, error) {
	if opts.AccessKey == "" || opts.SecretKey == "" {
		return nil, fmt.Errorf("AWS access key and secret key are required")
	}
	account := strings.TrimSpace(strings.SplitN(opts.Accounts, ",", 2)[0])
	return &AWSS3{base: newAWSBase(account, opts)}, nil
}

func (s *AWSS3) ListObjects(region, bucket, prefix string) ([]byte, error) {
	keys, err := s.base.keys()
	if err != nil {
		return nil, err
	}
	rawURL := fmt.Sprintf("https://s3.%s.amazonaws.com/%s?list-type=2", region, bucket)
	if prefix != "" {
		rawURL += "&prefix=" + url.QueryEscape(prefix)
	}
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-amz-content-sha256", fmt.Sprintf("%x", sha256.Sum256(nil)))
	resp, err := s.base.do(req, keys)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("S3 ListObjects: status %d: %s", resp.StatusCode, string(body))
	}

	type s3Object struct {
		Key          string `xml:"Key"`
		Size         int64  `xml:"Size"`
		LastModified string `xml:"LastModified"`
	}
	type listResult struct {
		Contents []s3Object `xml:"Contents"`
	}
	var result listResult
	if err := xml.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("S3 ListObjects: failed to parse response: %w", err)
	}
	return json.Marshal(result.Contents)
}

// GetObject downloads the object at s3://{bucket}/{key} and returns its body.
func (s *AWSS3) GetObject(region, bucket, key string) ([]byte, error) {
	keys, err := s.base.keys()
	if err != nil {
		return nil, err
	}
	rawURL := fmt.Sprintf("https://s3.%s.amazonaws.com/%s/%s", region, bucket, key)
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("x-amz-content-sha256", fmt.Sprintf("%x", sha256.Sum256(nil)))
	resp, err := s.base.do(req, keys)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("S3 GetObject: status %d: %s", resp.StatusCode, string(body))
	}
	return body, nil
}

// PutObject uploads body to s3://{bucket}/{key} in the given region.
func (s *AWSS3) PutObject(region, bucket, key, contentType string, body []byte) ([]byte, error) {
	keys, err := s.base.keys()
	if err != nil {
		return nil, err
	}
	rawURL := fmt.Sprintf("https://s3.%s.amazonaws.com/%s/%s", region, bucket, key)
	req, err := http.NewRequest("PUT", rawURL, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	if contentType != "" {
		req.Header.Set("Content-Type", contentType)
	}
	h := sha256.Sum256(body)
	req.Header.Set("x-amz-content-sha256", fmt.Sprintf("%x", h))
	resp, err := s.base.do(req, keys)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("S3 PutObject: status %d: %s", resp.StatusCode, string(respBody))
	}
	return respBody, nil
}

// ---- XML response types ------------------------------------------------

type AWSRegion struct {
	RegionName     string `xml:"regionName"`
	RegionEndpoint string `xml:"regionEndpoint"`
}

type awsAssumeRoleResponse struct {
	XMLName      xml.Name `xml:"https://sts.amazonaws.com/doc/2011-06-15/ AssumeRoleResponse"`
	AccessKey    string   `xml:"AssumeRoleResult>Credentials>AccessKeyId"`
	SecretKey    string   `xml:"AssumeRoleResult>Credentials>SecretAccessKey"`
	SessionToken string   `xml:"AssumeRoleResult>Credentials>SessionToken"`
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
	InstanceId       string `xml:"instanceId"`
	KeyName          string `xml:"keyName"`
	PrivateIpAddress string `xml:"privateIpAddress"`
	IpAddress        string `xml:"ipAddress"`
	Platform         string `xml:"platformDetails"`
	Tags             []struct {
		Key   string `xml:"key"`
		Value string `xml:"value"`
	} `xml:"tagSet>item"`
}

type awsAccountMetadata struct {
	XMLName   xml.Name `xml:"https://sts.amazonaws.com/doc/2011-06-15/ GetCallerIdentityResponse"`
	AccountID string   `xml:"GetCallerIdentityResult>Account"`
}

// ---- AWS Signature V4 --------------------------------------------------

type awsService struct {
	Name   string
	Region string
}

// awsSign derives the service/region from the request host and applies AWS Signature V4.
func awsSign(keys *AWSKeys, r *http.Request) error {
	svc, err := awsServiceFromHost(r.Host)
	if err != nil {
		return err
	}
	return svc.sign(keys, r)
}

// awsServiceFromHost parses the AWS service name and region from a hostname.
// It handles both path-style (s3.region.amazonaws.com) and
// virtual-hosted-style (bucket.s3.region.amazonaws.com) by scanning for a
// known service token rather than assuming a fixed position.
func awsServiceFromHost(host string) (*awsService, error) {
	parts := strings.Split(host, ".")
	knownServices := map[string]bool{"s3": true, "ec2": true, "sts": true, "iam": true}
	for i, p := range parts {
		if knownServices[p] && i+1 < len(parts) {
			return &awsService{Name: p, Region: parts[i+1]}, nil
		}
	}
	if len(parts) < 4 {
		return nil, fmt.Errorf("cannot parse AWS service from host: %s", host)
	}
	return &awsService{Name: parts[0], Region: parts[1]}, nil
}

func (s *awsService) sign(keys *AWSKeys, r *http.Request) error {
	t := time.Now().UTC()
	if date := r.Header.Get("Date"); date != "" {
		var err error
		t, err = time.Parse(http.TimeFormat, date)
		if err != nil {
			return err
		}
	}
	r.Header.Set("Date", t.Format(iSO8601BasicFormat))

	k := awsDeriveKey(keys.SecretKey, s, t)
	h := hmac.New(sha256.New, k)
	s.writeStringToSign(h, t, r)

	var auth bytes.Buffer
	auth.WriteString("AWS4-HMAC-SHA256 ")
	fmt.Fprintf(&auth, "Credential=%s/%s, ", keys.AccessKey, s.creds(t))
	auth.WriteString("SignedHeaders=")
	s.writeHeaderList(&auth, r)
	fmt.Fprintf(&auth, ", Signature=%x", h.Sum(nil))

	r.Header.Set("Authorization", auth.String())
	return nil
}

func (s *awsService) creds(t time.Time) string {
	return t.Format(iSO8601BasicFormatShort) + "/" + s.Region + "/" + s.Name + "/aws4_request"
}

func (s *awsService) writeStringToSign(w io.Writer, t time.Time, r *http.Request) {
	fmt.Fprintf(w, "AWS4-HMAC-SHA256\n%s\n%s\n", t.Format(iSO8601BasicFormat), s.creds(t))
	h := sha256.New()
	s.writeRequest(h, r)
	fmt.Fprintf(w, "%x", h.Sum(nil))
}

func (s *awsService) writeRequest(w io.Writer, r *http.Request) {
	r.Header.Set("host", r.Host)
	fmt.Fprintf(w, "%s\n", r.Method)
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

func (s *awsService) writeURI(w io.Writer, r *http.Request) {
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

func (s *awsService) writeQuery(w io.Writer, r *http.Request) {
	var a []string
	for k, vs := range r.URL.Query() {
		k = url.QueryEscape(k)
		for _, v := range vs {
			if v == "" {
				a = append(a, k)
			} else {
				a = append(a, k+"="+url.QueryEscape(v))
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

func (s *awsService) writeHeader(w io.Writer, r *http.Request) {
	a := make([]string, 0, len(r.Header))
	for k, v := range r.Header {
		sort.Strings(v)
		a = append(a, strings.ToLower(k)+":"+strings.Join(v, ","))
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			w.Write(lf)
		}
		io.WriteString(w, s)
	}
}

func (s *awsService) writeHeaderList(w io.Writer, r *http.Request) {
	a := make([]string, 0, len(r.Header))
	for k := range r.Header {
		a = append(a, strings.ToLower(k))
	}
	sort.Strings(a)
	for i, s := range a {
		if i > 0 {
			w.Write([]byte{';'})
		}
		w.Write([]byte(s))
	}
}

func (s *awsService) writeBody(w io.Writer, r *http.Request) {
	var b []byte
	if r.Body != nil {
		var err error
		b, err = io.ReadAll(r.Body)
		if err != nil {
			b = []byte{}
		} else {
			r.Body = io.NopCloser(bytes.NewBuffer(b))
		}
	}
	h := sha256.New()
	h.Write(b)
	fmt.Fprintf(w, "%x", h.Sum(nil))
}

func awsDeriveKey(secret string, s *awsService, t time.Time) []byte {
	h := ghmac([]byte("AWS4"+secret), []byte(t.Format(iSO8601BasicFormatShort)))
	h = ghmac(h, []byte(s.Region))
	h = ghmac(h, []byte(s.Name))
	return ghmac(h, []byte("aws4_request"))
}

func ghmac(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}
