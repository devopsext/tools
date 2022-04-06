package messaging

import (
	"bytes"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"strings"

	"github.com/devopsext/utils"
)

type TelegramOptions struct {
	URL                 string
	Timeout             int
	DisableNotification string
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

func (t *Telegram) post(URL, contentType string, body bytes.Buffer, message string) (error, []byte) {

	reader := bytes.NewReader(body.Bytes())

	req, err := http.NewRequest("POST", URL, reader)
	if err != nil {
		return err, nil
	}

	req.Header.Set("Content-Type", contentType)

	resp, err := t.client.Do(req)
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

func (t *Telegram) SendMessage(URL, message string) (error, []byte) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("text", message); err != nil {
		return err, nil
	}

	if err := w.WriteField("parse_mode", "HTML"); err != nil {
		return err, nil
	}

	if err := w.WriteField("disable_web_page_preview", "true"); err != nil {
		return err, nil
	}

	if err := w.WriteField("disable_notification", t.options.DisableNotification); err != nil {
		return err, nil
	}

	if err := w.Close(); err != nil {
		return err, nil
	}

	return t.post(URL, w.FormDataContentType(), body, message)
}

func (t *Telegram) SendPhoto(URL, message, fileName string, photo []byte) (error, []byte) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("caption", message); err != nil {
		return err, nil
	}

	if err := w.WriteField("parse_mode", "HTML"); err != nil {
		return err, nil
	}

	if err := w.WriteField("disable_web_page_preview", "true"); err != nil {
		return err, nil
	}

	if err := w.WriteField("disable_notification", t.options.DisableNotification); err != nil {
		return err, nil
	}

	fw, err := w.CreateFormFile("photo", fileName)
	if err != nil {
		return err, nil
	}

	if _, err := fw.Write(photo); err != nil {
		return err, nil
	}

	if err := w.Close(); err != nil {
		return err, nil
	}

	return t.post(URL, w.FormDataContentType(), body, message)
}

func NewTelegram(options TelegramOptions) *Telegram {

	return &Telegram{
		client:  utils.NewHttpInsecureClient(options.Timeout),
		options: options,
	}
}
