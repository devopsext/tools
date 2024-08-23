package vendors

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"path"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
	"golang.org/x/oauth2"
)

type Site24x7WebsiteMonitorOptions struct {
	Name          string
	URL           string
	Method        string
	Frequency     string
	Timeout       int
	Countries     []string
	UserAgent     string
	UseNameServer bool
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

type Site24x7Options struct {
	Timeout               int
	Insecure              bool
	ClientID              string
	ClientSecret          string
	RefreshToken          string
	NotificationProfileID string
	UserGroupIDs          []string
}

type Site24x7 struct {
	client  *http.Client
	options Site24x7Options
	logger  common.Logger
}

const (
	zohoOAuthV2TokenURL = "https://accounts.zoho.com/oauth/v2/token"
	site24x7ApiURL      = "https://www.site24x7.com/api"
	site24x7Monitors    = "/monitors"
	site24x7ContentType = "application/json"
)

func (s *Site24x7) CustomCreateWebsiteMonitor(site24x7Options Site24x7Options, createMonitorOptions Site24x7WebsiteMonitorOptions) ([]byte, error) {

	u, err := url.Parse(site24x7ApiURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, site24x7Monitors)

	r := &Site24x7WebsiteMonitor{
		DisplayName:           createMonitorOptions.Name,
		Type:                  "URL",
		Website:               createMonitorOptions.URL,
		CheckFrequency:        createMonitorOptions.Frequency,
		Timeout:               createMonitorOptions.Timeout,
		LocationProfileID:     "",
		NotificationProfileID: site24x7Options.NotificationProfileID,
		ThresholdProfileID:    "",
		UserGroupIDs:          site24x7Options.UserGroupIDs,
		HttpMethod:            createMonitorOptions.Method,
		UserAgent:             createMonitorOptions.UserAgent,
	}

	req, err := json.Marshal(&r)
	if err != nil {
		return nil, err
	}

	return utils.HttpPostRaw(s.client, u.String(), site24x7ContentType, "", req)
}

func (s *Site24x7) CreateWebsiteMonitor(options Site24x7WebsiteMonitorOptions) ([]byte, error) {
	return s.CustomCreateWebsiteMonitor(s.options, options)
}

func NewSite24x7(options Site24x7Options, logger common.Logger) *Site24x7 {

	config := &oauth2.Config{
		ClientID:     options.ClientID,
		ClientSecret: options.ClientSecret,

		Endpoint: oauth2.Endpoint{
			TokenURL:  zohoOAuthV2TokenURL,
			AuthStyle: oauth2.AuthStyleInParams,
		},
	}

	t := &oauth2.Token{
		RefreshToken: options.RefreshToken,
		TokenType:    "Zoho-oauthtoken",
	}

	client := utils.NewHttpClient(options.Timeout, options.Insecure)
	ctx := context.WithValue(context.Background(), oauth2.HTTPClient, client)

	return &Site24x7{
		client:  config.Client(ctx, t),
		options: options,
		logger:  logger,
	}
}
