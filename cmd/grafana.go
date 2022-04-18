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
	ApiKey:      envGet("GRAFANA_API_KEY", "").(string),
	OrgID:       envGet("GRAFANA_ORG_ID", "1").(string),
	UID:         envGet("GRAFANA_UID", "").(string),
	Slug:        envGet("GRAFANA_SLUG", "").(string),
	PanelID:     envGet("GRAFANA_PANEL_ID", "").(string),
	From:        envGet("GRAFANA_FROM", "").(string),
	To:          envGet("GRAFANA_TO", "").(string),
	ImageWidth:  envGet("GRAFANA_IMAGE_WIDTH", 1280).(int),
	ImageHeight: envGet("GRAFANA_IMAGE_HEIGHT", 640).(int),
}

var grafanaOutput = common.OutputOptions{
	Output: envGet("GRAFANA_OUTPUT", "").(string),
	Query:  envGet("GRAFANA_OUTPUT_QUERY", "").(string),
}

func grafanaNew(stdout *common.Stdout) *vendors.Grafana {

	common.Debug("Grafana", grafanaOptions, stdout)
	common.Debug("Grafana", grafanaOutput, stdout)

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
	flags.StringVar(&grafanaOptions.ApiKey, "grafana-api-key", grafanaOptions.ApiKey, "Grafana api key")
	flags.StringVar(&grafanaOptions.OrgID, "grafana-org-id", grafanaOptions.OrgID, "Grafana org id")
	flags.StringVar(&grafanaOptions.UID, "grafana-uid", grafanaOptions.UID, "Grafana dashboard uid")
	flags.StringVar(&grafanaOptions.Slug, "grafana-slug", grafanaOptions.Slug, "Grafana slug")
	flags.StringVar(&grafanaOptions.PanelID, "grafana-panel-id", grafanaOptions.PanelID, "Grafana panel id")
	flags.StringVar(&grafanaOptions.From, "grafana-from", grafanaOptions.From, "Grafana from")
	flags.StringVar(&grafanaOptions.To, "grafana-to", grafanaOptions.To, "Grafana to")
	flags.IntVar(&grafanaOptions.ImageWidth, "grafana-image-width", grafanaOptions.ImageWidth, "Grafana image width")
	flags.IntVar(&grafanaOptions.ImageHeight, "grafana-image-height", grafanaOptions.ImageHeight, "Grafana image height")
	flags.StringVar(&grafanaOutput.Output, "grafana-output", grafanaOutput.Output, "Grafana output")
	flags.StringVar(&grafanaOutput.Query, "grafana-output-query", grafanaOutput.Query, "Grafana output query")

	grafanaCmd.AddCommand(&cobra.Command{
		Use:   "get-image",
		Short: "Get image",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Grafana getting image...")
			bytes, err := grafanaNew(stdout).GetImage()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputRaw(grafanaOutput, bytes, stdout)
		},
	})

	return &grafanaCmd
}
