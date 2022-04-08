package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/spf13/cobra"
)

var graylogOptions = vendors.GraylogOptions{
	URL:      envGet("GRAYLOG_URL", "").(string),
	Timeout:  envGet("GRAYLOG_TIMEOUT", 30).(int),
	Insecure: envGet("GRAYLOG_INSECURE", false).(bool),

	Output: envGet("GRAYLOG_OUTPUT", "").(string),
	Query:  envGet("GRAYLOG_QUERY", "").(string),
}

func graylogNew(stdout *common.Stdout) common.LogManagement {
	graylog := vendors.NewGraylog(graylogOptions)
	if graylog == nil {
		stdout.Panic("No graylog")
	}
	return graylog
}

func NewGraylogCommand() *cobra.Command {

	graylogCmd := cobra.Command{
		Use:   "graylog",
		Short: "Graylog tools",
	}

	flags := graylogCmd.PersistentFlags()
	flags.StringVar(&graylogOptions.URL, "graylog-url", graylogOptions.URL, "Graylog URL")
	flags.IntVar(&graylogOptions.Timeout, "graylog-timeout", graylogOptions.Timeout, "Graylog timeout")
	flags.BoolVar(&graylogOptions.Insecure, "graylog-insecure", graylogOptions.Insecure, "Graylog insecure")
	flags.StringVar(&graylogOptions.Output, "graylog-output", graylogOptions.Output, "Graylog output")
	flags.StringVar(&graylogOptions.Query, "graylog-query", graylogOptions.Query, "Graylog query")

	graylogCmd.AddCommand(&cobra.Command{
		Use:   "logs",
		Short: "Getting logs",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Graylog getting logs...")
			bytes, err := graylogNew(stdout).Logs()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.Output(slackOptions.Query, slackOptions.Output, bytes, stdout)
		},
	})

	return &graylogCmd
}
