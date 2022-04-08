package cmd

import (
	"path/filepath"
	"strings"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var telegramOptions = vendors.TelegramOptions{
	URL:                 envGet("TELEGRAM_URL", "").(string),
	Timeout:             envGet("TELEGRAM_TIMEOUT", 30).(int),
	DisableNotification: envGet("TELEGRAM_DISABLE_NOTIFICATION", false).(bool),
	Message:             envGet("TELEGRAM_MESSAGE", "").(string),
	FileName:            envGet("TELEGRAM_FILENAME", "").(string),
	Content:             envGet("TELEGRAM_CONTENT", "").(string),
	Output:              envGet("TELEGRAM_OUTPUT", "").(string),
	OutputQuery:         envGet("TELEGRAM_OUTPUT_QUERY", "").(string),
}

func telegramNew(stdout *common.Stdout) common.Messenger {

	messageBytes, err := utils.Content(telegramOptions.Message)
	if err != nil {
		stdout.Panic(err)
	}
	telegramOptions.Message = string(messageBytes)

	contentBytes, err := utils.Content(telegramOptions.Content)
	if err != nil {
		stdout.Panic(err)
	}
	telegramOptions.Content = string(contentBytes)

	if utils.IsEmpty(telegramOptions.FileName) && utils.FileExists(telegramOptions.Content) {
		telegramOptions.FileName = strings.TrimSuffix(telegramOptions.Content, filepath.Ext(telegramOptions.Content))
	}

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
	flags.StringVar(&telegramOptions.URL, "telegram-url", telegramOptions.URL, "Telegram URL")
	flags.IntVar(&telegramOptions.Timeout, "telegram-timeout", telegramOptions.Timeout, "Telegram timeout")
	flags.BoolVar(&telegramOptions.Insecure, "telegram-insecure", telegramOptions.Insecure, "Telegram insecure")
	flags.BoolVar(&telegramOptions.DisableNotification, "telegram-disable-notification", telegramOptions.DisableNotification, "Telegram disable notification")
	flags.StringVar(&telegramOptions.Message, "telegram-message", telegramOptions.Message, "Telegram message")
	flags.StringVar(&telegramOptions.FileName, "telegram-filename", telegramOptions.FileName, "Telegram file name")
	flags.StringVar(&telegramOptions.Content, "telegram-content", telegramOptions.Content, "Telegram content")
	flags.StringVar(&telegramOptions.Output, "telegram-output", telegramOptions.Output, "Telegram output")
	flags.StringVar(&telegramOptions.OutputQuery, "telegram-output-query", telegramOptions.OutputQuery, "Telegram output query")

	telegramCmd.AddCommand(&cobra.Command{
		Use:   "send",
		Short: "Send text message",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Telegram sending message...")
			bytes, err := telegramNew(stdout).Send()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.Output(telegramOptions.OutputQuery, telegramOptions.Output, bytes, stdout)
		},
	})

	telegramCmd.AddCommand(&cobra.Command{
		Use:   "send-file",
		Short: "Send file",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Telegram sending file...")
			bytes, err := telegramNew(stdout).SendFile()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.Output(telegramOptions.OutputQuery, telegramOptions.Output, bytes, stdout)
		},
	})
	return &telegramCmd
}
