package cmd

import (
	"github.com/devopsext/tools/logging"
	"github.com/spf13/cobra"
)

var graylogOptions = logging.GraylogOptions{
	URL:     envGet("GRAYLOG_URL", "").(string),
	Timeout: envGet("GRAYLOG_TIMEOUT", 30).(int),
}

func NewGraylogCommand() *cobra.Command {

	graylogCmd := cobra.Command{
		Use:   "graylog",
		Short: "Graylog tools",
		Run: func(cmd *cobra.Command, args []string) {
			//
		},
	}

	flags := graylogCmd.PersistentFlags()
	flags.StringVar(&graylogOptions.URL, "graylog-url", graylogOptions.URL, "Graylog URL")
	flags.IntVar(&graylogOptions.Timeout, "graylog-timeout", graylogOptions.Timeout, "Graylog timeout")

	return &graylogCmd
}
