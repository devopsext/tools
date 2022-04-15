package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/spf13/cobra"
)

var grafanaOptions = vendors.GrafanaOptions{
	URL:         envGet("GRAFANA_URL", "").(string),
	Timeout:     envGet("GRAFANA_TIMEOUT", 30).(int),
	Insecure:    envGet("GRAFANA_INSECURE", false).(bool),
	User:        envGet("GRAFANA_USER", "").(string),
	Password:    envGet("GRAFANA_PASSWORD", "").(string),
	Output:      envGet("GRAFANA_OUTPUT", "").(string),
	OutputQuery: envGet("GRAFANA_OUTPUT_QUERY", "").(string),
}

func grafanaNew(stdout *common.Stdout) common.Dashboard {

	grafana := vendors.NewGrafana(grafanaOptions)
	if grafana == nil {
		stdout.Panic("No grafana")
	}
	return grafana
}

func NewGrafanaCommand() *cobra.Command {

	grafanaCmd := cobra.Command{
		Use:   "grafana",
		Short: "Grafana tools",
	}

	flags := grafanaCmd.PersistentFlags()
	flags.StringVar(&grafanaOptions.URL, "grafana-url", grafanaOptions.URL, "Grafana URL")
	flags.IntVar(&grafanaOptions.Timeout, "grafana-timeout", grafanaOptions.Timeout, "Grafana timeout")
	flags.BoolVar(&grafanaOptions.Insecure, "grafana-insecure", grafanaOptions.Insecure, "Grafana insecure")
	flags.StringVar(&grafanaOptions.User, "grafana-user", grafanaOptions.User, "Grafana user")
	flags.StringVar(&grafanaOptions.Password, "grafana-password", grafanaOptions.Password, "Grafana password")
	flags.StringVar(&grafanaOptions.Output, "grafana-output", grafanaOptions.Output, "Grafana output")
	flags.StringVar(&grafanaOptions.OutputQuery, "grafana-output-query", grafanaOptions.OutputQuery, "Grafana output query")

	grafanaCmd.AddCommand(&cobra.Command{
		Use:   "get-image",
		Short: "Get image",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Grafana image...")
			bytes, err := grafanaNew(stdout).GetImage()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.Output(grafanaOptions.OutputQuery, grafanaOptions.Output, "Grafana", grafanaOptions, bytes, stdout)
		},
	})

	return &grafanaCmd
}
