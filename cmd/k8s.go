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

var k8sResourceDeleteOptions = vendors.K8sResourceDeleteOptions{}

var k8sResourceScaleOptions = vendors.K8sResourceScaleOptions{
	Replicas: envGet("K8S_RESOURCE_REPLICAS", -1).(int),
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
	flags.IntVar(&k8sResourceScaleOptions.Replicas, "k8s-resource-replicas", k8sResourceScaleOptions.Replicas, "K8s Resource replicas")
	resourceCmd.AddCommand(resourceScaleCmd)

	return k8sCmd
}
