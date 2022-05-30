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
	APIKey:   envGet("GRAFANA_API_KEY", "").(string),
	OrgID:    envGet("GRAFANA_ORG_ID", "1").(string),
	UID:      envGet("GRAFANA_UID", "").(string),
	Slug:     envGet("GRAFANA_SLUG", "").(string),
	From:     envGet("GRAFANA_FROM", "").(string),
	To:       envGet("GRAFANA_TO", "").(string),
	PanelID:  envGet("GRAFANA_PANEL_ID", "").(string),
}

var grafanaRenderImageOptions = vendors.GrafanaRenderImageOptions{
	Width:  envGet("GRAFANA_IMAGE_WIDTH", 1280).(int),
	Height: envGet("GRAFANA_IMAGE_HEIGHT", 640).(int),
}

var grafanaGetAnnotationsOptions = vendors.GrafanaGetAnnotationsOptions{
	Tags: envGet("GRAFANA_ANNOTATION_TAGS", "").(string),
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
	flags.StringVar(&grafanaOptions.APIKey, "grafana-api-key", grafanaOptions.APIKey, "Grafana api key")
	flags.StringVar(&grafanaOptions.OrgID, "grafana-org-id", grafanaOptions.OrgID, "Grafana org id")
	flags.StringVar(&grafanaOptions.UID, "grafana-uid", grafanaOptions.UID, "Grafana dashboard uid")
	flags.StringVar(&grafanaOptions.Slug, "grafana-slug", grafanaOptions.Slug, "Grafana dashboard slug")
	flags.StringVar(&grafanaOptions.From, "grafana-from", grafanaOptions.From, "Grafana from")
	flags.StringVar(&grafanaOptions.To, "grafana-to", grafanaOptions.To, "Grafana to")
	flags.StringVar(&grafanaOptions.PanelID, "grafana-panel-id", grafanaOptions.PanelID, "Grafana panel id")

	flags.StringVar(&grafanaOutput.Output, "grafana-output", grafanaOutput.Output, "Grafana output")
	flags.StringVar(&grafanaOutput.Query, "grafana-output-query", grafanaOutput.Query, "Grafana output query")

	renderImageCmd := cobra.Command{
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
	}

	flags = renderImageCmd.PersistentFlags()
	flags.IntVar(&grafanaRenderImageOptions.Width, "grafana-image-width", grafanaRenderImageOptions.Width, "Grafana image width")
	flags.IntVar(&grafanaRenderImageOptions.Height, "grafana-image-height", grafanaRenderImageOptions.Height, "Grafana image height")

	grafanaCmd.AddCommand(&renderImageCmd)

	getDashboardCmd := cobra.Command{
		Use:   "get-dashboards",
		Short: "Get dashboards",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Grafana getting dashboards...")

			bytes, err := grafanaNew(stdout).GetDashboards()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(grafanaOutput, "Grafana", []interface{}{grafanaOptions}, bytes, stdout)
		},
	}

	grafanaCmd.AddCommand(&getDashboardCmd)

	getAnnotationsCmd := cobra.Command{
		Use:   "get-annotations",
		Short: "Get annotations",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Grafana getting annotations...")
			common.Debug("Grafana", grafanaGetAnnotationsOptions, stdout)

			grafanaOptions.GetAnnotationsOptions = &grafanaGetAnnotationsOptions
			bytes, err := grafanaNew(stdout).GetAnnotations()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(grafanaOutput, "Grafana", []interface{}{grafanaOptions}, bytes, stdout)
		},
	}

	flags = getAnnotationsCmd.PersistentFlags()
	flags.StringVar(&grafanaGetAnnotationsOptions.Tags, "grafana-annotations-tags", grafanaGetAnnotationsOptions.Tags, "Grafana annotations tags (comma separated)")

	grafanaCmd.AddCommand(&getAnnotationsCmd)

	return &grafanaCmd
}
