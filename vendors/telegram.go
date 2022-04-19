package vendors

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/devopsext/utils"
)

type TelegramOptions struct {
	URL                 string
	Timeout             int
	Insecure            bool
	DisableNotification bool
	Message             string
	FileName            string
	Content             string // content or path to file
}

type Telegram struct {
	client  *http.Client
	options TelegramOptions
}

// assume that url is => https://api.telegram.org/botID:botToken/sendMessage?chat_id=%s
func (t *Telegram) getBotID(URL string) string {

	arr := strings.Split(URL, "/bot")
	if len(arr) > 1 {
		arr = strings.Split(arr[1], ":")
		if len(arr) > 0 {
			return arr[0]
		}
	}
	return ""
}

func (t *Telegram) getBotToken(URL string) string {

	arr := strings.Split(URL, "/bot")
	if len(arr) > 1 {
		arr = strings.Split(arr[1], ":")
		if len(arr) > 1 {
			return arr[1]
		}
	}
	return ""
}

func (t *Telegram) getChatID(URL string) string {

	u, err := url.Parse(URL)
	if err != nil {
		return ""
	}
	return u.Query().Get("chat_id")
}

func (t *Telegram) getSendPhotoURL(URL string) string {
	return strings.Replace(URL, "sendMessage", "sendPhoto", -1)
}

func (t *Telegram) getSendDocumentURL(URL string) string {
	return strings.Replace(URL, "sendMessage", "sendDocument", -1)
}

func (t *Telegram) post(URL, contentType string, body bytes.Buffer, message string) ([]byte, error) {

	reader := bytes.NewReader(body.Bytes())

	req, err := http.NewRequest("POST", URL, reader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", contentType)

	resp, err := t.client.Do(req)
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

func (t *Telegram) SendCustom(opts TelegramOptions) ([]byte, error) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("text", opts.Message); err != nil {
		return nil, err
	}

	if err := w.WriteField("parse_mode", "HTML"); err != nil {
		return nil, err
	}

	if err := w.WriteField("disable_web_page_preview", "true"); err != nil {
		return nil, err
	}

	if err := w.WriteField("disable_notification", strconv.FormatBool(t.options.DisableNotification)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}

	return t.post(opts.URL, w.FormDataContentType(), body, opts.Message)
}

func (t *Telegram) SendCustomPhoto(opts TelegramOptions) ([]byte, error) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("caption", opts.Message); err != nil {
		return nil, err
	}

	if err := w.WriteField("parse_mode", "HTML"); err != nil {
		return nil, err
	}

	if err := w.WriteField("disable_web_page_preview", "true"); err != nil {
		return nil, err
	}

	if err := w.WriteField("disable_notification", strconv.FormatBool(t.options.DisableNotification)); err != nil {
		return nil, err
	}

	fw, err := w.CreateFormFile("photo", opts.FileName)
	if err != nil {
		return nil, err
	}

	if _, err := fw.Write([]byte(opts.Content)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}
	return t.post(t.getSendPhotoURL(opts.URL), w.FormDataContentType(), body, opts.Message)
}

func (t *Telegram) SendCustomDocument(opts TelegramOptions) ([]byte, error) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("caption", opts.Message); err != nil {
		return nil, err
	}

	if err := w.WriteField("parse_mode", "HTML"); err != nil {
		return nil, err
	}

	if err := w.WriteField("disable_web_page_preview", "true"); err != nil {
		return nil, err
	}

	if err := w.WriteField("disable_notification", strconv.FormatBool(t.options.DisableNotification)); err != nil {
		return nil, err
	}

	fw, err := w.CreateFormFile("document", opts.FileName)
	if err != nil {
		return nil, err
	}

	if _, err := fw.Write([]byte(opts.Content)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}
	return t.post(t.getSendDocumentURL(opts.URL), w.FormDataContentType(), body, opts.Message)
}

func (t *Telegram) Send() ([]byte, error) {
	return t.SendCustom(t.options)
}

func (t *Telegram) SendPhoto() ([]byte, error) {
	return t.SendCustomPhoto(t.options)
}

func (t *Telegram) SendDocument() ([]byte, error) {
	return t.SendCustomDocument(t.options)
}

func NewTelegram(options TelegramOptions) *Telegram {

	return &Telegram{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
