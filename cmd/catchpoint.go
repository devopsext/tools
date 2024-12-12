package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/spf13/cobra"
)

var catchpointOptions = vendors.CatchpointOptions{
	Timeout:  envGet("CATCHPOINT_TIMEOUT", 30).(int),
	Insecure: envGet("CATCHPOINT_INSECURE", false).(bool),
	APIToken: envGet("CATCHPOINT_API_TOKEN", "").(string),
}

var catchpointInstantTestOptions = vendors.CatchpointInstantTestOptions{
	URL:             envGet("CATCHPOINT_URL", "").(string),
	NodesIds:        envGet("CATCHPOINT_NODE_IDS", "").(string),
	InstantTestType: envGet("CATCHPOINT_TEST_TYPE_ID", 0).(int),
	HTTPMethodType:  envGet("CATCHPOINT_HTTP_METHOD_TYPE_ID", 0).(int),
	MonitorType:     envGet("CATCHPOINT_MONITOR_TYPE_ID", 2).(int),
	OnDemand:        envGet("CATCHPOINT_ON_DEMAND", false).(bool),
}

var catchpointInstantTestWithNodeGroupOptions = vendors.CatchpointInstantTestWithNodeGroupOptions{
	URL:             envGet("CATCHPOINT_URL", "").(string),
	NodeGroupID:     envGet("CATCHPOINT_NODE_GROUP_ID", 0).(int),
	InstantTestType: envGet("CATCHPOINT_TEST_TYPE_ID", 0).(int),
	HTTPMethodType:  envGet("CATCHPOINT_HTTP_METHOD_TYPE_ID", 0).(int),
	MonitorType:     envGet("CATCHPOINT_MONITOR_TYPE_ID", 2).(int),
	OnDemand:        envGet("CATCHPOINT_ON_DEMAND", false).(bool),
}

var catchpointOutput = common.OutputOptions{
	Output: envGet("CATCHPOINT_OUTPUT", "").(string),
	Query:  envGet("CATCHPOINT_OUTPUT_QUERY", "").(string),
}

func catchpointNew(stdout *common.Stdout) *vendors.Catchpoint {

	common.Debug("Catchpoint", catchpointOptions, stdout)
	common.Debug("Catchpoint", catchpointOutput, stdout)

	catchpoint := vendors.NewCatchpoint(catchpointOptions, stdout)
	if catchpoint == nil {
		stdout.Panic("No site24x7")
	}

	return catchpoint
}

func NewCatchpointCommand() *cobra.Command {
	catchpointCmd := &cobra.Command{
		Use:   "catchpoint",
		Short: "Catchpoint tools",
	}

	flags := catchpointCmd.PersistentFlags()
	flags.IntVar(&catchpointOptions.Timeout, "timeout", catchpointOptions.Timeout, "Timeout")
	flags.BoolVar(&catchpointOptions.Insecure, "insecure", catchpointOptions.Insecure, "Insecure")
	flags.StringVar(&catchpointOptions.APIToken, "api-token", catchpointOptions.APIToken, "API token")
	flags.StringVar(&catchpointOutput.Output, "output", catchpointOutput.Output, "Output")
	flags.StringVar(&catchpointOutput.Query, "query", catchpointOutput.Query, "Query")

	instantTest := &cobra.Command{
		Use:   "instant-test",
		Short: "Run Instant test",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Catchpoint instant test...")
			common.Debug("Catchpoint", catchpointInstantTestOptions, stdout)

			bytes, err := catchpointNew(stdout).InstantTest(catchpointInstantTestOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(catchpointOutput, "Catchpoint", []interface{}{catchpointOptions, catchpointInstantTestOptions}, bytes, stdout)
		},
	}
	flags = instantTest.PersistentFlags()
	flags.BoolVar(&catchpointInstantTestOptions.OnDemand, "on-demand", catchpointInstantTestOptions.OnDemand, "On demand")
	flags.StringVar(&catchpointInstantTestOptions.URL, "url", catchpointInstantTestOptions.URL, "URL")
	flags.StringVar(&catchpointInstantTestOptions.NodesIds, "node-ids", catchpointInstantTestOptions.NodesIds, "Node IDs")
	flags.IntVar(&catchpointInstantTestOptions.InstantTestType, "test-type-id", catchpointInstantTestOptions.InstantTestType, "Test type ID")
	flags.IntVar(&catchpointInstantTestOptions.HTTPMethodType, "http-method-type-id", catchpointInstantTestOptions.HTTPMethodType, "HTTP method type ID")
	flags.IntVar(&catchpointInstantTestOptions.MonitorType, "monitor-type-id", catchpointInstantTestOptions.MonitorType, "Monitor type ID")
	catchpointCmd.AddCommand(instantTest)

	instantTestWithNodeGroup := &cobra.Command{
		Use:   "instant-test-with-node-group",
		Short: "Run Instant test with node group",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Catchpoint instant test with node group...")
			common.Debug("Catchpoint", catchpointInstantTestWithNodeGroupOptions, stdout)

			bytes, err := catchpointNew(stdout).InstantTestWithNodeGroup(catchpointInstantTestWithNodeGroupOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(catchpointOutput, "Catchpoint", []interface{}{catchpointOptions, catchpointInstantTestWithNodeGroupOptions}, bytes, stdout)
		},
	}
	flags = instantTestWithNodeGroup.PersistentFlags()
	flags.BoolVar(&catchpointInstantTestWithNodeGroupOptions.OnDemand, "on-demand", catchpointInstantTestWithNodeGroupOptions.OnDemand, "On demand")
	flags.StringVar(&catchpointInstantTestWithNodeGroupOptions.URL, "url", catchpointInstantTestWithNodeGroupOptions.URL, "URL")
	flags.IntVar(&catchpointInstantTestWithNodeGroupOptions.NodeGroupID, "node-group-id", catchpointInstantTestWithNodeGroupOptions.NodeGroupID, "Node group ID")
	flags.IntVar(&catchpointInstantTestWithNodeGroupOptions.InstantTestType, "test-type-id", catchpointInstantTestWithNodeGroupOptions.InstantTestType, "Test type ID")
	flags.IntVar(&catchpointInstantTestWithNodeGroupOptions.HTTPMethodType, "http-method-type-id", catchpointInstantTestWithNodeGroupOptions.HTTPMethodType, "HTTP method type ID")
	flags.IntVar(&catchpointInstantTestWithNodeGroupOptions.MonitorType, "monitor-type-id", catchpointInstantTestWithNodeGroupOptions.MonitorType, "Monitor type ID")
	catchpointCmd.AddCommand(instantTestWithNodeGroup)

	return catchpointCmd
}
