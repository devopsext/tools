package cmd

import (
	"strings"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/spf13/cobra"
)

var site24x7WebsiteMonitorOptions = vendors.Site24x7WebsiteMonitorOptions{
	Name:          envGet("SITE24X7_WEBSITE_MONITOR_NAME", "").(string),
	URL:           envGet("SITE24X7_WEBSITE_MONITOR_URL", "").(string),
	Method:        envGet("SITE24X7_WEBSITE_MONITOR_METHOD", "GET").(string),
	Frequency:     envGet("SITE24X7_WEBSITE_MONITOR_FERQUENCY", "1440").(string),
	Timeout:       envGet("SITE24X7_WEBSITE_MONITOR_TIMEOUT", 30).(int),
	Countries:     strings.Split(envGet("SITE24X7_WEBSITE_MONITOR_COUNTRIES", "").(string), ","),
	UserAgent:     envGet("SITE24X7_WEBSITE_MONITOR_USER_AGENT", "").(string),
	UseNameServer: envGet("SITE24X7_WEBSITE_MONITOR_USE_NAME_SERVER", false).(bool),
}

var site24x7Options = vendors.Site24x7Options{
	Timeout:      envGet("SITE24X7_TIMEOUT", 30).(int),
	Insecure:     envGet("SITE24X7_INSECURE", false).(bool),
	ClientID:     envGet("SITE24X7_CLIENT_ID", "").(string),
	ClientSecret: envGet("SITE24X7_CLIENT_SECRET", "").(string),
	RefreshToken: envGet("SITE24X7_REFRESH_TOKEN", "").(string),
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
				stdout.Error(err)
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
	site24x7Cmd.AddCommand(site24x7CreateMonitorCmd)

	return site24x7Cmd
}
