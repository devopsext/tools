package cmd

import (
	"path/filepath"
	"strings"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var slackOptions = vendors.SlackOptions{
	Timeout:     envGet("SLACK_TIMEOUT", 30).(int),
	Insecure:    envGet("SLACK_INSECURE", false).(bool),
	Token:       envGet("SLACK_TOKEN", "").(string),
	Channel:     envGet("SLACK_CHANNEL", "").(string),
	Title:       envGet("SLACK_TITLE", "").(string),
	Message:     envGet("SLACK_MESSAGE", "").(string),
	ImageURL:    envGet("SLACK_IMAGE_URL", "").(string),
	FileName:    envGet("SLACK_FILENAME", "").(string),
	FileContent: envGet("SLACK_CONTENT", "").(string),
}

var slackOutput = common.OutputOptions{
	Output: envGet("SLACK_OUTPUT", "").(string),
	Query:  envGet("SLACK_OUTPUT_QUERY", "").(string),
}

func slackNew(stdout *common.Stdout) *vendors.Slack {

	common.Debug("Slack", slackOptions, stdout)
	common.Debug("Slack", slackOutput, stdout)

	messageBytes, err := utils.Content(slackOptions.Message)
	if err != nil {
		stdout.Panic(err)
	}
	slackOptions.Message = string(messageBytes)

	contentBytes, err := utils.Content(slackOptions.FileContent)
	if err != nil {
		stdout.Panic(err)
	}
	slackOptions.FileContent = string(contentBytes)

	if utils.IsEmpty(slackOptions.FileName) && utils.FileExists(slackOptions.FileContent) {
		slackOptions.FileName = strings.TrimSuffix(slackOptions.FileContent, filepath.Ext(slackOptions.FileContent))
	}

	slack := vendors.NewSlack(slackOptions)
	if slack == nil {
		stdout.Panic("No slack")
	}
	return slack
}

func NewSlackCommand() *cobra.Command {

	slackCmd := &cobra.Command{
		Use:   "slack",
		Short: "Slack tools",
	}

	flags := slackCmd.PersistentFlags()
	flags.IntVar(&slackOptions.Timeout, "slack-timeout", slackOptions.Timeout, "Slack timeout")
	flags.BoolVar(&slackOptions.Insecure, "slack-insecure", slackOptions.Insecure, "Slack insecure")
	flags.StringVar(&slackOptions.Message, "slack-message", slackOptions.Message, "Slack message")
	flags.StringVar(&slackOptions.FileName, "slack-filename", slackOptions.FileName, "Slack file name")
	flags.StringVar(&slackOptions.ImageURL, "slack-image-url", slackOptions.ImageURL, "Slack image url")
	flags.StringVar(&slackOptions.Title, "slack-title", slackOptions.Title, "Slack title")
	flags.StringVar(&slackOptions.FileContent, "slack-content", slackOptions.FileContent, "Slack content")
	flags.StringVar(&slackOptions.Token, "slack-token", slackOptions.Token, "Slack token")
	flags.StringVar(&slackOptions.Channel, "slack-channel", slackOptions.Channel, "Slack channel")
	flags.StringVar(&slackOutput.Output, "slack-output", slackOutput.Output, "Slack output")
	flags.StringVar(&slackOutput.Query, "slack-output-query", slackOutput.Query, "Slack output query")

	slackCmd.AddCommand(&cobra.Command{
		Use:   "send",
		Short: "Send text message",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Slack sending message...")
			bytes, err := slackNew(stdout).SendMessage()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(slackOutput, "Slack", slackOptions, bytes, stdout)
		},
	})

	slackCmd.AddCommand(&cobra.Command{
		Use:   "send-file",
		Short: "Send file",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Slack sending file...")
			s := slackNew(stdout)
			bytes, err := s.SendFile()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(slackOutput, "Slack", slackOptions, bytes, stdout)
		},
	})
	return slackCmd
}
