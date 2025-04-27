package vendors

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime/debug"
	"testing"

	"github.com/devopsext/utils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestHttpGetStream(t *testing.T) {
	j := NewJira(JiraOptions{
		URL:      "http://example.com",
		Timeout:  10,
		Insecure: false,
		User:     "user",
		Password: "password",
	})

	tests := []struct {
		name           string
		serverHandler  http.HandlerFunc
		expectedResult string
		expectedError  string
	}{
		{
			name: "Returns data on successful request",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(`{"key": "value"}`))
			},
			expectedResult: `{"key": "value"}`,
			expectedError:  "",
		},
		{
			name: "Retries on too many requests",
			serverHandler: func() http.HandlerFunc {
				attempts := 0
				return func(w http.ResponseWriter, r *http.Request) {
					attempts++
					if attempts < 3 {
						w.Header().Set("Retry-After", "1")
						w.WriteHeader(http.StatusTooManyRequests)
					} else {
						w.WriteHeader(http.StatusOK)
						w.Write([]byte(`{"key": "value"}`))
					}
				}
			}(),
			expectedResult: `{"key": "value"}`,
			expectedError:  "",
		},
		{
			name: "Fail on too many requests",
			serverHandler: func() http.HandlerFunc {
				attempts := 0
				return func(w http.ResponseWriter, r *http.Request) {
					attempts++
					w.Header().Set("Retry-After", "1")
					w.WriteHeader(http.StatusTooManyRequests)
				}
			}(),
			expectedResult: "",
			expectedError:  "too many requests",
		},
		{
			name: "Returns error on non-OK status",
			serverHandler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusInternalServerError)
			},
			expectedResult: "",
			expectedError:  "500 Internal Server Error",
		},
		{
			name:           "Returns error on request failure",
			serverHandler:  nil, // No server, simulating invalid URL
			expectedResult: "",
			expectedError:  "unsupported protocol scheme",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var server *httptest.Server
			if tt.serverHandler != nil {
				server = httptest.NewServer(tt.serverHandler)
				defer server.Close()
			}

			url := "blabla://invalid-url"
			if server != nil {
				url = server.URL
			}

			res, err := j.httpGetStream(url)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, res.String())
			}
		})
	}
}
