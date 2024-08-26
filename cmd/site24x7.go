package cmd

import (
	"strings"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/spf13/cobra"
)

var site24x7MonitorOptions = vendors.Site24x7MonitorOptions{
	ID: envGet("SITE24X7_MONITOR_ID", "").(string),
}

var site24x7WebsiteMonitorOptions = vendors.Site24x7WebsiteMonitorOptions{
	Name:                  envGet("SITE24X7_WEBSITE_MONITOR_NAME", "").(string),
	URL:                   envGet("SITE24X7_WEBSITE_MONITOR_URL", "").(string),
	Method:                envGet("SITE24X7_WEBSITE_MONITOR_METHOD", "GET").(string),
	Frequency:             envGet("SITE24X7_WEBSITE_MONITOR_FERQUENCY", "1440").(string),
	Timeout:               envGet("SITE24X7_WEBSITE_MONITOR_TIMEOUT", 30).(int),
	Countries:             strings.Split(envGet("SITE24X7_WEBSITE_MONITOR_COUNTRIES", "").(string), ","),
	UserAgent:             envGet("SITE24X7_WEBSITE_MONITOR_USER_AGENT", "").(string),
	UseNameServer:         envGet("SITE24X7_WEBSITE_MONITOR_USE_NAME_SERVER", false).(bool),
	NotificationProfileID: envGet("SITE24X7_WEBSITE_MONITOR_USER_AGENT", "").(string),
}

var site24x7LogReportOptions = vendors.Site24x7LogReportOptions{
	Site24x7MonitorOptions: vendors.Site24x7MonitorOptions{
		ID: envGet("SITE24X7_MONITOR_ID", "").(string),
	},
	StartDate: envGet("SITE24X7_LOG_REPORT_START_DATE", "").(string),
	EndDate:   envGet("SITE24X7_LOG_REPORT_END_DATE", "").(string),
}

var site24x7Options = vendors.Site24x7Options{
	Timeout:      envGet("SITE24X7_TIMEOUT", 30).(int),
	Insecure:     envGet("SITE24X7_INSECURE", false).(bool),
	ClientID:     envGet("SITE24X7_CLIENT_ID", "").(string),
	ClientSecret: envGet("SITE24X7_CLIENT_SECRET", "").(string),
	RefreshToken: envGet("SITE24X7_REFRESH_TOKEN", "").(string),
	AccessToken:  envGet("SITE24X7_ACCESS_TOKEN", "").(string),
}

var site24x7Output = common.OutputOptions{
	Output: envGet("SITE24X7_OUTPUT", "").(string),
	Query:  envGet("SITE24X7_OUTPUT_QUERY", "").(string),
}

func site24x7New(stdout *common.Stdout) *vendors.Site24x7 {

	common.Debug("Site24x7", site24x7Options, stdout)
	common.Debug("Site24x7", site24x7Output, stdout)

	site24x7 := vendors.NewSite24x7(site24x7Options, stdout)
	if site24x7 == nil {
		stdout.Panic("No site24x7")
	}
	return site24x7
}

