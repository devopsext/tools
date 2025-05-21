package vendors

import (
	"encoding/json"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

// Mock server that returns a fixed response
func setupNetboxMockServer() *httptest.Server {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Respond with a fixed JSON response or appropriate logic to simulate a Netbox server
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// You need to add a test netbox_test.json with multiple devices
		jsonData, err := os.ReadFile("netbox_test.json")
		if err != nil {
			log.Fatal(err)
		}
		w.Write(jsonData) // Simplified response
	})
	return httptest.NewServer(handler)
}

func BenchmarkCustomGetDevices(b *testing.B) {
	// Setup the mock server
	server := setupNetboxMockServer()
	defer server.Close()

	// Assume Netbox and related structs are defined elsewhere
	netbox := &Netbox{
		client: server.Client(), // Use the mock server's client
		options: NetboxOptions{
			URL: server.URL, // Use the mock server's URL
		},
	}

	netboxDeviceOptions := NetboxDeviceOptions{
		// Fill with appropriate zero values or test cases
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
	// Setup the mock server
	server := setupNetboxMockServer()
	defer server.Close()

	// Mock netbox with a fake client
	// Assume Netbox and related structs are defined elsewhere
	netbox := &Netbox{
		client: server.Client(), // Use the mock server's client
		options: NetboxOptions{
			URL: server.URL, // Use the mock server's URL
		},
	}

	options := NetboxOptions{
		URL: server.URL,
	}
	netboxDeviceOptions := NetboxDeviceOptions{
		DeviceID: "",
	}

	got, err := netbox.CustomGetDevices(options, netboxDeviceOptions)
	if err != nil {
		t.Fatalf("CustomGetDevices returned an error: %v", err)
	}

	var data []NetboxDevice

	json.Unmarshal(got, &data)
	if err != nil {
		t.Fatalf("CustomGetDevices returned an error: %v", err)
	}

	if data[0].ID != 11 {
		t.Errorf("CustomGetDevices returned ID for the first device = %d, want %d", data[0].ID, 11)
	}
	if data[1].ID != 12 {
		t.Errorf("CustomGetDevices  returned ID for the first device = %d, want %d", data[0].ID, 12)
	}
}
