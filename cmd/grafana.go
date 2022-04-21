package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/spf13/cobra"
)

var grafanaOptions = vendors.GrafanaOptions{
	URL:      envGet("GRAFANA_URL", "").(string),
	Timeout:  envGet("GRAFANA_TIMEOUT", 30).(int),
	Insecure: envGet("GRAFANA_INSECURE", false).(bool),
	ApiKey:   envGet("GRAFANA_API_KEY", "").(string),
	OrgID:    envGet("GRAFANA_ORG_ID", "1").(string),
	UID:      envGet("GRAFANA_UID", "").(string),
	Slug:     envGet("GRAFANA_SLUG", "").(string),
}

var grafanaRenderImageOptions = vendors.GrafanaRenderImageOptions{
	PanelID: envGet("GRAFANA_IMAGE_PANEL_ID", "").(string),
	From:    envGet("GRAFANA_IMAGE_FROM", "").(string),
	To:      envGet("GRAFANA_IMAGE_TO", "").(string),
	Width:   envGet("GRAFANA_IMAGE_WIDTH", 1280).(int),
	Height:  envGet("GRAFANA_IMAGE_HEIGHT", 640).(int),
}

var grafanaGetDashboardsOptions = vendors.GrafanaGetDashboardsOptions{
	PanelID: envGet("GRAFANA_DASHBOARD_PANEL_ID", "").(string),
	From:    envGet("GRAFANA_DASHBOARD_FROM", "").(string),
	To:      envGet("GRAFANA_DASHBOARD_TO", "").(string),
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
	flags.StringVar(&grafanaOptions.Slug, "grafana-slug", grafanaOptions.Slug, "Grafana dashboard slug")

	flags.StringVar(&grafanaRenderImageOptions.PanelID, "grafana-image-panel-id", grafanaRenderImageOptions.PanelID, "Grafana image panel id")
	flags.StringVar(&grafanaRenderImageOptions.From, "grafana-image-from", grafanaRenderImageOptions.From, "Grafana image from")
	flags.StringVar(&grafanaRenderImageOptions.To, "grafana-image-to", grafanaRenderImageOptions.To, "Grafana image to")
	flags.IntVar(&grafanaRenderImageOptions.Width, "grafana-image-width", grafanaRenderImageOptions.Width, "Grafana image width")
	flags.IntVar(&grafanaRenderImageOptions.Height, "grafana-image-height", grafanaRenderImageOptions.Height, "Grafana image height")

	flags.StringVar(&grafanaGetDashboardsOptions.PanelID, "grafana-dashboard-panel-id", grafanaGetDashboardsOptions.PanelID, "Grafana dashboard panel id")
	flags.StringVar(&grafanaGetDashboardsOptions.From, "grafana-dashboard-from", grafanaGetDashboardsOptions.From, "Grafana dashboard from")
	flags.StringVar(&grafanaGetDashboardsOptions.To, "grafana-dashboard-to", grafanaGetDashboardsOptions.To, "Grafana dashboard to")

	flags.StringVar(&grafanaOutput.Output, "grafana-output", grafanaOutput.Output, "Grafana output")
	flags.StringVar(&grafanaOutput.Query, "grafana-output-query", grafanaOutput.Query, "Grafana output query")

	grafanaCmd.AddCommand(&cobra.Command{
		Use:   "render-image",
		Short: "Render image",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Grafana rendering image...")
			common.Debug("Grafana", grafanaRenderImageOptions, stdout)

			grafanaOptions.RenderImageOptions = &grafanaRenderImageOptions
			bytes, err := grafanaNew(stdout).RenderImage()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputRaw(grafanaOutput.Output, bytes, stdout)
		},
	})

	grafanaCmd.AddCommand(&cobra.Command{
		Use:   "get-dashboards",
		Short: "Get dashboards",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Grafana getting dashboards...")
			common.Debug("Grafana", grafanaGetDashboardsOptions, stdout)

			grafanaOptions.GetDashboardsOptions = &grafanaGetDashboardsOptions
			bytes, err := grafanaNew(stdout).GetDashboards()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(grafanaOutput, "Grafana", []interface{}{grafanaOptions, grafanaGetDashboardsOptions}, bytes, stdout)
		},
	})

	return &grafanaCmd
}
