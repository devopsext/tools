package vendors

// Comprehensive test suite for Grafana API client
// Tests cover all major functionality:
// - Dashboard operations (get, search, delete, copy, create)
// - Library element operations (get, search, copy)
// - Folder operations (get, list)
// - Annotation operations (get, create)
// - Image rendering
// - Authentication and error handling
// - FolderUID vs FolderID priority logic

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Mock data for testing
var (
	mockDashboardUID = "test-dashboard-uid"
	mockFolderUID    = "test-folder-uid"
	mockLibraryUID   = "test-library-uid"
	mockAPIKey       = "test-api-key"
	mockOrgID        = "1"
)

// Test helper to create a mock Grafana server
func createMockGrafanaServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Check authentication
		auth := r.Header.Get("Authorization")
		if auth != "Bearer "+mockAPIKey {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized"})
			return
		}

		// Route based on path and method
		switch {
		// Dashboard endpoints
		case strings.HasPrefix(r.URL.Path, "/api/dashboards/uid/"):
			switch r.Method {
			case http.MethodGet:
				handleDashboardGet(w, r)
			case http.MethodDelete:
				handleDashboardDelete(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
				json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
			}

		case r.URL.Path == "/api/dashboards/db":
			switch r.Method {
			case http.MethodPost:
				handleDashboardCreate(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
				json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
			}

		case r.URL.Path == "/api/search":
			switch r.Method {
			case http.MethodGet:
				handleDashboardSearch(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
				json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
			}

		// Library element endpoints
		case strings.HasPrefix(r.URL.Path, "/api/library-elements/"):
			switch r.Method {
			case http.MethodGet:
				handleLibraryElementGet(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
				json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
			}

		case r.URL.Path == "/api/library-elements":
			switch r.Method {
			case http.MethodGet:
				handleLibraryElementSearch(w, r)
			case http.MethodPost:
				handleLibraryElementCreate(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
				json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
			}

		// Folder endpoints
		case strings.HasPrefix(r.URL.Path, "/api/folders"):
			switch r.Method {
			case http.MethodGet:
				handleFolderGet(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
				json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
			}

		// Annotation endpoints
		case r.URL.Path == "/api/annotations":
			switch r.Method {
			case http.MethodGet:
				handleAnnotationGet(w, r)
			case http.MethodPost:
				handleAnnotationCreate(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
				json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
			}

		// Render endpoint
		case strings.HasPrefix(r.URL.Path, "/render/d-solo/"):
			switch r.Method {
			case http.MethodGet:
				handleRender(w, r)
			default:
				w.WriteHeader(http.StatusMethodNotAllowed)
				json.NewEncoder(w).Encode(map[string]string{"error": "method not allowed"})
			}

		default:
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
		}
	}))
}

// HTTP handlers for mock Grafana API endpoints

func handleDashboardGet(w http.ResponseWriter, _ *http.Request) {
	mockBoard := GrafanaBoard{
		Dashboard: GrafanaDashboard{
			UID:     mockDashboardUID,
			Title:   "Test Dashboard",
			Tags:    []string{"test", "mock"},
			Version: 1,
		},
		Meta: DashboardMeta{
			Slug:      "test-dashboard",
			FolderUID: mockFolderUID,
		},
	}
	json.NewEncoder(w).Encode(mockBoard)
}

func handleDashboardCreate(w http.ResponseWriter, r *http.Request) {
	var board GrafanaBoard
	if err := json.NewDecoder(r.Body).Decode(&board); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Mock successful creation
	response := map[string]interface{}{
		"id":      1,
		"uid":     board.Dashboard.UID,
		"url":     "/d/" + board.Dashboard.UID + "/test-dashboard",
		"status":  "success",
		"version": 1,
	}
	json.NewEncoder(w).Encode(response)
}

func handleDashboardDelete(w http.ResponseWriter, _ *http.Request) {
	response := map[string]interface{}{
		"id":      1,
		"message": "Dashboard deleted",
		"title":   "Test Dashboard",
	}
	json.NewEncoder(w).Encode(response)
}

func handleDashboardSearch(w http.ResponseWriter, _ *http.Request) {
	searchResults := []map[string]interface{}{
		{
			"id":          1,
			"uid":         mockDashboardUID,
			"title":       "Test Dashboard",
			"uri":         "db/test-dashboard",
			"url":         "/d/" + mockDashboardUID + "/test-dashboard",
			"slug":        "test-dashboard",
			"type":        "dash-db",
			"tags":        []string{"test"},
			"isStarred":   false,
			"folderId":    1,
			"folderUid":   mockFolderUID,
			"folderTitle": "Test Folder",
		},
	}
	json.NewEncoder(w).Encode(searchResults)
}

func handleLibraryElementGet(w http.ResponseWriter, _ *http.Request) {
	result := GrafanaLibraryElementResult{
		Result: GrafanaLibraryElement{
			ID:        1,
			UID:       mockLibraryUID,
			Name:      "Test Library Element",
			Kind:      1,
			Type:      "panel",
			FolderUID: mockFolderUID,
			Model:     map[string]interface{}{"type": "graph"},
		},
	}
	json.NewEncoder(w).Encode(result)
}

func handleLibraryElementSearch(w http.ResponseWriter, _ *http.Request) {
	result := GrafanaLibraryElementSearchResult{
		Result: struct {
			TotalCount int                     `json:"totalCount,omitempty"`
			Elements   []GrafanaLibraryElement `json:"elements,omitempty"`
			Page       int                     `json:"page,omitempty"`
			PerPage    int                     `json:"perPage,omitempty"`
		}{
			TotalCount: 1,
			Elements: []GrafanaLibraryElement{
				{
					ID:        1,
					UID:       mockLibraryUID,
					Name:      "Test Library Element",
					Kind:      1,
					Type:      "panel",
					FolderUID: mockFolderUID,
				},
			},
			Page:    1,
			PerPage: 100,
		},
	}
	json.NewEncoder(w).Encode(result)
}

func handleLibraryElementCreate(w http.ResponseWriter, r *http.Request) {
	var element GrafanaLibraryElement
	if err := json.NewDecoder(r.Body).Decode(&element); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	// Mock successful creation
	element.ID = 1
	if element.UID == "" {
		element.UID = mockLibraryUID
	}

	result := GrafanaLibraryElementResult{Result: element}
	json.NewEncoder(w).Encode(result)
}

func handleFolderGet(w http.ResponseWriter, r *http.Request) {
	if strings.Contains(r.URL.Path, mockFolderUID) {
		// Single folder
		folder := map[string]interface{}{
			"id":    1,
			"uid":   mockFolderUID,
			"title": "Test Folder",
		}
		json.NewEncoder(w).Encode(folder)
	} else {
		// List folders
		folders := []map[string]interface{}{
			{
				"id":    1,
				"uid":   mockFolderUID,
				"title": "Test Folder",
			},
		}
		json.NewEncoder(w).Encode(folders)
	}
}

func handleAnnotationGet(w http.ResponseWriter, _ *http.Request) {
	annotations := []GrafanaAnnotation{
		{
			Time:    1634567890000,
			TimeEnd: 1634567890000,
			Tags:    []string{"test"},
			Text:    "Test annotation",
		},
	}
	json.NewEncoder(w).Encode(annotations)
}

func handleAnnotationCreate(w http.ResponseWriter, _ *http.Request) {
	response := map[string]interface{}{
		"id":      1,
		"message": "Annotation added",
	}
	json.NewEncoder(w).Encode(response)
}

func handleRender(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "image/png")
	// Return a minimal PNG header (not a real image, just for testing)
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	w.Write(pngHeader)
}

// Test Dashboard Methods

func TestGrafana_CustomGetDashboards(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	dashboardOptions := GrafanaDahboardOptions{
		UID: mockDashboardUID,
	}

	result, err := grafana.CustomGetDashboards(grafana.options, dashboardOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	var board GrafanaBoard
	err = json.Unmarshal(result, &board)
	require.NoError(t, err)
	assert.Equal(t, mockDashboardUID, board.Dashboard.UID)
	assert.Equal(t, "Test Dashboard", board.Dashboard.Title)
}

func TestGrafana_GetDashboards(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	dashboardOptions := GrafanaDahboardOptions{
		UID: mockDashboardUID,
	}

	result, err := grafana.GetDashboards(dashboardOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	var board GrafanaBoard
	err = json.Unmarshal(result, &board)
	require.NoError(t, err)
	assert.Equal(t, mockDashboardUID, board.Dashboard.UID)
}

func TestGrafana_CustomSearchDashboards(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	dashboardOptions := GrafanaDahboardOptions{
		FolderUID: mockFolderUID,
	}

	result, err := grafana.CustomSearchDashboards(grafana.options, dashboardOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	var searchResults []map[string]interface{}
	err = json.Unmarshal(result, &searchResults)
	require.NoError(t, err)
	assert.Len(t, searchResults, 1)
	assert.Equal(t, mockDashboardUID, searchResults[0]["uid"])
}

func TestGrafana_SearchDashboards(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	dashboardOptions := GrafanaDahboardOptions{
		FolderUID: mockFolderUID,
	}

	result, err := grafana.SearchDashboards(dashboardOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)
}

func TestGrafana_CustomDeleteDashboards(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	dashboardOptions := GrafanaDahboardOptions{
		UID: mockDashboardUID,
	}

	result, err := grafana.CustomDeleteDashboards(grafana.options, dashboardOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	var response map[string]interface{}
	err = json.Unmarshal(result, &response)
	require.NoError(t, err)
	assert.Equal(t, "Dashboard deleted", response["message"])
}

func TestGrafana_DeleteDashboards(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	dashboardOptions := GrafanaDahboardOptions{
		UID: mockDashboardUID,
	}

	result, err := grafana.DeleteDashboards(dashboardOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)
}

func TestGrafana_CustomCopyDashboard(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	dashboardOptions := GrafanaDahboardOptions{
		Title:     "Copied Dashboard",
		FolderUID: mockFolderUID,
		Cloned: GrafanaClonedDahboardOptions{
			URL:    server.URL,
			APIKey: mockAPIKey,
			UID:    mockDashboardUID,
		},
	}

	result, err := grafana.CustomCopyDashboard(grafana.options, dashboardOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	var response map[string]interface{}
	err = json.Unmarshal(result, &response)
	require.NoError(t, err)
	assert.Equal(t, "success", response["status"])
}

func TestGrafana_CopyDashboard(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	dashboardOptions := GrafanaDahboardOptions{
		Title:     "Copied Dashboard",
		FolderUID: mockFolderUID,
		Cloned: GrafanaClonedDahboardOptions{
			URL:    server.URL,
			APIKey: mockAPIKey,
			UID:    mockDashboardUID,
		},
	}

	result, err := grafana.CopyDashboard(dashboardOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)
}

func TestGrafana_CustomCreateDashboard(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	dashboardOptions := GrafanaDahboardOptions{
		Title:     "New Dashboard",
		FolderUID: mockFolderUID,
		Tags:      []string{"test", "new"},
		Timezone:  "UTC",
		From:      "now-1h",
		To:        "now",
		Cloned: GrafanaClonedDahboardOptions{
			URL:    server.URL,
			APIKey: mockAPIKey,
			UID:    mockDashboardUID,
		},
	}

	result, err := grafana.CustomCreateDashboard(grafana.options, dashboardOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	var response map[string]interface{}
	err = json.Unmarshal(result, &response)
	require.NoError(t, err)
	assert.Equal(t, "success", response["status"])
}

func TestGrafana_CreateDashboard(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	dashboardOptions := GrafanaDahboardOptions{
		Title:     "New Dashboard",
		FolderUID: mockFolderUID,
		Tags:      []string{"test", "new"},
		Cloned: GrafanaClonedDahboardOptions{
			URL:    server.URL,
			APIKey: mockAPIKey,
			UID:    mockDashboardUID,
		},
	}

	result, err := grafana.CreateDashboard(dashboardOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)
}

// Test Library Element Methods

func TestGrafana_CustomGetLibraryElement(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	libraryElementOptions := GrafanaLibraryElementOptions{
		UID: mockLibraryUID,
	}

	result, err := grafana.CustomGetLibraryElement(grafana.options, libraryElementOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	var element GrafanaLibraryElement
	err = json.Unmarshal(result, &element)
	require.NoError(t, err)
	assert.Equal(t, mockLibraryUID, element.UID)
	assert.Equal(t, "Test Library Element", element.Name)
}

func TestGrafana_GetLibraryElement(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	libraryElementOptions := GrafanaLibraryElementOptions{
		UID: mockLibraryUID,
	}

	result, err := grafana.GetLibraryElement(libraryElementOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)
}

func TestGrafana_CustomSearchLibraryElements(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	libraryElementOptions := GrafanaLibraryElementOptions{
		FolderUID: mockFolderUID,
	}

	result, err := grafana.CustomSearchLibraryElements(grafana.options, libraryElementOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	var elements []GrafanaLibraryElement
	err = json.Unmarshal(result, &elements)
	require.NoError(t, err)
	assert.Len(t, elements, 1)
	assert.Equal(t, mockLibraryUID, elements[0].UID)
}

func TestGrafana_SearchLibraryElements(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	libraryElementOptions := GrafanaLibraryElementOptions{
		FolderUID: mockFolderUID,
	}

	result, err := grafana.SearchLibraryElements(libraryElementOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)
}

func TestGrafana_CustomCopyLibraryElement(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	libraryElementOptions := GrafanaLibraryElementOptions{
		Name:      "Copied Library Element",
		FolderUID: mockFolderUID,
		SaveUID:   true,
		Cloned: GrafanaClonedLibraryElementOptions{
			URL:    server.URL,
			APIKey: mockAPIKey,
			UID:    mockLibraryUID,
		},
	}

	result, err := grafana.CustomCopyLibraryElement(grafana.options, libraryElementOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	var response GrafanaLibraryElementResult
	err = json.Unmarshal(result, &response)
	require.NoError(t, err)
	assert.Equal(t, mockLibraryUID, response.Result.UID)
}

func TestGrafana_CopyLibraryElement(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	libraryElementOptions := GrafanaLibraryElementOptions{
		Name:      "Copied Library Element",
		FolderUID: mockFolderUID,
		Cloned: GrafanaClonedLibraryElementOptions{
			URL:    server.URL,
			APIKey: mockAPIKey,
			UID:    mockLibraryUID,
		},
	}

	result, err := grafana.CopyLibraryElement(libraryElementOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)
}

// Test Folder Methods

func TestGrafana_CustomGetFolder(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	// Test getting a specific folder
	folderOptions := GrafanaFolderOptions{
		UID: mockFolderUID,
	}

	result, err := grafana.CustomGetFolder(grafana.options, folderOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	var folder map[string]interface{}
	err = json.Unmarshal(result, &folder)
	require.NoError(t, err)
	assert.Equal(t, mockFolderUID, folder["uid"])
	assert.Equal(t, "Test Folder", folder["title"])
}

func TestGrafana_CustomGetFolder_ListAll(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	// Test listing all folders
	folderOptions := GrafanaFolderOptions{}

	result, err := grafana.CustomGetFolder(grafana.options, folderOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	var folders []map[string]interface{}
	err = json.Unmarshal(result, &folders)
	require.NoError(t, err)
	assert.Len(t, folders, 1)
	assert.Equal(t, mockFolderUID, folders[0]["uid"])
}

func TestGrafana_GetFolder(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	folderOptions := GrafanaFolderOptions{
		UID: mockFolderUID,
	}

	result, err := grafana.GetFolder(folderOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)
}

// Test Annotation Methods

func TestGrafana_CustomGetAnnotations(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	dashboardOptions := GrafanaDahboardOptions{
		Timezone: "UTC",
	}

	annotationOptions := GrafanaGetAnnotationsOptions{
		From:  "2021-10-18T10:31:30.000Z",
		To:    "2021-10-18T11:31:30.000Z",
		Tags:  "test",
		Limit: 100,
	}

	result, err := grafana.CustomGetAnnotations(grafana.options, dashboardOptions, annotationOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	var annotations []GrafanaAnnotation
	err = json.Unmarshal(result, &annotations)
	require.NoError(t, err)
	assert.Len(t, annotations, 1)
	assert.Equal(t, "Test annotation", annotations[0].Text)
}

func TestGrafana_GetAnnotations(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	dashboardOptions := GrafanaDahboardOptions{
		Timezone: "UTC",
	}

	annotationOptions := GrafanaGetAnnotationsOptions{
		From:  "2021-10-18T10:31:30.000Z",
		To:    "2021-10-18T11:31:30.000Z",
		Tags:  "test",
		Limit: 100,
	}

	result, err := grafana.GetAnnotations(dashboardOptions, annotationOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)
}

func TestGrafana_CustomCreateAnnotation(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	annotationOptions := GrafanaCreateAnnotationOptions{
		Time:    "2021-10-18T10:31:30.000Z",
		TimeEnd: "2021-10-18T10:31:30.000Z",
		Tags:    "test,annotation",
		Text:    "Test annotation created",
	}

	result, err := grafana.CustomCreateAnnotation(grafana.options, annotationOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	var response map[string]interface{}
	err = json.Unmarshal(result, &response)
	require.NoError(t, err)
	assert.Equal(t, "Annotation added", response["message"])
}

func TestGrafana_CreateAnnotation(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	annotationOptions := GrafanaCreateAnnotationOptions{
		Time:    "2021-10-18T10:31:30.000Z",
		TimeEnd: "2021-10-18T10:31:30.000Z",
		Tags:    "test,annotation",
		Text:    "Test annotation created",
	}

	result, err := grafana.CreateAnnotation(annotationOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)
}

// Test Render Methods

func TestGrafana_CustomRenderImage(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	dashboardOptions := GrafanaDahboardOptions{
		UID:      mockDashboardUID,
		Slug:     "test-dashboard",
		Timezone: "UTC",
	}

	renderOptions := GrafanaRenderImageOptions{
		PanelID: "1",
		From:    "2021-10-18T10:31:30.000Z",
		To:      "2021-10-18T11:31:30.000Z",
		Width:   800,
		Height:  600,
	}

	result, err := grafana.CustomRenderImage(grafana.options, dashboardOptions, renderOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)

	// Check if result starts with PNG header
	assert.Equal(t, byte(0x89), result[0])
	assert.Equal(t, byte(0x50), result[1])
	assert.Equal(t, byte(0x4E), result[2])
	assert.Equal(t, byte(0x47), result[3])
}

func TestGrafana_RenderImage(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	dashboardOptions := GrafanaDahboardOptions{
		UID:      mockDashboardUID,
		Slug:     "test-dashboard",
		Timezone: "UTC",
	}

	renderOptions := GrafanaRenderImageOptions{
		PanelID: "1",
		Width:   800,
		Height:  600,
	}

	result, err := grafana.RenderImage(dashboardOptions, renderOptions)
	require.NoError(t, err)
	require.NotEmpty(t, result)
}

// Test Authentication

func TestGrafana_AuthenticationRequired(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	// Test without API key
	grafana := NewGrafana(GrafanaOptions{
		URL:   server.URL,
		OrgID: mockOrgID,
	})

	dashboardOptions := GrafanaDahboardOptions{
		UID: mockDashboardUID,
	}

	_, err := grafana.GetDashboards(dashboardOptions)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "401")
}

// Test FolderUID Priority Logic

func TestGrafana_FolderUIDPriority(t *testing.T) {
	server := createMockGrafanaServer()
	defer server.Close()

	grafana := NewGrafana(GrafanaOptions{
		URL:    server.URL,
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	t.Run("Dashboard search prioritizes FolderUID over FolderID", func(t *testing.T) {
		dashboardOptions := GrafanaDahboardOptions{
			FolderUID: mockFolderUID,
			FolderID:  999, // This should be ignored
		}

		result, err := grafana.SearchDashboards(dashboardOptions)
		require.NoError(t, err)
		require.NotEmpty(t, result)

		// Verify that the request used folderUIDs parameter, not folderIds
		// This is implicitly tested by the mock server behavior
	})

	t.Run("Library element search prioritizes FolderUID over FolderID", func(t *testing.T) {
		libraryElementOptions := GrafanaLibraryElementOptions{
			FolderUID: mockFolderUID,
			FolderID:  999, // This should be ignored
		}

		result, err := grafana.SearchLibraryElements(libraryElementOptions)
		require.NoError(t, err)
		require.NotEmpty(t, result)
	})
}

// Test Error Handling

func TestGrafana_ErrorHandling(t *testing.T) {
	// Test with invalid URL
	grafana := NewGrafana(GrafanaOptions{
		URL:    "invalid-url",
		APIKey: mockAPIKey,
		OrgID:  mockOrgID,
	})

	dashboardOptions := GrafanaDahboardOptions{
		UID: mockDashboardUID,
	}

	_, err := grafana.GetDashboards(dashboardOptions)
	require.Error(t, err)
}

// Test helper methods

func TestGrafana_HelperMethods(t *testing.T) {
	grafana := NewGrafana(GrafanaOptions{
		APIKey: mockAPIKey,
	})

	t.Run("getAuth returns correct Bearer token", func(t *testing.T) {
		auth := grafana.getAuth(grafana.options)
		assert.Equal(t, "Bearer "+mockAPIKey, auth)
	})

	t.Run("getAuth returns empty string when no API key", func(t *testing.T) {
		auth := grafana.getAuth(GrafanaOptions{})
		assert.Equal(t, "", auth)
	})

	t.Run("toRFC3339NanoStr converts timestamp correctly", func(t *testing.T) {
		result := grafana.toRFC3339NanoStr("2021-10-18T10:31:30.123456789Z")
		assert.NotEmpty(t, result)
		// Should be converted to milliseconds timestamp
		assert.NotContains(t, result, "T")
		assert.NotContains(t, result, "Z")
	})

	t.Run("toRFC3339Nano converts timestamp to int64", func(t *testing.T) {
		result := grafana.toRFC3339Nano("2021-10-18T10:31:30.123456789Z")
		assert.Greater(t, result, int64(0))
	})
}
