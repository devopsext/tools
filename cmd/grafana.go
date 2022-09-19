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
}

var grafanaRenderImageOptions = vendors.GrafanaRenderImageOptions{
	PanelID: envGet("GRAFANA_IMAGE_PANEL_ID", "").(string),
	From:    envGet("GRAFANA_IMAGE_FROM", "").(string),
	To:      envGet("GRAFANA_IMAGE_TO", "").(string),
	Width:   envGet("GRAFANA_IMAGE_WIDTH", 1280).(int),
	Height:  envGet("GRAFANA_IMAGE_HEIGHT", 640).(int),
}

var grafanaGetAnnotationsOptions = vendors.GrafanaGetAnnotationsOptions{
	Tags:        envGet("GRAFANA_ANNOTATION_TAGS", "").(string),
	From:        envGet("GRAFANA_ANNOTATION_FROM", "").(string),
	To:          envGet("GRAFANA_ANNOTATION_TO", "").(string),
	Type:        envGet("GRAFANA_ANNOTATION_TYPE", "").(string),
	Limit:       envGet("GRAFANA_ANNOTATION_LIMIT", 10).(int),
	AlertID:     envGet("GRAFANA_ANNOTATION_ALERT_ID", 0).(int),
	DashboardID: envGet("GRAFANA_ANNOTATION_DASHBOARD_ID", 0).(int),
	PanelID:     envGet("GRAFANA_ANNOTATION_PANEL_ID", 0).(int),
	MatchAny:    envGet("GRAFANA_ANNOTATION_MATCH_ANY", false).(bool),
}

var grafanaCreateAnnotationOptions = vendors.GrafanaCreateAnnotationOptions{
	Time:    envGet("GRAFANA_ANNOTATION_TIME", "").(string),
	TimeEnd: envGet("GRAFANA_ANNOTATION_TIME_END", "").(string),
	Tags:    envGet("GRAFANA_ANNOTATION_TAGS", "").(string),
	Text:    envGet("GRAFANA_ANNOTATION_TEXT", "").(string),
}

var grafanaOutput = common.OutputOptions{
	Output: envGet("GRAFANA_OUTPUT", "").(string),
	Query:  envGet("GRAFANA_OUTPUT_QUERY", "").(string),
}

func grafanaNew(stdout *common.Stdout) *vendors.Grafana {
	common.Debug("Grafana", grafanaOptions, stdout)
	common.Debug("Grafana", grafanaOutput, stdout)

	grafana, err := vendors.NewGrafana(grafanaOptions)
	if err != nil {
		stdout.Panic(err)
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

	flags.StringVar(&grafanaOutput.Output, "grafana-output", grafanaOutput.Output, "Grafana output")
	flags.StringVar(&grafanaOutput.Query, "grafana-output-query", grafanaOutput.Query, "Grafana output query")

	renderImageCmd := cobra.Command{
		Use:   "render-image",
		Short: "Render image",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Grafana rendering image...")
			common.Debug("Grafana", grafanaRenderImageOptions, stdout)

			bytes, err := grafanaNew(stdout).RenderImage(grafanaRenderImageOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputRaw(grafanaOutput.Output, bytes, stdout)
		},
	}

	flags = renderImageCmd.PersistentFlags()
	flags.StringVar(&grafanaRenderImageOptions.PanelID, "grafana-image-panel-id", grafanaRenderImageOptions.PanelID, "Grafana image panel id")
	flags.StringVar(&grafanaRenderImageOptions.From, "grafana-image-from", grafanaRenderImageOptions.From, "Grafana image from")
	flags.StringVar(&grafanaRenderImageOptions.To, "grafana-image-to", grafanaRenderImageOptions.To, "Grafana image to")
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

			bytes, err := grafanaNew(stdout).GetAnnotations(grafanaGetAnnotationsOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(grafanaOutput, "Grafana", []interface{}{grafanaOptions, grafanaGetAnnotationsOptions}, bytes, stdout)
		},
	}

	flags = getAnnotationsCmd.PersistentFlags()

	flags.StringVar(&grafanaGetAnnotationsOptions.From, "grafana-annotation-from", grafanaGetAnnotationsOptions.From, "Grafana annotation date from")
	flags.StringVar(&grafanaGetAnnotationsOptions.To, "grafana-annotation-to", grafanaGetAnnotationsOptions.To, "Grafana annotation date to")
	flags.StringVar(&grafanaGetAnnotationsOptions.Tags, "grafana-annotation-tags", grafanaGetAnnotationsOptions.Tags, "Grafana annotations tags (comma separated, optional)")
	flags.StringVar(&grafanaGetAnnotationsOptions.Type, "grafana-annotation-type", grafanaGetAnnotationsOptions.Type, "Grafana annotations type (alert|annotation, default both)")
	flags.IntVar(&grafanaGetAnnotationsOptions.Limit, "grafana-annotation-limit", grafanaGetAnnotationsOptions.Limit, "Grafana annotations limit (default: 10)")
	flags.IntVar(&grafanaGetAnnotationsOptions.AlertID, "grafana-annotation-alert", grafanaGetAnnotationsOptions.AlertID, "Grafana annotations alert")
	flags.IntVar(&grafanaGetAnnotationsOptions.DashboardID, "grafana-annotation-dashboard", grafanaGetAnnotationsOptions.DashboardID, "Grafana annotations dashboard")
	flags.IntVar(&grafanaGetAnnotationsOptions.PanelID, "grafana-annotation-panel", grafanaGetAnnotationsOptions.PanelID, "Grafana annotations panel")
	flags.BoolVar(&grafanaGetAnnotationsOptions.MatchAny, "grafana-annotation-match-any", grafanaGetAnnotationsOptions.MatchAny, "Grafana annotations match any tag")

	grafanaCmd.AddCommand(&getAnnotationsCmd)

	createAnnotationCmd := cobra.Command{
		Use:   "create-annotation",
		Short: "Create grafana annotation",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Grafana creating annotation...")
			common.Debug("Grafana", grafanaCreateAnnotationOptions, stdout)

			if grafanaCreateAnnotationOptions.Text == "" {
				stdout.Error("Grafana annotation text is required")
				return
			}

			bytes, err := grafanaNew(stdout).CreateAnnotation(grafanaCreateAnnotationOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(grafanaOutput, "Grafana", []interface{}{grafanaOptions, grafanaCreateAnnotationOptions}, bytes, stdout)
		},
	}

	flags = createAnnotationCmd.PersistentFlags()
	flags.StringVar(&grafanaCreateAnnotationOptions.Text, "grafana-annotation-text", grafanaCreateAnnotationOptions.Text, "Grafana annotation text")
	flags.StringVar(&grafanaCreateAnnotationOptions.Time, "grafana-annotation-time", grafanaCreateAnnotationOptions.Time, "Grafana annotation time")
	flags.StringVar(&grafanaCreateAnnotationOptions.TimeEnd, "grafana-annotation-time-end", grafanaCreateAnnotationOptions.Tags, "Grafana annotation end time")
	flags.StringVar(&grafanaCreateAnnotationOptions.Tags, "grafana-annotation-tags", grafanaCreateAnnotationOptions.Tags, "Grafana annotation tags (comma separated)")

	grafanaCmd.AddCommand(&createAnnotationCmd)

	return &grafanaCmd
}
