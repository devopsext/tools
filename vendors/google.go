package vendors

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/url"
	"path"
	"strconv"

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

type GoogleCalendarGetEventsOptions struct {
	TimeMin            string
	TimeMax            string
	AlwaysIncludeEmail bool
	OrderBy            string
	Q                  string
	SingleEvents       bool
}

type GoogleCalendarOptions struct {
	ID string
}

type GoogleOptions struct {
	Timeout           int
	Insecure          bool
	OAuthClientID     string
	OAuthClientSecret string
	RefreshToken      string
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

const (
	googleOAuthURL       = "https://oauth2.googleapis.com"
	googleCalendarURL    = "https://www.googleapis.com/calendar/v3"
	googleCalendarEvents = "/calendars/%s/events"
	googleMeetURL        = "https://meet.google.com/%s"
	googleMeetLabel      = "meet.google.com/%s"
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

// https://developers.google.com/calendar/api/v3/reference/events/get
// https://stackoverflow.com/questions/75785196/create-a-google-calendar-event-with-a-specified-google-meet-id-conferencedata-c

func (g *Google) CustomCalendarGetEvents(googleOptions GoogleOptions, calendarOptions GoogleCalendarOptions, calendarGetEventsOptions GoogleCalendarGetEventsOptions) ([]byte, error) {

	r, err := g.refreshToken(googleOptions)
	if err != nil {
		return nil, err
	}
	g.logger.Debug("Access token => %s", r.AccessToken)

	params := make(url.Values)
	params.Add("access_token", r.AccessToken)
	if !utils.IsEmpty(calendarGetEventsOptions.TimeMin) {
		params.Add("timeMin", calendarGetEventsOptions.TimeMin)
	}
	if !utils.IsEmpty(calendarGetEventsOptions.TimeMax) {
		params.Add("timeMax", calendarGetEventsOptions.TimeMax)
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

	params.Add("alwaysIncludeEmail", strconv.FormatBool(calendarGetEventsOptions.AlwaysIncludeEmail))

	u, err := url.Parse(googleCalendarURL)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, fmt.Sprintf(googleCalendarEvents, calendarOptions.ID))
	u.RawQuery = params.Encode()

	return utils.HttpGetRawWithHeaders(g.client, u.String(), nil)
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

func NewGoogle(options GoogleOptions, logger common.Logger) *Google {

	google := &Google{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
		logger:  logger,
	}
	return google
}
