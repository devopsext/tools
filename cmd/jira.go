package cmd

import (
	"path/filepath"
	"strings"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var jiraOptions = vendors.JiraOptions{
	URL:         envGet("JIRA_URL", "").(string),
	Timeout:     envGet("JIRA_TIMEOUT", 30).(int),
	Insecure:    envGet("JIRA_INSECURE", false).(bool),
	User:        envGet("JIRA_USER", "").(string),
	Password:    envGet("JIRA_PASSWORD", "").(string),
	AccessToken: envGet("JIRA_ACCESS_TOKEN", "").(string),
}

var jiraIssueCreateOptions = vendors.JiraIssueCreateOptions{
	ProjectKey: envGet("JIRA_ISSUE_PROJECT_KEY", "").(string),
	Type:       envGet("JIRA_ISSUE_TYPE", "").(string),
	Priority:   envGet("JIRA_ISSUE_PRIORITY", "").(string),
	Assignee:   envGet("JIRA_ISSUE_ASSIGNEE", "").(string),
	Reporter:   envGet("JIRA_ISSUE_REPORTER", "").(string),
}

var jiraIssueOptions = vendors.JiraIssueOptions{
	IdOrKey:      envGet("JIRA_ISSUE_ID_OR_KEY", "").(string),
	Summary:      envGet("JIRA_ISSUE_SUMMARY", "").(string),
	Description:  envGet("JIRA_ISSUE_DESCRIPTION", "").(string),
	CustomFields: envGet("JIRA_ISSUE_CUSTOM_FIELDS", "").(string),
	Labels:       strings.Split(envGet("JIRA_ISSUE_LABELS", "").(string), ","),
}

var jiraIssueAddCommentOptions = vendors.JiraIssueAddCommentOptions{
	Body: envGet("JIRA_ISSUE_COMMENT_BODY", "").(string),
}

var jiraIssueAddAttachmentOptions = vendors.JiraIssueAddAttachmentOptions{
	File: envGet("JIRA_ISSUE_ATTACHMENT_FILE", "").(string),
	Name: envGet("JIRA_ISSUE_ATTACHMENT_NAME", "").(string),
}

var jiraOutput = common.OutputOptions{
	Output: envGet("JIRA_OUTPUT", "").(string),
	Query:  envGet("JIRA_OUTPUT_QUERY", "").(string),
}

func jiraNew(stdout *common.Stdout) *vendors.Jira {

	common.Debug("Jira", jiraOptions, stdout)
	common.Debug("Jira", jiraOutput, stdout)

	jira, err := vendors.NewJira(jiraOptions)
	if err != nil {
		stdout.Panic(err)
	}
	return jira
}

func NewJiraCommand() *cobra.Command {

	jiraCmd := cobra.Command{
		Use:   "jira",
		Short: "Jira tools",
	}
	flags := jiraCmd.PersistentFlags()
	flags.StringVar(&jiraOptions.URL, "jira-url", jiraOptions.URL, "Jira URL")
	flags.IntVar(&jiraOptions.Timeout, "jira-timeout", jiraOptions.Timeout, "Jira timeout")
	flags.BoolVar(&jiraOptions.Insecure, "jira-insecure", jiraOptions.Insecure, "Jira insecure")
	flags.StringVar(&jiraOptions.User, "jira-user", jiraOptions.User, "Jira user")
	flags.StringVar(&jiraOptions.Password, "jira-password", jiraOptions.Password, "Jira password")
	flags.StringVar(&jiraOptions.AccessToken, "jira-access-token", jiraOptions.AccessToken, "Jira Personal Access Token")
	flags.StringVar(&jiraOutput.Output, "jira-output", jiraOutput.Output, "Jira output")
	flags.StringVar(&jiraOutput.Query, "jira-output-query", jiraOutput.Query, "Jira output query")

	issueCmd := &cobra.Command{
		Use:   "issue",
		Short: "Issue methods",
	}
	flags = issueCmd.PersistentFlags()
	flags.StringVar(&jiraIssueOptions.IdOrKey, "jira-issue-id-or-key", jiraIssueOptions.IdOrKey, "Jira issue ID or key")
	flags.StringVar(&jiraIssueOptions.Summary, "jira-issue-summary", jiraIssueOptions.Summary, "Jira issue summary")
	flags.StringVar(&jiraIssueOptions.Description, "jira-issue-description", jiraIssueOptions.Description, "Jira issue description")
	flags.StringVar(&jiraIssueOptions.CustomFields, "jira-issue-custom-fields", jiraIssueOptions.CustomFields, "Jira issue custom fields file")
	flags.StringSliceVar(&jiraIssueOptions.Labels, "jira-issue-labels", jiraIssueOptions.Labels, "Jira issue labels")
	jiraCmd.AddCommand(issueCmd)

	// tools jira issue create --jira-params --create-issue-params
	issueCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create issue",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Jira creating issue...")
			common.Debug("Jira", jiraIssueOptions, stdout)
			common.Debug("Jira", jiraIssueCreateOptions, stdout)

			descriptionBytes, err := utils.Content(jiraIssueOptions.Description)
			if err != nil {
				stdout.Panic(err)
			}
			jiraIssueOptions.Description = string(descriptionBytes)

			bytes, err := jiraNew(stdout).IssueCreate(jiraIssueOptions, jiraIssueCreateOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jiraOutput, "Jira", []interface{}{jiraOptions, jiraIssueOptions, jiraIssueCreateOptions}, bytes, stdout)
		},
	}
	flags = issueCreateCmd.PersistentFlags()
	flags.StringVar(&jiraIssueCreateOptions.ProjectKey, "jira-issue-project-key", jiraIssueCreateOptions.ProjectKey, "Jira issue project key")
	flags.StringVar(&jiraIssueCreateOptions.Type, "jira-issue-type", jiraIssueCreateOptions.Type, "Jira issue type")
	flags.StringVar(&jiraIssueCreateOptions.Priority, "jira-issue-priority", jiraIssueCreateOptions.Priority, "Jira issue priority")
	flags.StringVar(&jiraIssueCreateOptions.Assignee, "jira-issue-assignee", jiraIssueCreateOptions.Assignee, "Jira issue assignee")
	flags.StringVar(&jiraIssueCreateOptions.Reporter, "jira-issue-reporter", jiraIssueCreateOptions.Reporter, "Jira issue reporter")
	issueCmd.AddCommand(issueCreateCmd)

	// tools jira issue add-comment --jira-params --issue-params --add-comment-params
	issueAddCommentCmd := &cobra.Command{
		Use:   "add-comment",
		Short: "Issue add comment",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Jira issue adding comment...")
			common.Debug("Jira", jiraIssueOptions, stdout)
			common.Debug("Jira", jiraIssueAddCommentOptions, stdout)

			bodyBytes, err := utils.Content(jiraIssueAddCommentOptions.Body)
			if err != nil {
				stdout.Panic(err)
			}
			jiraIssueAddCommentOptions.Body = string(bodyBytes)

			bytes, err := jiraNew(stdout).IssueAddComment(jiraIssueOptions, jiraIssueAddCommentOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jiraOutput, "Jira", []interface{}{jiraOptions, jiraIssueOptions, jiraIssueAddCommentOptions}, bytes, stdout)
		},
	}
	flags = issueAddCommentCmd.PersistentFlags()
	flags.StringVar(&jiraIssueAddCommentOptions.Body, "jira-issue-comment-body", jiraIssueAddCommentOptions.Body, "Jira issue comment body")
	issueCmd.AddCommand(issueAddCommentCmd)

	// tools jira issue add-attachment --jira-params --issue-params --add-attachment-params
	issueAddAttachmentCmd := &cobra.Command{
		Use:   "add-attachment",
		Short: "Issue add attachment",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Jira issue adding attachment...")
			common.Debug("Jira", jiraIssueOptions, stdout)
			common.Debug("Jira", jiraIssueAddAttachmentOptions, stdout)

			if utils.IsEmpty(jiraIssueAddAttachmentOptions.Name) && utils.FileExists(jiraIssueAddAttachmentOptions.File) {
				jiraIssueAddAttachmentOptions.Name = filepath.Base(jiraIssueAddAttachmentOptions.File)
			}

			fileBytes, err := utils.Content(jiraIssueAddAttachmentOptions.File)
			if err != nil {
				stdout.Panic(err)
			}
			jiraIssueAddAttachmentOptions.File = string(fileBytes)

			bytes, err := jiraNew(stdout).IssueAddAttachment(jiraIssueOptions, jiraIssueAddAttachmentOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jiraOutput, "Jira", []interface{}{jiraOptions, jiraIssueOptions, jiraIssueAddAttachmentOptions}, bytes, stdout)
		},
	}
	flags = issueAddAttachmentCmd.PersistentFlags()
	flags.StringVar(&jiraIssueAddAttachmentOptions.File, "jira-issue-attachment-file", jiraIssueAddAttachmentOptions.File, "Jira issue attachment file")
	flags.StringVar(&jiraIssueAddAttachmentOptions.Name, "jira-issue-attachment-name", jiraIssueAddAttachmentOptions.Name, "Jira issue attachment name")
	issueCmd.AddCommand(issueAddAttachmentCmd)

	// tools jira issue update --jira-params --issue-params
	issueUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "Issue update",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Jira issue updating...")
			common.Debug("Jira", jiraIssueOptions, stdout)

			descriptionBytes, err := utils.Content(jiraIssueOptions.Description)
			if err != nil {
				stdout.Panic(err)
			}
			jiraIssueOptions.Description = string(descriptionBytes)

			bytes, err := jiraNew(stdout).IssueUpdate(jiraIssueOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jiraOutput, "Jira", []interface{}{jiraOptions, jiraIssueOptions}, bytes, stdout)
		},
	}
	issueCmd.AddCommand(issueUpdateCmd)

	return &jiraCmd
}
