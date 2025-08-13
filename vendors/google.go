package vendors

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
	"github.com/google/uuid"
)

type GoogleConferenceDataSolutionKey struct {
	Type string `json:"type"`
}

type GoogleConferenceDataSolution struct {
	Key     GoogleConferenceDataSolutionKey `json:"key"`
	Name    string                          `json:"name"`
	IconURI string                          `json:"iconUri"`
}

type GoogleConferenceDataCreateRequest struct {
	ConferenceSolutionKey GoogleConferenceDataSolutionKey `json:"conferenceSolutionKey"`
	RequestID             string                          `json:"requestId"`
}

type GoogleConferenceDataEntryPoint struct {
	EntryPointType string `json:"entryPointType,omitempty"`
	URI            string `json:"uri,omitempty"`
	Label          string `json:"label,omitempty"`
}

type GoogleConferenceData struct {
	ConferenceSolution *GoogleConferenceDataSolution      `json:"conferenceSolution,omitempty"`
	CreateRequest      *GoogleConferenceDataCreateRequest `json:"createRequest,omitempty"`
	EntryPoints        []*GoogleConferenceDataEntryPoint  `json:"entryPoints,omitempty"`
	ConferenceID       string                             `json:"conferenceId,omitempty"`
}

type GoogleCalendarEventDataTime struct {
	Date     string `json:"date,omitempty"`
	DateTime string `json:"dateTime,omitempty"`
	TimeZone string `json:"timeZone,omitempty"`
}

type GoogleCalendarEventAttendee struct {
	Email    string `json:"email"`
	Optional string `json:"optional,omitempty"`
}

type GoogleCalendarEventSource struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type GoogleCalendarEvent struct {
	ID                      string                         `json:"id,omitempty"`
	Summary                 string                         `json:"summary"`
	Description             string                         `json:"description"`
	EventType               string                         `json:"eventType"`
	Location                string                         `json:"location,omitempty"`
	Transparency            string                         `json:"transparency,omitempty"`
	Visibility              string                         `json:"visibility,omitempty"`
	Start                   GoogleCalendarEventDataTime    `json:"start"`
	End                     GoogleCalendarEventDataTime    `json:"end"`
	Attendees               []*GoogleCalendarEventAttendee `json:"attendees"`
	GuestsCanInviteOthers   bool                           `json:"guestsCanInviteOthers"`
	GuestsCanModify         bool                           `json:"guestsCanModify"`
	GuestsCanSeeOtherGuests bool                           `json:"guestsCanSeeOtherGuests"`
	Source                  *GoogleCalendarEventSource     `json:"source,omitempty"`
	ConferenceData          *GoogleConferenceData          `json:"conferenceData,omitempty"`
}

type GoogleCalendarEvents struct {
	Kind     string                 `json:"kind"`
	Summary  string                 `json:"summary,omitempty"`
	TimeZone string                 `json:"timeZone,omitempty"`
	Items    []*GoogleCalendarEvent `json:"items,omitempty"`
}

type GoogleCalendarInsertEventOptions struct {
	Summary             string
	Description         string
	Start               string
	End                 string
	TimeZone            string
	Visibility          string
	SendUpdates         string
	SupportsAttachments bool
	SourceTitle         string
	SourceURL           string
	ConferenceID        string
}

type GoogleCalendarDeleteEventOptions struct {
	ID          string
	SendUpdates string
}

type GoogleCalendarDeleteEventsOptions struct {
	TimeMin string
	TimeMax string
}

type GoogleCalendarGetEventsOptions struct {
	TimeMin      string
	TimeMax      string
	TimeZone     string
	OrderBy      string
	Q            string
	SingleEvents bool
}

type GoogleCalendarOptions struct {
	ID string
}

type GoogleDocsOptions struct {
	ID string
}

type GoogleOptions struct {
	Timeout           int
	Insecure          bool
	OAuthClientID     string
	OAuthClientSecret string
	RefreshToken      string
	ServiceAccountKey string
	ImpersonateEmail  string
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
	logger  common.Logger
}

