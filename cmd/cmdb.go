package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var cmdbOptions = vendors.CmdbOptions{
	Timeout:  envGet("CMDB_TIMEOUT", 30).(int),
	Insecure: envGet("CMDB_INSECURE", false).(bool),
	APIURL:   envGet("CMDB_API_URL", "").(string), // TODO: remove default
}

var cmdbOutput = common.OutputOptions{
	Output: envGet("CMDB_OUTPUT", "").(string),
	Query:  envGet("CMDB_OUTPUT_QUERY", "").(string),
}

func cmdbNew(stdout *common.Stdout) *vendors.Cmdb {
	common.Debug("Cmdb", cmdbOptions, stdout)
	common.Debug("Cmdb", cmdbOutput, stdout)

	if utils.IsEmpty(cmdbOptions.APIURL) {
		stdout.Panic("No CMDB API URL")
	}

	cmdb := vendors.NewCmdb(cmdbOptions)
	if cmdb == nil {
		stdout.Panic("No cmdb")
	}
	return cmdb
}

func NewCmdbCommand() *cobra.Command {
	cmdbCmd := &cobra.Command{
		Use:   "cmdb",
		Short: "CMDB tools",
	}

	flags := cmdbCmd.PersistentFlags()
	flags.IntVar(&cmdbOptions.Timeout, "cmdb-timeout", cmdbOptions.Timeout, "CMDB Timeout in seconds")
	flags.BoolVar(&cmdbOptions.Insecure, "cmdb-insecure", cmdbOptions.Insecure, "CMDB Insecure")
	flags.StringVar(&cmdbOptions.APIURL, "cmdb-api-url", cmdbOptions.APIURL, "CMDB API URL")
	flags.StringVar(&cmdbOutput.Output, "cmdb-output", cmdbOutput.Output, "CMDB Output")
	flags.StringVar(&cmdbOutput.Query, "cmdb-output-query", cmdbOutput.Query, "CMDB Output Query")

	cmdbCmd.AddCommand(&cobra.Command{
		Use:   "get-component",
		Short: "Get component manifest",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Getting component manifest from cmdb…")
			if len(args) == 0 {
				stdout.Error("No component name")
				return
			}
			bytes, err := cmdbNew(stdout).GetComponentManifest(args[0])
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(cmdbOutput, "CMDB", []interface{}{cmdbOptions}, bytes, stdout)
		},
	})

	return cmdbCmd
}
