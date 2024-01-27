package cmd

import (
	"encoding/json"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/spf13/cobra"
)

var EC2Options = vendors.EC2Options{
	Region:    envGet("AWS_EC2_REGION", "").(string),
	AccessKey: envGet("AWS_EC2_ACCESSKEY", "").(string),
	SecretKey: envGet("AWS_EC2_SECRETKEY", "").(string),
}

var EC2Output = common.OutputOptions{
	Output: envGet("AWS_EC2_OUTPUT", "").(string),
	Query:  envGet("AWS_EC2_OUTPUT_QUERY", "").(string),
}

func EC2New(stdout *common.Stdout) *vendors.EC2 {
	common.Debug("EC2", EC2Options, stdout)
	common.Debug("EC2", EC2Output, stdout)

	ec2, err := vendors.NewEC2(EC2Options)
	if ec2 == nil || err != nil {
		stdout.Panic("unable to generate EC2 object", err)
	}
	return ec2
}

func NewAWSCommand() *cobra.Command {
	awsCmd := &cobra.Command{
		Use:   "aws",
		Short: "AWS tools",
	}

	awsCmd.AddCommand(NewEC2Subcommand())

	return awsCmd
}

func NewEC2Subcommand() *cobra.Command {
	EC2Cmd := &cobra.Command{
		Use:   "ec2",
		Short: "EC2 tools",
	}
	flags := EC2Cmd.PersistentFlags()
	flags.StringVar(&EC2Options.Region, "aws-region", EC2Options.Region, "The AWS region to perform the initial query")
	flags.StringVar(&EC2Options.AccessKey, "aws-accesskey", EC2Options.AccessKey, "Access key for AWS")
	flags.StringVar(&EC2Options.SecretKey, "aws-secretkey", EC2Options.SecretKey, "Secret key for AWS")
	flags.StringVar(&EC2Output.Output, "ec2-output", EC2Output.Output, "EC2 output")
	flags.StringVar(&EC2Output.Query, "ec2-output-query", EC2Output.Query, "EC2 output query")

	ec2GetInstancesCmd := &cobra.Command{
		Use:   "get-instances",
		Short: "Get EC2 instance by querying AWS",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("Getting EC2 instances...")

			instances, err := EC2New(stdout).GetAllEC2Instances()
			if err != nil {
				stdout.Error(err)
				return
			}
			bytes, err := json.Marshal(instances)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(EC2Output, "EC2", []interface{}{EC2Options}, bytes, stdout)
		},
	}
	EC2Cmd.AddCommand(ec2GetInstancesCmd)

	return EC2Cmd
}
