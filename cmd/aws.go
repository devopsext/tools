package cmd

import (
	"encoding/json"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/spf13/cobra"
)

var awsOptions = vendors.AWSOptions{
	AWSKeys: vendors.AWSKeys{
		AccessKey: envGet("AWS_ACCESS_KEY", "").(string),
		SecretKey: envGet("AWS_SECRET_KEY", "").(string),
	},
	Accounts:        envGet("AWS_ACCOUNTS", "").(string),
	Role:            envGet("AWS_ROLE", "").(string),
	RoleTimeout:     envGet("AWS_ROLE_TIMEOUT", 3600).(int),
	RoleSessionName: envGet("AWS_ROLE_SESSION_NAME", "tools_session").(string),
	Timeout:         envGet("AWS_TIMEOUT", 30).(int),
	Insecure:        envGet("AWS_INSECURE", false).(bool),
}

var EC2Output = common.OutputOptions{
	Output: envGet("AWS_EC2_OUTPUT", "").(string),
	Query:  envGet("AWS_EC2_OUTPUT_QUERY", "").(string),
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
	ec2Cmd := &cobra.Command{
		Use:   "ec2",
		Short: "EC2 tools",
	}
	flags := ec2Cmd.PersistentFlags()
	flags.StringVar(&awsOptions.AccessKey, "aws-accesskey", awsOptions.AccessKey, "AWS access key")
	flags.StringVar(&awsOptions.SecretKey, "aws-secretkey", awsOptions.SecretKey, "AWS secret key")
	flags.StringVar(&awsOptions.Accounts, "aws-accounts", awsOptions.Accounts, "AWS account numbers, comma-separated")
	flags.StringVar(&awsOptions.Role, "aws-role", awsOptions.Role, "IAM role to assume")
	flags.IntVar(&awsOptions.RoleTimeout, "aws-role-timeout", awsOptions.RoleTimeout, "Assumed role duration in seconds")
	flags.StringVar(&awsOptions.RoleSessionName, "aws-role-session-name", awsOptions.RoleSessionName, "STS session name")
	flags.IntVar(&awsOptions.Timeout, "aws-timeout", awsOptions.Timeout, "HTTP timeout in seconds")
	flags.BoolVar(&awsOptions.Insecure, "aws-insecure", awsOptions.Insecure, "Skip TLS verification")
	flags.StringVar(&EC2Output.Output, "ec2-output", EC2Output.Output, "EC2 output file")
	flags.StringVar(&EC2Output.Query, "ec2-output-query", EC2Output.Query, "EC2 output JSONata query")

	getInstancesCmd := &cobra.Command{
		Use:   "get-instances",
		Short: "Get all EC2 instances across accounts and regions",
		Run: func(cmd *cobra.Command, args []string) {
			stdout.Debug("Getting EC2 instances...")
			common.Debug("EC2", awsOptions, stdout)

			ec2, err := vendors.NewAWSEC2(awsOptions)
			if err != nil {
				stdout.Panic("unable to create EC2 client", err)
			}
			instances, err := ec2.GetAllAWSEC2Instances()
			if err != nil {
				stdout.Error(err)
				return
			}
			b, err := json.Marshal(instances)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(EC2Output, "EC2", []interface{}{awsOptions}, b, stdout)
		},
	}
	ec2Cmd.AddCommand(getInstancesCmd)
	return ec2Cmd
}
