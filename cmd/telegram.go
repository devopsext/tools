package cmd

import (
	"github.com/devopsext/tools/messaging"
	"github.com/spf13/cobra"
)

var telegramOptions = messaging.TelegramOptions{
	URL:                 envGet("TELEGRAM_URL", "").(string),
	Timeout:             envGet("TELEGRAM_TIMEOUT", 30).(int),
	DisableNotification: envGet("TELEGRAM_DISABLE_NOTIFICATION", "false").(string),
}

func NewTelegramCommand() *cobra.Command {

	telegramCmd := cobra.Command{
		Use:   "telegram",
		Short: "Telegram tools",
		Run: func(cmd *cobra.Command, args []string) {
			//
		},
	}

	flags := telegramCmd.PersistentFlags()
	flags.StringVar(&telegramOptions.URL, "telegram-url", telegramOptions.URL, "Telegram URL")
	flags.IntVar(&telegramOptions.Timeout, "telegram-timeout", telegramOptions.Timeout, "Telegram timeout")
	flags.StringVar(&telegramOptions.DisableNotification, "telegram-disable-notification", telegramOptions.DisableNotification, "Telegram disable notification")

	return &telegramCmd
}
