package vendors

import (
	"bytes"
	"github.com/devopsext/utils"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"
)

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

type Slack struct {
	client  *http.Client
	options SlackOptions
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

	URL = s.filesUploadURL()

	return s.postBody(URL, q, w.FormDataContentType(), body)
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

	URL = s.filesUploadURL()

	return s.postBody(URL, q, w.FormDataContentType(), body)
}

func (s *Slack) Send() ([]byte, error) {
	return s.SendCustom("", s.options.Message, s.options.Title, s.options.Content)
}

func (s *Slack) SendFile() ([]byte, error) {
	return s.SendCustomFile("", s.options.Message, s.options.FileName, s.options.Title, []byte(s.options.Content))
}

func (s *Slack) filesUploadURL() string {
	return baseURL + filesUpload
}

func (s *Slack) chatPostMessageURL() string {
	return baseURL + chatPostMessage
}

func NewSlack(options SlackOptions) *Slack {

	return &Slack{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
