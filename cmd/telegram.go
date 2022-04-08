package cmd

import (
	"io/ioutil"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/messaging"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var telegramOptions = messaging.TelegramOptions{
	URL:                 envGet("TELEGRAM_URL", "").(string),
	Timeout:             envGet("TELEGRAM_TIMEOUT", 30).(int),
	DisableNotification: envGet("TELEGRAM_DISABLE_NOTIFICATION", false).(bool),
	Message:             envGet("TELEGRAM_MESSAGE", "").(string),
	FileName:            envGet("TELEGRAM_FILENAME", "").(string),
	Content:             envGet("TELEGRAM_CONTENT", "").(string),
	Output:              envGet("TELEGRAM_OUTPUT", "").(string),
}

func telegramNew(stdout *common.Stdout) common.Messenger {
	telegram := messaging.NewTelegram(telegramOptions)
	if telegram == nil {
		stdout.Panic("No telegram")
	}
	return telegram
}

func telegramOutput(stdout *common.Stdout, bytes []byte) {

	if utils.IsEmpty(telegramOptions.Output) {
		stdout.Info(string(bytes))
	} else {
		stdout.Debug("Telegram writing output to %s...", telegramOptions.Output)
		err := ioutil.WriteFile(telegramOptions.Output, bytes, 0644)
		if err != nil {
			stdout.Error(err)
		}
	}
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
			telegramOutput(stdout, bytes)
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
			telegramOutput(stdout, bytes)
		},
	})
	return &telegramCmd
}
