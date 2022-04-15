package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/spf13/cobra"
)

var jiraOptions = vendors.JiraOptions{
	URL:      envGet("JIRA_URL", "").(string),
	Timeout:  envGet("JIRA_TIMEOUT", 30).(int),
	Insecure: envGet("JIRA_INSECURE", false).(bool),
	User:     envGet("JIRA_USER", "").(string),
	Password: envGet("JIRA_PASSWORD", "").(string),
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

	jiraCmd.AddCommand(&cobra.Command{
		Use:   "create-task",
		Short: "Create task",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Jira creating task...")
			bytes, err := jiraNew(stdout).CreateTask()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jiraOutput, "Jira", jiraOptions, bytes, stdout)
		},
	})

	return &jiraCmd
}
