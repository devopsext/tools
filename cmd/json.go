package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var jsonOptions = vendors.JSONOptions{
	Timeout:  envGet("JSON_TIMEOUT", 30).(int),
	Insecure: envGet("JSON_INSECURE", false).(bool),
	URL:      envGet("JSON_URL", "").(string),
}

var jsonOutput = common.OutputOptions{
	Output: envGet("JSON_OUTPUT", "").(string),
	Query:  envGet("JSON_OUTPUT_QUERY", "").(string),
}

func jsonNew(stdout *common.Stdout) *vendors.JSON {
	common.Debug("JSON", jsonOptions, stdout)
	common.Debug("JSON", jsonOutput, stdout)

	if utils.IsEmpty(jsonOptions.URL) {
		stdout.Panic("No JSON URL")
	}

	json := vendors.NewJSON(jsonOptions)
	if json == nil {
		stdout.Panic("No json")
	}
	return json
}

func NewJSONCommand() *cobra.Command {
	jsonCmd := &cobra.Command{
		Use:   "json",
		Short: "JSON tools",
	}

	flags := jsonCmd.PersistentFlags()
	flags.IntVar(&jsonOptions.Timeout, "json-timeout", jsonOptions.Timeout, "JSON Timeout in seconds")
	flags.BoolVar(&jsonOptions.Insecure, "json-insecure", jsonOptions.Insecure, "JSON Insecure")
	flags.StringVar(&jsonOptions.URL, "json-url", jsonOptions.URL, "JSON URL")
	flags.StringVar(&jsonOutput.Output, "json-output", jsonOutput.Output, "JSON Output")
	flags.StringVar(&jsonOutput.Query, "json-output-query", jsonOutput.Query, "JSON Output Query")

	jsonCmd.AddCommand(&cobra.Command{
		Use:   "get",
		Short: "Get json from URL",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Getting JSON from URL…")
			bytes, err := jsonNew(stdout).Get()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(jsonOutput, "JSON", []interface{}{jsonOptions}, bytes, stdout)
		},
	})

	return jsonCmd
}
