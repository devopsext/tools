package vendors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strings"

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

type Site24x7WebsiteMonitorData struct {
	MonitorID         string `json:"monitor_id"`
	LocationProfileID string `json:"location_profile_id"`
}

type Site24x7WebsiteMonitorResponse struct {
	Site24x7Reponse
	Data *Site24x7WebsiteMonitorData `json:"data,omitempty"`
}

type Site24x7PollingStatusData struct {
	Status     string `json:"status"`
	MonmitorID string `json:"monitor_id"`
}

type Site24x7PollingStatusReponse struct {
	Site24x7Reponse
	Data *Site24x7PollingStatusData `json:"data,omitempty"`
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
	Site24x7LogReports           = "/reports/log_reports"
	Site24x7LocationTemplate     = "/location_template"
	Site24x7LocationProfiles     = "/location_profiles"
	Site24x7ContentType          = "application/json"
)

func (s *Site24x7) CheckResponse(resp Site24x7Reponse) error {

	if resp.Code != 0 {
		return fmt.Errorf("%s [%d]", resp.Message, resp.Code)
	}
	return nil
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

	bytes, err := utils.HttpPostRaw(s.client, u.String(), w.FormDataContentType(), "", body.Bytes())
	if err != nil {
		return nil, err
	}

	var r Site24x7AuthReponse
	err = json.Unmarshal(bytes, &r)
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

func (s *Site24x7) CustomCreateWebsiteMonitor(site24x7Options Site24x7Options, createMonitorOptions Site24x7WebsiteMonitorOptions) ([]byte, error) {

	at, err := s.CustomGetAccessToken(site24x7Options)
	if err != nil {
		return nil, err
	}

	opts := s.cloneSite24x7Options(site24x7Options)
	opts.AccessToken = at

	d, err := s.CustomGetLocationTemplate(opts)
	if err != nil {
		return nil, err
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

	if len(createMonitorOptions.Countries) == 0 {
		return nil, fmt.Errorf("no countries defined")
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

	countries := strings.Join(createMonitorOptions.Countries, ",")
	name := fmt.Sprintf("%s [%s]", createMonitorOptions.Name, countries)

	lopts := Site24x7LocationProfileOptions{
		Name:                 name,
		LocationID:           primary.LocationID,
		SecondaryLocationIDs: secondaryIDs,
	}

	d, err = s.CustomCreateLocationProfile(opts, lopts)
	if err != nil {
		return nil, err
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

	u, err := url.Parse(Site24x7ApiURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, Site24x7Monitors)

	r := &Site24x7WebsiteMonitor{
		DisplayName:           createMonitorOptions.Name,
		Type:                  "URL",
		Website:               createMonitorOptions.URL,
		CheckFrequency:        createMonitorOptions.Frequency,
		Timeout:               createMonitorOptions.Timeout,
		LocationProfileID:     lr.Data.ProfileID,
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

func (s *Site24x7) CustomDeleteMonitor(site24x7Options Site24x7Options, deleteMonitorOptions Site24x7MonitorOptions) ([]byte, error) {

	at, err := s.CustomGetAccessToken(site24x7Options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(Site24x7ApiURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, Site24x7Monitors, deleteMonitorOptions.ID)

	return utils.HttpDeleteRaw(s.client, u.String(), Site24x7ContentType, s.getAuth(at), nil)
}

func (s *Site24x7) DeleteMonitor(options Site24x7MonitorOptions) ([]byte, error) {
	return s.CustomDeleteMonitor(s.options, options)
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

func (s *Site24x7) CustomGetPollingStatus(site24x7Options Site24x7Options, pollMonitorOptions Site24x7MonitorOptions) ([]byte, error) {

	at, err := s.CustomGetAccessToken(site24x7Options)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(Site24x7ApiURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, Site24x7MonitorStatusPollNow, pollMonitorOptions.ID)

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
