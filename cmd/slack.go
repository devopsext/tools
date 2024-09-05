package cmd

import (
	"strings"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var slackOptions = vendors.SlackOptions{
	Timeout:  envGet("SLACK_TIMEOUT", 30).(int),
	Insecure: envGet("SLACK_INSECURE", false).(bool),
	Token:    envGet("SLACK_TOKEN", "").(string),
}

var slackMessageOptions = vendors.SlackMessageOptions{
	Channel:     envGet("SLACK_CHANNEL", "").(string),
	Thread:      envGet("SLACK_THREAD", "").(string),
	Title:       envGet("SLACK_TITLE", "").(string),
	Text:        envGet("SLACK_TEXT", "").(string),
	Attachments: envGet("SLACK_ATTACHMENTS", "").(string),
	Blocks:      envGet("SLACK_BLOCKS", "").(string),
}

var slackFileOptions = vendors.SlackFileOptions{
	Channel: envGet("SLACK_CHANNEL", "").(string),
	Thread:  envGet("SLACK_THREAD", "").(string),
	Title:   envGet("SLACK_TITLE", "").(string),
	Text:    envGet("SLACK_TEXT", "").(string),
	Name:    envGet("SLACK_NAME", "").(string),
	Content: envGet("SLACK_CONTENT", "").(string),
	Type:    envGet("SLACK_TYPE", "auto").(string),
}

var slackReactionOptions = vendors.SlackReactionOptions{
	Name: envGet("SLACK_REACTION_NAME", "").(string),
}

var slackUserEmail = vendors.SlackUserEmail{
	Email: envGet("SLACK_USER_EMAIL", "").(string),
}

var slackUsergroupUsers = vendors.SlackUsergroupUsers{
	Usergroup: envGet("SLACK_USERGROUP", "").(string),
	Users:     strings.Split(envGet("SLACK_USERS", "").(string), " "),
}

var slackOutput = common.OutputOptions{
	Output: envGet("SLACK_OUTPUT", "").(string),
	Query:  envGet("SLACK_OUTPUT_QUERY", "").(string),
}

func slackNew(stdout *common.Stdout) *vendors.Slack {

	common.Debug("Slack", slackOptions, stdout)
	common.Debug("Slack", slackOutput, stdout)

	return vendors.NewSlack(slackOptions)
}

func NewSlackCommand() *cobra.Command {

	slackCmd := &cobra.Command{
		Use:   "slack",
		Short: "Slack tools",
	}

	flags := slackCmd.PersistentFlags()
	flags.IntVar(&slackOptions.Timeout, "slack-timeout", slackOptions.Timeout, "Slack timeout")
	flags.BoolVar(&slackOptions.Insecure, "slack-insecure", slackOptions.Insecure, "Slack insecure")
	flags.StringVar(&slackOptions.Token, "slack-token", slackOptions.Token, "Slack token")
	flags.StringVar(&slackOutput.Output, "slack-output", slackOutput.Output, "Slack output")
	flags.StringVar(&slackOutput.Query, "slack-output-query", slackOutput.Query, "Slack output query")

	sendMessage := &cobra.Command{
		Use:   "send-message",
		Short: "Send text message",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Slack sending message...")
			common.Debug("Slack", slackMessageOptions, stdout)

			textBytes, err := utils.Content(slackMessageOptions.Text)
			if err != nil {
				stdout.Panic(err)
			}
			slackMessageOptions.Text = string(textBytes)

			bytes, err := slackNew(stdout).SendMessage(slackMessageOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(slackOutput, "Slack", []interface{}{slackOptions, slackMessageOptions}, bytes, stdout)
		},
	}
	flags = sendMessage.PersistentFlags()
	flags.StringVar(&slackMessageOptions.Channel, "slack-channel", slackMessageOptions.Channel, "Slack channel")
	flags.StringVar(&slackMessageOptions.Thread, "slack-thread", slackMessageOptions.Thread, "Slack thread")
	flags.StringVar(&slackMessageOptions.Title, "slack-title", slackMessageOptions.Title, "Slack title")
	flags.StringVar(&slackMessageOptions.Text, "slack-text", slackMessageOptions.Text, "Slack text")
	flags.StringVar(&slackMessageOptions.Attachments, "slack-attachments", slackMessageOptions.Attachments, "Slack attachments json")
	flags.StringVar(&slackMessageOptions.Blocks, "slack-blocks", slackMessageOptions.Blocks, "Slack blocks json")
	slackCmd.AddCommand(sendMessage)

	sendFile := &cobra.Command{
		Use:   "send-file",
		Short: "Send file",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Slack sending file...")
			common.Debug("Slack", slackFileOptions, stdout)

			textBytes, err := utils.Content(slackFileOptions.Text)
			if err != nil {
				stdout.Panic(err)
			}
			slackFileOptions.Text = string(textBytes)

			contentBytes, err := utils.Content(slackFileOptions.Content)
			if err != nil {
				stdout.Panic(err)
			}
			slackFileOptions.Content = string(contentBytes)

			bytes, err := slackNew(stdout).SendFile(slackFileOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(slackOutput, "Slack", []interface{}{slackOptions, slackFileOptions}, bytes, stdout)
		},
	}
	flags = sendFile.PersistentFlags()
	flags.StringVar(&slackFileOptions.Channel, "slack-channel", slackFileOptions.Channel, "Slack channel")
	flags.StringVar(&slackFileOptions.Thread, "slack-thread", slackFileOptions.Thread, "Slack thread")
	flags.StringVar(&slackFileOptions.Title, "slack-title", slackFileOptions.Title, "Slack title")
	flags.StringVar(&slackFileOptions.Text, "slack-text", slackFileOptions.Text, "Slack text")
	flags.StringVar(&slackFileOptions.Name, "slack-name", slackFileOptions.Name, "Slack file name")
	flags.StringVar(&slackFileOptions.Content, "slack-content", slackFileOptions.Content, "Slack file content")
	flags.StringVar(&slackFileOptions.Type, "slack-type", slackFileOptions.Type, "Slack file type")
	slackCmd.AddCommand(sendFile)

	addReactionCmd := &cobra.Command{
		Use:   "add-reaction",
		Short: "Add reaction",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Slack add reaction...")
			common.Debug("Slack", slackReactionOptions, stdout)

			bytes, err := slackNew(stdout).AddReaction(slackReactionOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(slackOutput, "Slack", []interface{}{slackOptions, slackReactionOptions}, bytes, stdout)
		},
	}
	flags = addReactionCmd.PersistentFlags()
	flags.StringVar(&slackReactionOptions.Name, "slack-reaction-name", slackReactionOptions.Name, "Slack reaction name")
	slackCmd.AddCommand(addReactionCmd)

	lookupByEmailCmd := &cobra.Command{
		Use:   "lookup-by-email",
		Short: "Lookup by email",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Getting User...")
			common.Debug("Slack", slackUserEmail, stdout)

			bytes, err := slackNew(stdout).GetUser(slackUserEmail)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(slackOutput, "Slack", []interface{}{slackOptions, slackUserEmail}, bytes, stdout)
		},
	}
	flags = lookupByEmailCmd.PersistentFlags()
	flags.StringVar(&slackUserEmail.Email, "slack-user-email", slackUserEmail.Email, "Slack user email")
	slackCmd.AddCommand(lookupByEmailCmd)

	usergroupUpdateCmd := &cobra.Command{
		Use:   "usergroup-update",
		Short: "Usergroup update",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Updateing usergroup...")
			common.Debug("Slack", slackUsergroupUsers, stdout)

			bytes, err := slackNew(stdout).UpdateUsergroup(slackUsergroupUsers)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(slackOutput, "Slack", []interface{}{slackOptions, slackUsergroupUsers}, bytes, stdout)
		},
	}
	flags = usergroupUpdateCmd.PersistentFlags()
	flags.StringVar(&slackUsergroupUsers.Usergroup, "slack-usergroup", slackUsergroupUsers.Usergroup, "Slack usergroup")
	flags.StringSliceVar(&slackUsergroupUsers.Users, "slack-users", slackUsergroupUsers.Users, "Slack usergroup")
	slackCmd.AddCommand(usergroupUpdateCmd)

	return slackCmd
}
