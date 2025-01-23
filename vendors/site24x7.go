package vendors

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
	"github.com/jinzhu/copier"
)

type Site24x7MonitorOptions struct {
	ID string
}

type Site24x7WebsiteMonitorOptions struct {
	Name                  string
	URL                   string
	Method                string
	Frequency             string
	Timeout               int
	Countries             []string
	UserAgent             string
	UseNameServer         bool
	NotificationProfileID string
	ThresholdProfileID    string
	UserGroupIDs          []string
}

type Site24x7LocationProfileOptions struct {
	ID                   string
	Name                 string
	LocationID           string
	SecondaryLocationIDs []string
}

type Site24x7LogReportOptions struct {
	Site24x7MonitorOptions
	StartDate string
	EndDate   string
}

type Site24x7LocationProfile struct {
	ProfileName                 string   `json:"profile_name"`
	PrimaryLocation             string   `json:"primary_location"`
	SecondaryCheckFrequency     string   `json:"secondary_check_frequency,omitempty"`
	SecondaryLocations          []string `json:"secondary_locations,omitempty"`
	RestrictAltLoc              bool     `json:"restrict_alt_loc,omitempty"`
	OuterRegionsLOcationConsent bool     `json:"outer_regions_location_consent,omitempty"`
}

type Site24x7WebsiteMonitor struct {
	DisplayName           string   `json:"display_name"`
	Type                  string   `json:"type"`
	Website               string   `json:"website"`
	CheckFrequency        string   `json:"check_frequency"`
	Timeout               int      `json:"timeout"`
	LocationProfileID     string   `json:"location_profile_id"`
	NotificationProfileID string   `json:"notification_profile_id"`
	ThresholdProfileID    string   `json:"threshold_profile_id"`
	UserGroupIDs          []string `json:"user_group_ids"`
	HttpMethod            string   `json:"http_method"`
	IPType                int      `json:"ip_type,omitempty"`
	UserAgent             string   `json:"user_agent,omitempty"`
	MonitorGroups         []string `json:"monitor_groups,omitempty"`
	UseNameServer         bool     `json:"use_name_server,omitempty"`
}

