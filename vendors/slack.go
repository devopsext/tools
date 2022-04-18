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
	"strings"
)

//go:embed slack.tmpl
var msgTemplate string

const baseURL = "https://slack.com/api/"

const (
	filesUpload     = "files.upload"
	chatPostMessage = "chat.postMessage"
)

type SlackOptions struct {
	Token       string
	Channels    []string
	Timeout     int
	Insecure    bool
	Message     string
	Title       string
	FileName    string
	Content     string // content or path to file
	Output      string // path to output if empty to stdout
	OutputQuery string
}

type Message struct {
	Channel  string
	Title    string
	Message  string
	ImageURL string
	Content  string
}

type Slack struct {
	client  *http.Client
	options SlackOptions
}

func (s *Slack) Send() ([]byte, error) {
	return s.SendCustom("", s.options.Message, s.options.Title, s.options.Content)
	//return s.SendMessage(s.options.Channels[0], s.options.Title, s.options.Message, s.options.Content)
}

func (s *Slack) SendFile() ([]byte, error) {
	return s.SendCustomFile("", s.options.Message, s.options.FileName, s.options.Title, []byte(s.options.Content))
}

func (s *Slack) SendMessage(channel string) ([]byte, error) {
	return s.sendMessage(channel, s.options.Title, s.options.Message, s.options.Content)
}

func (s *Slack) sendMessage(channel, title, message, imageUrl string) ([]byte, error) {
	m, err := s.prepareMessage(channel, title, message, imageUrl)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	b, err := s.postJson(s.apiURL(chatPostMessage), q, *m)
	if err != nil {
		return nil, err
	}
	return b, nil
}

func (s *Slack) SendCustom(URL, message, title, content string) ([]byte, error) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("initial_comment", message); err != nil {
		return nil, err
	}

	if err := w.WriteField("title", title); err != nil {
		return nil, err
	}

	if err := w.WriteField("content", content); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Add("channels", strings.Join(s.options.Channels, ","))

	return s.postBody(s.apiURL(filesUpload), q, w.FormDataContentType(), body)
}

func (s *Slack) SendCustomFile(URL, message, fileName, title string, content []byte) ([]byte, error) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("initial_comment", message); err != nil {
		return nil, err
	}

	if err := w.WriteField("title", title); err != nil {
		return nil, err
	}

	fw, err := w.CreateFormFile("file", fileName)
	if err != nil {
		return nil, err
	}

	if _, err := fw.Write(content); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	q := url.Values{}
	q.Add("channels", strings.Join(s.options.Channels, ","))

	return s.postBody(s.apiURL(filesUpload), q, w.FormDataContentType(), body)
}

func NewSlack(options SlackOptions) *Slack {

	return &Slack{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}

func (s *Slack) prepareMessage(channel, title, message, imageUrl string) (*bytes.Buffer, error) {
	m := &Message{
		Message:  message,
		Title:    title,
		ImageURL: imageUrl,
		Channel:  channel,
	}

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

// assume that url is => https://slack.com/api/files.upload?token=%s&channels=%s
func (s *Slack) getToken(URL string) string {

	u, err := url.Parse(URL)
	if err != nil {
		return ""
	}
	return u.Query().Get("token")
}

func (s *Slack) getChannel(URL string) string {

	u, err := url.Parse(URL)
	if err != nil {
		return ""
	}
	return u.Query().Get("channels")
}

func (s *Slack) postBody(URL string, query url.Values, contentType string, body bytes.Buffer) ([]byte, error) {

	reader := bytes.NewReader(body.Bytes())

	req, err := http.NewRequest("POST", URL, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Authorization", "Bearer "+s.options.Token)
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

func (s *Slack) postJson(URL string, query url.Values, body bytes.Buffer) ([]byte, error) {
	return s.postBody(URL, query, "application/json; charset=utf-8", body)
}

func (s *Slack) apiURL(cmd string) string {
	return baseURL + cmd
}
