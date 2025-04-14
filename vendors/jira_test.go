package vendors

import (
	"fmt"
	"runtime/debug"
	"testing"

	"github.com/devopsext/utils"
)

func envGet(s string, def interface{}) interface{} {
	return utils.EnvGet(fmt.Sprintf("%s_%s", "TOOLS", s), def)
}

func BenchmarkSearchAssets(b *testing.B) {
	// Set memory limit to 1GB
	debug.SetMemoryLimit(1024 * 1024 * 1024)

	j := NewJira(JiraOptions{
		URL:         envGet("JIRA_URL", "").(string),
		Timeout:     envGet("JIRA_TIMEOUT", 30).(int),
		Insecure:    envGet("JIRA_INSECURE", false).(bool),
		User:        envGet("JIRA_USER", "").(string),
		Password:    envGet("JIRA_PASSWORD", "").(string),
		AccessToken: envGet("JIRA_ACCESS_TOKEN", "").(string),
	})

	options := JiraSearchAssetOptions{
		SearchPattern: "objectType = \"Virtual Machine\" AND \"Status\" = \"In Use\" AND \"VM Cluster\" IN (\"jb-dta\",\"ld7-dta\",\"mi-dta\",\"nl-dta\",\"nl-dev\",\"nl-stage\",\"sg3-dta\",\"sl1-dta\",\"vsan-01\")",
		ResultPerPage: 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := j.SearchAssets(options)
		if err != nil {
			b.Fatalf("unexpected error: %v", err)
		}
	}
}
