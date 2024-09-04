package vendors

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/devopsext/utils"
)

type VCenterHostOptions struct {
	Cluster string
}

type VCenterVMGuestIdentityOptions struct {
	VM string
}

type VCenterVMOptions struct {
	Cluster string
	Host    string
}

type VCenterOptions struct {
	Timeout  int
	Insecure bool
	URL      string
	User     string
	Password string
	Session  string
}

type VCenterVMNameOptions struct {
	Names []string
}

type VMInfo struct {
	MemorySizeMiB int    `json:"memory_size_MiB"`
	VM            string `json:"vm"`
	Name          string `json:"name"`
	PowerState    string `json:"power_state"`
	CPUCount      int    `json:"cpu_count"`
}
type VMsResponse struct {
	Value []VMInfo `json:"value"`
}
type VCenter struct {
	client  *http.Client
	options VCenterOptions
}

type VCenterSessionResponse struct {
	Value string `json:"value"`
}

const (
	VCenterContentType                = "application/json"
	VCenterRestSessionPath            = "/rest/com/vmware/cis/session"
	VCenterRestClusterPath            = "/rest/vcenter/cluster"
	VCenterRestHostPath               = "/rest/vcenter/host"
	VCenterRestVMPath                 = "/rest/vcenter/vm"
	VCenterRestVMGuestIdentityPathFmt = "/rest/vcenter/vm/%s/guest/identity"
	VCenterRestVMPowerPathFmt         = "/rest/vcenter/vm/%s/power"
)

func (vc *VCenter) getAuth(opts VCenterOptions) string {
	auth := ""
	if !utils.IsEmpty(opts.User) {
		userPass := fmt.Sprintf("%s:%s", opts.User, opts.Password)
		auth = fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(userPass)))
		return auth
	}
	return auth
}

func (vc *VCenter) getSession(opts VCenterOptions) (string, error) {
	u, err := url.Parse(opts.URL)
	if err != nil {
		return "", err
	}
	u.Path = path.Join(u.Path, VCenterRestSessionPath)

	res, err := utils.HttpPostRaw(vc.client, u.String(), VCenterContentType, vc.getAuth(opts), nil)
	if err != nil {
		return "", err
	}

	var sr VCenterSessionResponse
	err = json.Unmarshal(res, &sr)
	if err != nil {
		return "", err
	}

	return sr.Value, nil
}

func (vc *VCenter) getHeaders(session string) map[string]string {
	headers := make(map[string]string)
	headers["Content-type"] = VCenterContentType
	headers["vmware-api-session-id"] = session
	return headers
}

func (vc *VCenter) CustomGetSession(options VCenterOptions) (string, error) {
	if utils.IsEmpty(options.Session) {
		s, err := vc.getSession(options)
		if err != nil {
			return "", err
		}
		return s, nil
	}
	return options.Session, nil
}

