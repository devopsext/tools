package vendors

import (
	"bytes"
	_ "embed"
	"github.com/devopsext/utils"
	"html/template"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
)

//go:embed slack.tmpl
var msgTemplate string

const baseURL = "https://slack.com/api/"

const (
	filesUpload     = "files.upload"
	chatPostMessage = "chat.postMessage"
)

type SlackOptions struct {
	Timeout     int
	Insecure    bool
	Token       string
	Channel     string
	Title       string
	Message     string
	FileName    string
	FileContent string // content or path to file
	ImageURL    string
}

type SlackOutputOptions struct {
	Output      string // path to output if empty to stdout
	OutputQuery string
}

type SlackMessage struct {
	Token       string
	Channel     string
	ParentTS    string
	Title       string
	Message     string
	ImageURL    string
	FileName    string
	FileContent string
}

type Slack struct {
	client  *http.Client
	options SlackOptions
}

func (s *Slack) Send() ([]byte, error) {
	m := SlackMessage{
		Token:       s.options.Token,
		Channel:     s.options.Channel,
		Title:       s.options.Title,
		Message:     s.options.Message,
		FileContent: s.options.FileContent,
	}
	return s.SendCustom(m)
}

func (s *Slack) SendFile() ([]byte, error) {
	m := SlackMessage{
		Token:       s.options.Token,
		Channel:     s.options.Channel,
		Title:       s.options.Title,
		Message:     s.options.Message,
		FileContent: s.options.FileContent,
	}
	return s.SendCustomFile(m)
}

func (s *Slack) SendMessage() ([]byte, error) {
	m := SlackMessage{
		Token:    s.options.Token,
		Channel:  s.options.Channel,
		Title:    s.options.Title,
		Message:  s.options.Message,
		ImageURL: s.options.ImageURL,
	}
	return s.sendMessage(m)
}

func (s *Slack) SendMessageCustom(m SlackMessage) ([]byte, error) {
	return s.sendMessage(m)
}

func (s *Slack) sendMessage(m SlackMessage) ([]byte, error) {
	jsonMsg, err := s.prepareMessage(m)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	b, err := s.post(m.Token, s.apiURL(chatPostMessage), q, "application/json; charset=utf-8", *jsonMsg)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Slack) SendCustom(m SlackMessage) ([]byte, error) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("initial_comment", m.Message); err != nil {
		return nil, err
	}

	if err := w.WriteField("title", m.Title); err != nil {
		return nil, err
	}

	if err := w.WriteField("content", m.FileContent); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Add("channels", s.options.Channel)

	return s.post(m.Token, s.apiURL(filesUpload), q, w.FormDataContentType(), body)
}

func (s *Slack) SendCustomFile(m SlackMessage) ([]byte, error) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("initial_comment", m.Message); err != nil {
		return nil, err
	}

	if err := w.WriteField("title", m.Title); err != nil {
		return nil, err
	}

	fw, err := w.CreateFormFile("file", m.FileName)
	if err != nil {
		return nil, err
	}

	if _, err := fw.Write([]byte(m.FileContent)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Add("channels", m.Channel)

	return s.post(m.Token, s.apiURL(filesUpload), q, w.FormDataContentType(), body)
}

func NewSlack(options SlackOptions) *Slack {

	return &Slack{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}

func (s *Slack) prepareMessage(m SlackMessage) (*bytes.Buffer, error) {

	t, err := template.New("slack").Parse(msgTemplate)
	if err != nil {
		return nil, err
	}

	b := &bytes.Buffer{}
	if err := t.Execute(b, m); err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Slack) post(token string, URL string, query url.Values, contentType string, body bytes.Buffer) ([]byte, error) {

	reader := bytes.NewReader(body.Bytes())

	req, err := http.NewRequest("POST", URL, reader)
	if err != nil {
		return nil, err
	}

	if token == "" {
		token = s.options.Token
	}
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+token)
	req.URL.RawQuery = query.Encode()

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Slack) apiURL(cmd string) string {
	return baseURL + cmd
}
