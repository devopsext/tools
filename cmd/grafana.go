package cmd

import (
	"strings"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var grafanaOptions = vendors.GrafanaOptions{
	URL:      envGet("GRAFANA_URL", "").(string),
	Timeout:  envGet("GRAFANA_TIMEOUT", 30).(int),
	Insecure: envGet("GRAFANA_INSECURE", false).(bool),
	APIKey:   envGet("GRAFANA_API_KEY", "").(string),
	OrgID:    envGet("GRAFANA_ORG_ID", "1").(string),
}

var grafanaDashboardOptions = vendors.GrafanaDahboardOptions{
	Title:     envGet("GRAFANA_DASHBOARD_TITLE", "").(string),
	UID:       envGet("GRAFANA_DASHBOARD_UID", "").(string),
	Slug:      envGet("GRAFANA_DASHBOARD_SLUG", "").(string),
	Timezone:  envGet("GRAFANA_DASHBOARD_TIMEZONE", "UTC").(string),
	FolderUID: envGet("GRAFANA_DASHBOARD_FOLDER_UID", "").(string),
	Tags:      strings.Split(envGet("GRAFANA_DASHBOARD_TAGS", "").(string), ","),
	From:      envGet("GRAFANA_DASHBOARD_FROM", "now-1h").(string),
	To:        envGet("GRAFANA_DASHBOARD_TO", "now").(string),
	SaveUID:   envGet("GRAFANA_DASHBOARD_SAVE_UID", true).(bool),
	Cloned: vendors.GrafanaClonedDahboardOptions{
		URL:         envGet("GRAFANA_DASHBOARD_CLONED_URL", "").(string),
		Timeout:     envGet("GRAFANA_DASHBOARD_CLONED_TIMEOUT", 30).(int),
		Insecure:    envGet("GRAFANA_DASHBOARD_CLONED_INSECURE", false).(bool),
		APIKey:      envGet("GRAFANA_DASHBOARD_CLONED_API_KEY", "").(string),
		OrgID:       envGet("GRAFANA_DASHBOARD_CLONED_ORG_ID", "1").(string),
		UID:         envGet("GRAFANA_DASHBOARD_CLONED_UID", "").(string),
		Annotations: strings.Split(envGet("GRAFANA_DASHBOARD_CLONED_ANNOTATIONS", "").(string), ","),
		PanelIDs:    strings.Split(envGet("GRAFANA_DASHBOARD_CLONED_PANEL_IDS", "").(string), ","),
		PanelTitles: strings.Split(envGet("GRAFANA_DASHBOARD_CLONED_PANEL_TITLES", "").(string), ","),
		PanelSeries: strings.Split(envGet("GRAFANA_DASHBOARD_CLONED_PANEL_SERIES", "").(string), ","),
		LegendRight: envGet("GRAFANA_DASHBOARD_CLONED_LEGEND_RIGHT", false).(bool),
		Arrange:     envGet("GRAFANA_DASHBOARD_CLONED_ARRANGE", false).(bool),
		Count:       envGet("GRAFANA_DASHBOARD_CLONED_COUNT", 3).(int),
		Width:       envGet("GRAFANA_DASHBOARD_CLONED_WIDTH", 6).(int),
		Height:      envGet("GRAFANA_DASHBOARD_CLONED_HEIGHT", 7).(int),
	},
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

	return vendors.NewGrafana(grafanaOptions)
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

	flags.StringVar(&grafanaOutput.Output, "grafana-output", grafanaOutput.Output, "Grafana output")
	flags.StringVar(&grafanaOutput.Query, "grafana-output-query", grafanaOutput.Query, "Grafana output query")

	getDashboardCmd := cobra.Command{
		Use:   "get-dashboards",
		Short: "Get dashboards by uid",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Grafana getting dashboards...")

			bytes, err := grafanaNew(stdout).GetDashboards(grafanaDashboardOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(grafanaOutput, "Grafana", []interface{}{grafanaOptions}, bytes, stdout)
		},
	}
	grafanaCmd.AddCommand(&getDashboardCmd)

	searchDashboardCmd := cobra.Command{
		Use:   "search-dashboards",
		Short: "search dashboards by folder/dashboard UID",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Grafana searching dashboard...")
			common.Debug("Grafana", grafanaDashboardOptions, stdout)

			bytes, err := grafanaNew(stdout).SearchDashboards(grafanaDashboardOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(grafanaOutput, "Grafana", []interface{}{grafanaOptions, grafanaDashboardOptions}, bytes, stdout)
		},
	}
	flags = searchDashboardCmd.PersistentFlags()
	flags.StringVar(&grafanaDashboardOptions.UID, "grafana-dashboard-uid", grafanaDashboardOptions.UID, "Grafana dashboard uid")
	flags.StringVar(&grafanaDashboardOptions.FolderUID, "grafana-dashboard-folder-uid", grafanaDashboardOptions.FolderUID, "Grafana dashboard folder uid")
	grafanaCmd.AddCommand(&searchDashboardCmd)

	copyDashboardCmd := cobra.Command{
		Use:   "copy-dashboard",
		Short: "copy dashboard",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Grafana copiyng dashboard...")
			common.Debug("Grafana", grafanaDashboardOptions, stdout)

			bytes, err := grafanaNew(stdout).CopyDashboard(grafanaDashboardOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(grafanaOutput, "Grafana", []interface{}{grafanaOptions, grafanaDashboardOptions}, bytes, stdout)
		},
	}
	flags = copyDashboardCmd.PersistentFlags()
	flags.StringVar(&grafanaDashboardOptions.Title, "grafana-dashboard-title", grafanaDashboardOptions.Title, "Grafana dashboard title")
	flags.StringVar(&grafanaDashboardOptions.FolderUID, "grafana-dashboard-folder-uid", grafanaDashboardOptions.FolderUID, "Grafana dashboard folder uid")
	flags.StringSliceVar(&grafanaDashboardOptions.Tags, "grafana-dashboard-tags", grafanaDashboardOptions.Tags, "Grafana dashboard tags")
	flags.StringVar(&grafanaDashboardOptions.From, "grafana-dashboard-from", grafanaDashboardOptions.From, "Grafana dashboard time from")
	flags.StringVar(&grafanaDashboardOptions.To, "grafana-dashboard-to", grafanaDashboardOptions.To, "Grafana dashboard time to")
	flags.BoolVar(&grafanaDashboardOptions.SaveUID, "grafana-dashboard-save-uid", grafanaDashboardOptions.SaveUID, "Save UID for copied Grafana dashboard")
	flags.StringVar(&grafanaDashboardOptions.Cloned.URL, "grafana-dashboard-cloned-url", grafanaDashboardOptions.Cloned.URL, "Grafana Dashboard cloned URL exist")
	flags.IntVar(&grafanaDashboardOptions.Cloned.Timeout, "grafana-dashboard-cloned-timeout", grafanaDashboardOptions.Cloned.Timeout, "Grafana Dashboard cloned timeout")
	flags.BoolVar(&grafanaDashboardOptions.Cloned.Insecure, "grafana-dashboard-cloned-insecure", grafanaDashboardOptions.Cloned.Insecure, "Grafana Dashboard cloned insecure")
	flags.StringVar(&grafanaDashboardOptions.Cloned.APIKey, "grafana-dashboard-cloned-api-key", grafanaDashboardOptions.Cloned.APIKey, "Grafana Dashboard cloned api-key")
	flags.StringVar(&grafanaDashboardOptions.Cloned.UID, "grafana-dashboard-cloned-uid", grafanaDashboardOptions.Cloned.UID, "Grafana Dashboard cloned UID")
	grafanaCmd.AddCommand(&copyDashboardCmd)

	createDashboardCmd := cobra.Command{
		Use:   "create-dashboard",
		Short: "Create dashboard",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Grafana creating dashboard...")
			common.Debug("Grafana", grafanaDashboardOptions, stdout)

			if utils.IsEmpty(grafanaDashboardOptions.Title) {
				stdout.Error("Grafana create title is required")
				return
			}

			bytes, err := grafanaNew(stdout).CreateDashboard(grafanaDashboardOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(grafanaOutput, "Grafana", []interface{}{grafanaOptions, grafanaDashboardOptions}, bytes, stdout)
		},
	}
	flags = createDashboardCmd.PersistentFlags()
	flags.StringVar(&grafanaDashboardOptions.Title, "grafana-dashboard-title", grafanaDashboardOptions.Title, "Grafana dashboard title")
	flags.StringVar(&grafanaDashboardOptions.Timezone, "grafana-dashboard-timezone", grafanaDashboardOptions.Timezone, "Grafana dashboard timezone")
	flags.StringVar(&grafanaDashboardOptions.FolderUID, "grafana-dashboard-folder-uid", grafanaDashboardOptions.FolderUID, "Grafana dashboard folder uid")
	flags.StringSliceVar(&grafanaDashboardOptions.Tags, "grafana-dashboard-tags", grafanaDashboardOptions.Tags, "Grafana dashboard tags")
	flags.StringVar(&grafanaDashboardOptions.From, "grafana-dashboard-from", grafanaDashboardOptions.From, "Grafana dashboard time from")
	flags.StringVar(&grafanaDashboardOptions.To, "grafana-dashboard-to", grafanaDashboardOptions.To, "Grafana dashboard time to")
	flags.StringVar(&grafanaDashboardOptions.Cloned.URL, "grafana-dashboard-cloned-url", grafanaDashboardOptions.Cloned.URL, "Grafana Dashboard cloned URL exist")
	flags.IntVar(&grafanaDashboardOptions.Cloned.Timeout, "grafana-dashboard-cloned-timeout", grafanaDashboardOptions.Cloned.Timeout, "Grafana Dashboard cloned timeout")
	flags.BoolVar(&grafanaDashboardOptions.Cloned.Insecure, "grafana-dashboard-cloned-insecure", grafanaDashboardOptions.Cloned.Insecure, "Grafana Dashboard cloned insecure")
	flags.StringVar(&grafanaDashboardOptions.Cloned.APIKey, "grafana-dashboard-cloned-api-key", grafanaDashboardOptions.Cloned.APIKey, "Grafana Dashboard cloned api-key")
	flags.StringVar(&grafanaDashboardOptions.Cloned.UID, "grafana-dashboard-cloned-uid", grafanaDashboardOptions.Cloned.UID, "Grafana dashboard cloned uuid")
	flags.StringSliceVar(&grafanaDashboardOptions.Cloned.Annotations, "grafana-dashboard-cloned-annotations", grafanaDashboardOptions.Cloned.Annotations, "Grafana dashboard cloned annotations")
	flags.StringSliceVar(&grafanaDashboardOptions.Cloned.PanelIDs, "grafana-dashboard-cloned-panel-ids", grafanaDashboardOptions.Cloned.PanelIDs, "Grafana dashboard cloned panel ids")
	flags.StringSliceVar(&grafanaDashboardOptions.Cloned.PanelTitles, "grafana-dashboard-cloned-panel-titles", grafanaDashboardOptions.Cloned.PanelTitles, "Grafana dashboard cloned panel titles")
	flags.StringSliceVar(&grafanaDashboardOptions.Cloned.PanelSeries, "grafana-dashboard-cloned-panel-series", grafanaDashboardOptions.Cloned.PanelSeries, "Grafana dashboard cloned panel series")
	flags.BoolVar(&grafanaDashboardOptions.Cloned.LegendRight, "grafana-dashboard-cloned-legend-right", grafanaDashboardOptions.Cloned.LegendRight, "Grafana dashboard cloned legend right")
	flags.BoolVar(&grafanaDashboardOptions.Cloned.Arrange, "grafana-dashboard-cloned-arrange", grafanaDashboardOptions.Cloned.Arrange, "Grafana dashboard cloned arrange")
	flags.IntVar(&grafanaDashboardOptions.Cloned.Count, "grafana-dashboard-cloned-count", grafanaDashboardOptions.Cloned.Count, "Grafana dashboard cloned count per line")
	flags.IntVar(&grafanaDashboardOptions.Cloned.Width, "grafana-dashboard-cloned-width", grafanaDashboardOptions.Cloned.Width, "Grafana dashboard cloned width")
	flags.IntVar(&grafanaDashboardOptions.Cloned.Height, "grafana-dashboard-cloned-height", grafanaDashboardOptions.Cloned.Height, "Grafana dashboard cloned height")
	grafanaCmd.AddCommand(&createDashboardCmd)

	deleteDashboardCmd := cobra.Command{
		Use:   "delete-dashboard",
		Short: "delete dashboard by uid",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Grafana deleting dashboard...")
			common.Debug("Grafana", grafanaDashboardOptions, stdout)

			bytes, err := grafanaNew(stdout).DeleteDashboards(grafanaDashboardOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(grafanaOutput, "Grafana", []interface{}{grafanaOptions, grafanaDashboardOptions}, bytes, stdout)
		},
	}
	flags = deleteDashboardCmd.PersistentFlags()
	flags.StringVar(&grafanaDashboardOptions.UID, "grafana-dashboard-uid", grafanaDashboardOptions.UID, "Grafana dashboard uid")
	grafanaCmd.AddCommand(&deleteDashboardCmd)

	renderImageCmd := cobra.Command{
		Use:   "render-image",
		Short: "Render image",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Grafana rendering image...")
			common.Debug("Grafana", []interface{}{grafanaDashboardOptions, grafanaRenderImageOptions}, stdout)

			bytes, err := grafanaNew(stdout).RenderImage(grafanaDashboardOptions, grafanaRenderImageOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputRaw(grafanaOutput.Output, bytes, stdout)
		},
	}
	flags = renderImageCmd.PersistentFlags()
	flags.StringVar(&grafanaDashboardOptions.UID, "grafana-dashboard-uid", grafanaDashboardOptions.UID, "Grafana dashboard uid")
	flags.StringVar(&grafanaDashboardOptions.Slug, "grafana-dashboard-slug", grafanaDashboardOptions.Slug, "Grafana dashboard slug")
	flags.StringVar(&grafanaDashboardOptions.Timezone, "grafana-dashboard-timezone", grafanaDashboardOptions.Timezone, "Grafana dashboard timezone")
	flags.StringVar(&grafanaRenderImageOptions.PanelID, "grafana-image-panel-id", grafanaRenderImageOptions.PanelID, "Grafana image panel id")
	flags.StringVar(&grafanaRenderImageOptions.From, "grafana-image-from", grafanaRenderImageOptions.From, "Grafana image from")
	flags.StringVar(&grafanaRenderImageOptions.To, "grafana-image-to", grafanaRenderImageOptions.To, "Grafana image to")
	flags.IntVar(&grafanaRenderImageOptions.Width, "grafana-image-width", grafanaRenderImageOptions.Width, "Grafana image width")
	flags.IntVar(&grafanaRenderImageOptions.Height, "grafana-image-height", grafanaRenderImageOptions.Height, "Grafana image height")
	grafanaCmd.AddCommand(&renderImageCmd)

	getAnnotationsCmd := cobra.Command{
		Use:   "get-annotations",
		Short: "Get annotations",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Grafana getting annotations...")
			common.Debug("Grafana", grafanaGetAnnotationsOptions, stdout)

			bytes, err := grafanaNew(stdout).GetAnnotations(grafanaDashboardOptions, grafanaGetAnnotationsOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(grafanaOutput, "Grafana", []interface{}{grafanaOptions, grafanaDashboardOptions, grafanaGetAnnotationsOptions}, bytes, stdout)
		},
	}
	flags = getAnnotationsCmd.PersistentFlags()
	flags.StringVar(&grafanaDashboardOptions.Timezone, "grafana-dashboard-timezone", grafanaDashboardOptions.Timezone, "Grafana dashboard timezone")
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