type GoogleMeetSpaceConfig struct {
	AccessType string `json:"accessType"`
}

type GoogleMeetSpaceRequest struct {
	Config GoogleMeetSpaceConfig `json:"config"`
}

type GoogleMeetSpaceResponse struct {
	Name        string `json:"name"`
	MeetingUri  string `json:"meetingUri"`
	MeetingCode string `json:"meetingCode"`
}

type GoogleMeetOptions struct {
	AccessType string
}

// Service Account JSON structure
type GoogleServiceAccount struct {
	Type                    string `json:"type"`
	ProjectID               string `json:"project_id"`
	PrivateKeyID            string `json:"private_key_id"`
	PrivateKey              string `json:"private_key"`
	ClientEmail             string `json:"client_email"`
	ClientID                string `json:"client_id"`
	AuthURI                 string `json:"auth_uri"`
	TokenURI                string `json:"token_uri"`
	AuthProviderX509CertURL string `json:"auth_provider_x509_cert_url"`
	ClientX509CertURL       string `json:"client_x509_cert_url"`
}

// JWT Claims for service account
type GoogleJWTClaims struct {
	Iss   string `json:"iss"`   // service account email
	Sub   string `json:"sub"`   // impersonation email (for domain-wide delegation)
	Scope string `json:"scope"` // required scopes
	Aud   string `json:"aud"`   // token endpoint
	Exp   int64  `json:"exp"`   // expiration time
	Iat   int64  `json:"iat"`   // issued at time
}

// JWT Header
type GoogleJWTHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

const (
	googleOAuthURL            = "https://oauth2.googleapis.com"
	googleCalendarURL         = "https://www.googleapis.com/calendar/v3"
	googleMeetAPIURL          = "https://meet.googleapis.com/v2"
	googleCalendarEvents      = "/calendars/%s/events"
	googleCalendarDeleteEvent = "/calendars/%s/events/%s"
	googleMeetURL             = "https://meet.google.com/%s"
	googleMeetLabel           = "meet.google.com/%s"
	googleDocsURL             = "https://www.googleapis.com/drive/v3"
)

// go to https://developers.google.com/oauthplayground
// set options to use OAuth Client ID and OAuth Client secret
// choose Access type => Online
// select API => https://www.googleapis.com/auth/calendar,https://www.googleapis.com/auth/calendar.events
// clieck Autorize Api, and Allow for your user
// use refresh token

