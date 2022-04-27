package vendors

import (
	"bytes"
	"fmt"
	"mime/multipart"
	"net/http"
	"strconv"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
)

// assume that url is => https://api.telegram.org/botID:botToken/sendMessage?chat_id=%s

const (
	telegramSendMessageURL  = "https://api.telegram.org/bot%s/sendMessage?chat_id=%s"
	telegramSendPhotoURL    = "https://api.telegram.org/bot%s/sendPhoto?chat_id=%s"
	telegramSendDocumentURL = "https://api.telegram.org/bot%s/sendDocument?chat_id=%s"
)

type TelegramMessageOptions struct {
	Text string
}

type TelegramPhotoOptions struct {
	Caption string
	Name    string
	Content string
}

type TelegramDocumentOptions struct {
	Caption string
	Name    string
	Content string
}

type TelegramOptions struct {
	IDToken               string
	ChatID                string
	Timeout               int
	Insecure              bool
	DisableNotification   bool
	ParseMode             string
	DisableWebPagePreview bool
	MessageOptions        *TelegramMessageOptions
	PhotoOptions          *TelegramPhotoOptions
	DocumentOptions       *TelegramDocumentOptions
}

type Telegram struct {
	client  *http.Client
	options TelegramOptions
}

func (t *Telegram) getSendMessageURL(opts TelegramOptions) string {
	return fmt.Sprintf(telegramSendMessageURL, opts.IDToken, opts.ChatID)
}

func (t *Telegram) getSendPhotoURL(opts TelegramOptions) string {
	return fmt.Sprintf(telegramSendPhotoURL, opts.IDToken, opts.ChatID)
}

func (t *Telegram) getSendDocumentURL(opts TelegramOptions) string {
	return fmt.Sprintf(telegramSendDocumentURL, opts.IDToken, opts.ChatID)
}

func (t *Telegram) getDefaultParseMode(parseMode string) string {

	if utils.IsEmpty(parseMode) {
		return "HTML"
	}
	return parseMode
}

func (t *Telegram) SendCustomMessage(opts TelegramOptions) ([]byte, error) {

	if opts.MessageOptions == nil {
		return nil, fmt.Errorf("options are not enough")
	}

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("text", opts.MessageOptions.Text); err != nil {
		return nil, err
	}

	if err := w.WriteField("parse_mode", t.getDefaultParseMode(opts.ParseMode)); err != nil {
		return nil, err
	}

	if err := w.WriteField("disable_web_page_preview", strconv.FormatBool(opts.DisableWebPagePreview)); err != nil {
		return nil, err
	}

	if err := w.WriteField("disable_notification", strconv.FormatBool(opts.DisableNotification)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}
	return common.HttpPostRaw(t.client, t.getSendMessageURL(opts), w.FormDataContentType(), "", body.Bytes())
}

func (t *Telegram) SendCustomPhoto(opts TelegramOptions) ([]byte, error) {

	if opts.PhotoOptions == nil {
		return nil, fmt.Errorf("options are not enough")
	}

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("caption", opts.PhotoOptions.Caption); err != nil {
		return nil, err
	}

	if err := w.WriteField("parse_mode", t.getDefaultParseMode(opts.ParseMode)); err != nil {
		return nil, err
	}

	if err := w.WriteField("disable_web_page_preview", strconv.FormatBool(opts.DisableWebPagePreview)); err != nil {
		return nil, err
	}

	if err := w.WriteField("disable_notification", strconv.FormatBool(opts.DisableNotification)); err != nil {
		return nil, err
	}

	fw, err := w.CreateFormFile("photo", opts.PhotoOptions.Name)
	if err != nil {
		return nil, err
	}

	if _, err := fw.Write([]byte(opts.PhotoOptions.Content)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}
	return common.HttpPostRaw(t.client, t.getSendPhotoURL(opts), w.FormDataContentType(), "", body.Bytes())
}

func (t *Telegram) SendCustomDocument(opts TelegramOptions) ([]byte, error) {

	if opts.DocumentOptions == nil {
		return nil, fmt.Errorf("options are not enough")
	}
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("caption", opts.DocumentOptions.Caption); err != nil {
		return nil, err
	}

	if err := w.WriteField("parse_mode", t.getDefaultParseMode(opts.ParseMode)); err != nil {
		return nil, err
	}

	if err := w.WriteField("disable_web_page_preview", strconv.FormatBool(opts.DisableWebPagePreview)); err != nil {
		return nil, err
	}

	if err := w.WriteField("disable_notification", strconv.FormatBool(opts.DisableNotification)); err != nil {
		return nil, err
	}

	fw, err := w.CreateFormFile("document", opts.DocumentOptions.Name)
	if err != nil {
		return nil, err
	}

	if _, err := fw.Write([]byte(opts.DocumentOptions.Content)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}
	return common.HttpPostRaw(t.client, t.getSendDocumentURL(opts), w.FormDataContentType(), "", body.Bytes())
}

func (t *Telegram) SendMessage() ([]byte, error) {
	return t.SendCustomMessage(t.options)
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
