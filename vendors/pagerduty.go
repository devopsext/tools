package vendors

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
)

type PagerDutyCreateIncidentOptions struct {
	From string
}

type PagerDutyIncidentOptions struct {
	Title        string
	Body         string
	Urgency      string
	ServiceID    string
	PriorityID   string
	IncidentType string
}
type PagerDutyIncidentNoteOptions struct {
	IncidentID  string
	NoteContent string
}

type PagerDutyGetIncidentsOptions struct {
	Key   string
	Limit int
}

type PagerDutyService struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type PagerDutyPriority struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type PagerDutyBody struct {
	Type    string `json:"type"`
	Details string `json:"details"`
}

type PagerDutyIncidentType struct {
	Name string `json:"name"`
}

type PagerDutyIncident struct {
	Type         string                 `json:"type"`
	Title        string                 `json:"title"`
	Urgency      string                 `json:"urgency,omitempty"`
	Service      *PagerDutyService      `json:"service"`
	Priority     *PagerDutyPriority     `json:"priority,omitempty"`
	Body         *PagerDutyBody         `json:"body,omitempty"`
	IncidentType *PagerDutyIncidentType `json:"incident_type,omitempty"`
}

type PagerDutyIncidentRequest struct {
	Incident *PagerDutyIncident `json:"incident"`
}

type PagerDutyIncidentNote struct {
	Content string `json:"content,omitempty"`
}

type PagerDutyIncidentNoteRequest struct {
	Note *PagerDutyIncidentNote `json:"note"`
}

type PagerDutyOptions struct {
	Timeout  int
	Insecure bool
	URL      string
	Token    string
}

type PagerDuty struct {
	client  *http.Client
	options PagerDutyOptions
	logger  common.Logger
}

const (
	pagerDutyContentType       = "application/json"
	pagerDutyIncidentsPath     = "/incidents"
	pagerDutyIncidentNotesPath = "/notes"
)

func (pd *PagerDuty) getAuth(options PagerDutyOptions) string {
	auth := ""
	if !utils.IsEmpty(options.Token) {
		auth = fmt.Sprintf("Token token=%s", options.Token)
	}
	return auth
}

func (pd *PagerDuty) CustomCreateIncident(options PagerDutyOptions, incidentOptions PagerDutyIncidentOptions, createOptions PagerDutyCreateIncidentOptions) ([]byte, error) {

	u, err := url.Parse(options.URL)
	if err != nil {
		return nil, err
	}

	var params = make(url.Values)
	if !utils.IsEmpty(createOptions.From) {
		params.Add("from", createOptions.From)
	}
	u.RawQuery = params.Encode()
	u.Path = path.Join(u.Path, pagerDutyIncidentsPath)

	var priority *PagerDutyPriority
	if !utils.IsEmpty(incidentOptions.PriorityID) {
		priority = &PagerDutyPriority{
			Type: "priority_reference",
			ID:   incidentOptions.PriorityID,
		}
	}

	var body *PagerDutyBody
	if !utils.IsEmpty(incidentOptions.Body) {
		body = &PagerDutyBody{
			Type:    "incident_body",
			Details: incidentOptions.Body,
		}
	}

	var incidentType *PagerDutyIncidentType
	if !utils.IsEmpty(incidentOptions.IncidentType) {
		incidentType = &PagerDutyIncidentType{
			Name: incidentOptions.IncidentType,
		}
	}

	incident := &PagerDutyIncident{
		Type:    "incident",
		Title:   incidentOptions.Title,
		Urgency: incidentOptions.Urgency,
		Service: &PagerDutyService{
			Type: "service_reference",
			ID:   incidentOptions.ServiceID,
		},
		Priority:     priority,
		Body:         body,
		IncidentType: incidentType,
	}

	request := &PagerDutyIncidentRequest{
		Incident: incident,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	return utils.HttpPostRaw(pd.client, u.String(), pagerDutyContentType, pd.getAuth(options), data)
}

func (pd *PagerDuty) CreateIncident(incidentOptions PagerDutyIncidentOptions, createOptions PagerDutyCreateIncidentOptions) ([]byte, error) {
	return pd.CustomCreateIncident(pd.options, incidentOptions, createOptions)
}

func (pd *PagerDuty) CustomCreateIncidentNote(options PagerDutyOptions, noteOptions PagerDutyIncidentNoteOptions, createOptions PagerDutyCreateIncidentOptions) ([]byte, error) {

	u, err := url.Parse(options.URL)
	if err != nil {
		return nil, err
	}

	var params = make(url.Values)
	if !utils.IsEmpty(createOptions.From) {
		params.Add("from", createOptions.From)
	}
	u.RawQuery = params.Encode()
	u.Path = path.Join(u.Path, pagerDutyIncidentsPath, noteOptions.IncidentID, pagerDutyIncidentNotesPath)

	incidentNote := &PagerDutyIncidentNote{
		Content: noteOptions.NoteContent,
	}

	request := &PagerDutyIncidentNoteRequest{
		Note: incidentNote,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return nil, err
	}

	return utils.HttpPostRaw(pd.client, u.String(), pagerDutyContentType, pd.getAuth(options), data)
}

func (pd *PagerDuty) CreateIncidentNote(noteOptions PagerDutyIncidentNoteOptions, createOptions PagerDutyCreateIncidentOptions) ([]byte, error) {
	return pd.CustomCreateIncidentNote(pd.options, noteOptions, createOptions)
}

func (pd *PagerDuty) CustomGetIncidents(options PagerDutyOptions, getOptions PagerDutyGetIncidentsOptions) ([]byte, error) {

	u, err := url.Parse(options.URL)
	if err != nil {
		return nil, err
	}

	var params = make(url.Values)
	if !utils.IsEmpty(getOptions.Key) {
		params.Add("incident_key", getOptions.Key)
	}
	if getOptions.Limit > 0 {
		params.Add("limit", fmt.Sprintf("%d", getOptions.Limit))
	}
	u.RawQuery = params.Encode()
	u.Path = path.Join(u.Path, pagerDutyIncidentsPath)

	return utils.HttpGetRaw(pd.client, u.String(), pagerDutyContentType, pd.getAuth(options))
}
func (pd *PagerDuty) GetIncidents(getOptions PagerDutyGetIncidentsOptions) ([]byte, error) {
	return pd.CustomGetIncidents(pd.options, getOptions)
}

func NewPagerDuty(options PagerDutyOptions, logger common.Logger) *PagerDuty {

	return &PagerDuty{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
		logger:  logger,
	}
}
