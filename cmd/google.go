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
}

var googleCalendarOptions = vendors.GoogleCalendarOptions{
	ID: envGet("GOOGLE_CALENDAR_ID", "").(string),
}

var googleCalendarGetEventsOptions = vendors.GoogleCalendarGetEventsOptions{
	TimeMin:      envGet("GOOGLE_CALENDAR_TIME_MIN", "").(string),
	TimeMax:      envGet("GOOGLE_CALENDAR_TIME_MAX", "").(string),
	TimeZone:     envGet("GOOGLE_CALENDAR_TIMEZONE", "").(string),
	OrderBy:      envGet("GOOGLE_CALENDAR_ORDER_BY", "").(string),
	Q:            envGet("GOOGLE_CALENDAR_Q", "").(string),
	SingleEvents: envGet("GOOGLE_CALENDAR_SINGLE_EVENTS", false).(bool),
}

var googleCalendarInsertEventOptions = vendors.GoogleCalendarInsertEventOptions{
	Summary:             envGet("GOOGLE_CALENDAR_EVENT_SUMMARY", "").(string),
	Description:         envGet("GOOGLE_CALENDAR_EVENT_DESCRIPTION", "").(string),
	Start:               envGet("GOOGLE_CALENDAR_EVENT_START", "").(string),
	End:                 envGet("GOOGLE_CALENDAR_EVENT_END", "").(string),
	TimeZone:            envGet("GOOGLE_CALENDAR_EVENT_TIMEZONE", "").(string),
	Visibility:          envGet("GOOGLE_CALENDAR_EVENT_VISIBILITY", "public").(string),
	SendUpdates:         envGet("GOOGLE_CALENDAR_EVENT_SEND_UPDATES", "all").(string),
	SupportsAttachments: envGet("GOOGLE_CALENDAR_EVENT_SUPPORTS_ATTACHMENTS", false).(bool),
	SourceTitle:         envGet("GOOGLE_CALENDAR_EVENT_SOURCE_TITLE", "").(string),
	SourceURL:           envGet("GOOGLE_CALENDAR_EVENT_SOURCE_URL", "").(string),
	ConferenceID:        envGet("GOOGLE_CALENDAR_EVENT_CONFERENCE_ID", "").(string),
}

var googleCalendarDeleteEventOptions = vendors.GoogleCalendarDeleteEventOptions{
	ID: envGet("GOOGLE_CALENDAR_EVENT_ID", "").(string),
}

type GoogleCalendarInsertEventOptions struct {
	SupportsAttachments bool
	SourceTitle         string
	SourceURL           string
}

var googleOutput = common.OutputOptions{
	Output: envGet("GOOGLE_OUTPUT", "").(string),
	Query:  envGet("GOOGLE_OUTPUT_QUERY", "").(string),
}

