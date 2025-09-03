package vendors

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock server that returns a fixed response
func setupNetboxMockServer() *httptest.Server {
	// Hardcoded JSON response that matches the NetxboxAPIResponse structure
	mockResponse := `{
		"count": 2,
		"next": null,
		"previous": null,
		"results": [
			{
				"id": 11,
				"name": "device-11",
				"display": "Device 11"
			},
			{
				"id": 12,
				"name": "device-12",
				"display": "Device 12"
			}
		]
	}`

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(mockResponse))
	})
	return httptest.NewServer(handler)
}

func BenchmarkCustomGetDevices(b *testing.B) {
	// Setup the mock server
	server := setupNetboxMockServer()
	defer server.Close()

	// Create Netbox client with mock server
	netbox := &Netbox{
		client: server.Client(),
		options: NetboxOptions{
			URL: server.URL,
		},
	}

	netboxDeviceOptions := NetboxDeviceOptions{
		// Empty options for this test
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := netbox.CustomGetDevices(netbox.options, netboxDeviceOptions)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestCustomGetDevices(t *testing.T) {
	// Define test cases
	tests := []struct {
		name                string
		deviceID            string
		expectedDeviceCount int
		expectedFirstID     int
		expectedSecondID    int
	}{
		{
			name:                "Get all devices",
			deviceID:            "",
			expectedDeviceCount: 2,
			expectedFirstID:     11,
			expectedSecondID:    12,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup the mock server
			server := setupNetboxMockServer()
			defer server.Close()

			// Create Netbox client with mock server
			netbox := &Netbox{
				client: server.Client(),
				options: NetboxOptions{
					URL: server.URL,
				},
			}

			options := NetboxOptions{
				URL: server.URL,
			}
			netboxDeviceOptions := NetboxDeviceOptions{
				DeviceID: tt.deviceID,
			}

			// Call the function being tested
			got, err := netbox.CustomGetDevices(options, netboxDeviceOptions)

			// Assert no error occurred
			require.NoError(t, err, "CustomGetDevices should not return an error")

			// Unmarshal the response
			var data []NetboxDevice
			err = json.Unmarshal(got, &data)
			require.NoError(t, err, "Failed to unmarshal response")

			// Assert the expected results
			assert.Equal(t, tt.expectedDeviceCount, len(data), "Unexpected number of devices")
			assert.Equal(t, tt.expectedFirstID, data[0].ID, "First device has unexpected ID")
			assert.Equal(t, tt.expectedSecondID, data[1].ID, "Second device has unexpected ID")
		})
	}
}
