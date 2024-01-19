package cmd

import (
	"github.com/devopsext/tools/common"
	"github.com/devopsext/tools/vendors"
	"github.com/devopsext/utils"
	"github.com/spf13/cobra"
)

var pagerDutyOptions = vendors.PagerDutyOptions{
	Timeout:  envGet("PAGERDUTY_TIMEOUT", 30).(int),
	Insecure: envGet("PAGERDUTY_INSECURE", false).(bool),
	URL:      envGet("PAGERDUTY_URL", "").(string),
	Token:    envGet("PAGERDUTY_TOKEN", "").(string),
}

var pagerDutyGetIncidentsOptions = vendors.PagerDutyGetIncidentsOptions{
	Key:   envGet("PAGERDUTY_INCIDENT_KEY", "").(string),
	Limit: envGet("PAGERDUTY_INCIDENTS_LIMIT", 10).(int),
}

var pagerDutyCreateIncidentOptions = vendors.PagerDutyCreateIncidentOptions{
	From: envGet("PAGERDUTY_INCIDENT_FROM", "").(string),
}

var pagerDutyIncidentOptions = vendors.PagerDutyIncidentOptions{
	Title:      envGet("PAGERDUTY_INCIDENT_TITLE", "").(string),
	Body:       envGet("PAGERDUTY_INCIDENT_BODY", "").(string),
	Urgency:    envGet("PAGERDUTY_INCIDENT_URGENCY", "").(string),
	ServiceID:  envGet("PAGERDUTY_INCIDENT_SERVICE_ID", "").(string),
	PriorityID: envGet("PAGERDUTY_INCIDENT_PRIORITY_ID", "").(string),
}

var pagerDutyOutput = common.OutputOptions{
	Output: envGet("PAGERDUTY_OUTPUT", "").(string),
	Query:  envGet("PAGERDUTY_OUTPUT_QUERY", "").(string),
}

func pagerDutyNew(stdout *common.Stdout) *vendors.PagerDuty {

	common.Debug("PagerDuty", pagerDutyOptions, stdout)
	common.Debug("PagerDuty", pagerDutyOutput, stdout)

	pagerDuty := vendors.NewPagerDuty(pagerDutyOptions, stdout)
	if pagerDuty == nil {
		stdout.Panic("No PagerDuty")
	}
	return pagerDuty
}

func NewPagerDutyCommand() *cobra.Command {

	pagerDutyCmd := &cobra.Command{
		Use:   "pagerduty",
		Short: "PagerDuty tools",
	}
	flags := pagerDutyCmd.PersistentFlags()
	flags.IntVar(&pagerDutyOptions.Timeout, "pagerduty-timeout", pagerDutyOptions.Timeout, "pagerDuty timeout in seconds")
	flags.BoolVar(&pagerDutyOptions.Insecure, "pagerduty-insecure", pagerDutyOptions.Insecure, "pagerDuty insecure")
	flags.StringVar(&pagerDutyOptions.URL, "pagerduty-url", pagerDutyOptions.URL, "pagerDuty URL")
	flags.StringVar(&pagerDutyOptions.Token, "pagerduty-token", pagerDutyOptions.Token, "pagerDuty token")
	flags.StringVar(&pagerDutyOutput.Output, "pagerduty-output", pagerDutyOutput.Output, "pagerDuty output")
	flags.StringVar(&pagerDutyOutput.Query, "pagerduty-output-query", pagerDutyOutput.Query, "pagerDuty output query")

	// tools pagerduty get-incidents
	getIncidentsCmd := &cobra.Command{
		Use:   "get-incidents",
		Short: "Get incidents",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("PagerDuty getting incident...")
			common.Debug("PagerDuty", pagerDutyGetIncidentsOptions, stdout)

			bytes, err := pagerDutyNew(stdout).GetIncidents(pagerDutyGetIncidentsOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(pagerDutyOutput, "PagerDuty", []interface{}{pagerDutyOptions}, bytes, stdout)
		},
	}
	flags = getIncidentsCmd.PersistentFlags()
	flags.StringVar(&pagerDutyGetIncidentsOptions.Key, "pagerduty-incident-key", pagerDutyGetIncidentsOptions.Key, "PagerDuty incident key")
	flags.IntVar(&pagerDutyGetIncidentsOptions.Limit, "pagerduty-incidents-limit", pagerDutyGetIncidentsOptions.Limit, "PagerDuty incidents limit")
	pagerDutyCmd.AddCommand(getIncidentsCmd)

	incidentCmd := &cobra.Command{
		Use:   "incident",
		Short: "Incident methods",
	}
	flags = incidentCmd.PersistentFlags()
	flags.StringVar(&pagerDutyIncidentOptions.Title, "pagerduty-incident-title", pagerDutyIncidentOptions.Title, "PagerDuty incident title")
	flags.StringVar(&pagerDutyIncidentOptions.Body, "pagerduty-incident-body", pagerDutyIncidentOptions.Body, "PagerDuty incident body")
	flags.StringVar(&pagerDutyIncidentOptions.Urgency, "pagerduty-incident-urgency", pagerDutyIncidentOptions.Urgency, "PagerDuty incident urgency")
	flags.StringVar(&pagerDutyIncidentOptions.ServiceID, "pagerduty-incident-service-id", pagerDutyIncidentOptions.ServiceID, "PagerDuty incident service ID")
	flags.StringVar(&pagerDutyIncidentOptions.PriorityID, "pagerduty-incident-priority-id", pagerDutyIncidentOptions.PriorityID, "PagerDuty incident priority ID")
	pagerDutyCmd.AddCommand(incidentCmd)

	// tools pagerduty incident create --incident-params --create-incident-params
	createIncidentCmd := &cobra.Command{
		Use:   "create",
		Short: "Create incidnet",
		Run: func(cmd *cobra.Command, args []string) {

			stdout.Debug("PagerDuty creating incident...")
			common.Debug("PagerDuty", pagerDutyIncidentOptions, stdout)
			common.Debug("PagerDuty", pagerDutyCreateIncidentOptions, stdout)

			bodyBytes, err := utils.Content(pagerDutyIncidentOptions.Body)
			if err != nil {
				stdout.Panic(err)
			}
			pagerDutyIncidentOptions.Body = string(bodyBytes)

			bytes, err := pagerDutyNew(stdout).CreateIncident(pagerDutyIncidentOptions, pagerDutyCreateIncidentOptions)
			if err != nil {
				stdout.Error(err)
				return
			}
			common.OutputJson(pagerDutyOutput, "PagerDuty", []interface{}{pagerDutyOptions, pagerDutyIncidentOptions, pagerDutyCreateIncidentOptions}, bytes, stdout)
		},
	}
	flags = createIncidentCmd.PersistentFlags()
	flags.StringVar(&pagerDutyCreateIncidentOptions.From, "pagerduty-incident-from", pagerDutyCreateIncidentOptions.From, "PagerDuty incident from")
	incidentCmd.AddCommand(createIncidentCmd)

	return pagerDutyCmd
}
