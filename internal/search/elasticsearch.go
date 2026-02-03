package search

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"beauty-salons/internal/domain"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
)

// ===========================================
// ELASTICSEARCH SEARCH ENGINE
// ===========================================
// - Full-text search with relevance scoring
// - Fuzzy matching (typo tolerance)
// - Faceted search (aggregations)
// - Geo-spatial queries

const (
	SalonIndex = "salons"
)

// ElasticsearchClient wraps the Elasticsearch client
type ElasticsearchClient struct {
	client *elasticsearch.Client
}

// NewElasticsearchClient creates a new Elasticsearch connection
func NewElasticsearchClient(addresses []string) (*ElasticsearchClient, error) {
	cfg := elasticsearch.Config{
		Addresses: addresses,
	}

	client, err := elasticsearch.NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create elasticsearch client: %w", err)
	}

	// Test the connection
	res, err := client.Info()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to elasticsearch: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, fmt.Errorf("elasticsearch error: %s", res.String())
	}

	log.Println("Connected to Elasticsearch cluster")
	return &ElasticsearchClient{client: client}, nil
}

// CreateIndex creates the salons index with proper mappings.
// The mapping defines HOW each field is indexed and searched.
func (es *ElasticsearchClient) CreateIndex(ctx context.Context) error {
	// Check if index already exists
	res, err := es.client.Indices.Exists([]string{SalonIndex})
	if err != nil {
		return fmt.Errorf("failed to check index existence: %w", err)
	}
	defer res.Body.Close()

	if res.StatusCode == 200 {
		log.Printf("Index %s already exists", SalonIndex)
		return nil
	}

	// Define the index mapping
	mapping := map[string]interface{}{
		"settings": map[string]interface{}{
			// NUMBER OF SHARDS: How data is distributed
			"number_of_shards": 1,
			// NUMBER OF REPLICAS:
			"number_of_replicas": 0,
			"analysis": map[string]interface{}{
				"analyzer": map[string]interface{}{
					// Custom analyzer for Spanish text
					"spanish_analyzer": map[string]interface{}{
						"type":      "custom",
						"tokenizer": "standard",
						"filter":    []string{"lowercase", "spanish_stemmer", "asciifolding"},
					},
				},
				"filter": map[string]interface{}{
					"spanish_stemmer": map[string]interface{}{
						"type":     "stemmer",
						"language": "spanish",
					},
				},
			},
		},
		"mappings": map[string]interface{}{
			"properties": map[string]interface{}{
				// Text fields are analyzed for full-text search
				"name": map[string]interface{}{
					"type":     "text",
					"analyzer": "spanish_analyzer",
					"fields": map[string]interface{}{
						// Also store as keyword for exact matching & sorting
						"keyword": map[string]interface{}{
							"type": "keyword",
						},
					},
				},
				"description": map[string]interface{}{
					"type":     "text",
					"analyzer": "spanish_analyzer",
				},
				// Keyword fields are for exact matches and aggregations
				"slug": map[string]interface{}{
					"type": "keyword",
				},
				"city": map[string]interface{}{
					"type": "keyword",
				},
				"state": map[string]interface{}{
					"type": "keyword",
				},
				"country": map[string]interface{}{
					"type": "keyword",
				},
				"category_name": map[string]interface{}{
					"type": "keyword",
				},
				// Numeric fields
				"price_range": map[string]interface{}{
					"type": "integer",
				},
				"rating": map[string]interface{}{
					"type": "float",
				},
				"review_count": map[string]interface{}{
					"type": "integer",
				},
				// Boolean fields
				"is_active": map[string]interface{}{
					"type": "boolean",
				},
				"is_verified": map[string]interface{}{
					"type": "boolean",
				},
				// Geo-point for location-based search!
				"location": map[string]interface{}{
					"type": "geo_point",
				},
				// Nested array of services
				"services": map[string]interface{}{
					"type": "nested",
					"properties": map[string]interface{}{
						"name": map[string]interface{}{
							"type":     "text",
							"analyzer": "spanish_analyzer",
						},
						"price_min": map[string]interface{}{
							"type": "float",
						},
						"price_max": map[string]interface{}{
							"type": "float",
						},
					},
				},
				// Array of amenity names
				"amenities": map[string]interface{}{
					"type": "keyword",
				},
			},
		},
	}

	body, _ := json.Marshal(mapping)
	res, err = es.client.Indices.Create(
		SalonIndex,
		es.client.Indices.Create.WithBody(bytes.NewReader(body)),
		es.client.Indices.Create.WithContext(ctx),
	)
	if err != nil {
		return fmt.Errorf("failed to create index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to create index: %s", res.String())
	}

	log.Printf("Created index %s", SalonIndex)
	return nil
}

