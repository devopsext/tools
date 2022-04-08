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
	Output              string // path to output if empty to stdout
	OutputQuery         string
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

func (t *Telegram) SendCustom(URL, message, title, content string) ([]byte, error) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("text", message); err != nil {
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

	return t.post(URL, w.FormDataContentType(), body, message)
}

func (t *Telegram) SendCustomFile(URL, message, fileName, title string, file []byte) ([]byte, error) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("caption", message); err != nil {
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

	fw, err := w.CreateFormFile("photo", fileName)
	if err != nil {
		return nil, err
	}

	if _, err := fw.Write(file); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}
	return t.post(t.getSendPhotoURL(URL), w.FormDataContentType(), body, message)
}

func (t *Telegram) Send() ([]byte, error) {
	return t.SendCustom(t.options.URL, t.options.Message, "", t.options.Content)
}

func (t *Telegram) SendFile() ([]byte, error) {
	return t.SendCustomFile(t.options.URL, t.options.Message, t.options.FileName, "", []byte(t.options.Content))
}

func NewTelegram(options TelegramOptions) *Telegram {

	return &Telegram{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
