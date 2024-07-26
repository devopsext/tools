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

var JiraIssueOptions = vendors.JiraIssueOptions{
	ProjectKey:   envGet("JIRA_ISSUE_PROJECT_KEY", "").(string),
	Type:         envGet("JIRA_ISSUE_TYPE", "").(string),
	Priority:     envGet("JIRA_ISSUE_PRIORITY", "").(string),
	Assignee:     envGet("JIRA_ISSUE_ASSIGNEE", "").(string),
	Reporter:     envGet("JIRA_ISSUE_REPORTER", "").(string),
	IdOrKey:      envGet("JIRA_ISSUE_ID_OR_KEY", "").(string),
	Summary:      envGet("JIRA_ISSUE_SUMMARY", "").(string),
	Description:  envGet("JIRA_ISSUE_DESCRIPTION", "").(string),
	CustomFields: envGet("JIRA_ISSUE_CUSTOM_FIELDS", "").(string),
	Labels:       strings.Split(envGet("JIRA_ISSUE_LABELS", "").(string), ","),
	TransitionID: envGet("JIRA_ISSUE_STATUS", "").(string),
}

var jiraIssueAddCommentOptions = vendors.JiraAddIssueCommentOptions{
	Body: envGet("JIRA_ISSUE_COMMENT_BODY", "").(string),
}

var jiraIssueAddAttachmentOptions = vendors.JiraAddIssueAttachmentOptions{
	File: envGet("JIRA_ISSUE_ATTACHMENT_FILE", "").(string),
	Name: envGet("JIRA_ISSUE_ATTACHMENT_NAME", "").(string),
}

var jiraIssueSearchOptions = vendors.JiraSearchIssueOptions{
	SearchPattern: envGet("JIRA_ISSUE_SEARCH_PATTERN", "").(string),
	MaxResults:    envGet("JIRA_ISSUE_SEARCH_MAX_RESULTS", 50).(int),
}

var jiraAssetSearchOptions = vendors.JiraSearchAssetOptions{
	SearchPattern: envGet("JIRA_ASSET_SEARCH_PATTERN", "").(string),
	ResultPerPage: envGet("JIRA_ASSET_SEARCH_RESULT_PER_PAGE", 50).(int),
}

var jiraAssetCreateOptions = vendors.JiraCreateAssetOptions{
	Name:        envGet("JIRA_ASSET_CREATE_NAME", "").(string),
	Description: envGet("JIRA_ASSET_CREATE_DESCRIPTION", "").(string),
}

var jiraAssetUpdateOptions = vendors.JiraUpdateAssetOptions{
	ObjectId: envGet("JIRA_ASSET_OBJECT_ID", "").(string),
	Json:     envGet("JIRA_ASSET_JSON", "").(string),
}

var jiraOutput = common.OutputOptions{
	Output: envGet("JIRA_OUTPUT", "").(string),
	Query:  envGet("JIRA_OUTPUT_QUERY", "").(string),
}