func (vc *VCenter) CustomGetClusters(options VCenterOptions) ([]byte, error) {
	session, err := vc.CustomGetSession(options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(options.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, VCenterRestClusterPath)
	return utils.HttpGetRawWithHeaders(vc.client, u.String(), vc.getHeaders(session))
}

func (vc *VCenter) GetClusters() ([]byte, error) {
	return vc.CustomGetClusters(vc.options)
}

func (vc *VCenter) CustomGetHosts(options VCenterOptions, hostOptions VCenterHostOptions) ([]byte, error) {
	session, err := vc.CustomGetSession(options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(options.URL)
	if err != nil {
		return nil, err
	}

	if !utils.IsEmpty(hostOptions.Cluster) {
		var params = make(url.Values)
		params.Add("filter.clusters", hostOptions.Cluster)
		u.RawQuery = params.Encode()
	}

	u.Path = path.Join(u.Path, VCenterRestHostPath)

	return utils.HttpGetRawWithHeaders(vc.client, u.String(), vc.getHeaders(session))
}

func (vc *VCenter) GetHosts(options VCenterHostOptions) ([]byte, error) {
	return vc.CustomGetHosts(vc.options, options)
}

func (vc *VCenter) CustomGetVMs(options VCenterOptions, vmOptions VCenterVMOptions) ([]byte, error) {
	session, err := vc.CustomGetSession(options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(options.URL)
	if err != nil {
		return nil, err
	}

	if !utils.IsEmpty(vmOptions.Cluster) {
		var params = make(url.Values)
		params.Add("filter.clusters", vmOptions.Cluster)
		params.Add("filter.hosts", vmOptions.Host)
		u.RawQuery = params.Encode()
	}

	u.Path = path.Join(u.Path, VCenterRestVMPath)

	return utils.HttpGetRawWithHeaders(vc.client, u.String(), vc.getHeaders(session))
}

func (vc *VCenter) GetVMs(options VCenterVMOptions) ([]byte, error) {
	return vc.CustomGetVMs(vc.options, options)
}

func (vc *VCenter) CustomGetVMGuestIdentity(options VCenterOptions, vmGuestidentity VCenterVMGuestIdentityOptions) ([]byte, error) {
	session, err := vc.CustomGetSession(options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(options.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf(VCenterRestVMGuestIdentityPathFmt, vmGuestidentity.VM))

	return utils.HttpGetRawWithHeaders(vc.client, u.String(), vc.getHeaders(session))
}

func (vc *VCenter) GetVMGuestIdentity(options VCenterVMGuestIdentityOptions) ([]byte, error) {
	return vc.CustomGetVMGuestIdentity(vc.options, options)
}

func (vc *VCenter) GetVMsByName(options VCenterVMNameOptions) ([]byte, error) {

	return vc.CustomGetVMsByName(vc.options, options)
}

func (vc *VCenter) CustomGetVMsByName(options VCenterOptions, vmNameOptions VCenterVMNameOptions) ([]byte, error) {
	session, err := vc.CustomGetSession(options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(options.URL)
	if err != nil {
		return nil, err
	}

	if len(vmNameOptions.Names) > 0 {
		var params = make(url.Values)
		for _, name := range vmNameOptions.Names {
			params.Add("filter.names", name)
		}
		u.RawQuery = params.Encode()
	}

	u.Path = path.Join(u.Path, VCenterRestVMPath)

	return utils.HttpGetRawWithHeaders(vc.client, u.String(), vc.getHeaders(session))
}

func (vc *VCenter) CustomControlVMPower(options VCenterOptions, vmID string, action string) ([]byte, error) {
	session, err := vc.CustomGetSession(options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(options.URL)
	if err != nil {
		return nil, err
	}

	p := fmt.Sprintf(VCenterRestVMPowerPathFmt, vmID)
	u.Path = path.Join(u.Path, fmt.Sprintf("%s/%s", p, action))

	return utils.HttpPostRawWithHeaders(vc.client, u.String(), vc.getHeaders(session), nil)
}

func (vc *VCenter) CustomGetVM(options VCenterOptions, vmID string) ([]byte, error) {
	session, err := vc.CustomGetSession(options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(options.URL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("%s/%s", VCenterRestVMPath, vmID))

	return utils.HttpGetRawWithHeaders(vc.client, u.String(), vc.getHeaders(session))
}

func (vc *VCenter) StartVM(vmID string) ([]byte, error) {
	return vc.CustomControlVMPower(vc.options, vmID, "start")
}

func (vc *VCenter) StopVM(vmID string) ([]byte, error) {
	return vc.CustomControlVMPower(vc.options, vmID, "stop")
}

func (vc *VCenter) RestartVM(vmID string) ([]byte, error) {
	return vc.CustomControlVMPower(vc.options, vmID, "reset")
}

func (vc *VCenter) GetVM(vmID string) ([]byte, error) {
	return vc.CustomGetVM(vc.options, vmID)
}

func NewVCenter(options VCenterOptions) *VCenter {

	return &VCenter{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}

func InitializeVCenterSession(options VCenterOptions) (VCenterOptions, error) {
	client := utils.NewHttpClient(options.Timeout, options.Insecure)

	tempVC := &VCenter{
		client:  client,
		options: options,
	}

	session, err := tempVC.CustomGetSession(options)
	if err != nil {

		return options, err
	}

	options.Session = session
	return options, nil
}
