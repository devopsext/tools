package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/spf13/cobra"
)

var observiumOptions = vendors.ObserviumOptions{
	Timeout:  envGet("OBSERVIUM_TIMEOUT", 30).(int),
	Insecure: envGet("OBSERVIUM_INSECURE", false).(bool),
	URL:      envGet("OBSERVIUM_URL", "").(string),
	User:     envGet("OBSERVIUM_USER", "").(string),
	Password: envGet("OBSERVIUM_PASSWORD", "").(string),
	Token:    envGet("OBSERVIUM_TOKEN", "").(string),
}

var observiumOutput = common.OutputOptions{
	Output: envGet("OBSERVIUM_OUTPUT", "").(string),
	Query:  envGet("OBSERVIUM_OUTPUT_QUERY", "").(string),
}

func observiumNew(stdout *common.Stdout) *vendors.Observium {

	common.Debug("Observium", observiumOptions, stdout)
	common.Debug("Observium", observiumOutput, stdout)

	observium := vendors.NewObservium(observiumOptions)
	if observium == nil {
		stdout.Panic("No observium")
	}
	return observium
}

func NewObserviumCommand() *cobra.Command {

	observiumCmd := &cobra.Command{
		Use:   "observium",
		Short: "Observium tools",
	}
	flags := observiumCmd.PersistentFlags()
	flags.IntVar(&observiumOptions.Timeout, "observium-timeout", observiumOptions.Timeout, "Observium timeout in seconds")
	flags.BoolVar(&observiumOptions.Insecure, "observium-insecure", observiumOptions.Insecure, "Observium insecure")
	flags.StringVar(&observiumOptions.URL, "observium-url", observiumOptions.URL, "Observium URL")
	flags.StringVar(&observiumOptions.User, "observium-user", observiumOptions.User, "Observium user")
	flags.StringVar(&observiumOptions.Password, "observium-password", observiumOptions.Password, "Observium password")
	flags.StringVar(&observiumOptions.Token, "observium-token", observiumOptions.Token, "Observium token")
	flags.StringVar(&observiumOutput.Output, "observium-output", observiumOutput.Output, "Observium output")
	flags.StringVar(&observiumOutput.Query, "observium-output-query", observiumOutput.Query, "Observium output query")

	observiumCmd.AddCommand(&cobra.Command{
		Use:   "get-devices",
		Short: "Get devices from URL",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Getting devices from URL...")

			bytes, err := observiumNew(stdout).GetDevices()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(observiumOutput, "Observium", []interface{}{observiumOptions}, bytes, stdout)
		},
	})

	return observiumCmd
}
