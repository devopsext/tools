package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var netboxOptions = vendors.NetboxOptions{
	Timeout:  envGet("NETBOX_TIMEOUT", 30).(int),
	Insecure: envGet("NETBOX_INSECURE", false).(bool),
	URL:      envGet("NETBOX_URL", "").(string),
	Token:    envGet("NETBOX_TOKEN", "").(string),
	Limit:    envGet("NETBOX_LIMIT", "50").(string),
	Brief:    envGet("NETBOX_BRIEF", false).(bool),
	Filter:   utils.MapGetKeyValues(envGet("NETBOX_FILTER", "").(string)),
}

var netboxDeviceOptions = vendors.NetboxDeviceOptions{
	DeviceID: envGet("NETBOX_DEVICE_ID", "").(string),
}

var netboxOutput = common.OutputOptions{
	Output: envGet("NETBOX_OUTPUT", "").(string),
	Query:  envGet("NETBOX_OUTPUT_QUERY", "").(string),
}

func netboxNew(stdout *common.Stdout) *vendors.Netbox {

	common.Debug("Netbox", netboxOptions, stdout)
	common.Debug("Netbox", netboxOutput, stdout)

	netbox := vendors.NewNetbox(netboxOptions)
	if netbox == nil {
		stdout.Panic("No netbox")
	}
	return netbox
}

func NewNetboxCommand() *cobra.Command {

	netboxCmd := &cobra.Command{
		Use:   "netbox",
		Short: "Netbox tools",
	}
	flags := netboxCmd.PersistentFlags()
	flags.IntVar(&netboxOptions.Timeout, "netbox-timeout", netboxOptions.Timeout, "Netbox timeout in seconds")
	flags.BoolVar(&netboxOptions.Insecure, "netbox-insecure", netboxOptions.Insecure, "Netbox insecure")
	flags.StringVar(&netboxOptions.URL, "netbox-url", netboxOptions.URL, "Netbox URL")
	flags.StringVar(&netboxOptions.Token, "netbox-token", netboxOptions.Token, "Netbox token")
	flags.StringVar(&netboxOptions.Limit, "netbox-limit", netboxOptions.Limit, "Netbox API limit")
	flags.BoolVar(&netboxOptions.Brief, "netbox-brief", netboxOptions.Brief, "Netbox API brief param")
	flags.StringVar(&netboxOutput.Output, "netbox-output", netboxOutput.Output, "Netbox output")
	flags.StringVar(&netboxOutput.Query, "netbox-output-query", netboxOutput.Query, "Netbox output query")

	getDeviceCmd := cobra.Command{
		Use:   "get-devices",
		Short: "Get devices from URL",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Getting devices from URL...")

			bytes, err := netboxNew(stdout).GetDevices(netboxDeviceOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(netboxOutput, "Netbox", []interface{}{netboxOptions}, bytes, stdout)
		},
	}
	flags = getDeviceCmd.PersistentFlags()
	flags.StringVar(&netboxDeviceOptions.DeviceID, "netbox-device-id", netboxDeviceOptions.DeviceID, "Netbox device id")
	netboxCmd.AddCommand(&getDeviceCmd)

	return netboxCmd
}