func (g *Google) refreshToken(opts GoogleOptions) (*GoogleTokenReponse, error) {

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

	bytes, err := utils.HttpPostRaw(g.client, u.String(), w.FormDataContentType(), "", body.Bytes())
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

// Service account authentication with domain-wide delegation
func (g *Google) getServiceAccountToken(opts GoogleOptions) (string, error) {
	if utils.IsEmpty(opts.ServiceAccountKey) {
		return "", errors.New("service account key is required")
	}

	// Parse service account JSON
	var serviceAccount GoogleServiceAccount
	if strings.TrimSpace(opts.ServiceAccountKey)[0] == '{' {
		// JSON string
		err := json.Unmarshal([]byte(opts.ServiceAccountKey), &serviceAccount)
		if err != nil {
			return "", fmt.Errorf("failed to parse service account JSON: %v", err)
		}
	} else {
		// File path - read file
		data, err := utils.Content(opts.ServiceAccountKey)
		if err != nil {
			return "", fmt.Errorf("failed to read service account file: %v", err)
		}
		err = json.Unmarshal(data, &serviceAccount)
		if err != nil {
			return "", fmt.Errorf("failed to parse service account file: %v", err)
		}
	}

	g.logger.Debug("Service account loaded: %s", serviceAccount.ClientEmail)

	// Create JWT
	jwt, err := g.createServiceAccountJWT(serviceAccount, opts.ImpersonateEmail)
	if err != nil {
		return "", fmt.Errorf("failed to create JWT: %v", err)
	}

	g.logger.Debug("JWT created successfully")

	// Exchange JWT for access token
	token, err := g.exchangeJWTForToken(jwt, serviceAccount.TokenURI)
	if err != nil {
		return "", fmt.Errorf("failed to exchange JWT for token: %v", err)
	}

	g.logger.Debug("Access token obtained successfully")
	return token, nil
}

// Create and sign JWT for service account
func (g *Google) createServiceAccountJWT(serviceAccount GoogleServiceAccount, impersonateEmail string) (string, error) {
	// JWT Header
	header := GoogleJWTHeader{
		Alg: "RS256",
		Typ: "JWT",
	}

	headerJSON, err := json.Marshal(header)
	if err != nil {
		return "", err
	}
	headerEncoded := base64.RawURLEncoding.EncodeToString(headerJSON)

	// JWT Claims
	now := time.Now()
	claims := GoogleJWTClaims{
		Iss:   serviceAccount.ClientEmail,
		Scope: "https://www.googleapis.com/auth/meetings.space.created",
		Aud:   serviceAccount.TokenURI,
		Exp:   now.Add(time.Hour).Unix(),
		Iat:   now.Unix(),
	}

	// Add impersonation for domain-wide delegation
	if !utils.IsEmpty(impersonateEmail) {
		claims.Sub = impersonateEmail
		g.logger.Debug("Impersonating user: %s", impersonateEmail)
	}

	claimsJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	claimsEncoded := base64.RawURLEncoding.EncodeToString(claimsJSON)

	// Create signature
	message := headerEncoded + "." + claimsEncoded
	signature, err := g.signJWT(message, serviceAccount.PrivateKey)
	if err != nil {
		return "", err
	}

	// Complete JWT
	jwt := message + "." + signature
	return jwt, nil
}

// Sign JWT using RSA-SHA256
func (g *Google) signJWT(message, privateKeyPEM string) (string, error) {
	// Parse private key
	block, _ := pem.Decode([]byte(privateKeyPEM))
	if block == nil {
		return "", errors.New("failed to decode PEM block containing private key")
	}

	var privateKey *rsa.PrivateKey
	var err error

	// Try PKCS#1 format first, then PKCS#8
	if block.Type == "RSA PRIVATE KEY" {
		privateKey, err = x509.ParsePKCS1PrivateKey(block.Bytes)
	} else if block.Type == "PRIVATE KEY" {
		key, err := x509.ParsePKCS8PrivateKey(block.Bytes)
		if err != nil {
			return "", err
		}
		var ok bool
		privateKey, ok = key.(*rsa.PrivateKey)
		if !ok {
			return "", errors.New("not an RSA private key")
		}
	} else {
		return "", fmt.Errorf("unsupported key type: %s", block.Type)
	}

	if err != nil {
		return "", err
	}

	// Sign the message
	hashed := sha256.Sum256([]byte(message))
	signature, err := rsa.SignPKCS1v15(rand.Reader, privateKey, crypto.SHA256, hashed[:])
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(signature), nil
}

// Exchange JWT for access token
func (g *Google) exchangeJWTForToken(jwt, tokenURI string) (string, error) {
	// Prepare form data
	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:jwt-bearer")
	data.Set("assertion", jwt)

	// Make request
	resp, err := g.client.Post(tokenURI, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Parse response
	var tokenResponse GoogleTokenReponse
	err = json.NewDecoder(resp.Body).Decode(&tokenResponse)
	if err != nil {
		return "", err
	}

	if resp.StatusCode != 200 {
		return "", fmt.Errorf("token exchange failed: %d - %s", resp.StatusCode, tokenResponse.AccessToken)
	}

	if utils.IsEmpty(tokenResponse.AccessToken) {
		return "", errors.New("no access token received")
	}

	return tokenResponse.AccessToken, nil
}

// Get access token using either OAuth refresh or service account
func (g *Google) getAccessToken(opts GoogleOptions) (string, error) {
	// Try service account first if configured
	if !utils.IsEmpty(opts.ServiceAccountKey) {
		return g.getServiceAccountToken(opts)
	}

	// Fall back to OAuth refresh token
	if !utils.IsEmpty(opts.RefreshToken) {
		r, err := g.refreshToken(opts)
		if err != nil {
			return "", err
		}
		return r.AccessToken, nil
	}

	return "", errors.New("no authentication method configured. Need either service account key or OAuth refresh token")
}

// https://developers.google.com/calendar/api/v3/reference/events/get
// https://stackoverflow.com/questions/75785196/create-a-google-calendar-event-with-a-specified-google-meet-id-conferencedata-c

func (g *Google) calendarGetEvents(token string, calendarOptions GoogleCalendarOptions, calendarGetEventsOptions GoogleCalendarGetEventsOptions) ([]byte, error) {

	params := make(url.Values)
	params.Add("access_token", token)
	if !utils.IsEmpty(calendarGetEventsOptions.TimeMin) {
		params.Add("timeMin", calendarGetEventsOptions.TimeMin)
	}
	if !utils.IsEmpty(calendarGetEventsOptions.TimeMax) {
		params.Add("timeMax", calendarGetEventsOptions.TimeMax)
	}
	if !utils.IsEmpty(calendarGetEventsOptions.TimeZone) {
		params.Add("timeZone", calendarGetEventsOptions.TimeZone)
	}

	params.Add("singleEvents", strconv.FormatBool(calendarGetEventsOptions.SingleEvents))

	if !utils.IsEmpty(calendarGetEventsOptions.OrderBy) {
		if calendarGetEventsOptions.OrderBy == "startTime" {
			if calendarGetEventsOptions.SingleEvents {
				params.Add("orderBy", calendarGetEventsOptions.OrderBy)
			} else {
				return nil, errors.New("if orderBy=startTime singleEvents must be true")
			}

		} else {
			params.Add("orderBy", calendarGetEventsOptions.OrderBy)
		}
	}
	if !utils.IsEmpty(calendarGetEventsOptions.Q) {
		params.Add("q", calendarGetEventsOptions.Q)
	}

	u, err := url.Parse(googleCalendarURL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf(googleCalendarEvents, calendarOptions.ID))
	u.RawQuery = params.Encode()

	return utils.HttpGetRawWithHeaders(g.client, u.String(), nil)
}

func (g *Google) CustomCalendarGetEvents(googleOptions GoogleOptions, calendarOptions GoogleCalendarOptions, calendarGetEventsOptions GoogleCalendarGetEventsOptions) ([]byte, error) {

	r, err := g.refreshToken(googleOptions)
	if err != nil {
		return nil, err
	}
	g.logger.Debug("Access token => %s", r.AccessToken)

	return g.calendarGetEvents(r.AccessToken, calendarOptions, calendarGetEventsOptions)
}

func (g *Google) CalendarGetEvents(calendarOptions GoogleCalendarOptions, calendarGetEventsOptions GoogleCalendarGetEventsOptions) ([]byte, error) {
	return g.CustomCalendarGetEvents(g.options, calendarOptions, calendarGetEventsOptions)
}

// https://developers.google.com/calendar/api/v3/reference/events/insert

func (g *Google) CustomCalendarInsertEvent(googleOptions GoogleOptions, calendarOptions GoogleCalendarOptions, calendarInsertEventOptions GoogleCalendarInsertEventOptions) ([]byte, error) {

	r, err := g.refreshToken(googleOptions)
	if err != nil {
		return nil, err
	}
	g.logger.Debug("Access token => %s", r.AccessToken)

	params := make(url.Values)
	params.Add("access_token", r.AccessToken)
	if !utils.IsEmpty(calendarInsertEventOptions.SendUpdates) {
		params.Add("sendUpdates", calendarInsertEventOptions.SendUpdates)
	}
	params.Add("supportsAttachments", strconv.FormatBool(calendarInsertEventOptions.SupportsAttachments))
	params.Add("conferenceDataVersion", "1")

	var source *GoogleCalendarEventSource
	if !utils.IsEmpty(calendarInsertEventOptions.SourceTitle) || !utils.IsEmpty(calendarInsertEventOptions.SourceURL) {
		source = &GoogleCalendarEventSource{
			Title: calendarInsertEventOptions.SourceTitle,
			URL:   calendarInsertEventOptions.SourceURL,
		}
	}

	var conference *GoogleConferenceData
	if !utils.IsEmpty(calendarInsertEventOptions.ConferenceID) {

		entryVideo := &GoogleConferenceDataEntryPoint{
			EntryPointType: "video",
			URI:            fmt.Sprintf(googleMeetURL, calendarInsertEventOptions.ConferenceID),
			Label:          fmt.Sprintf(googleMeetLabel, calendarInsertEventOptions.ConferenceID),
		}
		conference = &GoogleConferenceData{
			ConferenceSolution: &GoogleConferenceDataSolution{
				Key: GoogleConferenceDataSolutionKey{
					Type: "hangoutsMeet",
				},
			},
			EntryPoints:  []*GoogleConferenceDataEntryPoint{entryVideo},
			ConferenceID: calendarInsertEventOptions.ConferenceID,
		}
	} else {
		requestID := uuid.New().String()
		conference = &GoogleConferenceData{
			CreateRequest: &GoogleConferenceDataCreateRequest{
				ConferenceSolutionKey: GoogleConferenceDataSolutionKey{
					Type: "hangoutsMeet",
				},
				RequestID: requestID,
			},
		}
	}

	event := &GoogleCalendarEvent{
		Summary:     calendarInsertEventOptions.Summary,
		Description: calendarInsertEventOptions.Description,
		Start: GoogleCalendarEventDataTime{
			DateTime: calendarInsertEventOptions.Start,
			TimeZone: calendarInsertEventOptions.TimeZone,
		},
		End: GoogleCalendarEventDataTime{
			DateTime: calendarInsertEventOptions.End,
			TimeZone: calendarInsertEventOptions.TimeZone,
		},
		EventType:               "default",
		Transparency:            "transparent",
		Visibility:              calendarInsertEventOptions.Visibility,
		Attendees:               []*GoogleCalendarEventAttendee{},
		GuestsCanInviteOthers:   true,
		GuestsCanModify:         false,
		GuestsCanSeeOtherGuests: true,
		Source:                  source,
		ConferenceData:          conference,
	}

	data, err := json.Marshal(event)
	if err != nil {
		return nil, err
	}

	u, err := url.Parse(googleCalendarURL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf(googleCalendarEvents, calendarOptions.ID))
	u.RawQuery = params.Encode()

	return utils.HttpPostRawWithHeaders(g.client, u.String(), nil, data)
}

func (g *Google) CalendarInsertEvent(calendarOptions GoogleCalendarOptions, calendarInsertEventOptions GoogleCalendarInsertEventOptions) ([]byte, error) {
	return g.CustomCalendarInsertEvent(g.options, calendarOptions, calendarInsertEventOptions)
}

// https://developers.google.com/calendar/api/v3/reference/events/delete

func (g *Google) calendarDeleteEvent(token string, calendarOptions GoogleCalendarOptions, calendarDeleteEventOptions GoogleCalendarDeleteEventOptions) ([]byte, error) {

	params := make(url.Values)
	params.Add("access_token", token)
	if !utils.IsEmpty(calendarDeleteEventOptions.SendUpdates) {
		params.Add("sendUpdates", calendarDeleteEventOptions.SendUpdates)
	}
	u, err := url.Parse(googleCalendarURL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf(googleCalendarDeleteEvent, calendarOptions.ID, calendarDeleteEventOptions.ID))
	u.RawQuery = params.Encode()

	return utils.HttpDeleteRawWithHeaders(g.client, u.String(), nil, nil)
}

func (g *Google) CustomCalendarDeleteEvent(googleOptions GoogleOptions, calendarOptions GoogleCalendarOptions, calendarDeleteEventOptions GoogleCalendarDeleteEventOptions) ([]byte, error) {

	r, err := g.refreshToken(googleOptions)
	if err != nil {
		return nil, err
	}
	g.logger.Debug("Access token => %s", r.AccessToken)

	return g.calendarDeleteEvent(r.AccessToken, calendarOptions, calendarDeleteEventOptions)
}

func (g *Google) CalendarDeleteEvent(calendarOptions GoogleCalendarOptions, calendarDeleteEventOptions GoogleCalendarDeleteEventOptions) ([]byte, error) {
	return g.CustomCalendarDeleteEvent(g.options, calendarOptions, calendarDeleteEventOptions)
}

func (g *Google) CustomCalendarDeleteEvents(googleOptions GoogleOptions, calendarOptions GoogleCalendarOptions, calendarGetEventsOptions GoogleCalendarGetEventsOptions) ([]byte, error) {

	r, err := g.refreshToken(googleOptions)
	if err != nil {
		return nil, err
	}
	g.logger.Debug("Access token => %s", r.AccessToken)

	data, err := g.calendarGetEvents(r.AccessToken, calendarOptions, calendarGetEventsOptions)
	if err != nil {
		return data, err
	}

	g.logger.Debug(string(data))

	var events GoogleCalendarEvents
	err = json.Unmarshal(data, &events)
	if err != nil {
		return data, err
	}

	for _, e := range events.Items {

		data, err = g.calendarDeleteEvent(r.AccessToken, calendarOptions, GoogleCalendarDeleteEventOptions{ID: e.ID})
		if err != nil {
			return data, err
		}
	}
	return nil, nil
}

func (g *Google) CalendarDeleteEvents(calendarOptions GoogleCalendarOptions, calendarGetEventsOptions GoogleCalendarGetEventsOptions) ([]byte, error) {
	return g.CustomCalendarDeleteEvents(g.options, calendarOptions, calendarGetEventsOptions)
}

// Google Meet REST API methods

func (g *Google) createMeetSpace(token string, meetOptions GoogleMeetOptions) ([]byte, error) {

	requestData := GoogleMeetSpaceRequest{
		Config: GoogleMeetSpaceConfig{
			AccessType: meetOptions.AccessType, // "TRUSTED" for org-only
		},
	}

	data, err := json.Marshal(requestData)
	if err != nil {
		return nil, err
	}

	headers := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", token),
		"Content-Type":  "application/json",
	}

	u, err := url.Parse(googleMeetAPIURL)
	if err != nil {
		return nil, err
	}
	u.Path = path.Join(u.Path, "/spaces")

	return utils.HttpPostRawWithHeaders(g.client, u.String(), headers, data)
}

func (g *Google) CustomCreateMeetSpace(googleOptions GoogleOptions, meetOptions GoogleMeetOptions) (*GoogleMeetSpaceResponse, error) {

	accessToken, err := g.getAccessToken(googleOptions)
	if err != nil {
		return nil, err
	}
	g.logger.Debug("Access token obtained successfully (length: %d chars)", len(accessToken))

	responseBytes, err := g.createMeetSpace(accessToken, meetOptions)
	if err != nil {
		return nil, err
	}

	var meetResponse GoogleMeetSpaceResponse
	err = json.Unmarshal(responseBytes, &meetResponse)
	if err != nil {
		return nil, err
	}

	return &meetResponse, nil
}

func (g *Google) CreateMeetSpace(meetOptions GoogleMeetOptions) (*GoogleMeetSpaceResponse, error) {
	return g.CustomCreateMeetSpace(g.options, meetOptions)
}

func (g *Google) DocsCopyDocument(calendarOptions GoogleDocsOptions) ([]byte, error) {
	r, err := g.refreshToken(g.options)
	if err != nil {
		return nil, err
	}
	g.logger.Debug("Access token => %s", r.AccessToken)

	params := make(url.Values)
	params.Add("access_token", r.AccessToken)

	u, err := url.Parse(googleDocsURL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, "files", calendarOptions.ID, "copy")
	u.RawQuery = params.Encode()

	return utils.HttpPostRawWithHeaders(g.client, u.String(), nil, nil)
}

func NewGoogle(options GoogleOptions, logger common.Logger) *Google {

	google := &Google{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
		logger:  logger,
	}
	return google
}
