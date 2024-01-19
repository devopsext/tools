package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/spf13/cobra"
)

var googleOptions = vendors.GoogleOptions{
	Timeout:           envGet("GOOGLE_TIMEOUT", 30).(int),
	Insecure:          envGet("GOOGLE_INSECURE", false).(bool),
	OAuthClientID:     envGet("GOOGLE_OAUTH_CLIENT_ID", "").(string),
	OAuthClientSecret: envGet("GOOGLE_OAUTH_CLIENT_SECRET", "").(string),
	RefreshToken:      envGet("GOOGLE_REFRESH_TOKEN", "").(string),
	AccessToken:       envGet("GOOGLE_ACCESS_TOKEN", "").(string),
}

var googleCalendarOptions = vendors.GoogleCalendarOptions{
	ID:                 envGet("GOOGLE_CALENDAR_ID", "").(string),
	TimeMin:            envGet("GOOGLE_CALENDAR_TIME_MIN", "").(string),
	TimeMax:            envGet("GOOGLE_CALENDAR_TIME_MAX", "").(string),
	AlwaysIncludeEmail: envGet("GOOGLE_CALENDAR_ALWAYS_INCLUDE_EMAIL", true).(bool),
	OrderBy:            envGet("GOOGLE_CALENDAR_ORDER_BY", "").(string),
	Q:                  envGet("GOOGLE_CALENDAR_Q", "").(string),
	SingleEvents:       envGet("GOOGLE_CALENDAR_SINGLE_EVENTS", false).(bool),
}

var googleOutput = common.OutputOptions{
	Output: envGet("GOOGLE_OUTPUT", "").(string),
	Query:  envGet("GOOGLE_OUTPUT_QUERY", "").(string),
}

func googleNew(stdout *common.Stdout) *vendors.Google {

	common.Debug("Google", googleOptions, stdout)
	common.Debug("Google", googleOutput, stdout)

	google, err := vendors.NewGoogle(googleOptions, stdout)
	if err != nil {
		stdout.Panic(err)
	}
	return google
}

func NewGoogleCommand() *cobra.Command {

	googleCmd := cobra.Command{
		Use:   "google",
		Short: "Google tools",
	}
	flags := googleCmd.PersistentFlags()
	flags.IntVar(&googleOptions.Timeout, "google-timeout", googleOptions.Timeout, "Google timeout")
	flags.BoolVar(&googleOptions.Insecure, "google-insecure", googleOptions.Insecure, "Google insecure")
	flags.StringVar(&googleOptions.OAuthClientID, "google-oauth-client-id", googleOptions.OAuthClientID, "Google OAuth client id")
	flags.StringVar(&googleOptions.OAuthClientSecret, "google-oauth-client-secret", googleOptions.OAuthClientSecret, "Google OAuth client secret")
	flags.StringVar(&googleOptions.RefreshToken, "google-refresh-token", googleOptions.RefreshToken, "Google refresh token")
	flags.StringVar(&googleOptions.AccessToken, "google-access-token", googleOptions.AccessToken, "Google access token")
	flags.StringVar(&googleOutput.Output, "google-output", googleOutput.Output, "Google output")
	flags.StringVar(&googleOutput.Query, "google-output-query", googleOutput.Query, "Google output query")

	calendarCmd := &cobra.Command{
		Use:   "calendar",
		Short: "Calendar methods",
	}
	flags = calendarCmd.PersistentFlags()
	flags.StringVar(&googleCalendarOptions.ID, "google-calendar-id", googleCalendarOptions.ID, "Google calendar id")
	googleCmd.AddCommand(calendarCmd)

	calendarEventsCmd := &cobra.Command{
		Use:   "get-events",
		Short: "Calendar get events",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Google callendar getting events...")
			common.Debug("Google", googleCalendarOptions, stdout)

			bytes, err := googleNew(stdout).GetCalendarEvents(googleCalendarOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(googleOutput, "Google", []interface{}{googleOptions, googleCalendarOptions}, bytes, stdout)
		},
	}
	flags = calendarEventsCmd.PersistentFlags()
	flags.StringVar(&googleCalendarOptions.TimeMin, "google-calendar-time-min", googleCalendarOptions.TimeMin, "Google calendar time min")
	flags.StringVar(&googleCalendarOptions.TimeMax, "google-calendar-time-max", googleCalendarOptions.TimeMax, "Google calendar time max")
	flags.StringVar(&googleCalendarOptions.OrderBy, "google-calendar-oreder-by", googleCalendarOptions.OrderBy, "Google calendar oreder by")
	flags.StringVar(&googleCalendarOptions.Q, "google-calendar-q", googleCalendarOptions.Q, "Google calendar q")
	flags.BoolVar(&googleCalendarOptions.AlwaysIncludeEmail, "google-calendar-always-include-email", googleCalendarOptions.AlwaysIncludeEmail, "Google calendar always include email")
	flags.BoolVar(&googleCalendarOptions.SingleEvents, "google-calendar-single-events", googleCalendarOptions.SingleEvents, "Google calendar single events")
	calendarCmd.AddCommand(calendarEventsCmd)

	return &googleCmd
}
