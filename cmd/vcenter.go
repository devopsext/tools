package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/spf13/cobra"
)

var vcenterHostOptions = vendors.VCenterHostOptions{
	Cluster: envGet("VCENTER_HOST_CLUSTER", "").(string),
}

var vcenterVMOptions = vendors.VCenterVMOptions{
	Cluster: envGet("VCENTER_VM_CLUSTER", "").(string),
	Host:    envGet("VCENTER_VM_HOST", "").(string),
}

var vcenterOptions = vendors.VCenterOptions{
	Timeout:  envGet("VCENTER_TIMEOUT", 30).(int),
	Insecure: envGet("VCENTER_INSECURE", false).(bool),
	URL:      envGet("VCENTER_URL", "").(string),
	User:     envGet("VCENTER_USER", "").(string),
	Password: envGet("VCENTER_PASSWORD", "").(string),
	Session:  envGet("VCENTER_SESSION", "").(string),
}

var vcenterOutput = common.OutputOptions{
	Output: envGet("VCENTER_OUTPUT", "").(string),
	Query:  envGet("VCENTER_OUTPUT_QUERY", "").(string),
}

func vcenterNew(stdout *common.Stdout) *vendors.VCenter {

	common.Debug("VCenter", vcenterOptions, stdout)
	common.Debug("VCenter", vcenterOutput, stdout)

	vcenter := vendors.NewVCenter(vcenterOptions)
	if vcenter == nil {
		stdout.Panic("No VCenter")
	}
	return vcenter
}

func NewVCenterCommand() *cobra.Command {

	vcenterCmd := &cobra.Command{
		Use:   "vcenter",
		Short: "VCenter tools",
	}
	flags := vcenterCmd.PersistentFlags()
	flags.IntVar(&vcenterOptions.Timeout, "vcenter-timeout", vcenterOptions.Timeout, "VCenter timeout in seconds")
	flags.BoolVar(&vcenterOptions.Insecure, "vcenter-insecure", vcenterOptions.Insecure, "VCenter insecure")
	flags.StringVar(&vcenterOptions.URL, "vcenter-url", vcenterOptions.URL, "VCenter URL")
	flags.StringVar(&vcenterOptions.User, "vcenter-user", vcenterOptions.User, "VCenter user")
	flags.StringVar(&vcenterOptions.Password, "vcenter-password", vcenterOptions.Password, "VCenter password")
	flags.StringVar(&vcenterOptions.Session, "vcenter-session", vcenterOptions.Session, "VCenter session")
	flags.StringVar(&vcenterOutput.Output, "vcenter-output", vcenterOutput.Output, "VCenter output")
	flags.StringVar(&vcenterOutput.Query, "vcenter-output-query", vcenterOutput.Query, "VCenter output query")

	vcenterGetClustersCmd := &cobra.Command{
		Use:   "get-clusters",
		Short: "Get clusters from URL",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Getting clusters from URL...")

			bytes, err := vcenterNew(stdout).GetClusters()
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(vcenterOutput, "VCenter", []interface{}{vcenterOptions}, bytes, stdout)
		},
	}
	vcenterCmd.AddCommand(vcenterGetClustersCmd)

	vcenterGetHostsCmd := &cobra.Command{
		Use:   "get-hosts",
		Short: "Get hosts from URL",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Getting hosts from URL...")
			common.Debug("vcenter", vcenterHostOptions, stdout)

			bytes, err := vcenterNew(stdout).GetHosts(vcenterHostOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(vcenterOutput, "VCenter", []interface{}{vcenterOptions, vcenterHostOptions}, bytes, stdout)
		},
	}
	flags = vcenterGetHostsCmd.PersistentFlags()
	flags.StringVar(&vcenterHostOptions.Cluster, "vcenter-host-cluster", vcenterHostOptions.Cluster, "VCenter get host cluster")
	vcenterCmd.AddCommand(vcenterGetHostsCmd)

	vcenterGetVMsCmd := &cobra.Command{
		Use:   "get-vms",
		Short: "Get vms from URL",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Getting vms from URL...")
			common.Debug("vcenter", vcenterVMOptions, stdout)

			bytes, err := vcenterNew(stdout).GetVMs(vcenterVMOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(vcenterOutput, "VCenter", []interface{}{vcenterOptions, vcenterVMOptions}, bytes, stdout)
		},
	}
	flags = vcenterGetVMsCmd.PersistentFlags()
	flags.StringVar(&vcenterVMOptions.Cluster, "vcenter-vm-cluster", vcenterVMOptions.Cluster, "VCenter get vm cluster")
	flags.StringVar(&vcenterVMOptions.Host, "vcenter-vm-host", vcenterVMOptions.Host, "VCenter get vm host")
	vcenterCmd.AddCommand(vcenterGetVMsCmd)

	return vcenterCmd
}
