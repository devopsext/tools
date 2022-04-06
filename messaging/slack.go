package messaging

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"

	"github.com/devopsext/utils"
)

type SlackOptions struct {
	URL     string
	Timeout int
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

func (s *Slack) post(URL, contentType string, body bytes.Buffer, message string) (error, []byte) {

	reader := bytes.NewReader(body.Bytes())

	req, err := http.NewRequest("POST", URL, reader)
	if err != nil {
		return err, nil
	}

	req.Header.Set("Content-Type", contentType)

	resp, err := s.client.Do(req)
	if err != nil {
		return err, nil
	}
	defer resp.Body.Close()

	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err, nil
	}
	return nil, b
}

func (s *Slack) SendMessage(URL, message, title, content string) (error, []byte) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("initial_comment", message); err != nil {
		return err, nil
	}

	if err := w.WriteField("title", title); err != nil {
		return err, nil
	}

	if err := w.WriteField("content", content); err != nil {
		return err, nil
	}

	if err := w.Close(); err != nil {
		return err, nil
	}
	return s.post(URL, w.FormDataContentType(), body, message)
}

func (s *Slack) SendPhoto(URL, message, fileName, title string, photo []byte) (error, []byte) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("initial_comment", message); err != nil {
		return err, nil
	}

	if err := w.WriteField("title", title); err != nil {
		return err, nil
	}

	fw, err := w.CreateFormFile("file", fileName)
	if err != nil {
		return err, nil
	}

	if _, err := fw.Write(photo); err != nil {
		return err, nil
	}

	if err := w.Close(); err != nil {
		return err, nil
	}
	return s.post(URL, w.FormDataContentType(), body, message)
}

func NewSlack(options SlackOptions) *Slack {

	return &Slack{
		client:  utils.NewHttpClient(options.Timeout, false),
		options: options,
	}
}
