package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/render"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var templateOptions = render.TemplateOptions{
	Name:       envGet("TEMPLATE_NAME", "").(string),
	Content:    envGet("TEMPLATE_CONTENT", "").(string),
	Object:     envGet("TEMPLATE_OBJECT", "").(string),
	TimeFormat: envGet("TEMPLATE_TIME_FORMAT", "").(string),
}

var templateOutput = common.OutputOptions{
	Output: envGet("TEMPLATE_OUTPUT", "").(string),
	Query:  envGet("TEMPLATE_OUTPUT_QUERY", "").(string),
}

func textTemplateNew(stdout *common.Stdout) *render.TextTemplate {

	common.Debug("Template", templateOutput, stdout)

	contentBytes, err := utils.Content(templateOptions.Content)
	if err != nil {
		stdout.Panic(err)
	}
	templateOptions.Content = string(contentBytes)

	objectBytes, err := utils.Content(templateOptions.Object)
	if err != nil {
		stdout.Panic(err)
	}
	templateOptions.Object = string(objectBytes)

	template := render.NewTextTemplate(templateOptions)
	if template == nil {
		stdout.Panic("No text template")
	}
	return template
}

func htmlTemplateNew(stdout *common.Stdout) *render.HtmlTemplate {

	common.Debug("Template", templateOutput, stdout)

	contentBytes, err := utils.Content(templateOptions.Content)
	if err != nil {
		stdout.Panic(err)
	}
	templateOptions.Content = string(contentBytes)

	objectBytes, err := utils.Content(templateOptions.Object)
	if err != nil {
		stdout.Panic(err)
	}
	templateOptions.Object = string(objectBytes)

	template := render.NewHtmlTemplate(templateOptions)
	if template == nil {
		stdout.Panic("No html template")
	}
	return template
}

func NewTemplateCommand() *cobra.Command {

	templateCmd := &cobra.Command{
		Use:   "template",
		Short: "Template tools",
	}
	flags := templateCmd.PersistentFlags()
	flags.StringVar(&templateOptions.Name, "template-name", templateOptions.Name, "Template name")
	flags.StringVar(&templateOptions.Content, "template-content", templateOptions.Content, "Template content")
	flags.StringVar(&templateOptions.Object, "template-object", templateOptions.Object, "Template object: json")
	flags.StringVar(&templateOptions.TimeFormat, "template-time-format", templateOptions.TimeFormat, "Template time format")
	flags.StringVar(&templateOutput.Output, "template-output", templateOutput.Output, "Template output")
	flags.StringVar(&templateOutput.Query, "template-output-query", templateOutput.Query, "Template output query")

	templateCmd.AddCommand(&cobra.Command{
		Use:   "render-text",
		Short: "Render text",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Template text rendering...")

			bytes, err := textTemplateNew(stdout).Render()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(templateOutput, "template", []interface{}{templateOptions}, bytes, stdout)
		},
	})

	templateCmd.AddCommand(&cobra.Command{
		Use:   "render-html",
		Short: "Render html",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Template html rendering...")

			bytes, err := htmlTemplateNew(stdout).Render()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(templateOutput, "template", []interface{}{templateOptions}, bytes, stdout)
		},
	})

	return templateCmd
}
