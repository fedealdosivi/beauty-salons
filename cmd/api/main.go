package main

import (
	"context"
	"log"
	"os"

	"beauty-salons/internal/api/handlers"
	"beauty-salons/internal/repository"
	"beauty-salons/internal/search"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file if it exists
	godotenv.Load()

	// Get configuration from environment
	pgConn := getEnv("DATABASE_URL", "postgres://beauty:beauty123@localhost:5432/beauty_salons?sslmode=disable")
	esURL := getEnv("ELASTICSEARCH_URL", "http://localhost:9200")
	port := getEnv("PORT", "8080")

	log.Println("===========================================")
	log.Println("Beauty Salons Search API")
	log.Println("===========================================")

	// Connect to PostgreSQL (Source of Truth)
	log.Println("Connecting to PostgreSQL...")
	repo, err := repository.NewPostgresRepository(pgConn)
	if err != nil {
		log.Fatalf("Failed to connect to PostgreSQL: %v", err)
	}
	defer repo.Close()
	log.Println("✓ Connected to PostgreSQL")

	// Connect to Elasticsearch (Search Cluster)
	log.Println("Connecting to Elasticsearch...")
	esClient, err := search.NewElasticsearchClient([]string{esURL})
	if err != nil {
		log.Fatalf("Failed to connect to Elasticsearch: %v", err)
	}
	log.Println("✓ Connected to Elasticsearch")

	// Create the search index if it doesn't exist
	if err := esClient.CreateIndex(context.Background()); err != nil {
		log.Printf("Warning: Could not create index: %v", err)
	}

	// Set up HTTP handlers
	handler := handlers.NewHandler(repo, esClient)

	// Set up Gin router
	r := gin.Default()

	// API routes
	v1 := r.Group("/api/v1")
	{
		// Search endpoints
		v1.GET("/search", handler.SearchSalons)                // Elasticsearch search
		v1.GET("/search/postgres", handler.SearchSalonsPostgres) // PostgreSQL search (for comparison)

		// Resource endpoints
		v1.GET("/salons/:id", handler.GetSalon)
		v1.GET("/categories", handler.GetCategories)

		// Admin endpoints (for learning/testing)
		admin := v1.Group("/admin")
		{
			admin.POST("/sync", handler.SyncToElasticsearch)        // Sync data to ES
			admin.GET("/cluster/health", handler.GetClusterHealth)  // ES cluster health
			admin.GET("/cluster/stats", handler.GetIndexStats)      // ES index stats
		}
	}

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// Print usage info
	log.Println("")
	log.Println("Available endpoints:")
	log.Println("  GET  /api/v1/search          - Search salons (Elasticsearch)")
	log.Println("  GET  /api/v1/search/postgres - Search salons (PostgreSQL)")
	log.Println("  GET  /api/v1/salons/:id      - Get salon by ID")
	log.Println("  GET  /api/v1/categories      - List categories")
	log.Println("  POST /api/v1/admin/sync      - Sync data to Elasticsearch")
	log.Println("  GET  /api/v1/admin/cluster/health - ES cluster health")
	log.Println("  GET  /api/v1/admin/cluster/stats  - ES index stats")
	log.Println("")
	log.Printf("Starting server on :%s", port)
	log.Println("===========================================")

	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
