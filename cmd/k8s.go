package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/spf13/cobra"
)

var k8sResourceOptions = vendors.K8sResourceOptions{
	Kind:      envGet("K8S_RESOURCE_KIND", "").(string),
	Namespace: envGet("K8S_RESOURCE_NAMESPACE", "").(string),
	Name:      envGet("K8S_RESOURCE_NAME", "").(string),
}

var k8sResourceDescribeOptions = vendors.K8sResourceDescribeOptions{}

var k8sResourceDeleteOptions = vendors.K8sResourceDeleteOptions{}

var k8sResourceScaleOptions = vendors.K8sResourceScaleOptions{
	Replicas:    envGet("K8S_RESOURCE_SCALE_REPLICAS", -1).(int),
	WaitTimeout: envGet("K8S_RESOURCE_SCALE_WAIT_TIMEOUT", 30).(int),
	PollTimeout: envGet("K8S_RESOURCE_SCALE_POLL_TIMEOUT", 1).(int),
}

var k8sResourceRestartOptions = vendors.K8sResourceRestartOptions{
	WaitTimeout: envGet("K8S_RESOURCE_RESTART_WAIT_TIMEOUT", 30).(int),
	PollTimeout: envGet("K8S_RESOURCE_RESTART_POLL_TIMEOUT", 1).(int),
}

var k8sOptions = vendors.K8sOptions{
	Config:  envGet("K8S_CONFIG", "").(string),
	Timeout: envGet("K8S_TIMEOUT", 30).(int),
}

var k8sOutput = common.OutputOptions{
	Output: envGet("K8S_OUTPUT", "").(string),
	Query:  envGet("K8S_OUTPUT_QUERY", "").(string),
}

func k8sNew(stdout *common.Stdout) *vendors.K8s {

	common.Debug("K8s", k8sOutput, stdout)
	return vendors.NewK8s(k8sOptions, stdout)
}

func NewK8sCommand() *cobra.Command {

	k8sCmd := &cobra.Command{
		Use:   "k8s",
		Short: "K8s tools",
	}
	flags := k8sCmd.PersistentFlags()
	flags.StringVar(&k8sOptions.Config, "k8s-config", k8sOptions.Config, "K8s config")
	flags.IntVar(&k8sOptions.Timeout, "k8s-timeout", k8sOptions.Timeout, "K8s timeout")
	flags.StringVar(&k8sOutput.Output, "k8s-output", k8sOutput.Output, "K8s output")
	flags.StringVar(&k8sOutput.Query, "k8s-output-query", k8sOutput.Query, "K8s output query")

	resourceCmd := &cobra.Command{
		Use:   "resource",
		Short: "K8s Resource tools",
	}
	flags = resourceCmd.PersistentFlags()
	flags.StringVar(&k8sResourceOptions.Kind, "k8s-resource-kind", k8sResourceOptions.Kind, "K8s Resource kind")
	flags.StringVar(&k8sResourceOptions.Namespace, "k8s-resource-namespace", k8sResourceOptions.Namespace, "K8s Resource namespace")
	flags.StringVar(&k8sResourceOptions.Name, "k8s-resource-name", k8sResourceOptions.Name, "K8s Resource name")
	k8sCmd.AddCommand(resourceCmd)

	resourceDescribeCmd := &cobra.Command{
		Use:   "describe",
		Short: "K8s Resource describe",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("K8s resource describing...")

			k8sResourceDescribeOptions.K8sResourceOptions = k8sResourceOptions
			common.Debug("K8s", k8sResourceDescribeOptions, stdout)

			bytes, err := k8sNew(stdout).ResourceDescribe(k8sResourceDescribeOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(k8sOutput, "K8s", []interface{}{k8sOptions, k8sResourceDescribeOptions}, bytes, stdout)
		},
	}
	resourceCmd.AddCommand(resourceDescribeCmd)

	resourceDeleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "K8s Resource delete",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("K8s resource deleting...")

			k8sResourceDeleteOptions.K8sResourceOptions = k8sResourceOptions
			common.Debug("K8s", k8sResourceDeleteOptions, stdout)

			bytes, err := k8sNew(stdout).ResourceDelete(k8sResourceDeleteOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(k8sOutput, "K8s", []interface{}{k8sOptions, k8sResourceDeleteOptions}, bytes, stdout)
		},
	}
	resourceCmd.AddCommand(resourceDeleteCmd)

	resourceScaleCmd := &cobra.Command{
		Use:   "scale",
		Short: "K8s Resource scale",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("K8s resource scaling...")

			k8sResourceScaleOptions.K8sResourceOptions = k8sResourceOptions
			common.Debug("K8s", k8sResourceScaleOptions, stdout)

			bytes, err := k8sNew(stdout).ResourceScale(k8sResourceScaleOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(k8sOutput, "K8s", []interface{}{k8sOptions, k8sResourceScaleOptions}, bytes, stdout)
		},
	}
	flags = resourceScaleCmd.PersistentFlags()
	flags.IntVar(&k8sResourceScaleOptions.Replicas, "k8s-resource-scale-replicas", k8sResourceScaleOptions.Replicas, "K8s Resource scale replicas")
	flags.IntVar(&k8sResourceScaleOptions.WaitTimeout, "k8s-resource-scale-wait-timeout", k8sResourceScaleOptions.WaitTimeout, "K8s Resource scale wait timeout")
	flags.IntVar(&k8sResourceScaleOptions.PollTimeout, "k8s-resource-scale-poll-timeout", k8sResourceScaleOptions.PollTimeout, "K8s Resource scale poll timeout")
	resourceCmd.AddCommand(resourceScaleCmd)

	resourceRestartCmd := &cobra.Command{
		Use:   "restart",
		Short: "K8s Resource restart",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("K8s resource restarting...")

			k8sResourceRestartOptions.K8sResourceOptions = k8sResourceOptions
			common.Debug("K8s", k8sResourceRestartOptions, stdout)

			bytes, err := k8sNew(stdout).ResourceRestart(k8sResourceRestartOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(k8sOutput, "K8s", []interface{}{k8sOptions, k8sResourceRestartOptions}, bytes, stdout)
		},
	}
	flags = resourceRestartCmd.PersistentFlags()
	flags.IntVar(&k8sResourceRestartOptions.WaitTimeout, "k8s-resource-restart-wait-timeout", k8sResourceRestartOptions.WaitTimeout, "K8s Resource restart wait timeout")
	flags.IntVar(&k8sResourceRestartOptions.PollTimeout, "k8s-resource-restart-poll-timeout", k8sResourceRestartOptions.PollTimeout, "K8s Resource restart poll timeout")
	resourceCmd.AddCommand(resourceRestartCmd)

	return k8sCmd
}