// IndexSalon indexes a single salon document
func (es *ElasticsearchClient) IndexSalon(ctx context.Context, salon *domain.Salon) error {
	// Transform to ES document format
	doc := es.salonToDocument(salon)

	body, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("failed to marshal salon: %w", err)
	}

	req := esapi.IndexRequest{
		Index:      SalonIndex,
		DocumentID: fmt.Sprintf("%d", salon.ID),
		Body:       bytes.NewReader(body),
		Refresh:    "true", // Make immediately searchable (slower)
	}

	res, err := req.Do(ctx, es.client)
	if err != nil {
		return fmt.Errorf("failed to index salon: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("failed to index salon: %s", res.String())
	}

	return nil
}

// BulkIndexSalons indexes multiple salons at once
func (es *ElasticsearchClient) BulkIndexSalons(ctx context.Context, salons []domain.Salon) error {
	if len(salons) == 0 {
		return nil
	}

	var buf bytes.Buffer
	for _, salon := range salons {
		// Action line
		meta := map[string]interface{}{
			"index": map[string]interface{}{
				"_index": SalonIndex,
				"_id":    fmt.Sprintf("%d", salon.ID),
			},
		}
		metaBytes, _ := json.Marshal(meta)
		buf.Write(metaBytes)
		buf.WriteByte('\n')

		// Document line
		doc := es.salonToDocument(&salon)
		docBytes, _ := json.Marshal(doc)
		buf.Write(docBytes)
		buf.WriteByte('\n')
	}

	res, err := es.client.Bulk(
		bytes.NewReader(buf.Bytes()),
		es.client.Bulk.WithContext(ctx),
		es.client.Bulk.WithRefresh("true"),
	)
	if err != nil {
		return fmt.Errorf("failed to bulk index: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return fmt.Errorf("bulk index error: %s", res.String())
	}

	log.Printf("Indexed %d salons", len(salons))
	return nil
}

// Search performs a search query against Elasticsearch
func (es *ElasticsearchClient) Search(ctx context.Context, params domain.SalonSearchParams) ([]domain.Salon, int, error) {
	// Build the query
	query := es.buildQuery(params)

	body, _ := json.Marshal(query)

	res, err := es.client.Search(
		es.client.Search.WithContext(ctx),
		es.client.Search.WithIndex(SalonIndex),
		es.client.Search.WithBody(bytes.NewReader(body)),
	)
	if err != nil {
		return nil, 0, fmt.Errorf("search failed: %w", err)
	}
	defer res.Body.Close()

	if res.IsError() {
		return nil, 0, fmt.Errorf("search error: %s", res.String())
	}

	// Parse response
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		return nil, 0, fmt.Errorf("failed to parse response: %w", err)
	}

	hits := result["hits"].(map[string]interface{})
	total := int(hits["total"].(map[string]interface{})["value"].(float64))

	hitsList := hits["hits"].([]interface{})
	salons := make([]domain.Salon, 0, len(hitsList))

	for _, hit := range hitsList {
		hitMap := hit.(map[string]interface{})
		source := hitMap["_source"].(map[string]interface{})

		salon := es.documentToSalon(source)
		salons = append(salons, salon)
	}

	return salons, total, nil
}

// GetClusterHealth returns cluster health information
func (es *ElasticsearchClient) GetClusterHealth(ctx context.Context) (map[string]interface{}, error) {
	res, err := es.client.Cluster.Health(
		es.client.Cluster.Health.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var health map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&health); err != nil {
		return nil, err
	}

	return health, nil
}

