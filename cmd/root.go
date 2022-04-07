package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/messaging"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var version = "unknown"
var APPNAME = "TOOLS"
var appName = strings.ToLower(APPNAME)

var stdoutOptions = common.StdoutOptions{
	Format:          "text",
	Level:           "info",
	Template:        "{{.file}} {{.msg}}",
	TimestampFormat: time.RFC3339Nano,
	TextColors:      true,
	Debug:           false,
}

var slackOptions = messaging.SlackOptions{
	URL:     envGet("SLACK_URL", "").(string),
	Timeout: envGet("SLACK_TIMEOUT", 30).(int),
}

var telegramOptions = messaging.TelegramOptions{
	URL:                 envGet("TELEGRAM_URL", "").(string),
	Timeout:             envGet("TELEGRAM_TIMEOUT", 30).(int),
	DisableNotification: envGet("TELEGRAM_DISABLE_NOTIFICATION", "false").(string),
}

func envGet(s string, d interface{}) interface{} {
	return utils.EnvGet(fmt.Sprintf("%s_%s", APPNAME, s), d)
}

var stdout *common.Stdout

func Execute() {

	rootCmd := &cobra.Command{
		Use:   "tools",
		Short: "Tools",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			stdoutOptions.Version = version
			stdout = common.NewStdout(stdoutOptions)
			stdout.SetCallerOffset(1)
		},
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Info("Log message...")

			messengers := make(map[string]common.Messenger)
			messengers["slack"] = messaging.NewSlack(slackOptions)
			messengers["telegram"] = messaging.NewTelegram(telegramOptions)
		},
	}

	flags := rootCmd.PersistentFlags()

	flags.StringVar(&stdoutOptions.Format, "stdout-format", stdoutOptions.Format, "Stdout format: json, text, template")
	flags.StringVar(&stdoutOptions.Level, "stdout-level", stdoutOptions.Level, "Stdout level: info, warn, error, debug, panic")
	flags.StringVar(&stdoutOptions.Template, "stdout-template", stdoutOptions.Template, "Stdout template")
	flags.StringVar(&stdoutOptions.TimestampFormat, "stdout-timestamp-format", stdoutOptions.TimestampFormat, "Stdout timestamp format")
	flags.BoolVar(&stdoutOptions.TextColors, "stdout-text-colors", stdoutOptions.TextColors, "Stdout text colors")
	flags.BoolVar(&stdoutOptions.Debug, "stdout-debug", stdoutOptions.Debug, "Stdout debug")

	flags.StringVar(&slackOptions.URL, "slack-url", slackOptions.URL, "Slack URL")
	flags.IntVar(&slackOptions.Timeout, "slack-timeout", slackOptions.Timeout, "Slack timeout")

	flags.StringVar(&telegramOptions.URL, "telegram-url", telegramOptions.URL, "Telegram URL")
	flags.IntVar(&telegramOptions.Timeout, "telegram-timeout", telegramOptions.Timeout, "Telegram timeout")
	flags.StringVar(&telegramOptions.DisableNotification, "telegram-disable-notification", telegramOptions.DisableNotification, "Telegram disable notification")

	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print the version number",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println(version)
		},
	})

	if err := rootCmd.Execute(); err != nil {
		stdout.Error(err)
		os.Exit(1)
	}
}