func googleNew(stdout *common.Stdout) *vendors.Google {

	common.Debug("Google", googleOptions, stdout)
	common.Debug("Google", googleOutput, stdout)

	return vendors.NewGoogle(googleOptions, stdout)
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
	flags.StringVar(&googleOutput.Output, "google-output", googleOutput.Output, "Google output")
	flags.StringVar(&googleOutput.Query, "google-output-query", googleOutput.Query, "Google output query")

	calendarCmd := &cobra.Command{
		Use:   "calendar",
		Short: "Calendar methods",
	}
	flags = calendarCmd.PersistentFlags()
	flags.StringVar(&googleCalendarOptions.ID, "google-calendar-id", googleCalendarOptions.ID, "Google calendar id")
	googleCmd.AddCommand(calendarCmd)

	calendarGetEventsCmd := &cobra.Command{
		Use:   "get-events",
		Short: "Calendar get events",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Google callendar getting events...")
			common.Debug("Google", googleCalendarOptions, stdout)
			common.Debug("Google", googleCalendarGetEventsOptions, stdout)

			bytes, err := googleNew(stdout).CalendarGetEvents(googleCalendarOptions, googleCalendarGetEventsOptions)
			if err != nil {
				stdout.Error("CalendarGetEvents error: %s", err)
			}
			common.OutputJson(googleOutput, "Google", []interface{}{googleOptions, googleCalendarOptions, googleCalendarGetEventsOptions}, bytes, stdout)
		},
	}
	flags = calendarGetEventsCmd.PersistentFlags()
	flags.StringVar(&googleCalendarGetEventsOptions.TimeMin, "google-calendar-time-min", googleCalendarGetEventsOptions.TimeMin, "Google calendar time min")
	flags.StringVar(&googleCalendarGetEventsOptions.TimeMax, "google-calendar-time-max", googleCalendarGetEventsOptions.TimeMax, "Google calendar time max")
	flags.StringVar(&googleCalendarGetEventsOptions.TimeZone, "google-calendar-timezone", googleCalendarGetEventsOptions.TimeZone, "Google calendar timezone")
	flags.StringVar(&googleCalendarGetEventsOptions.OrderBy, "google-calendar-order-by", googleCalendarGetEventsOptions.OrderBy, "Google calendar order by")
	flags.StringVar(&googleCalendarGetEventsOptions.Q, "google-calendar-q", googleCalendarGetEventsOptions.Q, "Google calendar q")
	flags.BoolVar(&googleCalendarGetEventsOptions.SingleEvents, "google-calendar-single-events", googleCalendarGetEventsOptions.SingleEvents, "Google calendar single events")
	calendarCmd.AddCommand(calendarGetEventsCmd)

	calendarInsertEventCmd := &cobra.Command{
		Use:   "insert-event",
		Short: "Calendar insert event",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Google callendar inserting event...")
			common.Debug("Google", googleCalendarOptions, stdout)
			common.Debug("Google", googleCalendarInsertEventOptions, stdout)

			bytes, err := googleNew(stdout).CalendarInsertEvent(googleCalendarOptions, googleCalendarInsertEventOptions)
			if err != nil {
				stdout.Error("CalendarInsertEvent error: %s", err)
			}
			common.OutputJson(googleOutput, "Google", []interface{}{googleOptions, googleCalendarOptions, googleCalendarInsertEventOptions}, bytes, stdout)
		},
	}
	flags = calendarInsertEventCmd.PersistentFlags()
	flags.StringVar(&googleCalendarInsertEventOptions.Summary, "google-calendar-event-summary", googleCalendarInsertEventOptions.Summary, "Google calendar event summary")
	flags.StringVar(&googleCalendarInsertEventOptions.Description, "google-calendar-event-description", googleCalendarInsertEventOptions.Description, "Google calendar event description")
	flags.StringVar(&googleCalendarInsertEventOptions.Start, "google-calendar-event-start", googleCalendarInsertEventOptions.Start, "Google calendar event start")
	flags.StringVar(&googleCalendarInsertEventOptions.End, "google-calendar-event-end", googleCalendarInsertEventOptions.End, "Google calendar event end")
	flags.StringVar(&googleCalendarInsertEventOptions.TimeZone, "google-calendar-event-timezone", googleCalendarInsertEventOptions.TimeZone, "Google calendar event timezone")
	flags.StringVar(&googleCalendarInsertEventOptions.Visibility, "google-calendar-event-visibility", googleCalendarInsertEventOptions.Visibility, "Google calendar event visibility")
	flags.StringVar(&googleCalendarInsertEventOptions.SendUpdates, "google-calendar-event-send-updates", googleCalendarInsertEventOptions.SendUpdates, "Google calendar event send updates")
	flags.BoolVar(&googleCalendarInsertEventOptions.SupportsAttachments, "google-calendar-event-supports-attachments", googleCalendarInsertEventOptions.SupportsAttachments, "Google calendar event support attachments")
	flags.StringVar(&googleCalendarInsertEventOptions.SourceTitle, "google-calendar-event-source-title", googleCalendarInsertEventOptions.SourceTitle, "Google calendar event source title")
	flags.StringVar(&googleCalendarInsertEventOptions.SourceURL, "google-calendar-event-source-url", googleCalendarInsertEventOptions.SourceURL, "Google calendar event source URL")
	flags.StringVar(&googleCalendarInsertEventOptions.ConferenceID, "google-calendar-event-conference-id", googleCalendarInsertEventOptions.ConferenceID, "Google calendar conference ID")
	calendarCmd.AddCommand(calendarInsertEventCmd)

	calendarDeleteEventCmd := &cobra.Command{
		Use:   "delete-event",
		Short: "Calendar delete event",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Google callendar inserting event...")
			common.Debug("Google", googleCalendarOptions, stdout)
			common.Debug("Google", googleCalendarDeleteEventOptions, stdout)

			bytes, err := googleNew(stdout).CalendarDeleteEvent(googleCalendarOptions, googleCalendarDeleteEventOptions)
			if err != nil {
				stdout.Error("CalendarDeleteEvent error: %s", err)
			}
			common.OutputJson(googleOutput, "Google", []interface{}{googleOptions, googleCalendarOptions, googleCalendarDeleteEventOptions}, bytes, stdout)
		},
	}
	flags = calendarDeleteEventCmd.PersistentFlags()
	flags.StringVar(&googleCalendarDeleteEventOptions.ID, "google-calendar-event-id", googleCalendarDeleteEventOptions.ID, "Google calendar event ID")
	calendarCmd.AddCommand(calendarDeleteEventCmd)

	calendarDeleteEventsCmd := &cobra.Command{
		Use:   "delete-events",
		Short: "Calendar delete evenst",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Google callendar deleting events...")
			common.Debug("Google", googleCalendarOptions, stdout)
			common.Debug("Google", googleCalendarGetEventsOptions, stdout)

			bytes, err := googleNew(stdout).CalendarDeleteEvents(googleCalendarOptions, googleCalendarGetEventsOptions)
			if err != nil {
				stdout.Error("CalendarDeleteEvent error: %s", err)
			}
			common.OutputJson(googleOutput, "Google", []interface{}{googleOptions, googleCalendarOptions, googleCalendarGetEventsOptions}, bytes, stdout)
		},
	}
	flags = calendarDeleteEventsCmd.PersistentFlags()
	flags.StringVar(&googleCalendarGetEventsOptions.TimeMin, "google-calendar-time-min", googleCalendarGetEventsOptions.TimeMin, "Google calendar time min")
	flags.StringVar(&googleCalendarGetEventsOptions.TimeMax, "google-calendar-time-max", googleCalendarGetEventsOptions.TimeMax, "Google calendar time max")
	flags.StringVar(&googleCalendarGetEventsOptions.TimeZone, "google-calendar-timezone", googleCalendarGetEventsOptions.TimeZone, "Google calendar timezone")
	flags.StringVar(&googleCalendarGetEventsOptions.OrderBy, "google-calendar-order-by", googleCalendarGetEventsOptions.OrderBy, "Google calendar order by")
	flags.StringVar(&googleCalendarGetEventsOptions.Q, "google-calendar-q", googleCalendarGetEventsOptions.Q, "Google calendar q")
	flags.BoolVar(&googleCalendarGetEventsOptions.SingleEvents, "google-calendar-single-events", googleCalendarGetEventsOptions.SingleEvents, "Google calendar single events")
	calendarCmd.AddCommand(calendarDeleteEventsCmd)

	return &googleCmd
}
