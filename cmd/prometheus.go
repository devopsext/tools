package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var prometheusOptions = vendors.PrometheusOptions{
	URL:      envGet("PROMETHEUS_URL", "").(string),
	Timeout:  envGet("PROMETHEUS_TIMEOUT", 30).(int),
	Insecure: envGet("PROMETHEUS_INSECURE", false).(bool),
	Query:    envGet("PROMETHEUS_QUERY", "").(string),
	From:     envGet("PROMETHEUS_FROM", "").(string),
	To:       envGet("PROMETHEUS_TO", "").(string),
	Step:     envGet("PROMETHEUS_STEP", "60").(string),
	Params:   envGet("PROMETHEUS_PARAMS", "").(string),
	User:     envGet("PROMETHEUS_USER", "").(string),
	Password: envGet("PROMETHEUS_PASSWORD", "").(string),
}

var prometheusOutput = common.OutputOptions{
	Output: envGet("PROMETHEUS_OUTPUT", "").(string),
	Query:  envGet("PROMETHEUS_OUTPUT_QUERY", "").(string),
}

func prometheusNew(stdout *common.Stdout) *vendors.Prometheus {

	common.Debug("Prometheus", prometheusOptions, stdout)
	common.Debug("Prometheus", prometheusOutput, stdout)

	queryBytes, err := utils.Content(prometheusOptions.Query)
	if err != nil {
		stdout.Panic(err)
	}
	prometheusOptions.Query = string(queryBytes)

	prometheus := vendors.NewPrometheus(prometheusOptions)
	if prometheus == nil {
		stdout.Panic("No prometheus")
	}
	return prometheus
}

func NewPrometheusCommand() *cobra.Command {

	prometheusCmd := &cobra.Command{
		Use:   "prometheus",
		Short: "Prometheus tools",
	}

	flags := prometheusCmd.PersistentFlags()
	flags.StringVar(&prometheusOptions.URL, "prometheus-url", prometheusOptions.URL, "Prometheus URL")
	flags.IntVar(&prometheusOptions.Timeout, "prometheus-timeout", prometheusOptions.Timeout, "Prometheus timeout in seconds")
	flags.BoolVar(&prometheusOptions.Insecure, "prometheus-insecure", prometheusOptions.Insecure, "Prometheus insecure")
	flags.StringVar(&prometheusOptions.Query, "prometheus-query", prometheusOptions.Query, "Prometheus query")
	flags.StringVar(&prometheusOptions.From, "prometheus-from", prometheusOptions.From, "Prometheus from")
	flags.StringVar(&prometheusOptions.To, "prometheus-to", prometheusOptions.To, "Prometheus to")
	flags.StringVar(&prometheusOptions.Step, "prometheus-step", prometheusOptions.Step, "Prometheus step")
	flags.StringVar(&prometheusOptions.Params, "prometheus-params", prometheusOptions.Params, "Prometheus params")
	flags.StringVar(&prometheusOptions.User, "prometheus-user", prometheusOptions.User, "Prometheus user")
	flags.StringVar(&prometheusOptions.Password, "prometheus-password", prometheusOptions.Password, "Prometheus password")
	flags.StringVar(&prometheusOutput.Output, "prometheus-output", prometheusOutput.Output, "Prometheus output")
	flags.StringVar(&prometheusOutput.Query, "prometheus-output-query", prometheusOutput.Query, "Prometheus output query")

	prometheusCmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get data from URL",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Getting data from URL...")

			bytes, err := prometheusNew(stdout).Get()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(prometheusOutput, "Prometheus", []interface{}{prometheusOptions}, bytes, stdout)
		},
	})

	return prometheusCmd
}
