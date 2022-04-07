package cmd

import (
	"io/ioutil"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/messaging"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var slackOptions = messaging.SlackOptions{
	URL:      envGet("SLACK_URL", "").(string),
	Timeout:  envGet("SLACK_TIMEOUT", 30).(int),
	Insecure: envGet("SLACK_INSECURE", false).(bool),
	Message:  envGet("SLACK_MESSAGE", "").(string),
	FileName: envGet("SLACK_FILENAME", "").(string),
	Title:    envGet("SLACK_TITLE", "").(string),
	Content:  envGet("SLACK_CONTENT", "").(string),
	Output:   envGet("SLACK_OUTPUT", "").(string),
}

func slackNew(stdout *common.Stdout) common.Messenger {
	slack := messaging.NewSlack(slackOptions)
	if slack == nil {
		stdout.Panic("No slack")
	}
	return slack
}

func slackOutput(stdout *common.Stdout, bytes []byte) {

	if utils.IsEmpty(slackOptions.Output) {
		stdout.Info(string(bytes))
	} else {
		stdout.Debug("Slack writing output to %s...", slackOptions.Output)
		err := ioutil.WriteFile(slackOptions.Output, bytes, 0644)
		if err != nil {
			stdout.Error(err)
		}
	}
}

func NewSlackCommand() *cobra.Command {

	slackCmd := &cobra.Command{
		Use:   "slack",
		Short: "Slack tools",
	}

	flags := slackCmd.PersistentFlags()
	flags.StringVar(&slackOptions.URL, "slack-url", slackOptions.URL, "Slack URL")
	flags.IntVar(&slackOptions.Timeout, "slack-timeout", slackOptions.Timeout, "Slack timeout")
	flags.BoolVar(&slackOptions.Insecure, "slack-insecure", slackOptions.Insecure, "Slack insecure")
	flags.StringVar(&slackOptions.Message, "slack-message", slackOptions.Message, "Slack message")
	flags.StringVar(&slackOptions.FileName, "slack-filename", slackOptions.FileName, "Slack file name")
	flags.StringVar(&slackOptions.Title, "slack-title", slackOptions.Title, "Slack title")
	flags.StringVar(&slackOptions.Content, "slack-content", slackOptions.Content, "Slack content")
	flags.StringVar(&slackOptions.Output, "slack-output", slackOptions.Output, "Slack output")

	slackCmd.AddCommand(&cobra.Command{
		Use:   "send",
		Short: "Send text message",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Slack sending message...")
			bytes, err := slackNew(stdout).Send()
			if err != nil {
				stdout.Error(err)
				return
			}
			slackOutput(stdout, bytes)
		},
	})

	slackCmd.AddCommand(&cobra.Command{
		Use:   "send-file",
		Short: "Send file",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Slack sending file...")
			bytes, err := slackNew(stdout).SendFile()
			if err != nil {
				stdout.Error(err)
				return
			}
			slackOutput(stdout, bytes)
		},
	})
	return slackCmd
}
