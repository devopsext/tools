package vendors

import (
	"net/http"
	"net/url"
	"path"

	"github.com/devopsext/tools/common"
	"github.com/devopsext/utils"
)

const virustotalAPIURL = "https://www.virustotal.com/api/"
const virustotalAPIVersion = "v3"

const (
	virustotalGetDomainReport = "/domains"
)

type VirusTotal struct {
	client  *http.Client
	options VirusTotalOptions
}

type VirusTotalOptions struct {
	APIKey   string
	Timeout  int
	Insecure bool
}

type VirusTotalDomainReportOptions struct {
	Domain string
}

type VirusTotalDomainReportResult struct {
	DomainName *VirusTotalDomainNameData
}

type VirusTotalDomainNameData struct {
	ID                   string                                  `json:"id"`
	Type                 string                                  `json:"type"`
	LastAnalysisDate     int64                                   `json:"last_analysis_date"`
	LastModificationDate int64                                   `json:"last_modification_date"`
	LastAnalysisResult   map[string]VirusTotalLastAnalysisResult `json:"last_analysis_results"`
	Reputation           string                                  `json:"reputation"`
	TotalVotes           *VirusTotalTotalVotes
}

type VirusTotalLastAnalysisResult struct {
	Method     string `json:"method"`
	EngineName string `json:"engine_name"`
	Category   string `json:"category"`
	Result     string `json:"result"`
}

type VirusTotalTotalVotes struct {
	Harmless  int `json:"harmless"`
	Malicious int `json:"malicious"`
}

func (v *VirusTotal) DomainReport(options VirusTotalDomainReportOptions) ([]byte, error) {
	return v.CustomDomainReport(v.options, options)
}

func (v *VirusTotal) CustomDomainReport(virusTotalOptions VirusTotalOptions, virusTotalDomainReportOptions VirusTotalDomainReportOptions) ([]byte, error) {

	u, err := url.Parse(virustotalAPIURL + virustotalAPIVersion + virustotalGetDomainReport)
	if err != nil {
		return nil, err
	}

	u.Path = path.Join(u.Path, virusTotalDomainReportOptions.Domain)

	headers := make(map[string]string)
	headers["x-apikey"] = virusTotalOptions.APIKey
	headers["accept"] = "application/json"

	return utils.HttpGetRawWithHeaders(v.client, u.String(), headers)

}

func NewVirusTotal(options VirusTotalOptions, logger common.Logger) *VirusTotal {
	virustotal := &VirusTotal{
		client:  utils.NewHttpClient(options.Timeout, options.Insecure),
		options: options,
	}
	return virustotal
}