func jiraNew(stdout *common.Stdout) *vendors.Jira {

	common.Debug("Jira", jiraOptions, stdout)
	common.Debug("Jira", jiraOutput, stdout)

	return vendors.NewJira(jiraOptions)
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
	flags.StringVar(&JiraIssueOptions.IdOrKey, "jira-issue-id-or-key", JiraIssueOptions.IdOrKey, "Jira issue ID or key")
	flags.StringVar(&JiraIssueOptions.Summary, "jira-issue-summary", JiraIssueOptions.Summary, "Jira issue summary")
	flags.StringVar(&JiraIssueOptions.Description, "jira-issue-description", JiraIssueOptions.Description, "Jira issue description")
	flags.StringVar(&JiraIssueOptions.CustomFields, "jira-issue-custom-fields", JiraIssueOptions.CustomFields, "Jira issue custom fields file")
	flags.StringSliceVar(&JiraIssueOptions.Labels, "jira-issue-labels", JiraIssueOptions.Labels, "Jira issue labels")
	jiraCmd.AddCommand(issueCmd)

	// tools jira issue create --jira-params --create-issue-params
	issueCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create issue",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Jira creating issue...")
			common.Debug("Jira", JiraIssueOptions, stdout)
			common.Debug("Jira", JiraIssueOptions, stdout)

			descriptionBytes, err := utils.Content(JiraIssueOptions.Description)
			if err != nil {
				stdout.Panic(err)
			}
			JiraIssueOptions.Description = string(descriptionBytes)

			bytes, err := jiraNew(stdout).CreateIssue(JiraIssueOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jiraOutput, "Jira", []interface{}{jiraOptions, JiraIssueOptions}, bytes, stdout)
		},
	}
	flags = issueCreateCmd.PersistentFlags()
	flags.StringVar(&JiraIssueOptions.ProjectKey, "jira-issue-project-key", JiraIssueOptions.ProjectKey, "Jira issue project key")
	flags.StringVar(&JiraIssueOptions.Type, "jira-issue-type", JiraIssueOptions.Type, "Jira issue type")
	flags.StringVar(&JiraIssueOptions.Priority, "jira-issue-priority", JiraIssueOptions.Priority, "Jira issue priority")
	flags.StringVar(&JiraIssueOptions.Assignee, "jira-issue-assignee", JiraIssueOptions.Assignee, "Jira issue assignee")
	flags.StringVar(&JiraIssueOptions.Reporter, "jira-issue-reporter", JiraIssueOptions.Reporter, "Jira issue reporter")
	issueCmd.AddCommand(issueCreateCmd)

	// tools jira issue add-comment --jira-params --issue-params --add-comment-params
	issueAddCommentCmd := &cobra.Command{
		Use:   "add-comment",
		Short: "Issue add comment",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Jira issue adding comment...")
			common.Debug("Jira", JiraIssueOptions, stdout)
			common.Debug("Jira", jiraIssueAddCommentOptions, stdout)

			bodyBytes, err := utils.Content(jiraIssueAddCommentOptions.Body)
			if err != nil {
				stdout.Panic(err)
			}
			jiraIssueAddCommentOptions.Body = string(bodyBytes)

			bytes, err := jiraNew(stdout).IssueAddComment(JiraIssueOptions, jiraIssueAddCommentOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jiraOutput, "Jira", []interface{}{jiraOptions, JiraIssueOptions, jiraIssueAddCommentOptions}, bytes, stdout)
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
			common.Debug("Jira", JiraIssueOptions, stdout)
			common.Debug("Jira", jiraIssueAddAttachmentOptions, stdout)

			if utils.IsEmpty(jiraIssueAddAttachmentOptions.Name) && utils.FileExists(jiraIssueAddAttachmentOptions.File) {
				jiraIssueAddAttachmentOptions.Name = filepath.Base(jiraIssueAddAttachmentOptions.File)
			}

			fileBytes, err := utils.Content(jiraIssueAddAttachmentOptions.File)
			if err != nil {
				stdout.Panic(err)
			}
			jiraIssueAddAttachmentOptions.File = string(fileBytes)

			bytes, err := jiraNew(stdout).AddIssueAttachment(JiraIssueOptions, jiraIssueAddAttachmentOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jiraOutput, "Jira", []interface{}{jiraOptions, JiraIssueOptions, jiraIssueAddAttachmentOptions}, bytes, stdout)
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
			common.Debug("Jira", JiraIssueOptions, stdout)

			descriptionBytes, err := utils.Content(JiraIssueOptions.Description)
			if err != nil {
				stdout.Panic(err)
			}
			JiraIssueOptions.Description = string(descriptionBytes)

			bytes, err := jiraNew(stdout).UpdateIssue(JiraIssueOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jiraOutput, "Jira", []interface{}{jiraOptions, JiraIssueOptions}, bytes, stdout)
		},
	}
	issueCmd.AddCommand(issueUpdateCmd)

	// tools jira issue change-transitions --jira-params --issue-params
	issueChangeTransitionsCmd := &cobra.Command{
		Use:   "change-transitions",
		Short: "Transitions change",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Jira issue updating...")
			common.Debug("Jira", JiraIssueOptions, stdout)

			statusBytes, err := utils.Content(JiraIssueOptions.TransitionID)
			if err != nil {
				stdout.Panic(err)
			}
			JiraIssueOptions.TransitionID = string(statusBytes)

			bytes, err := jiraNew(stdout).ChangeIssueTransitions(JiraIssueOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jiraOutput, "Jira", []interface{}{jiraOptions, JiraIssueOptions}, bytes, stdout)
		},
	}
	flags = issueChangeTransitionsCmd.PersistentFlags()
	flags.StringVar(&JiraIssueOptions.TransitionID, "jira-issue-status", JiraIssueOptions.TransitionID, "Jira issue status")
	issueCmd.AddCommand(issueChangeTransitionsCmd)

	issueSearchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search issue",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Jira issue searching...")
			common.Debug("Jira", JiraIssueOptions, stdout)
			common.Debug("Jira", jiraIssueSearchOptions, stdout)

			searchBytes, err := utils.Content(jiraIssueSearchOptions.SearchPattern)
			if err != nil {
				stdout.Panic(err)
			}
			jiraIssueSearchOptions.SearchPattern = string(searchBytes)

			bytes, err := jiraNew(stdout).SearchIssue(jiraIssueSearchOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jiraOutput, "Jira", []interface{}{jiraOptions, jiraIssueSearchOptions}, bytes, stdout)
		},
	}
	flags = issueSearchCmd.PersistentFlags()
	flags.StringVar(&jiraIssueSearchOptions.SearchPattern, "jira-issue-search-pattern", jiraIssueSearchOptions.SearchPattern, "Jira issue search pattern")
	flags.IntVar(&jiraIssueSearchOptions.MaxResults, "jira-issue-search-max-results", jiraIssueSearchOptions.MaxResults, "Jira issue search max results")
	issueCmd.AddCommand((issueSearchCmd))

	assetCmd := &cobra.Command{
		Use:   "asset",
		Short: "Asset methods",
	}
	jiraCmd.AddCommand(assetCmd)

	assetSearchCmd := &cobra.Command{
		Use:   "search",
		Short: "Search assets",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Jira asset searching...")
			common.Debug("Jira", jiraAssetSearchOptions, stdout)

			searchBytes, err := utils.Content(jiraAssetSearchOptions.SearchPattern)
			if err != nil {
				stdout.Panic(err)
			}
			jiraAssetSearchOptions.SearchPattern = string(searchBytes)

			bytes, err := jiraNew(stdout).SearchAssets(jiraAssetSearchOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jiraOutput, "Jira", []interface{}{jiraOptions, jiraAssetSearchOptions}, bytes, stdout)
		},
	}
	flags = assetSearchCmd.PersistentFlags()
	flags.StringVar(&jiraAssetSearchOptions.SearchPattern, "jira-asset-search-pattern", jiraAssetSearchOptions.SearchPattern, "Jira asset search pattern")
	flags.IntVar(&jiraAssetSearchOptions.ResultPerPage, "jira-asset-search-results-per-page", jiraAssetSearchOptions.ResultPerPage, "Jira asset result per page")
	assetCmd.AddCommand(assetSearchCmd)

	assetCreateCmd := &cobra.Command{
		Use:   "create",
		Short: "Create asset",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Jira asset creating...")
			common.Debug("Jira", jiraAssetCreateOptions, stdout)

			bytes, err := jiraNew(stdout).CreateAsset(jiraAssetCreateOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jiraOutput, "Jira", []interface{}{jiraOptions, jiraAssetCreateOptions}, bytes, stdout)
		},
	}
	flags = assetCreateCmd.PersistentFlags()
	flags.StringVar(&jiraAssetCreateOptions.Name, "jira-asset-create-name", jiraAssetCreateOptions.Name, "Jira asset name")
	flags.StringVar(&jiraAssetCreateOptions.Description, "jira-asset-create-rdescription", jiraAssetCreateOptions.Description, "Jira asset description")
	// ... all options should be added, like value, etc.
	assetCmd.AddCommand(assetCreateCmd)

	assetUpdateCmd := &cobra.Command{
		Use:   "update",
		Short: "Update asset",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Jira asset updating...")
			common.Debug("Jira", jiraAssetUpdateOptions, stdout)

			jsonBytes, err := utils.Content(jiraAssetUpdateOptions.Json)
			if err != nil {
				stdout.Panic(err)
			}
			jiraAssetUpdateOptions.Json = string(jsonBytes)

			bytes, err := jiraNew(stdout).UpdateAsset(jiraAssetUpdateOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jiraOutput, "Jira", []interface{}{jiraOptions, jiraAssetUpdateOptions}, bytes, stdout)
		},
	}
	flags = assetUpdateCmd.PersistentFlags()
	flags.StringVar(&jiraAssetUpdateOptions.ObjectId, "jira-asset-update-object-id", jiraAssetUpdateOptions.ObjectId, "Jira asset object id")
	flags.StringVar(&jiraAssetUpdateOptions.Json, "jira-asset-update-json", jiraAssetUpdateOptions.Json, "Jira asset json")
	assetCmd.AddCommand(assetUpdateCmd)

	return &jiraCmd
}