// GetIndexStats returns index statistics
func (es *ElasticsearchClient) GetIndexStats(ctx context.Context) (map[string]interface{}, error) {
	res, err := es.client.Indices.Stats(
		es.client.Indices.Stats.WithIndex(SalonIndex),
		es.client.Indices.Stats.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	var stats map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&stats); err != nil {
		return nil, err
	}

	return stats, nil
}

// buildQuery constructs an Elasticsearch query from search params
func (es *ElasticsearchClient) buildQuery(params domain.SalonSearchParams) map[string]interface{} {
	must := []map[string]interface{}{}
	filter := []map[string]interface{}{}

	// Full-text search with multi_match
	if params.Query != "" {
		must = append(must, map[string]interface{}{
			"multi_match": map[string]interface{}{
				"query":     params.Query,
				"fields":    []string{"name^3", "description", "services.name"},
				"fuzziness": "AUTO", // Typo tolerance!
			},
		})
	}

	// Filters
	filter = append(filter, map[string]interface{}{
		"term": map[string]interface{}{
			"is_active": true,
		},
	})

	if params.City != "" {
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{
				"city": params.City,
			},
		})
	}

	if params.CategoryID != nil {
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{
				"category_id": *params.CategoryID,
			},
		})
	}

	if params.PriceRange != 0 {
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{
				"price_range": params.PriceRange,
			},
		})
	}

	if params.MinRating != nil {
		filter = append(filter, map[string]interface{}{
			"range": map[string]interface{}{
				"rating": map[string]interface{}{
					"gte": *params.MinRating,
				},
			},
		})
	}

	if params.IsVerified != nil && *params.IsVerified {
		filter = append(filter, map[string]interface{}{
			"term": map[string]interface{}{
				"is_verified": true,
			},
		})
	}

	// Geo-distance filter
	if params.Location != nil && params.RadiusKm != nil {
		filter = append(filter, map[string]interface{}{
			"geo_distance": map[string]interface{}{
				"distance": fmt.Sprintf("%fkm", *params.RadiusKm),
				"location": map[string]interface{}{
					"lat": params.Location.Latitude,
					"lon": params.Location.Longitude,
				},
			},
		})
	}

	// If no text query, match all
	if len(must) == 0 {
		must = append(must, map[string]interface{}{
			"match_all": map[string]interface{}{},
		})
	}

	// Pagination
	pageSize := params.PageSize
	if pageSize <= 0 {
		pageSize = 10
	}
	page := params.Page
	if page <= 0 {
		page = 1
	}
	from := (page - 1) * pageSize

	// Build sort
	sort := []map[string]interface{}{}
	switch params.SortBy {
	case domain.SortByRating:
		sort = append(sort, map[string]interface{}{"rating": map[string]interface{}{"order": "desc", "missing": "_last"}})
	case domain.SortByReviews:
		sort = append(sort, map[string]interface{}{"review_count": map[string]interface{}{"order": "desc"}})
	case domain.SortByNewest:
		sort = append(sort, map[string]interface{}{"created_at": map[string]interface{}{"order": "desc"}})
	case domain.SortByDistance:
		if params.Location != nil {
			sort = append(sort, map[string]interface{}{
				"_geo_distance": map[string]interface{}{
					"location": map[string]interface{}{
						"lat": params.Location.Latitude,
						"lon": params.Location.Longitude,
					},
					"order": "asc",
					"unit":  "km",
				},
			})
		}
	default:
		sort = append(sort, map[string]interface{}{"_score": "desc"})
		sort = append(sort, map[string]interface{}{"rating": map[string]interface{}{"order": "desc", "missing": "_last"}})
	}

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": map[string]interface{}{
				"must":   must,
				"filter": filter,
			},
		},
		"sort": sort,
		"from": from,
		"size": pageSize,
	}
}

