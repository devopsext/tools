package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/spf13/cobra"
)

var virusTotalOptions = vendors.VirusTotalOptions{
	Timeout:  envGet("VIRUSTOTAL_TIMEOUT", 30).(int),
	Insecure: envGet("VIRUSTOTAL_INSECURE", false).(bool),
	APIKey:   envGet("VIRUSTOTAL_API_KEY", "").(string),
}

var virusTotalDomainReportOptions = vendors.VirusTotalDomainReportOptions{
	Domain: envGet("VIRUSTOTAL_DOMAIN", "").(string),
}

var virusTotalOutput = common.OutputOptions{
	Output: envGet("VIRUSTOTAL_OUTPUT", "").(string),
	Query:  envGet("VIRUSTOTAL_OUTPUT_QUERY", "").(string),
}

func virusTotalNew(stdout *common.Stdout) *vendors.VirusTotal {

	common.Debug("VirusTotal", virusTotalOptions, stdout)
	common.Debug("VirusTotal", virusTotalOutput, stdout)

	virusTotal := vendors.NewVirusTotal(virusTotalOptions, stdout)
	if virusTotal == nil {
		stdout.Panic("No virustotal")
	}

	return virusTotal
}

func NewVirusTotalCommand() *cobra.Command {
	virusTotalCmd := &cobra.Command{
		Use:   "virustotal",
		Short: "VirusTotal tools",
	}

	flags := virusTotalCmd.PersistentFlags()
	flags.IntVar(&virusTotalOptions.Timeout, "timeout", virusTotalOptions.Timeout, "Timeout")
	flags.BoolVar(&virusTotalOptions.Insecure, "insecure", virusTotalOptions.Insecure, "Insecure")
	flags.StringVar(&virusTotalOptions.APIKey, "api-key", virusTotalOptions.APIKey, "API key")
	flags.StringVar(&virusTotalOutput.Output, "output", virusTotalOutput.Output, "Output")
	flags.StringVar(&virusTotalOutput.Query, "query", virusTotalOutput.Query, "Query")

	domainReport := &cobra.Command{
		Use:   "domain-report",
		Short: "Get domain report",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("VirusTotal get domain report...")
			common.Debug("VirusTotal", virusTotalOptions, stdout)

			bytes, err := virusTotalNew(stdout).DomainReport(virusTotalDomainReportOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(virusTotalOutput, "VirusTotal", []interface{}{virusTotalOptions, virusTotalDomainReportOptions}, bytes, stdout)
		},
	}
	flags = domainReport.PersistentFlags()

	flags.StringVar(&virusTotalDomainReportOptions.Domain, "domain", virusTotalDomainReportOptions.Domain, "Domains")

	virusTotalCmd.AddCommand(domainReport)

	return virusTotalCmd
}
