package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var graylogOptions = vendors.GraylogOptions{
	URL:       envGet("GRAYLOG_URL", "").(string),
	Timeout:   envGet("GRAYLOG_TIMEOUT", 30).(int),
	Insecure:  envGet("GRAYLOG_INSECURE", false).(bool),
	User:      envGet("GRAYLOG_USER", "").(string),
	Password:  envGet("GRAYLOG_PASSWORD", "").(string),
	Streams:   envGet("GRAYLOG_STREAMS", "").(string),
	Query:     envGet("GRAYLOG_QUERY", "").(string),
	RangeType: envGet("GRAYLOG_RANGE_TYPE", "absolute").(string),
	From:      envGet("GRAYLOG_FROM", "").(string),
	To:        envGet("GRAYLOG_TO", "").(string),
	Sort:      envGet("GRAYLOG_SORT", "").(string),
	Limit:     envGet("GRAYLOG_LIMIT", 100).(int),
	Range:     envGet("GRAYLOG_RANGE", "").(string),
}

var graylogOutput = common.OutputOptions{
	Output: envGet("GRAYLOG_OUTPUT", "").(string),
	Query:  envGet("GRAYLOG_OUTPUT_QUERY", "").(string),
}

func graylogNew(stdout *common.Stdout) *vendors.Graylog {

	common.Debug("Graylog", graylogOptions, stdout)
	common.Debug("Graylog", graylogOutput, stdout)

	queryBytes, err := utils.Content(graylogOptions.Query)
	if err != nil {
		stdout.Panic(err)
	}
	graylogOptions.Query = string(queryBytes)

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
	flags.StringVar(&graylogOptions.User, "graylog-user", graylogOptions.User, "Graylog user")
	flags.StringVar(&graylogOptions.Password, "graylog-password", graylogOptions.Password, "Graylog password")
	flags.StringVar(&graylogOptions.Streams, "graylog-streams", graylogOptions.Streams, "Graylog streams")
	flags.StringVar(&graylogOptions.Query, "graylog-query", graylogOptions.Query, "Graylog query")
	flags.StringVar(&graylogOptions.RangeType, "graylog-range-type", graylogOptions.RangeType, "Graylog range type")
	flags.StringVar(&graylogOptions.Sort, "graylog-sort", graylogOptions.Sort, "Graylog sort")
	flags.IntVar(&graylogOptions.Limit, "graylog-limit", graylogOptions.Limit, "Graylog limit")
	flags.StringVar(&graylogOptions.From, "graylog-from", graylogOptions.From, "Graylog from time")
	flags.StringVar(&graylogOptions.To, "graylog-to", graylogOptions.To, "Graylog to time")
	flags.StringVar(&graylogOutput.Output, "graylog-output", graylogOutput.Output, "Graylog output")
	flags.StringVar(&graylogOutput.Query, "graylog-output-query", graylogOutput.Query, "Graylog output query")

	graylogCmd.AddCommand(&cobra.Command{
		Use:   "get-logs",
		Short: "Get logs",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Graylog getting logs...")
			bytes, err := graylogNew(stdout).GetLogs()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(graylogOutput, "Graylog", []interface{}{graylogOptions}, bytes, stdout)
		},
	})

	return &graylogCmd
}
