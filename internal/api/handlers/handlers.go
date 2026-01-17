package handlers

import (
	"net/http"
	"strconv"

	"beauty-salons/internal/domain"
	"beauty-salons/internal/repository"
	"beauty-salons/internal/search"

	"github.com/gin-gonic/gin"
)

// Handler contains all HTTP handlers
type Handler struct {
	repo *repository.PostgresRepository
	es   *search.ElasticsearchClient
}

// NewHandler creates a new handler instance
func NewHandler(repo *repository.PostgresRepository, es *search.ElasticsearchClient) *Handler {
	return &Handler{
		repo: repo,
		es:   es,
	}
}

// SearchResponse represents the API response for search endpoints
type SearchResponse struct {
	Data       []domain.Salon `json:"data"`
	Total      int            `json:"total"`
	Page       int            `json:"page"`
	PageSize   int            `json:"page_size"`
	TotalPages int            `json:"total_pages"`
	Source     string         `json:"source"` // "elasticsearch" or "postgresql"
}

// SearchSalons handles search requests using Elasticsearch
// GET /api/v1/search?q=...&city=...&category=...&min_rating=...&verified=...
func (h *Handler) SearchSalons(c *gin.Context) {
	params := h.parseSearchParams(c)

	salons, total, err := h.es.Search(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Search failed: " + err.Error(),
		})
		return
	}

	h.sendSearchResponse(c, salons, total, params, "elasticsearch")
}

// SearchSalonsPostgres handles search using PostgreSQL (for comparison)
// GET /api/v1/search/postgres?q=...
func (h *Handler) SearchSalonsPostgres(c *gin.Context) {
	params := h.parseSearchParams(c)

	salons, total, err := h.repo.SearchSalons(c.Request.Context(), params)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Search failed: " + err.Error(),
		})
		return
	}

	h.sendSearchResponse(c, salons, total, params, "postgresql")
}

// GetSalon retrieves a single salon by ID
// GET /api/v1/salons/:id
func (h *Handler) GetSalon(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid salon ID"})
		return
	}

	salon, err := h.repo.GetSalonByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Salon not found"})
		return
	}

	c.JSON(http.StatusOK, salon)
}

// GetCategories retrieves all categories
// GET /api/v1/categories
func (h *Handler) GetCategories(c *gin.Context) {
	categories, err := h.repo.GetCategories(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, categories)
}

// SyncToElasticsearch syncs all data from PostgreSQL to Elasticsearch
// POST /api/v1/admin/sync
func (h *Handler) SyncToElasticsearch(c *gin.Context) {
	ctx := c.Request.Context()

	// Get all salons from PostgreSQL
	salons, err := h.repo.GetAllSalons(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch salons: " + err.Error()})
		return
	}

	// Enrich with services (in a real app, you'd batch this)
	for i := range salons {
		fullSalon, err := h.repo.GetSalonByID(ctx, salons[i].ID)
		if err == nil {
			salons[i].Services = fullSalon.Services
			salons[i].Amenities = fullSalon.Amenities
		}
	}

	// Delete and recreate index for clean sync
	if err := h.es.DeleteIndex(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete index: " + err.Error()})
		return
	}

	if err := h.es.CreateIndex(ctx); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create index: " + err.Error()})
		return
	}

	// Bulk index all salons
	if err := h.es.BulkIndexSalons(ctx, salons); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to index salons: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Sync completed successfully",
		"count":   len(salons),
	})
}

// GetClusterHealth returns Elasticsearch cluster health
// GET /api/v1/admin/cluster/health
func (h *Handler) GetClusterHealth(c *gin.Context) {
	health, err := h.es.GetClusterHealth(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, health)
}

// GetIndexStats returns Elasticsearch index statistics
// GET /api/v1/admin/cluster/stats
func (h *Handler) GetIndexStats(c *gin.Context) {
	stats, err := h.es.GetIndexStats(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// Helper methods

func (h *Handler) parseSearchParams(c *gin.Context) domain.SalonSearchParams {
	params := domain.SalonSearchParams{
		Query: c.Query("q"),
		City:  c.Query("city"),
	}

	if categoryStr := c.Query("category"); categoryStr != "" {
		if cat, err := strconv.ParseInt(categoryStr, 10, 64); err == nil {
			params.CategoryID = &cat
		}
	}

	if priceStr := c.Query("price_range"); priceStr != "" {
		if pr, err := strconv.Atoi(priceStr); err == nil {
			params.PriceRange = &pr
		}
	}

	if ratingStr := c.Query("min_rating"); ratingStr != "" {
		if r, err := strconv.ParseFloat(ratingStr, 64); err == nil {
			params.MinRating = &r
		}
	}

	if verifiedStr := c.Query("verified"); verifiedStr == "true" {
		v := true
		params.IsVerified = &v
	}

	// Geo search params
	if latStr := c.Query("lat"); latStr != "" {
		if lat, err := strconv.ParseFloat(latStr, 64); err == nil {
			params.Latitude = &lat
		}
	}
	if lonStr := c.Query("lon"); lonStr != "" {
		if lon, err := strconv.ParseFloat(lonStr, 64); err == nil {
			params.Longitude = &lon
		}
	}
	if radiusStr := c.Query("radius"); radiusStr != "" {
		if r, err := strconv.ParseFloat(radiusStr, 64); err == nil {
			params.RadiusKm = &r
		}
	}

	// Pagination
	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil {
			params.Page = p
		}
	}
	if sizeStr := c.Query("page_size"); sizeStr != "" {
		if s, err := strconv.Atoi(sizeStr); err == nil {
			params.PageSize = s
		}
	}

	// Defaults
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.PageSize <= 0 {
		params.PageSize = 10
	}

	return params
}

func (h *Handler) sendSearchResponse(c *gin.Context, salons []domain.Salon, total int, params domain.SalonSearchParams, source string) {
	totalPages := (total + params.PageSize - 1) / params.PageSize

	c.JSON(http.StatusOK, SearchResponse{
		Data:       salons,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
		Source:     source,
	})
}