// salonToDocument converts a Salon to an ES document
func (es *ElasticsearchClient) salonToDocument(salon *domain.Salon) map[string]interface{} {
	doc := map[string]interface{}{
		"id":           salon.ID,
		"name":         salon.Name,
		"slug":         salon.Slug,
		"description":  salon.Description,
		"address":      salon.Location.Address,
		"city":         salon.Location.City,
		"state":        salon.Location.State,
		"country":      salon.Location.Country,
		"phone":        salon.Contact.Phone,
		"email":        salon.Contact.Email,
		"website":      salon.Contact.Website,
		"category_id":  salon.CategoryID,
		"price_range":  salon.PriceRange,
		"rating":       salon.Rating,
		"review_count": salon.ReviewCount,
		"is_active":    salon.IsActive,
		"is_verified":  salon.IsVerified,
	}

	// Add geo-point if coordinates exist
	if salon.Location.GeoPoint != nil {
		doc["location"] = map[string]interface{}{
			"lat": salon.Location.GeoPoint.Latitude,
			"lon": salon.Location.GeoPoint.Longitude,
		}
	}

	if salon.Category != nil {
		doc["category_name"] = salon.Category.Name
	}

	if len(salon.Services) > 0 {
		services := make([]map[string]interface{}, len(salon.Services))
		for i, s := range salon.Services {
			services[i] = map[string]interface{}{
				"name":      s.Name,
				"price_min": s.PriceMin,
				"price_max": s.PriceMax,
			}
		}
		doc["services"] = services
	}

	if len(salon.Amenities) > 0 {
		amenityNames := make([]string, len(salon.Amenities))
		for i, a := range salon.Amenities {
			amenityNames[i] = a.Name
		}
		doc["amenities"] = amenityNames
	}

	return doc
}

// documentToSalon converts an ES document back to a Salon
func (es *ElasticsearchClient) documentToSalon(doc map[string]interface{}) domain.Salon {
	salon := domain.Salon{}

	if v, ok := doc["id"].(float64); ok {
		salon.ID = int64(v)
	}
	if v, ok := doc["name"].(string); ok {
		salon.Name = v
	}
	if v, ok := doc["slug"].(string); ok {
		salon.Slug = v
	}
	if v, ok := doc["description"].(string); ok {
		salon.Description = &v
	}

	// Location fields
	if v, ok := doc["address"].(string); ok {
		salon.Location.Address = v
	}
	if v, ok := doc["city"].(string); ok {
		salon.Location.City = v
	}
	if v, ok := doc["state"].(string); ok {
		salon.Location.State = v
	}
	if v, ok := doc["country"].(string); ok {
		salon.Location.Country = v
	}
	if loc, ok := doc["location"].(map[string]interface{}); ok {
		if lat, ok := loc["lat"].(float64); ok {
			if lon, ok := loc["lon"].(float64); ok {
				salon.Location.GeoPoint = &domain.GeoPoint{
					Latitude:  lat,
					Longitude: lon,
				}
			}
		}
	}

	// Contact fields
	if v, ok := doc["phone"].(string); ok {
		salon.Contact.Phone = v
	}
	if v, ok := doc["email"].(string); ok {
		salon.Contact.Email = v
	}
	if v, ok := doc["website"].(string); ok {
		salon.Contact.Website = v
	}

	// Business info
	if v, ok := doc["rating"].(float64); ok {
		salon.Rating = &v
	}
	if v, ok := doc["review_count"].(float64); ok {
		salon.ReviewCount = int(v)
	}
	if v, ok := doc["price_range"].(float64); ok {
		salon.PriceRange = domain.PriceRange(int(v))
	}
	if v, ok := doc["is_verified"].(bool); ok {
		salon.IsVerified = v
	}
	if v, ok := doc["is_active"].(bool); ok {
		salon.IsActive = v
	}
	if v, ok := doc["category_name"].(string); ok {
		salon.Category = &domain.Category{Name: v}
	}
	if v, ok := doc["category_id"].(float64); ok {
		catID := int64(v)
		salon.CategoryID = &catID
	}

	// Handle amenities array
	if v, ok := doc["amenities"].([]interface{}); ok {
		for _, a := range v {
			if s, ok := a.(string); ok {
				salon.Amenities = append(salon.Amenities, domain.Amenity{Name: s})
			}
		}
	}

	return salon
}

// DeleteIndex removes the index
func (es *ElasticsearchClient) DeleteIndex(ctx context.Context) error {
	res, err := es.client.Indices.Delete(
		[]string{SalonIndex},
		es.client.Indices.Delete.WithContext(ctx),
	)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.IsError() && !strings.Contains(res.String(), "index_not_found") {
		return fmt.Errorf("failed to delete index: %s", res.String())
	}

	return nil
}
