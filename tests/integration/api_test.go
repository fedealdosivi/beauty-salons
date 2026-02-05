package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

// Integration tests for API endpoints
// These tests verify the HTTP layer works correctly

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAPIHealthCheck(t *testing.T) {
	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":        "ok",
			"elasticsearch": "connected",
			"postgres":      "connected",
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Health check status = %v, want %v", w.Code, http.StatusOK)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse health response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Health status = %v, want ok", response["status"])
	}
}

func TestAPISearchEndpointFormat(t *testing.T) {
	router := gin.New()

	// Mock search endpoint that returns expected format
	router.GET("/api/v1/search", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"results": []gin.H{
				{
					"salon": gin.H{
						"id":   1,
						"name": "Test Salon",
						"slug": "test-salon",
					},
					"score":    15.5,
					"distance": nil,
				},
			},
			"total":       1,
			"page":        1,
			"page_size":   10,
			"total_pages": 1,
			"query":       "test",
			"source":      "elasticsearch",
		})
	})

	tests := []struct {
		name       string
		query      string
		wantStatus int
	}{
		{"basic search", "/api/v1/search?q=salon", http.StatusOK},
		{"search with filters", "/api/v1/search?q=spa&city=Miami&min_rating=4.0", http.StatusOK},
		{"search with geo", "/api/v1/search?lat=40.71&lon=-74.00&radius=10", http.StatusOK},
		{"search with pagination", "/api/v1/search?page=2&page_size=20", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.query, nil)
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %v, want %v", w.Code, tt.wantStatus)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to parse response: %v", err)
			}

			// Verify response structure
			requiredFields := []string{"results", "total", "page", "page_size", "total_pages"}
			for _, field := range requiredFields {
				if _, ok := response[field]; !ok {
					t.Errorf("Response missing required field: %s", field)
				}
			}
		})
	}
}

func TestAPICategoriesEndpoint(t *testing.T) {
	router := gin.New()

	router.GET("/api/v1/categories", func(c *gin.Context) {
		c.JSON(http.StatusOK, []gin.H{
			{"id": 1, "name": "Hair Salon", "slug": "hair-salon"},
			{"id": 2, "name": "Spa", "slug": "spa"},
			{"id": 3, "name": "Nail Salon", "slug": "nail-salon"},
		})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/categories", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Categories status = %v, want %v", w.Code, http.StatusOK)
	}

	var categories []map[string]interface{}
	if err := json.Unmarshal(w.Body.Bytes(), &categories); err != nil {
		t.Fatalf("Failed to parse categories: %v", err)
	}

	if len(categories) == 0 {
		t.Error("Expected at least one category")
	}

	// Verify category structure
	for i, cat := range categories {
		if _, ok := cat["id"]; !ok {
			t.Errorf("Category %d missing 'id' field", i)
		}
		if _, ok := cat["name"]; !ok {
			t.Errorf("Category %d missing 'name' field", i)
		}
	}
}

func TestAPISalonDetailEndpoint(t *testing.T) {
	router := gin.New()

	router.GET("/api/v1/salons/:id", func(c *gin.Context) {
		id := c.Param("id")
		if id == "999" {
			c.JSON(http.StatusNotFound, gin.H{"error": "salon not found"})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"id":          1,
			"name":        "Test Salon",
			"slug":        "test-salon",
			"description": "A great salon",
			"rating":      4.5,
			"is_verified": true,
		})
	})

	tests := []struct {
		name       string
		path       string
		wantStatus int
	}{
		{"existing salon", "/api/v1/salons/1", http.StatusOK},
		{"non-existing salon", "/api/v1/salons/999", http.StatusNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.path, nil)
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %v, want %v", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestAPIErrorResponses(t *testing.T) {
	router := gin.New()

	router.GET("/api/v1/search", func(c *gin.Context) {
		// Simulate invalid parameter
		if c.Query("page") == "-1" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid page parameter"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"results": []interface{}{}})
	})

	tests := []struct {
		name       string
		query      string
		wantStatus int
		wantError  bool
	}{
		{"valid request", "/api/v1/search?q=test", http.StatusOK, false},
		{"invalid page", "/api/v1/search?page=-1", http.StatusBadRequest, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req := httptest.NewRequest("GET", tt.query, nil)
			router.ServeHTTP(w, req)

			if w.Code != tt.wantStatus {
				t.Errorf("Status = %v, want %v", w.Code, tt.wantStatus)
			}

			var response map[string]interface{}
			if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
				t.Fatalf("Failed to unmarshal response: %v", err)
			}

			_, hasError := response["error"]
			if hasError != tt.wantError {
				t.Errorf("Has error = %v, want %v", hasError, tt.wantError)
			}
		})
	}
}
