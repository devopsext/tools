package cmd

import (
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
	ProjectKey:  envGet("JIRA_PROJECT_KEY", "").(string),
	IssueType:   envGet("JIRA_ISSUE_TYPE", "").(string),
	Summary:     envGet("JIRA_SUMMARY", "").(string),
	Description: envGet("JIRA_DESCRIPTION", "").(string),
	Labels:      strings.Split(envGet("JIRA_LABELS", "").(string), ","),
	Priority:    envGet("JIRA_PRIORITY", "").(string),
	Assignee:    envGet("JIRA_ASSIGNEE", "").(string),
	Reporter:    envGet("JIRA_REPORTER", "").(string),
}

var jiraOutput = common.OutputOptions{
	Output: envGet("JIRA_OUTPUT", "").(string),
	Query:  envGet("JIRA_OUTPUT_QUERY", "").(string),
}

func jiraNew(stdout *common.Stdout) *vendors.Jira {

	common.Debug("Jira", jiraOptions, stdout)
	common.Debug("Jira", jiraOutput, stdout)

	descriptionBytes, err := utils.Content(jiraOptions.Description)
	if err != nil {
		stdout.Panic(err)
	}
	jiraOptions.Description = string(descriptionBytes)

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
	flags.StringVar(&jiraOptions.ProjectKey, "jira-project-key", jiraOptions.ProjectKey, "Jira project key")
	flags.StringVar(&jiraOptions.IssueType, "jira-issue-type", jiraOptions.IssueType, "Jira issue type")
	flags.StringVar(&jiraOptions.Summary, "jira-summary", jiraOptions.Summary, "Jira summary")
	flags.StringVar(&jiraOptions.Description, "jira-description", jiraOptions.Description, "Jira description")
	flags.StringSliceVar(&jiraOptions.Labels, "jira-labels", jiraOptions.Labels, "Jira labels")
	flags.StringVar(&jiraOptions.Priority, "jira-priority", jiraOptions.Priority, "Jira priority")
	flags.StringVar(&jiraOptions.Assignee, "jira-assignee", jiraOptions.Assignee, "Jira assignee")
	flags.StringVar(&jiraOptions.Reporter, "jira-reporter", jiraOptions.Reporter, "Jira reporter")
	flags.StringVar(&jiraOutput.Output, "jira-output", jiraOutput.Output, "Jira output")
	flags.StringVar(&jiraOutput.Query, "jira-output-query", jiraOutput.Query, "Jira output query")

	jiraCmd.AddCommand(&cobra.Command{
		Use:   "create-issue",
		Short: "Create issue",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Jira creating issue...")
			bytes, err := jiraNew(stdout).CreateIssue()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jiraOutput, "Jira", jiraOptions, bytes, stdout)
		},
	})

	return &jiraCmd
}
