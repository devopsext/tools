package messaging

import (
	"bytes"
	"errors"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/devopsext/utils"
)

type SlackOptions struct {
	URL      string
	Timeout  int
	Insecure bool
	Message  string
	Title    string
	FileName string
	Content  string // content or path to file
	Output   string // path to output if empty to stdout
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

func (s *Slack) post(URL, contentType string, body bytes.Buffer, message string) ([]byte, error) {

	reader := bytes.NewReader(body.Bytes())

	req, err := http.NewRequest("POST", URL, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

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
	return s.post(URL, w.FormDataContentType(), body, message)
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
	return s.post(URL, w.FormDataContentType(), body, message)
}

func (s *Slack) Send() ([]byte, error) {
	return s.SendCustom(s.options.URL, s.options.Message, s.options.Title, s.options.Content)
}

func (s *Slack) SendFile() ([]byte, error) {

	var bytes []byte
	fileName := s.options.FileName

	_, err := os.Stat(s.options.Content)
	if err == nil {

		bytes, err = ioutil.ReadFile(s.options.Content)
		if err != nil {
			return nil, err
		}

		if utils.IsEmpty(fileName) {
			fileName = strings.TrimSuffix(s.options.Content, filepath.Ext(s.options.Content))
		}
	} else {
		bytes = []byte(s.options.Content)
	}

	if len(bytes) == 0 {
		return nil, errors.New("SendFile content is not defined")
	}
	return s.SendCustomFile(s.options.URL, s.options.Message, fileName, s.options.Title, bytes)
}

func NewSlack(options SlackOptions) *Slack {

	return &Slack{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
