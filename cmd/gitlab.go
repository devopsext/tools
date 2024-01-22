package cmd

import (
	"strings"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/spf13/cobra"
)

var gitlabOptions = vendors.GitlabOptions{
	Timeout:  envGet("GITLAB_TIMEOUT", 30).(int),
	Insecure: envGet("GITLAB_INSECURE", false).(bool),
	URL:      envGet("GITLAB_URL", "").(string),
	Token:    envGet("GITLAB_TOKEN", "").(string),
}

var gitlabOutput = common.OutputOptions{
	Output: envGet("GITLAB_OUTPUT", "").(string),
	Query:  envGet("GITLAB_OUTPUT_QUERY", "").(string),
}

var pipelineOptions = vendors.GitlabPipelineOptions{
	ProjectID: envGet("GITLAB_PIPELINE_PROJECT_ID", 0).(int),
	Scope:     envGet("GITLAB_PIPELINE_SCOPE", "finished").(string),
	Status:    envGet("GITLAB_PIPELINE_STATUS", "").(string),
	Source:    envGet("GITLAB_PIPELINE_SOURCE", "").(string),
	Ref:       envGet("GITLAB_PIPELINE_REF", "").(string),
	OrderBy:   envGet("GITLAB_PIPELINE_OREDR_BY", "updated_at").(string),
	Sort:      envGet("GITLAB_PIPELINE_SORT", "desc").(string),
	Limit:     envGet("GITLAB_PIPELINE_LIMIT", 1).(int),
}

var pipelineGetVariablesOptions = vendors.GitlabGetPipelineVariablesOptions{
	Query: strings.Split(envGet("GITLAB_PIPELINE_VARIABLE_QUERY", "").(string), ","),
}

func gitlabNew(stdout *common.Stdout) *vendors.Gitlab {

	common.Debug("Gitlab", gitlabOptions, stdout)
	common.Debug("Gitlab", gitlabOutput, stdout)

	return vendors.NewGitlab(gitlabOptions)
}

func NewGitlabCommand() *cobra.Command {

	gitlabCmd := &cobra.Command{
		Use:   "gitlab",
		Short: "Gitlab tools",
	}
	flags := gitlabCmd.PersistentFlags()
	flags.IntVar(&gitlabOptions.Timeout, "gitlab-timeout", gitlabOptions.Timeout, "Gitlab Timeout in seconds")
	flags.BoolVar(&gitlabOptions.Insecure, "gitlab-insecure", gitlabOptions.Insecure, "Gitlab Insecure")
	flags.StringVar(&gitlabOptions.URL, "gitlab-url", gitlabOptions.URL, "Gitlab URL")
	flags.StringVar(&gitlabOptions.Token, "gitlab-token", gitlabOptions.Token, "Gitlab Token")
	flags.StringVar(&gitlabOutput.Output, "gitlab-output", gitlabOutput.Output, "Gitlab Output")
	flags.StringVar(&gitlabOutput.Query, "gitlab-output-query", gitlabOutput.Query, "Gitlab Output Query")

	pipelineCmd := &cobra.Command{
		Use:   "pipeline",
		Short: "Get pipeline from Gitlab",
	}
	flags = pipelineCmd.PersistentFlags()
	flags.IntVar(&pipelineOptions.ProjectID, "gitlab-pipeline-project-id", pipelineOptions.ProjectID, "Gitlab pipeline project ID")
	flags.StringVar(&pipelineOptions.Scope, "gitlab-pipeline-scope", pipelineOptions.Scope, "Gitlab pipeline scope")
	flags.StringVar(&pipelineOptions.Status, "gitlab-pipeline-status", pipelineOptions.Status, "Gitlab pipeline status")
	flags.StringVar(&pipelineOptions.Source, "gitlab-pipeline-source", pipelineOptions.Source, "Gitlab pipeline source")
	flags.StringVar(&pipelineOptions.Ref, "gitlab-pipeline-ref", pipelineOptions.Ref, "Gitlab pipeline Ref")
	flags.StringVar(&pipelineOptions.OrderBy, "gitlab-pipeline-order-by", pipelineOptions.OrderBy, "Gitlab pipeline order by")
	flags.StringVar(&pipelineOptions.Sort, "gitlab-pipeline-sort", pipelineOptions.Sort, "Gitlab pipeline sort")
	flags.IntVar(&pipelineOptions.Limit, "gitlab-pipeline-limit", pipelineOptions.Limit, "Gitlab pipeline limit")
	gitlabCmd.AddCommand(pipelineCmd)

	pipelineCmd.AddCommand(&cobra.Command{
		Use:   "last",
		Short: "Get last successful gitlab pipeline",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Getting pipelines…")
			common.Debug("Gitlab", pipelineOptions, stdout)

			bytes, err := gitlabNew(stdout).GetLastPipeline(pipelineOptions.ProjectID, pipelineOptions.Ref)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(gitlabOutput, "Gitlab", []interface{}{gitlabOptions, pipelineOptions}, bytes, stdout)
		},
	})

	pipelineCmd.AddCommand(&cobra.Command{
		Use:   "last-variables",
		Short: "Get last successful gitlab pipeline variables",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Getting pipelines…")
			common.Debug("Gitlab", pipelineOptions, stdout)

			bytes, err := gitlabNew(stdout).GetLastPipelineVariables(pipelineOptions.ProjectID, pipelineOptions.Ref)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(gitlabOutput, "Gitlab", []interface{}{gitlabOptions, pipelineOptions}, bytes, stdout)
		},
	})

	pipelineGetVariablesCmd := &cobra.Command{
		Use:   "get-variables",
		Short: "Gitlab pipeline get variables",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Gitlab pipeline getting events...")
			common.Debug("Gitlab", pipelineOptions, stdout)
			common.Debug("Gitlab", pipelineGetVariablesOptions, stdout)

			bytes, err := gitlabNew(stdout).GetPipelineVariables(pipelineOptions, pipelineGetVariablesOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(gitlabOutput, "Gitlab", []interface{}{gitlabOptions, pipelineOptions, pipelineGetVariablesOptions}, bytes, stdout)
		},
	}
	flags = pipelineGetVariablesCmd.PersistentFlags()
	flags.StringSliceVar(&pipelineGetVariablesOptions.Query, "gitlab-pipeline-variable-query", pipelineGetVariablesOptions.Query, "Gitlab pipeline variable query")
	pipelineCmd.AddCommand(pipelineGetVariablesCmd)

	return gitlabCmd
}
