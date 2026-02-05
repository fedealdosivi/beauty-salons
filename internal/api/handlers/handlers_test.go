package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"beauty-salons/internal/domain"

	"github.com/gin-gonic/gin"
)

func TestParseSearchParams(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name        string
		queryString string
		check       func(t *testing.T, params domain.SalonSearchParams)
	}{
		{
			name:        "basic query",
			queryString: "q=haircut",
			check: func(t *testing.T, p domain.SalonSearchParams) {
				if p.Query != "haircut" {
					t.Errorf("Query = %v, want haircut", p.Query)
				}
			},
		},
		{
			name:        "city filter",
			queryString: "city=Miami",
			check: func(t *testing.T, p domain.SalonSearchParams) {
				if p.City != "Miami" {
					t.Errorf("City = %v, want Miami", p.City)
				}
			},
		},
		{
			name:        "category filter",
			queryString: "category=5",
			check: func(t *testing.T, p domain.SalonSearchParams) {
				if p.CategoryID == nil || *p.CategoryID != 5 {
					t.Errorf("CategoryID = %v, want 5", p.CategoryID)
				}
			},
		},
		{
			name:        "price range filter",
			queryString: "price_range=3",
			check: func(t *testing.T, p domain.SalonSearchParams) {
				if p.PriceRange != 3 {
					t.Errorf("PriceRange = %v, want 3", p.PriceRange)
				}
			},
		},
		{
			name:        "min rating filter",
			queryString: "min_rating=4.5",
			check: func(t *testing.T, p domain.SalonSearchParams) {
				if p.MinRating == nil || *p.MinRating != 4.5 {
					t.Errorf("MinRating = %v, want 4.5", p.MinRating)
				}
			},
		},
		{
			name:        "verified filter",
			queryString: "verified=true",
			check: func(t *testing.T, p domain.SalonSearchParams) {
				if p.IsVerified == nil || !*p.IsVerified {
					t.Error("IsVerified should be true")
				}
			},
		},
		{
			name:        "geo search",
			queryString: "lat=40.7128&lon=-74.0060&radius=10",
			check: func(t *testing.T, p domain.SalonSearchParams) {
				if p.Location == nil {
					t.Fatal("Location should not be nil")
				}
				if p.Location.Latitude != 40.7128 {
					t.Errorf("Latitude = %v, want 40.7128", p.Location.Latitude)
				}
				if p.Location.Longitude != -74.0060 {
					t.Errorf("Longitude = %v, want -74.0060", p.Location.Longitude)
				}
				if p.RadiusKm == nil || *p.RadiusKm != 10 {
					t.Errorf("RadiusKm = %v, want 10", p.RadiusKm)
				}
			},
		},
		{
			name:        "sort option",
			queryString: "sort=rating",
			check: func(t *testing.T, p domain.SalonSearchParams) {
				if p.SortBy != "rating" {
					t.Errorf("SortBy = %v, want rating", p.SortBy)
				}
			},
		},
		{
			name:        "pagination",
			queryString: "page=2&page_size=20",
			check: func(t *testing.T, p domain.SalonSearchParams) {
				if p.Page != 2 {
					t.Errorf("Page = %v, want 2", p.Page)
				}
				if p.PageSize != 20 {
					t.Errorf("PageSize = %v, want 20", p.PageSize)
				}
			},
		},
		{
			name:        "default pagination",
			queryString: "",
			check: func(t *testing.T, p domain.SalonSearchParams) {
				if p.Page != 1 {
					t.Errorf("Page = %v, want 1 (default)", p.Page)
				}
				if p.PageSize != 10 {
					t.Errorf("PageSize = %v, want 10 (default)", p.PageSize)
				}
			},
		},
		{
			name:        "combined filters",
			queryString: "q=spa&city=LA&min_rating=4.0&verified=true&page=1&page_size=5",
			check: func(t *testing.T, p domain.SalonSearchParams) {
				if p.Query != "spa" {
					t.Errorf("Query = %v, want spa", p.Query)
				}
				if p.City != "LA" {
					t.Errorf("City = %v, want LA", p.City)
				}
				if p.MinRating == nil || *p.MinRating != 4.0 {
					t.Errorf("MinRating = %v, want 4.0", p.MinRating)
				}
				if p.IsVerified == nil || !*p.IsVerified {
					t.Error("IsVerified should be true")
				}
				if p.PageSize != 5 {
					t.Errorf("PageSize = %v, want 5", p.PageSize)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/search?"+tt.queryString, nil)

			h := &Handler{}
			params := h.parseSearchParams(c)
			tt.check(t, params)
		})
	}
}

func TestSalonsToSearchResults(t *testing.T) {
	salons := []domain.Salon{
		{ID: 1, Name: "Salon A"},
		{ID: 2, Name: "Salon B"},
		{ID: 3, Name: "Salon C"},
	}

	results := salonsToSearchResults(salons)

	if len(results) != 3 {
		t.Fatalf("len(results) = %v, want 3", len(results))
	}

	for i, r := range results {
		if r.Salon.ID != salons[i].ID {
			t.Errorf("results[%d].Salon.ID = %v, want %v", i, r.Salon.ID, salons[i].ID)
		}
		if r.Score != 0 {
			t.Errorf("results[%d].Score = %v, want 0", i, r.Score)
		}
		if r.Distance != nil {
			t.Errorf("results[%d].Distance should be nil", i)
		}
	}
}

func TestSearchResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Create a test response
	results := []domain.SalonSearchResult{
		{
			Salon: domain.Salon{
				ID:   1,
				Name: "Test Salon",
				Slug: "test-salon",
			},
			Score: 15.5,
		},
	}

	params := domain.SalonSearchParams{
		Query:    "test",
		Page:     1,
		PageSize: 10,
	}

	response := domain.NewSearchResponse(results, 1, params)
	response.Source = "elasticsearch"

	// Marshal and check JSON structure
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		t.Fatalf("Failed to marshal response: %v", err)
	}

	var parsed map[string]interface{}
	if err := json.Unmarshal(jsonBytes, &parsed); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Check expected fields exist
	expectedFields := []string{"results", "total", "page", "page_size", "total_pages", "source"}
	for _, field := range expectedFields {
		if _, ok := parsed[field]; !ok {
			t.Errorf("Response missing field: %s", field)
		}
	}

	// Check results structure
	resultsArr, ok := parsed["results"].([]interface{})
	if !ok {
		t.Fatal("results should be an array")
	}
	if len(resultsArr) != 1 {
		t.Fatalf("len(results) = %v, want 1", len(resultsArr))
	}

	result := resultsArr[0].(map[string]interface{})
	if _, ok := result["salon"]; !ok {
		t.Error("Result missing 'salon' field")
	}
	if _, ok := result["score"]; !ok {
		t.Error("Result missing 'score' field")
	}
}

// Mock handler for integration tests
type mockHandler struct {
	*Handler
}

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Status code = %v, want %v", w.Code, http.StatusOK)
	}

	var response map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("status = %v, want ok", response["status"])
	}
}
