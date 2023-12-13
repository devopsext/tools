package cmd

import (
	"strings"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/spf13/cobra"
)

var zabbixHostGetOptions = vendors.ZabbixHostGetOptions{
	Fields:     strings.Split(envGet("ZABBIX_HOST_GET_FIELDS", "").(string), ","),
	Inventory:  strings.Split(envGet("ZABBIX_HOST_GET_INVENTORY", "").(string), ","),
	Interfaces: strings.Split(envGet("ZABBIX_HOST_GET_INTERFACSES", "").(string), ","),
}

var zabbixOptions = vendors.ZabbixOptions{
	Timeout:  envGet("ZABBIX_TIMEOUT", 30).(int),
	Insecure: envGet("ZABBIX_INSECURE", false).(bool),
	URL:      envGet("ZABBIX_URL", "").(string),
	User:     envGet("ZABBIX_USER", "").(string),
	Password: envGet("ZABBIX_PASSWORD", "").(string),
	Auth:     envGet("ZABBIX_AUTH", "").(string),
}

var zabbixOutput = common.OutputOptions{
	Output: envGet("ZABBIX_OUTPUT", "").(string),
	Query:  envGet("ZABBIX_OUTPUT_QUERY", "").(string),
}

func zabbixNew(stdout *common.Stdout) *vendors.Zabbix {

	common.Debug("Zabbix", zabbixOptions, stdout)
	common.Debug("Zabbix", zabbixOutput, stdout)

	zabbix := vendors.NewZabbix(zabbixOptions)
	if zabbix == nil {
		stdout.Panic("No zabbix")
	}
	return zabbix
}

func NewZabbixCommand() *cobra.Command {

	zabbixCmd := &cobra.Command{
		Use:   "zabbix",
		Short: "zabbix tools",
	}
	flags := zabbixCmd.PersistentFlags()
	flags.IntVar(&zabbixOptions.Timeout, "zabbix-timeout", zabbixOptions.Timeout, "Zabbix timeout in seconds")
	flags.BoolVar(&zabbixOptions.Insecure, "zabbix-insecure", zabbixOptions.Insecure, "Zabbix insecure")
	flags.StringVar(&zabbixOptions.URL, "zabbix-url", zabbixOptions.URL, "Zabbix URL")
	flags.StringVar(&zabbixOptions.User, "zabbix-user", zabbixOptions.User, "Zabbix user")
	flags.StringVar(&zabbixOptions.Password, "zabbix-password", zabbixOptions.Password, "Zabbix password")
	flags.StringVar(&zabbixOptions.Auth, "zabbix-auth", zabbixOptions.Auth, "Zabbix auth")
	flags.StringVar(&zabbixOutput.Output, "zabbix-output", zabbixOutput.Output, "Zabbix output")
	flags.StringVar(&zabbixOutput.Query, "zabbix-output-query", zabbixOutput.Query, "Zabbix output query")

	zabbixHostGetCmd := &cobra.Command{
		Use:   "get-hosts",
		Short: "Get hosts from URL",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Getting hosts from URL...")
			common.Debug("Zabbix", zabbixHostGetOptions, stdout)

			bytes, err := zabbixNew(stdout).GetHosts(zabbixHostGetOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(zabbixOutput, "Zabbix", []interface{}{zabbixOptions, zabbixHostGetOptions}, bytes, stdout)
		},
	}
	flags = zabbixHostGetCmd.PersistentFlags()
	flags.StringSliceVar(&zabbixHostGetOptions.Fields, "zabbix-host-get-fields", zabbixHostGetOptions.Fields, "Zabbix host get fields")
	flags.StringSliceVar(&zabbixHostGetOptions.Inventory, "zabbix-host-get-inventory", zabbixHostGetOptions.Inventory, "Zabbix host get inventory")
	flags.StringSliceVar(&zabbixHostGetOptions.Interfaces, "zabbix-host-get-interfaces", zabbixHostGetOptions.Interfaces, "Zabbix host get interfaces")
	zabbixCmd.AddCommand(zabbixHostGetCmd)

	return zabbixCmd
}
