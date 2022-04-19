package vendors

import (
	"bytes"
	"mime/multipart"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/devopsext/tools/common"
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

	return common.HttpPostRaw(t.client, opts.URL, w.FormDataContentType(), "", body.Bytes())
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
	return common.HttpPostRaw(t.client, t.getSendPhotoURL(opts.URL), w.FormDataContentType(), "", body.Bytes())
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
	return common.HttpPostRaw(t.client, t.getSendDocumentURL(opts.URL), w.FormDataContentType(), "", body.Bytes())
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
