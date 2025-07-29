package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/spf13/cobra"
)

var k8sPodOptions = vendors.K8sPodOptions{
	Namespace: envGet("K8S_POD_NAMESPACE", "").(string),
	Name:      envGet("K8S_POD_NAME", "").(string),
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
	flags.StringVar(&k8sOutput.Output, "k8s-output", k8sOutput.Output, "K8s output")
	flags.StringVar(&k8sOutput.Query, "k8s-output-query", k8sOutput.Query, "K8s output query")

	podCmd := &cobra.Command{
		Use:   "pod",
		Short: "K8s Pod tools",
	}
	flags = podCmd.PersistentFlags()
	flags.StringVar(&k8sPodOptions.Namespace, "k8s-namespace", k8sPodOptions.Namespace, "K8s Pod namespace")
	flags.StringVar(&k8sPodOptions.Name, "k8s-name", k8sPodOptions.Name, "K8s Pod name")
	k8sCmd.AddCommand(podCmd)

	podDeleteCmd := &cobra.Command{
		Use:   "delete",
		Short: "K8s Pod delete",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("K8s pod deleting...")
			common.Debug("K8s", k8sPodOptions, stdout)

			bytes, err := k8sNew(stdout).PodDelete(k8sPodOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(k8sOutput, "K8s", []interface{}{k8sOptions, k8sPodOptions}, bytes, stdout)
		},
	}
	podCmd.AddCommand(podDeleteCmd)

	return k8sCmd
}
