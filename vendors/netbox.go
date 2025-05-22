package vendors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path"
	"sync"
	"time"

	"github.com/devopsext/utils"
)

var bufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

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
	Count    int            `json:"count"`
	Next     string         `json:"next"`
	Previous string         `json:"previous"`
	Results  []NetboxDevice `json:"results"`
}

type NetboxDevice struct {
	ID         int    `json:"id,omitempty"`
	URL        string `json:"url,omitempty"`
	DisplayURL string `json:"display_url,omitempty"`
	Display    string `json:"display,omitempty"`
	Name       string `json:"name,omitempty"`
	DeviceType struct {
		ID           int    `json:"id,omitempty"`
		URL          string `json:"url,omitempty"`
		Display      string `json:"display,omitempty"`
		Manufacturer struct {
			ID          int    `json:"id,omitempty"`
			URL         string `json:"url,omitempty"`
			Display     string `json:"display,omitempty"`
			Name        string `json:"name,omitempty"`
			Slug        string `json:"slug,omitempty"`
			Description string `json:"description,omitempty"`
		} `json:"manufacturer,omitempty"`
		Model       string `json:"model,omitempty"`
		Slug        string `json:"slug,omitempty"`
		Description string `json:"description,omitempty"`
	} `json:"device_type,omitempty"`
	Role struct {
		ID          int    `json:"id,omitempty"`
		URL         string `json:"url,omitempty"`
		Display     string `json:"display,omitempty"`
		Name        string `json:"name,omitempty"`
		Slug        string `json:"slug,omitempty"`
		Description string `json:"description,omitempty"`
	} `json:"role,omitempty"`
	Tenant struct {
		ID          int    `json:"id,omitempty"`
		URL         string `json:"url,omitempty"`
		Display     string `json:"display,omitempty"`
		Name        string `json:"name,omitempty"`
		Slug        string `json:"slug,omitempty"`
		Description string `json:"description,omitempty"`
	} `json:"tenant,omitempty"`
	Platform struct {
		ID          int    `json:"id,omitempty"`
		URL         string `json:"url,omitempty"`
		Display     string `json:"display,omitempty"`
		Name        string `json:"name,omitempty"`
		Slug        string `json:"slug,omitempty"`
		Description string `json:"description,omitempty"`
	} `json:"platform,omitempty"`
	Serial   string `json:"serial,omitempty"`
	AssetTag any    `json:"asset_tag,omitempty"`
	Site     struct {
		ID          int    `json:"id,omitempty"`
		URL         string `json:"url,omitempty"`
		Display     string `json:"display,omitempty"`
		Name        string `json:"name,omitempty"`
		Slug        string `json:"slug,omitempty"`
		Description string `json:"description,omitempty"`
	} `json:"site,omitempty"`
	Location struct {
		ID          int    `json:"id,omitempty"`
		URL         string `json:"url,omitempty"`
		Display     string `json:"display,omitempty"`
		Name        string `json:"name,omitempty"`
		Slug        string `json:"slug,omitempty"`
		Description string `json:"description,omitempty"`
		RackCount   int    `json:"rack_count,omitempty"`
		Depth       int    `json:"_depth,omitempty"`
	} `json:"location,omitempty"`
	Rack struct {
		ID          int    `json:"id,omitempty"`
		URL         string `json:"url,omitempty"`
		Display     string `json:"display,omitempty"`
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
	} `json:"rack,omitempty"`
	Position float64 `json:"position,omitempty"`
	Face     struct {
		Value string `json:"value,omitempty"`
		Label string `json:"label,omitempty"`
	} `json:"face,omitempty"`
	Latitude     any `json:"latitude,omitempty"`
	Longitude    any `json:"longitude,omitempty"`
	ParentDevice any `json:"parent_device,omitempty"`
	Status       struct {
		Value string `json:"value,omitempty"`
		Label string `json:"label,omitempty"`
	} `json:"status,omitempty"`
	Airflow struct {
		Value string `json:"value,omitempty"`
		Label string `json:"label,omitempty"`
	} `json:"airflow,omitempty"`
	PrimaryIP struct {
		ID      int    `json:"id,omitempty"`
		URL     string `json:"url,omitempty"`
		Display string `json:"display,omitempty"`
		Family  struct {
			Value int    `json:"value,omitempty"`
			Label string `json:"label,omitempty"`
		} `json:"family,omitempty"`
		Address     string `json:"address,omitempty"`
		Description string `json:"description,omitempty"`
	} `json:"primary_ip,omitempty"`
	PrimaryIP4 struct {
		ID      int    `json:"id,omitempty"`
		URL     string `json:"url,omitempty"`
		Display string `json:"display,omitempty"`
		Family  struct {
			Value int    `json:"value,omitempty"`
			Label string `json:"label,omitempty"`
		} `json:"family,omitempty"`
		Address     string `json:"address,omitempty"`
		Description string `json:"description,omitempty"`
	} `json:"primary_ip4,omitempty"`
	PrimaryIP6       any    `json:"primary_ip6,omitempty"`
	OobIP            any    `json:"oob_ip,omitempty"`
	Cluster          any    `json:"cluster,omitempty"`
	VirtualChassis   any    `json:"virtual_chassis,omitempty"`
	VcPosition       any    `json:"vc_position,omitempty"`
	VcPriority       any    `json:"vc_priority,omitempty"`
	Description      string `json:"description,omitempty"`
	Comments         string `json:"comments,omitempty"`
	ConfigTemplate   any    `json:"config_template,omitempty"`
	LocalContextData any    `json:"local_context_data,omitempty"`
	Tags             []any  `json:"tags,omitempty"`
	CustomFields     struct {
		BrocadeFcOemSerialnumber any      `json:"brocade_fc_oem_serialnumber,omitempty"`
		BrocadeFcSwitchLid       any      `json:"brocade_fc_switch_lid,omitempty"`
		BrocadeFcSwitchPrincipal any      `json:"brocade_fc_switch_principal,omitempty"`
		BrocadeFcSwitchType      any      `json:"brocade_fc_switch_type,omitempty"`
		BrocadeFosDomainID       any      `json:"brocade_fos_domain_id,omitempty"`
		BrocadeFosLicensedPorts  any      `json:"brocade_fos_licensed_ports,omitempty"`
		JiraAssetsID             string   `json:"jira_assets_id,omitempty"`
		InstallDate              string   `json:"install_date,omitempty"`
		BiosVersion              string   `json:"bios_version,omitempty"`
		FirmwareVersion          any      `json:"firmware_version,omitempty"`
		SoftwareVersion          string   `json:"software_version,omitempty"`
		ReferencedAssets         any      `json:"referenced_assets,omitempty"`
		WarrantyContractNumber   any      `json:"warranty_contract_number,omitempty"`
		WarrantyContractType     any      `json:"warranty_contract_type,omitempty"`
		WarrantyExpirationDate   any      `json:"warranty_expiration_date,omitempty"`
		HumidSensors             any      `json:"humid_sensors,omitempty"`
		TempSensors              any      `json:"temp_sensors,omitempty"`
		HostCPUModel             any      `json:"host_cpu_model,omitempty"`
		HostCurrentCPUSockets    any      `json:"host_current_cpu_sockets,omitempty"`
		HostCurrentDimmSlots     any      `json:"host_current_dimm_slots,omitempty"`
		HostMaxCPUSockets        any      `json:"host_max_cpu_sockets,omitempty"`
		HostMaxDimmSlots         any      `json:"host_max_dimm_slots,omitempty"`
		HostMemory               any      `json:"host_memory,omitempty"`
		HostPartnumber           any      `json:"host_partnumber,omitempty"`
		HostSku                  any      `json:"host_sku,omitempty"`
		HostSmbiosUUID           any      `json:"host_smbios_uuid,omitempty"`
		HostUUID                 any      `json:"host_uuid,omitempty"`
		Hostname                 string   `json:"hostname,omitempty"`
		NetworkDeviceSystemMac   any      `json:"network_device_system_mac,omitempty"`
		NetworkDeviceType        []string `json:"network_device_type,omitempty"`
		NetworkDeviceObserviumID any      `json:"network_device_observium_id,omitempty"`
		PduDetails               any      `json:"pdu_details,omitempty"`
	} `json:"custom_fields,omitempty"`
	Created                time.Time `json:"created,omitempty"`
	LastUpdated            time.Time `json:"last_updated,omitempty"`
	ConsolePortCount       int       `json:"console_port_count,omitempty"`
	ConsoleServerPortCount int       `json:"console_server_port_count,omitempty"`
	PowerPortCount         int       `json:"power_port_count,omitempty"`
	PowerOutletCount       int       `json:"power_outlet_count,omitempty"`
	InterfaceCount         int       `json:"interface_count,omitempty"`
	FrontPortCount         int       `json:"front_port_count,omitempty"`
	RearPortCount          int       `json:"rear_port_count,omitempty"`
	DeviceBayCount         int       `json:"device_bay_count,omitempty"`
	ModuleBayCount         int       `json:"module_bay_count,omitempty"`
	InventoryItemCount     int       `json:"inventory_item_count,omitempty"`
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

	var devices []NetboxDevice

	for {
		buf := bufferPool.Get().(*bytes.Buffer)
		defer bufferPool.Put(buf)
		buf.Reset()

		req, err := http.NewRequest("GET", u.String(), nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", "application/json")
		if auth := n.getAuth(n.options); auth != "" {
			req.Header.Set("Authorization", auth)
		}

		resp, err := n.client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		_, err = io.Copy(buf, resp.Body)
		if err != nil {
			return nil, err
		}

		var apiResp NetxboxAPIResponse

		err = json.NewDecoder(buf).Decode(&apiResp)
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

	return json.Marshal(devices)
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
