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
	Timeout:    envGet("SLACK_TIMEOUT", 30).(int),
	Insecure:   envGet("SLACK_INSECURE", false).(bool),
	Token:      envGet("SLACK_TOKEN", "").(string),
	Channel:    envGet("SLACK_CHANNEL", "").(string),
	Title:      envGet("SLACK_TITLE", "").(string),
	Message:    envGet("SLACK_MESSAGE", "").(string),
	ImageURL:   envGet("SLACK_IMAGE_URL", "").(string),
	FileName:   envGet("SLACK_FILENAME", "").(string),
	File:       envGet("SLACK_FILE", "").(string),
	ParentTS:   envGet("SLACK_THREAD", "").(string),
	QuoteColor: envGet("SLACK_QUOTE_COLOR", "").(string),
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

	messageBytes, err := utils.Content(slackOptions.Message)
	if err != nil {
		stdout.Panic(err)
	}
	slackOptions.Message = string(messageBytes)

	if utils.IsEmpty(slackOptions.FileName) && utils.FileExists(slackOptions.File) {
		slackOptions.FileName = filepath.Base(slackOptions.File)
	}

	fileBytes, err := utils.Content(slackOptions.File)
	if err != nil {
		stdout.Panic(err)
	}
	slackOptions.File = string(fileBytes)

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
	flags.StringVar(&slackOptions.Message, "slack-message", slackOptions.Message, "Slack message")
	flags.StringVar(&slackOptions.FileName, "slack-filename", slackOptions.FileName, "Slack file name")
	flags.StringVar(&slackOptions.ImageURL, "slack-image-url", slackOptions.ImageURL, "Slack image url")
	flags.StringVar(&slackOptions.Title, "slack-title", slackOptions.Title, "Slack title")
	flags.StringVar(&slackOptions.File, "slack-file", slackOptions.File, "Slack file content or path")
	flags.StringVar(&slackOptions.Token, "slack-token", slackOptions.Token, "Slack token")
	flags.StringVar(&slackOptions.Channel, "slack-channel", slackOptions.Channel, "Slack channel")
	flags.StringVar(&slackOptions.ParentTS, "slack-thread", slackOptions.ParentTS, "Slack thread")
	flags.StringVar(&slackOptions.QuoteColor, "slack-quote-color", slackOptions.QuoteColor, "Slack quote color in hex format (#008000, no quote by default)")
	flags.StringVar(&slackOutput.Output, "slack-output", slackOutput.Output, "Slack output")
	flags.StringVar(&slackOutput.Query, "slack-output-query", slackOutput.Query, "Slack output query")

	slackCmd.AddCommand(&cobra.Command{
		Use:   "send-message",
		Short: "Send text message",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Slack sending message...")
			bytes, err := slackNew(stdout).SendMessage()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(slackOutput, "Slack", []interface{}{slackOptions}, bytes, stdout)
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
			common.OutputJson(slackOutput, "Slack", []interface{}{slackOptions}, bytes, stdout)
		},
	})

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