func NewSite24x7Command() *cobra.Command {

	site24x7Cmd := &cobra.Command{
		Use:   "site24x7",
		Short: "Site24x7 tools",
	}
	flags := site24x7Cmd.PersistentFlags()
	flags.IntVar(&site24x7Options.Timeout, "site24x7-timeout", site24x7Options.Timeout, "Site24x7 timeout in seconds")
	flags.BoolVar(&site24x7Options.Insecure, "site24x7-insecure", site24x7Options.Insecure, "Site24x7 insecure")
	flags.StringVar(&site24x7Options.ClientID, "site24x7-client-id", site24x7Options.ClientID, "Site24x7 client ID")
	flags.StringVar(&site24x7Options.ClientSecret, "site24x7-client-secret", site24x7Options.ClientSecret, "Site24x7 client secret")
	flags.StringVar(&site24x7Options.RefreshToken, "site24x7-refresh-token", site24x7Options.RefreshToken, "Site24x7 refresh token")
	flags.StringVar(&site24x7Options.AccessToken, "site24x7-access-token", site24x7Options.AccessToken, "Site24x7 access token")
	flags.StringVar(&site24x7Output.Output, "site24x7-output", site24x7Output.Output, "Site24x7 output")
	flags.StringVar(&site24x7Output.Query, "site24x7-output-query", site24x7Output.Query, "Site24x7 output query")

	site24x7CreateMonitorCmd := &cobra.Command{
		Use:   "create-website-monitor",
		Short: "Create website montitor",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Creating website monitor...")
			common.Debug("Site24x7", site24x7WebsiteMonitorOptions, stdout)

			bytes, err := site24x7New(stdout).CreateWebsiteMonitor(site24x7WebsiteMonitorOptions)
			if err != nil {
				stdout.Error("Error: %v %s", err, string(bytes))
				return
			}
			common.OutputJson(site24x7Output, "site24x7", []interface{}{site24x7Options, site24x7WebsiteMonitorOptions}, bytes, stdout)
		},
	}
	flags = site24x7CreateMonitorCmd.PersistentFlags()
	flags.StringVar(&site24x7WebsiteMonitorOptions.Name, "site24x7-website-monitor-name", site24x7WebsiteMonitorOptions.Name, "Site24x7 website monitor name")
	flags.StringVar(&site24x7WebsiteMonitorOptions.URL, "site24x7-website-monitor-url", site24x7WebsiteMonitorOptions.URL, "Site24x7 website monitor URL")
	flags.StringVar(&site24x7WebsiteMonitorOptions.Method, "site24x7-website-monitor-method", site24x7WebsiteMonitorOptions.Method, "Site24x7 website monitor method")
	flags.StringVar(&site24x7WebsiteMonitorOptions.Frequency, "site24x7-website-monitor-frequency", site24x7WebsiteMonitorOptions.Frequency, "Site24x7 website monitor frequency")
	flags.IntVar(&site24x7WebsiteMonitorOptions.Timeout, "site24x7-website-monitor-timeout", site24x7WebsiteMonitorOptions.Timeout, "Site24x7 website monitor timeout in seconds")
	flags.StringSliceVar(&site24x7WebsiteMonitorOptions.Countries, "site24x7-website-monitor-countries", site24x7WebsiteMonitorOptions.Countries, "Site24x7 website monitor countries")
	flags.StringVar(&site24x7WebsiteMonitorOptions.UserAgent, "site24x7-website-monitor-user-agent", site24x7WebsiteMonitorOptions.UserAgent, "Site24x7 website monitor user agent")
	flags.BoolVar(&site24x7WebsiteMonitorOptions.UseNameServer, "site24x7-website-monitor-use-name-server", site24x7WebsiteMonitorOptions.UseNameServer, "Site24x7 website monitor use name server")
	flags.StringSliceVar(&site24x7WebsiteMonitorOptions.UserGroupIDs, "site24x7-website-monitor-user-group-ids", site24x7WebsiteMonitorOptions.UserGroupIDs, "Site24x7 website monitor user group ids")
	site24x7Cmd.AddCommand(site24x7CreateMonitorCmd)

	site24x7DeleteMonitorCmd := &cobra.Command{
		Use:   "delete-monitor",
		Short: "Delete montitor",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Deleting monitor...")
			common.Debug("Site24x7", site24x7MonitorOptions, stdout)

			bytes, err := site24x7New(stdout).DeleteMonitor(site24x7MonitorOptions)
			if err != nil {
				stdout.Error("Error: %v %s", err, string(bytes))
				return
			}
			common.OutputJson(site24x7Output, "site24x7", []interface{}{site24x7Options, site24x7MonitorOptions}, bytes, stdout)
		},
	}
	flags = site24x7DeleteMonitorCmd.PersistentFlags()
	flags.StringVar(&site24x7MonitorOptions.ID, "site24x7-monitor-id", site24x7MonitorOptions.ID, "Site24x7 monitor id")
	site24x7Cmd.AddCommand(site24x7DeleteMonitorCmd)

	site24x7PollMonitorCmd := &cobra.Command{
		Use:   "poll-monitor",
		Short: "Poll montitor",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Polling monitor...")
			common.Debug("Site24x7", site24x7MonitorOptions, stdout)

			bytes, err := site24x7New(stdout).PollMonitor(site24x7MonitorOptions)
			if err != nil {
				stdout.Error("Error: %v %s", err, string(bytes))
				return
			}
			common.OutputJson(site24x7Output, "site24x7", []interface{}{site24x7Options, site24x7MonitorOptions}, bytes, stdout)
		},
	}
	flags = site24x7PollMonitorCmd.PersistentFlags()
	flags.StringVar(&site24x7MonitorOptions.ID, "site24x7-monitor-id", site24x7MonitorOptions.ID, "Site24x7 monitor id")
	site24x7Cmd.AddCommand(site24x7PollMonitorCmd)

	site24x7PollingStatusCmd := &cobra.Command{
		Use:   "polling-status",
		Short: "Polling status",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Polling status...")
			common.Debug("Site24x7", site24x7MonitorOptions, stdout)

			bytes, err := site24x7New(stdout).PollingStatus(site24x7MonitorOptions)
			if err != nil {
				stdout.Error("Error: %v %s", err, string(bytes))
				return
			}
			common.OutputJson(site24x7Output, "site24x7", []interface{}{site24x7Options, site24x7MonitorOptions}, bytes, stdout)
		},
	}
	flags = site24x7PollingStatusCmd.PersistentFlags()
	flags.StringVar(&site24x7MonitorOptions.ID, "site24x7-monitor-id", site24x7MonitorOptions.ID, "Site24x7 monitor id")
	site24x7Cmd.AddCommand(site24x7PollingStatusCmd)

	site24x7LogReportCmd := &cobra.Command{
		Use:   "log-report",
		Short: "Logging report",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Logging report...")
			common.Debug("Site24x7", site24x7LogReportOptions, stdout)

			bytes, err := site24x7New(stdout).LogReport(site24x7LogReportOptions)
			if err != nil {
				stdout.Error("Error: %v %s", err, string(bytes))
				return
			}
			common.OutputJson(site24x7Output, "site24x7", []interface{}{site24x7Options, site24x7LogReportOptions}, bytes, stdout)
		},
	}
	flags = site24x7LogReportCmd.PersistentFlags()
	flags.StringVar(&site24x7LogReportOptions.ID, "site24x7-monitor-id", site24x7LogReportOptions.ID, "Site24x7 monitor id")
	flags.StringVar(&site24x7LogReportOptions.StartDate, "site24x7-log-report-start-date", site24x7LogReportOptions.StartDate, "Site24x7 log report start date")
	flags.StringVar(&site24x7LogReportOptions.EndDate, "site24x7-log-report-end-date", site24x7LogReportOptions.EndDate, "Site24x7 log report end date")
	site24x7Cmd.AddCommand(site24x7LogReportCmd)

	return site24x7Cmd
}
