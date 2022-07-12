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

func (t *Telegram) CustomSendMessage(telegramOptions TelegramOptions, messageOptions TelegramMessageOptions) ([]byte, error) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("text", messageOptions.Text); err != nil {
		return nil, err
	}

	if err := w.WriteField("parse_mode", t.getDefaultParseMode(telegramOptions.ParseMode)); err != nil {
		return nil, err
	}

	if err := w.WriteField("disable_web_page_preview", strconv.FormatBool(telegramOptions.DisableWebPagePreview)); err != nil {
		return nil, err
	}

	if err := w.WriteField("disable_notification", strconv.FormatBool(telegramOptions.DisableNotification)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}
	return common.HttpPostRaw(t.client, t.getSendMessageURL(telegramOptions), w.FormDataContentType(), "", body.Bytes())
}

func (t *Telegram) SendMessage(options TelegramMessageOptions) ([]byte, error) {
	return t.CustomSendMessage(t.options, options)
}

func (t *Telegram) CustomSendPhoto(telegramOptions TelegramOptions, photoOptions TelegramPhotoOptions) ([]byte, error) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("caption", photoOptions.Caption); err != nil {
		return nil, err
	}

	if err := w.WriteField("parse_mode", t.getDefaultParseMode(telegramOptions.ParseMode)); err != nil {
		return nil, err
	}

	if err := w.WriteField("disable_web_page_preview", strconv.FormatBool(telegramOptions.DisableWebPagePreview)); err != nil {
		return nil, err
	}

	if err := w.WriteField("disable_notification", strconv.FormatBool(telegramOptions.DisableNotification)); err != nil {
		return nil, err
	}

	fw, err := w.CreateFormFile("photo", photoOptions.Name)
	if err != nil {
		return nil, err
	}

	if _, err := fw.Write([]byte(photoOptions.Content)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}
	return common.HttpPostRaw(t.client, t.getSendPhotoURL(telegramOptions), w.FormDataContentType(), "", body.Bytes())
}

func (t *Telegram) SendPhoto(options TelegramPhotoOptions) ([]byte, error) {
	return t.CustomSendPhoto(t.options, options)
}

func (t *Telegram) CustomSendDocument(telegramOptions TelegramOptions, documentOptions TelegramDocumentOptions) ([]byte, error) {

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	defer func() {
		w.Close()
	}()

	if err := w.WriteField("caption", documentOptions.Caption); err != nil {
		return nil, err
	}

	if err := w.WriteField("parse_mode", t.getDefaultParseMode(telegramOptions.ParseMode)); err != nil {
		return nil, err
	}

	if err := w.WriteField("disable_web_page_preview", strconv.FormatBool(telegramOptions.DisableWebPagePreview)); err != nil {
		return nil, err
	}

	if err := w.WriteField("disable_notification", strconv.FormatBool(telegramOptions.DisableNotification)); err != nil {
		return nil, err
	}

	fw, err := w.CreateFormFile("document", documentOptions.Name)
	if err != nil {
		return nil, err
	}

	if _, err := fw.Write([]byte(documentOptions.Content)); err != nil {
		return nil, err
	}

	if err := w.Close(); err != nil {
		return nil, err
	}
	return common.HttpPostRaw(t.client, t.getSendDocumentURL(telegramOptions), w.FormDataContentType(), "", body.Bytes())
}

func (t *Telegram) SendDocument(options TelegramDocumentOptions) ([]byte, error) {
	return t.CustomSendDocument(t.options, options)
}

func NewTelegram(options TelegramOptions) *Telegram {

	return &Telegram{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
}
