package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/spf13/cobra"
)

var teleportResourceOptions = vendors.TeleportResourceOptions{
	Kind: envGet("TELEPORT_RESOURCE_KIND", "").(string),
}

var teleportResourceListOptions = vendors.TeleportResourceListOptions{}

var teleportOptions = vendors.TeleportOptions{
	Address:  envGet("TELEPORT_ADDRESS", "").(string),
	Identity: envGet("TELEPORT_IDENTITY", "").(string),
	Timeout:  envGet("TELEPORT_TIMEOUT", 30).(int),
	Insecure: envGet("TELEPORT_INSECURE", false).(bool),
}

var teleportOutput = common.OutputOptions{
	Output: envGet("TELEPORT_OUTPUT", "").(string),
	Query:  envGet("TELEPORT_OUTPUT_QUERY", "").(string),
}

func teleportNew(stdout *common.Stdout) *vendors.Teleport {

	common.Debug("Teleport", teleportOutput, stdout)
	return vendors.NewTeleport(teleportOptions, stdout)
}

func NewTeleportCommand() *cobra.Command {

	teleportCmd := &cobra.Command{
		Use:   "teleport",
		Short: "Teleport tools",
	}
	flags := teleportCmd.PersistentFlags()
	flags.StringVar(&teleportOptions.Address, "teleport-address", teleportOptions.Address, "Teleport address")
	flags.StringVar(&teleportOptions.Identity, "teleport-identity", teleportOptions.Identity, "Teleport identity")
	flags.IntVar(&teleportOptions.Timeout, "teleport-timeout", teleportOptions.Timeout, "Teleport timeout")
	flags.BoolVar(&teleportOptions.Insecure, "teleport-insecure", teleportOptions.Insecure, "Teleport insecure")
	flags.StringVar(&teleportOutput.Output, "teleport-output", teleportOutput.Output, "Teleport output")
	flags.StringVar(&teleportOutput.Query, "teleport-output-query", teleportOutput.Query, "Teleport output query")

	// ping
	pingCmd := &cobra.Command{
		Use:   "ping",
		Short: "Teleport ping",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Teleport pinging...")

			bytes, err := teleportNew(stdout).CustomPing(teleportOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(teleportOutput, "Teleport", []interface{}{teleportOptions}, bytes, stdout)
		},
	}
	teleportCmd.AddCommand(pingCmd)

	// resource
	resourceCmd := &cobra.Command{
		Use:   "resource",
		Short: "Teleport Resource tools",
	}
	flags = resourceCmd.PersistentFlags()
	flags.StringVar(&teleportResourceOptions.Kind, "teleport-resource-kind", teleportResourceOptions.Kind, "Teleport Resource kind")
	teleportCmd.AddCommand(resourceCmd)

	resourceListCmd := &cobra.Command{
		Use:   "list",
		Short: "Teleport Resource list",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Teleport resource listing...")

			teleportResourceListOptions.TeleportResourceOptions = teleportResourceOptions
			common.Debug("Teleport", teleportResourceListOptions, stdout)

			bytes, err := teleportNew(stdout).ResourceList(teleportResourceListOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(teleportOutput, "Teleport", []interface{}{teleportOptions, teleportResourceOptions}, bytes, stdout)
		},
	}
	resourceCmd.AddCommand(resourceListCmd)

	return teleportCmd
}
