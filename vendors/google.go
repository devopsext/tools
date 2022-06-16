package vendors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strconv"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
)

type GoogleCalendarOptions struct {
	ID                 string
	TimeMin            string
	TimeMax            string
	AlwaysIncludeEmail bool
}

type GoogleOptions struct {
	Timeout           int
	Insecure          bool
	OAuthClientID     string
	OAuthClientSecret string
	RefreshToken      string
	AccessToken       string
	CalendarOptions   *GoogleCalendarOptions
}

type GoogleTokenReponse struct {
	AccessToken string `json:"access_token"`
	ExpiresIn   int    `json:"expires_in"`
	Scope       string `json:"scope"`
	TokenType   string `json:"token_type"`
}

type Google struct {
	client  *http.Client
	options GoogleOptions
	stdout  *common.Stdout
}

const (
	googleOAuthURL    = "https://oauth2.googleapis.com"
	googleCalendarURL = "https://www.googleapis.com/calendar/v3"
)

// https://developers.google.com/oauthplayground
func (g *Google) refreshCustomAccessToken(opts GoogleOptions) (*GoogleTokenReponse, error) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if !utils.IsEmpty(opts.OAuthClientID) {
		if err := w.WriteField("client_id", opts.OAuthClientID); err != nil {
			return nil, err
		}
	}
	if !utils.IsEmpty(opts.OAuthClientSecret) {
		if err := w.WriteField("client_secret", opts.OAuthClientSecret); err != nil {
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

	u, err := url.Parse(googleOAuthURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "/token")

	bytes, err := common.HttpPostRaw(g.client, u.String(), w.FormDataContentType(), "", body.Bytes())
	if err != nil {
		return nil, err
	}

	var r GoogleTokenReponse
	err = json.Unmarshal(bytes, &r)
	if err != nil {
		return nil, err
	}
	return &r, nil
}

func (g *Google) getCustomAccessToken(opts GoogleOptions) (string, error) {

	if !utils.IsEmpty(opts.AccessToken) {
		return opts.AccessToken, nil
	}
	r, err := g.refreshCustomAccessToken(opts)
	if err != nil {
		return "", err
	}
	return r.AccessToken, nil
}

func (g *Google) CalendarGetCustomEvents(opts GoogleOptions) ([]byte, error) {

	if opts.CalendarOptions == nil {
		return nil, fmt.Errorf("options are not enough")
	}

	accessToken, err := g.getCustomAccessToken(opts)
	if err != nil {
		return nil, err
	}
	g.stdout.Debug("Access token => %s", accessToken)

	params := make(url.Values)
	params.Add("access_token", accessToken)
	if !utils.IsEmpty(opts.CalendarOptions.TimeMin) {
		params.Add("timeMin", opts.CalendarOptions.TimeMin)
	}
	if !utils.IsEmpty(opts.CalendarOptions.TimeMax) {
		params.Add("timeMax", opts.CalendarOptions.TimeMax)
	}
	params.Add("alwaysIncludeEmail", strconv.FormatBool(opts.CalendarOptions.AlwaysIncludeEmail))

	u, err := url.Parse(googleCalendarURL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf("/calendars/%s/events", opts.CalendarOptions.ID))
	if params != nil {
		u.RawQuery = params.Encode()
	}
	return common.HttpGetRawWithHeaders(g.client, u.String(), nil)
}

func (g *Google) CalendarGetEvents() ([]byte, error) {
	return g.CalendarGetCustomEvents(g.options)
}

func NewGoogle(options GoogleOptions, stdout *common.Stdout) *Google {
	return &Google{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
		stdout:  stdout,
	}
}
