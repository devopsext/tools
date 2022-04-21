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
	URL:      envGet("JIRA_URL", "").(string),
	Timeout:  envGet("JIRA_TIMEOUT", 30).(int),
	Insecure: envGet("JIRA_INSECURE", false).(bool),
	User:     envGet("JIRA_USER", "").(string),
	Password: envGet("JIRA_PASSWORD", "").(string),
}

var jiraCreateIssueOptions = vendors.JiraCreateIssueOptions{
	ProjectKey:  envGet("JIRA_ISSUE_PROJECT_KEY", "").(string),
	Type:        envGet("JIRA_ISSUE_TYPE", "").(string),
	Summary:     envGet("JIRA_ISSUE_SUMMARY", "").(string),
	Description: envGet("JIRA_ISSUE_DESCRIPTION", "").(string),
	Labels:      strings.Split(envGet("JIRA_ISSUE_LABELS", "").(string), ","),
	Priority:    envGet("JIRA_ISSUE_PRIORITY", "").(string),
	Assignee:    envGet("JIRA_ISSUE_ASSIGNEE", "").(string),
	Reporter:    envGet("JIRA_ISSUE_REPORTER", "").(string),
}

var jiraIssueOptions = vendors.JiraIssueOptions{
	IdOrKey: envGet("JIRA_ISSUE_ID_OR_KEY", "").(string),
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

	jira := vendors.NewJira(jiraOptions)
	if jira == nil {
		stdout.Panic("No jira")
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
	flags.StringVar(&jiraOutput.Output, "jira-output", jiraOutput.Output, "Jira output")
	flags.StringVar(&jiraOutput.Query, "jira-output-query", jiraOutput.Query, "Jira output query")

	// tools jira create-issue --jira-params --create-issue-params
	createIssueCmd := &cobra.Command{
		Use:   "create-issue",
		Short: "Create issue",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Jira creating issue...")
			common.Debug("Jira", jiraCreateIssueOptions, stdout)

			descriptionBytes, err := utils.Content(jiraCreateIssueOptions.Description)
			if err != nil {
				stdout.Panic(err)
			}
			jiraCreateIssueOptions.Description = string(descriptionBytes)

			jiraOptions.CreateIssueOptions = &jiraCreateIssueOptions
			bytes, err := jiraNew(stdout).CreateIssue()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jiraOutput, "Jira", []interface{}{jiraOptions, jiraCreateIssueOptions}, bytes, stdout)
		},
	}
	flags = createIssueCmd.PersistentFlags()
	flags.StringVar(&jiraCreateIssueOptions.ProjectKey, "jira-issue-project-key", jiraCreateIssueOptions.ProjectKey, "Jira issue project key")
	flags.StringVar(&jiraCreateIssueOptions.Type, "jira-issue-type", jiraCreateIssueOptions.Type, "Jira issue type")
	flags.StringVar(&jiraCreateIssueOptions.Summary, "jira-issue-summary", jiraCreateIssueOptions.Summary, "Jira issue summary")
	flags.StringVar(&jiraCreateIssueOptions.Description, "jira-issue-description", jiraCreateIssueOptions.Description, "Jira issue description")
	flags.StringSliceVar(&jiraCreateIssueOptions.Labels, "jira-issue-labels", jiraCreateIssueOptions.Labels, "Jira issue labels")
	flags.StringVar(&jiraCreateIssueOptions.Priority, "jira-issue-priority", jiraCreateIssueOptions.Priority, "Jira issue priority")
	flags.StringVar(&jiraCreateIssueOptions.Assignee, "jira-issue-assignee", jiraCreateIssueOptions.Assignee, "Jira issue assignee")
	flags.StringVar(&jiraCreateIssueOptions.Reporter, "jira-issue-reporter", jiraCreateIssueOptions.Reporter, "Jira issue reporter")
	jiraCmd.AddCommand(createIssueCmd)

	issue := &cobra.Command{
		Use:   "issue",
		Short: "Issue methods",
	}
	flags = issue.PersistentFlags()
	flags.StringVar(&jiraIssueOptions.IdOrKey, "jira-issue-id-or-key", jiraIssueOptions.IdOrKey, "Jira issue ID or key")
	jiraCmd.AddCommand(issue)

	// tools jira issue add-comment --jira-params --issue-params --add-comment-params
	issueAddComment := &cobra.Command{
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

			jiraOptions.IssueOptions = &jiraIssueOptions
			jiraOptions.IssueAddCommentOptions = &jiraIssueAddCommentOptions
			bytes, err := jiraNew(stdout).IssueAddComment()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jiraOutput, "Jira", []interface{}{jiraOptions, jiraIssueOptions, jiraIssueAddCommentOptions}, bytes, stdout)
		},
	}
	flags = issueAddComment.PersistentFlags()
	flags.StringVar(&jiraIssueAddCommentOptions.Body, "jira-issue-comment-body", jiraIssueAddCommentOptions.Body, "Jira issue comment body")
	issue.AddCommand(issueAddComment)

	// tools jira issue add-attachment --jira-params --issue-params --add-attachment-params
	issueAddAttachment := &cobra.Command{
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

			jiraOptions.IssueOptions = &jiraIssueOptions
			jiraOptions.IssueAddAttachmentOptions = &jiraIssueAddAttachmentOptions
			bytes, err := jiraNew(stdout).IssueAddAttachment()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jiraOutput, "Jira", []interface{}{jiraOptions, jiraIssueOptions, jiraIssueAddAttachmentOptions}, bytes, stdout)
		},
	}
	flags = issueAddAttachment.PersistentFlags()
	flags.StringVar(&jiraIssueAddAttachmentOptions.File, "jira-issue-attachment-file", jiraIssueAddAttachmentOptions.File, "Jira issue attachment file")
	flags.StringVar(&jiraIssueAddAttachmentOptions.Name, "jira-issue-attachment-name", jiraIssueAddAttachmentOptions.Name, "Jira issue attachment name")
	issue.AddCommand(issueAddAttachment)

	return &jiraCmd
}
