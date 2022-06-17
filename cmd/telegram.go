package cmd

import (
	"path/filepath"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var telegramOptions = vendors.TelegramOptions{
	IDToken:               envGet("TELEGRAM_ID_TOKEN", "").(string),
	ChatID:                envGet("TELEGRAM_CHAT_ID", "").(string),
	Insecure:              envGet("TELEGRAM_INSECURE", false).(bool),
	Timeout:               envGet("TELEGRAM_TIMEOUT", 30).(int),
	DisableNotification:   envGet("TELEGRAM_DISABLE_NOTIFICATION", true).(bool),
	ParseMode:             envGet("TELEGRAM_PARSE_MODE", "HTML").(string),
	DisableWebPagePreview: envGet("TELEGRAM_DISABLE_WEB_PAGE_PREVIEW", true).(bool),
}

var telegramMessageOptions = vendors.TelegramMessageOptions{
	Text: envGet("TELEGRAM_MESSAGE_TEXT", "").(string),
}

var telegramPhotoOptions = vendors.TelegramPhotoOptions{
	Caption: envGet("TELEGRAM_PHOTO_CAPTION", "").(string),
	Name:    envGet("TELEGRAM_PHOTO_NAME", "").(string),
	Content: envGet("TELEGRAM_PHOTO_CONTENT", "").(string),
}

var telegramDocumentOptions = vendors.TelegramDocumentOptions{
	Caption: envGet("TELEGRAM_DOCUMENT_CAPTION", "").(string),
	Name:    envGet("TELEGRAM_DOCUMENT_NAME", "").(string),
	Content: envGet("TELEGRAM_DOCUMENT_CONTENT", "").(string),
}

var telegramOutput = common.OutputOptions{
	Output: envGet("TELEGRAM_OUTPUT", "").(string),
	Query:  envGet("TELEGRAM_OUTPUT_QUERY", "").(string),
}

func telegramNew(stdout *common.Stdout) *vendors.Telegram {

	common.Debug("Telegram", telegramOptions, stdout)
	common.Debug("Telegram", telegramOutput, stdout)

	telegram := vendors.NewTelegram(telegramOptions)
	if telegram == nil {
		stdout.Panic("No telegram")
	}
	return telegram
}

func NewTelegramCommand() *cobra.Command {

	telegramCmd := cobra.Command{
		Use:   "telegram",
		Short: "Telegram tools",
	}

	flags := telegramCmd.PersistentFlags()
	flags.StringVar(&telegramOptions.IDToken, "telegram-id-token", telegramOptions.IDToken, "Telegram bot ID token")
	flags.StringVar(&telegramOptions.ChatID, "telegram-chat-id", telegramOptions.ChatID, "Telegram chat ID")
	flags.IntVar(&telegramOptions.Timeout, "telegram-timeout", telegramOptions.Timeout, "Telegram timeout")
	flags.BoolVar(&telegramOptions.Insecure, "telegram-insecure", telegramOptions.Insecure, "Telegram insecure")
	flags.BoolVar(&telegramOptions.DisableNotification, "telegram-disable-notification", telegramOptions.DisableNotification, "Telegram disable notification")
	flags.StringVar(&telegramOptions.ParseMode, "telegram-parse-node", telegramOptions.ParseMode, "Telegram parse mode")
	flags.BoolVar(&telegramOptions.DisableWebPagePreview, "telegram-disable-webpage-preview", telegramOptions.DisableWebPagePreview, "Telegram disable webpage preview")
	flags.StringVar(&telegramOutput.Output, "telegram-output", telegramOutput.Output, "Telegram output")
	flags.StringVar(&telegramOutput.Query, "telegram-output-query", telegramOutput.Query, "Telegram output query")

	sendMessageCmd := &cobra.Command{
		Use:   "send-message",
		Short: "Send text message",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Telegram sending message...")
			common.Debug("Telegram", telegramMessageOptions, stdout)

			textBytes, err := utils.Content(telegramMessageOptions.Text)
			if err != nil {
				stdout.Panic(err)
			}
			telegramMessageOptions.Text = string(textBytes)

			telegramOptions.MessageOptions = &telegramMessageOptions
			bytes, err := telegramNew(stdout).SendMessage()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(telegramOutput, "Telegram", []interface{}{telegramOptions, telegramMessageOptions}, bytes, stdout)
		},
	}
	flags = sendMessageCmd.PersistentFlags()
	flags.StringVar(&telegramMessageOptions.Text, "telegram-message-text", telegramMessageOptions.Text, "Telegram message text")
	telegramCmd.AddCommand(sendMessageCmd)

	sendPhotoCmd := &cobra.Command{
		Use:   "send-photo",
		Short: "Send photo",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Telegram sending photo...")
			common.Debug("Telegram", telegramPhotoOptions, stdout)

			contentBytes, err := utils.Content(telegramPhotoOptions.Content)
			if err != nil {
				stdout.Panic(err)
			}
			telegramPhotoOptions.Content = string(contentBytes)

			if utils.IsEmpty(telegramPhotoOptions.Name) && utils.FileExists(telegramPhotoOptions.Content) {
				telegramPhotoOptions.Name = filepath.Base(telegramPhotoOptions.Content)
			}

			telegramOptions.PhotoOptions = &telegramPhotoOptions
			bytes, err := telegramNew(stdout).SendPhoto()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(telegramOutput, "Telegram", []interface{}{telegramOptions, telegramPhotoOptions}, bytes, stdout)
		},
	}
	flags = sendPhotoCmd.PersistentFlags()
	flags.StringVar(&telegramPhotoOptions.Caption, "telegram-photo-caption", telegramPhotoOptions.Caption, "Telegram photo caption")
	flags.StringVar(&telegramPhotoOptions.Name, "telegram-photo-name", telegramPhotoOptions.Name, "Telegram photo name")
	flags.StringVar(&telegramPhotoOptions.Content, "telegram-photo-content", telegramPhotoOptions.Content, "Telegram photo content")
	telegramCmd.AddCommand(sendPhotoCmd)

	sendDocumentCmd := &cobra.Command{
		Use:   "send-document",
		Short: "Send document",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Telegram sending document...")
			common.Debug("Telegram", telegramDocumentOptions, stdout)

			contentBytes, err := utils.Content(telegramDocumentOptions.Content)
			if err != nil {
				stdout.Panic(err)
			}
			telegramDocumentOptions.Content = string(contentBytes)

			if utils.IsEmpty(telegramDocumentOptions.Name) && utils.FileExists(telegramDocumentOptions.Content) {
				telegramDocumentOptions.Name = filepath.Base(telegramDocumentOptions.Content)
			}

			telegramOptions.DocumentOptions = &telegramDocumentOptions
			bytes, err := telegramNew(stdout).SendDocument()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(telegramOutput, "Telegram", []interface{}{telegramOptions, telegramDocumentOptions}, bytes, stdout)
		},
	}
	flags = sendDocumentCmd.PersistentFlags()
	flags.StringVar(&telegramDocumentOptions.Caption, "telegram-document-caption", telegramDocumentOptions.Caption, "Telegram document caption")
	flags.StringVar(&telegramDocumentOptions.Name, "telegram-document-name", telegramDocumentOptions.Name, "Telegram document name")
	flags.StringVar(&telegramDocumentOptions.Content, "telegram-document-content", telegramDocumentOptions.Content, "Telegram document content")
	telegramCmd.AddCommand(sendDocumentCmd)

	return &telegramCmd
}