type Site24x7AuthReponse struct {
	AccessToken string `json:"access_token"`
	ApiDomain   string `json:"api_domain"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type Site24x7Reponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type Site24x7ErrorReponse struct {
	ErrorCode int    `json:"error_code"`
	Message   string `json:"message"`
}

type Site24x7LocationTemplateDataLocation struct {
	LocationID  string `json:"location_id"`
	DisplayName string `json:"display_name"`
	CityName    string `json:"city_name"`
	CityShort   string `json:"city_short"`
	CountryName string `json:"country_name"`
	Continent   string `json:"continent"`
	UseIpv6     bool   `json:"use_ipv6"`
	ProbInfo    string `json:"probe_info"`
}

type Site24x7LocationTemplateData struct {
	Locations []*Site24x7LocationTemplateDataLocation `json:"locations,omitempty"`
}

type Site24x7LocationTemplateReponse struct {
	Site24x7Reponse
	Data *Site24x7LocationTemplateData `json:"data,omitempty"`
}

type Site24x7LocationProfileData struct {
	ProfileID               string   `json:"profile_id"`
	ProfileName             string   `json:"profile_name"`
	PrimaryLocation         string   `json:"primary_location"`
	SecondaryCheckFrequency string   `json:"secondary_check_frequency,omitempty"`
	SecondaryLocations      []string `json:"secondary_locations,omitempty"`
}

type Site24x7LocationProfileReponse struct {
	Site24x7Reponse
	Data *Site24x7LocationProfileData `json:"data,omitempty"`
}

type Site24x7LocationProfilesReponse struct {
	Site24x7Reponse
	Data []*Site24x7LocationProfileData `json:"data,omitempty"`
}

type Site24x7WebsiteMonitorData struct {
	MonitorID         string `json:"monitor_id"`
	LocationProfileID string `json:"location_profile_id"`
	DisplayName       string `json:"display_name"`
}

type Site24x7WebsiteMonitorResponse struct {
	Site24x7Reponse
	Data *Site24x7WebsiteMonitorData `json:"data,omitempty"`
}

type Site24x7PollStatusData struct {
	Status    string `json:"status"`
	MonitorID string `json:"monitor_id"`
}

type Site24x7PollStatusReponse struct {
	Site24x7Reponse
	Data *Site24x7PollStatusData `json:"data,omitempty"`
}

type Site24x7DeleteData struct {
	ResourceName string `json:"resource_name"`
}

type Site24x7DeleteResponse struct {
	Site24x7Reponse
	Data *Site24x7DeleteData `json:"data,omitempty"`
}

type Site24x7LogReportDataReport struct {
	ConnectionTime     string `json:"connection_time"`
	DnsTime            string `json:"dns_time"`
	SSLTime            string `json:"ssl_time"`
	ResponseCode       string `json:"response_code"`
	CollectionTime     string `json:"collection_time"`
	Availability       string `json:"availability"`
	ResponseTime       string `json:"response_time"`
	LocationID         string `json:"location_id"`
	Nameserver         string `json:"nameserver"`
	ResolvedIP         string `json:"resolved_ip"`
	Reason             string `json:"reason"`
	ContentLength      string `json:"content_length"`
	DataCollectionType string `json:"data_collection_type"`
}

type Site24x7LogReportData struct {
	Report []*Site24x7LogReportDataReport
}

type Site24x7LogReportReponse struct {
	Site24x7Reponse
	Data *Site24x7LogReportData `json:"data,omitempty"`
}

type Site24x7Options struct {
	Timeout      int
	Insecure     bool
	ClientID     string
	ClientSecret string
	RefreshToken string
	AccessToken  string
}

type Site24x7 struct {
	client  *http.Client
	options Site24x7Options
	logger  common.Logger
}

const (
	ZohoOAuthV2TokenURL          = "https://accounts.zoho.com/oauth/v2/token"
	Site24x7ApiURL               = "https://www.site24x7.com/api"
	Site24x7Monitors             = "/monitors"
	Site24x7MonitorPollNow       = "/monitor/poll_now"
	Site24x7MonitorStatusPollNow = "/monitor/status_poll_now"
	Site24x7MonitorsName         = "/monitors/name"
	Site24x7MonitorsActivate     = "/monitors/activate"
	Site24x7MonitorsSuspend      = "/monitors/suspend"
	Site24x7LogReports           = "/reports/log_reports"
	Site24x7LocationTemplate     = "/location_template"
	Site24x7LocationProfiles     = "/location_profiles"
	Site24x7ContentType          = "application/json"
)

const (
	Site24x7DataCollectionTypeNormal  = "1"
	Site24x7DataCollectionTypePollNow = "3"
)

func (s *Site24x7) CheckResponse(resp Site24x7Reponse) error {

	if resp.Code != 0 {
		return fmt.Errorf("%s [%d]", resp.Message, resp.Code)
	}
	return nil
}

func (s *Site24x7) CheckError(data []byte, e error) error {

	r := Site24x7ErrorReponse{}

	err := json.Unmarshal(data, &r)
	if err != nil {
		return err
	}

	if r.ErrorCode != 0 {
		return fmt.Errorf("%s [%d]", r.Message, r.ErrorCode)
	}
	return e
}

// go to https://api-console.zoho.com/ and generate code with Site24x7.Admin.All scope
// do curl with params according https://www.site24x7.com/help/api/#authentication page
// copy refresh token and put into the options

func (s *Site24x7) getTokenAuth(opts Site24x7Options) (*Site24x7AuthReponse, error) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if !utils.IsEmpty(opts.ClientID) {
		if err := w.WriteField("client_id", opts.ClientID); err != nil {
			return nil, err
		}
	}
	if !utils.IsEmpty(opts.ClientSecret) {
		if err := w.WriteField("client_secret", opts.ClientSecret); err != nil {
			return nil, err
		}
	}
	if !utils.IsEmpty(opts.RefreshToken) {
		if err := w.WriteField("refresh_token", opts.RefreshToken); err != nil {
			return nil, err
		}
	}
	if err := w.WriteField("grant_type", "refresh_token"); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	u, err := url.Parse(ZohoOAuthV2TokenURL)
	if err != nil {
		return nil, err
	}

	d, err := utils.HttpPostRaw(s.client, u.String(), w.FormDataContentType(), "", body.Bytes())
	if err != nil {
		return nil, s.CheckError(d, err)
	}

	var r Site24x7AuthReponse
	err = json.Unmarshal(d, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (s *Site24x7) getAuth(token string) string {
	return fmt.Sprintf("Zoho-oauthtoken %s", token)
}

func (s *Site24x7) getAccessToken(opts Site24x7Options) (string, error) {

	r, err := s.getTokenAuth(opts)
	if err != nil {
		return "", err
	}
	return r.AccessToken, nil
}

func (s *Site24x7) cloneSite24x7Options(opts Site24x7Options) Site24x7Options {

	r := Site24x7Options{}
	copier.Copy(&r, &opts)
	return r
}

func (s *Site24x7) CustomGetAccessToken(opts Site24x7Options) (string, error) {

	if !utils.IsEmpty(opts.AccessToken) {
		return opts.AccessToken, nil
	}
	return s.getAccessToken(opts)
}

func (s *Site24x7) CustomGetLocationTemplate(site24x7Options Site24x7Options) ([]byte, error) {

	at, err := s.CustomGetAccessToken(site24x7Options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(Site24x7ApiURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, Site24x7LocationTemplate)

	return utils.HttpGetRaw(s.client, u.String(), Site24x7ContentType, s.getAuth(at))
}

func (s *Site24x7) GetLocationTemplate() ([]byte, error) {
	return s.CustomGetLocationTemplate(s.options)
}

func (s *Site24x7) CustomGetLocationProfiles(site24x7Options Site24x7Options) ([]byte, error) {

	at, err := s.CustomGetAccessToken(site24x7Options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(Site24x7ApiURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, Site24x7LocationProfiles)

	return utils.HttpGetRaw(s.client, u.String(), Site24x7ContentType, s.getAuth(at))
}

func (s *Site24x7) GetLocationProfiles() ([]byte, error) {
	return s.CustomGetLocationProfiles(s.options)
}

func (s *Site24x7) FindLocationProfileByName(site24x7Options Site24x7Options, name string) (string, error) {

	d, err := s.CustomGetLocationProfiles(site24x7Options)
	if err != nil {
		return "", s.CheckError(d, err)
	}

	lps := Site24x7LocationProfilesReponse{}
	err = json.Unmarshal(d, &lps)
	if err != nil {
		return "", err
	}

	err = s.CheckResponse(lps.Site24x7Reponse)
	if err != nil {
		return "", err
	}

	var lp *Site24x7LocationProfileData
	for _, p := range lps.Data {

		if p.ProfileName == name {
			lp = p
			break
		}
	}
	if lp == nil {
		return "", nil
	}
	return lp.ProfileID, nil
}

func (s *Site24x7) CustomCreateLocationProfile(site24x7Options Site24x7Options, createLocationOptions Site24x7LocationProfileOptions) ([]byte, error) {

	at, err := s.CustomGetAccessToken(site24x7Options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(Site24x7ApiURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, Site24x7LocationProfiles)

	r := &Site24x7LocationProfile{
		ProfileName:        createLocationOptions.Name,
		PrimaryLocation:    createLocationOptions.LocationID,
		SecondaryLocations: createLocationOptions.SecondaryLocationIDs,
	}

	req, err := json.Marshal(&r)
	if err != nil {
		return nil, err
	}

	return utils.HttpPostRaw(s.client, u.String(), Site24x7ContentType, s.getAuth(at), req)
}

func (s *Site24x7) CreateLocationProfile(options Site24x7LocationProfileOptions) ([]byte, error) {
	return s.CustomCreateLocationProfile(s.options, options)
}

func (s *Site24x7) CustomDeleteLocationProfile(site24x7Options Site24x7Options, deleteLocationOptions Site24x7LocationProfileOptions) ([]byte, error) {

	at, err := s.CustomGetAccessToken(site24x7Options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(Site24x7ApiURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, Site24x7LocationProfiles, deleteLocationOptions.ID)

	return utils.HttpDeleteRaw(s.client, u.String(), Site24x7ContentType, s.getAuth(at), nil)
}

func (s *Site24x7) DeleteLocationProfile(options Site24x7LocationProfileOptions) ([]byte, error) {
	return s.CustomDeleteLocationProfile(s.options, options)
}

func (s *Site24x7) FindLocationByCountry(locations []*Site24x7LocationTemplateDataLocation, country string) *Site24x7LocationTemplateDataLocation {

	for _, l := range locations {

		short := common.CountryShort(l.CountryName)
		if short == country {
			return l
		}
	}
	return nil
}

func (s *Site24x7) CustomRetrieveMonitorByName(site24x7Options Site24x7Options, name string) ([]byte, error) {

	at, err := s.CustomGetAccessToken(site24x7Options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(Site24x7ApiURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, Site24x7MonitorsName, name)

	return utils.HttpGetRaw(s.client, u.String(), Site24x7ContentType, s.getAuth(at))
}

func (s *Site24x7) RetrieveMonitorByName(name string) ([]byte, error) {
	return s.CustomRetrieveMonitorByName(s.options, name)
}

func (s *Site24x7) CustomCreateWebsiteMonitor(site24x7Options Site24x7Options, createMonitorOptions Site24x7WebsiteMonitorOptions) ([]byte, error) {

	if len(createMonitorOptions.Countries) == 0 {
		return nil, fmt.Errorf("no countries defined")
	}

	at, err := s.CustomGetAccessToken(site24x7Options)
	if err != nil {
		return nil, err
	}

	opts := s.cloneSite24x7Options(site24x7Options)
	opts.AccessToken = at

	name := createMonitorOptions.Name

	profileID, _ := s.FindLocationProfileByName(opts, name)

	if utils.IsEmpty(profileID) {

		d, err := s.CustomGetLocationTemplate(opts)
		if err != nil {
			return nil, s.CheckError(d, err)
		}

		ltr := Site24x7LocationTemplateReponse{}
		err = json.Unmarshal(d, &ltr)
		if err != nil {
			return nil, err
		}

		err = s.CheckResponse(ltr.Site24x7Reponse)
		if err != nil {
			return nil, err
		}

		primary := s.FindLocationByCountry(ltr.Data.Locations, createMonitorOptions.Countries[0])
		if primary == nil {
			return nil, fmt.Errorf("no primary location found %s", createMonitorOptions.Countries[0])
		}

		secondaryIDs := []string{}
		for i := 1; i < len(createMonitorOptions.Countries); i++ {
			secondary := s.FindLocationByCountry(ltr.Data.Locations, createMonitorOptions.Countries[i])
			if secondary != nil {
				secondaryIDs = append(secondaryIDs, secondary.LocationID)
			}
		}

		lopts := Site24x7LocationProfileOptions{
			Name:                 name,
			LocationID:           primary.LocationID,
			SecondaryLocationIDs: secondaryIDs,
		}

		d, err = s.CustomCreateLocationProfile(opts, lopts)
		if err != nil {
			return nil, s.CheckError(d, err)
		}

		lr := Site24x7LocationProfileReponse{}
		err = json.Unmarshal(d, &lr)
		if err != nil {
			return nil, err
		}

		err = s.CheckResponse(ltr.Site24x7Reponse)
		if err != nil {
			return nil, err
		}
		profileID = lr.Data.ProfileID
	}

	d, err := s.CustomRetrieveMonitorByName(opts, name)
	if err == nil {
		rr := Site24x7WebsiteMonitorResponse{}
		err = json.Unmarshal(d, &rr)
		if err != nil {
			return nil, err
		}

		err = s.CheckResponse(rr.Site24x7Reponse)
		if err != nil {
			return nil, err
		}

		if rr.Data != nil {

			s.CustomActivateMonitor(opts, Site24x7MonitorOptions{ID: rr.Data.MonitorID})
			return d, nil
		}
	}

	u, err := url.Parse(Site24x7ApiURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, Site24x7Monitors)

	frequency := createMonitorOptions.Frequency
	if utils.IsEmpty(frequency) {
		frequency = "1440"
	}

	timeout := createMonitorOptions.Timeout
	if utils.IsEmpty(timeout) {
		timeout = 30
	}

	r := &Site24x7WebsiteMonitor{
		DisplayName:           name,
		Type:                  "URL",
		Website:               createMonitorOptions.URL,
		CheckFrequency:        frequency,
		Timeout:               timeout,
		LocationProfileID:     profileID,
		NotificationProfileID: createMonitorOptions.NotificationProfileID,
		ThresholdProfileID:    createMonitorOptions.ThresholdProfileID,
		UserGroupIDs:          createMonitorOptions.UserGroupIDs,
		HttpMethod:            createMonitorOptions.Method,
		UserAgent:             createMonitorOptions.UserAgent,
	}

	req, err := json.Marshal(&r)
	if err != nil {
		return nil, err
	}

	return utils.HttpPostRaw(s.client, u.String(), Site24x7ContentType, s.getAuth(at), req)
}

func (s *Site24x7) CreateWebsiteMonitor(options Site24x7WebsiteMonitorOptions) ([]byte, error) {
	return s.CustomCreateWebsiteMonitor(s.options, options)
}

func (s *Site24x7) CustomDeleteMonitor(site24x7Options Site24x7Options, monitorOptions Site24x7MonitorOptions) ([]byte, error) {

	at, err := s.CustomGetAccessToken(site24x7Options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(Site24x7ApiURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, Site24x7Monitors, monitorOptions.ID)

	return utils.HttpDeleteRaw(s.client, u.String(), Site24x7ContentType, s.getAuth(at), nil)
}

func (s *Site24x7) DeleteMonitor(options Site24x7MonitorOptions) ([]byte, error) {
	return s.CustomDeleteMonitor(s.options, options)
}

func (s *Site24x7) CustomActivateMonitor(site24x7Options Site24x7Options, monitorOptions Site24x7MonitorOptions) ([]byte, error) {

	at, err := s.CustomGetAccessToken(site24x7Options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(Site24x7ApiURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, Site24x7MonitorsActivate, monitorOptions.ID)

	return utils.HttpDeleteRaw(s.client, u.String(), Site24x7ContentType, s.getAuth(at), nil)
}

func (s *Site24x7) ActivateMonitor(options Site24x7MonitorOptions) ([]byte, error) {
	return s.CustomActivateMonitor(s.options, options)
}

func (s *Site24x7) CustomSuspendMonitor(site24x7Options Site24x7Options, monitorOptions Site24x7MonitorOptions) ([]byte, error) {

	at, err := s.CustomGetAccessToken(site24x7Options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(Site24x7ApiURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, Site24x7MonitorsSuspend, monitorOptions.ID)

	return utils.HttpDeleteRaw(s.client, u.String(), Site24x7ContentType, s.getAuth(at), nil)
}

func (s *Site24x7) SuspendMonitor(options Site24x7MonitorOptions) ([]byte, error) {
	return s.CustomSuspendMonitor(s.options, options)
}

func (s *Site24x7) CustomPollMonitor(site24x7Options Site24x7Options, pollMonitorOptions Site24x7MonitorOptions) ([]byte, error) {

	at, err := s.CustomGetAccessToken(site24x7Options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(Site24x7ApiURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, Site24x7MonitorPollNow, pollMonitorOptions.ID)

	return utils.HttpGetRaw(s.client, u.String(), Site24x7ContentType, s.getAuth(at))
}

func (s *Site24x7) PollMonitor(options Site24x7MonitorOptions) ([]byte, error) {
	return s.CustomPollMonitor(s.options, options)
}

func (s *Site24x7) PollMonitorWait(ctx context.Context, site24x7Options Site24x7Options, monitorOptions Site24x7MonitorOptions, delay int, statuses []string) bool {

	t := time.Duration(delay) * time.Second
	for {

		select {
		case <-ctx.Done():
			return false
		case <-time.After(t):

			d, err := s.CustomGetPollingStatus(site24x7Options, monitorOptions)
			if err != nil {
				continue
			}

			r := Site24x7PollStatusReponse{}
			err = json.Unmarshal(d, &r)
			if err != nil {
				continue
			}

			err = s.CheckResponse(r.Site24x7Reponse)
			if err != nil {
				continue
			}

			status := strings.ToLower(r.Data.Status)
			if utils.Contains(statuses, status) {
				return true
			}
		}
	}
}

func (s *Site24x7) CustomGetPollingStatus(site24x7Options Site24x7Options, monitorOptions Site24x7MonitorOptions) ([]byte, error) {

	at, err := s.CustomGetAccessToken(site24x7Options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(Site24x7ApiURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, Site24x7MonitorStatusPollNow, monitorOptions.ID)

	return utils.HttpGetRaw(s.client, u.String(), Site24x7ContentType, s.getAuth(at))
}

func (s *Site24x7) GetPollingStatus(options Site24x7MonitorOptions) ([]byte, error) {
	return s.CustomGetPollingStatus(s.options, options)
}

func (s *Site24x7) CustomGetLogReport(site24x7Options Site24x7Options, logReportOptions Site24x7LogReportOptions) ([]byte, error) {

	at, err := s.CustomGetAccessToken(site24x7Options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(Site24x7ApiURL)
	if err != nil {
		return nil, err
	}

	var params = make(url.Values)

	if !utils.IsEmpty(logReportOptions.StartDate) && !utils.IsEmpty(logReportOptions.EndDate) {
		params.Add("start_date", logReportOptions.StartDate)
		params.Add("end_date", logReportOptions.EndDate)
	}

	if !utils.IsEmpty(logReportOptions.StartDate) && utils.IsEmpty(logReportOptions.EndDate) {
		params.Add("date", logReportOptions.StartDate)
	}

	u.Path = path.Join(u.Path, Site24x7LogReports, logReportOptions.ID)
	u.RawQuery = params.Encode()
	return utils.HttpGetRaw(s.client, u.String(), Site24x7ContentType, s.getAuth(at))
}

func (s *Site24x7) GetLogReport(options Site24x7LogReportOptions) ([]byte, error) {
	return s.CustomGetLogReport(s.options, options)
}

func NewSite24x7(options Site24x7Options, logger common.Logger) *Site24x7 {

	return &Site24x7{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
		logger:  logger,
	}
}
