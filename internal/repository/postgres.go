package repository

import (
	"context"
	"fmt"

	"beauty-salons/internal/domain"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// PostgresRepository handles all database operations.
type PostgresRepository struct {
	db *sqlx.DB
}

// NewPostgresRepository creates a new PostgreSQL connection.
func NewPostgresRepository(connectionString string) (*PostgresRepository, error) {
	db, err := sqlx.Connect("postgres", connectionString)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	return &PostgresRepository{db: db}, nil
}

// Close closes the database connection
func (r *PostgresRepository) Close() error {
	return r.db.Close()
}

// GetAllSalons retrieves all salons (used for initial sync to Elasticsearch)
func (r *PostgresRepository) GetAllSalons(ctx context.Context) ([]domain.Salon, error) {
	query := `
		SELECT
			s.*,
			c.name as category_name
		FROM salons s
		LEFT JOIN categories c ON s.category_id = c.id
		WHERE s.is_active = true
		ORDER BY s.id
	`

	var salons []domain.Salon
	if err := r.db.SelectContext(ctx, &salons, query); err != nil {
		return nil, fmt.Errorf("failed to get salons: %w", err)
	}

	return salons, nil
}

// GetSalonByID retrieves a single salon by ID
func (r *PostgresRepository) GetSalonByID(ctx context.Context, id int64) (*domain.Salon, error) {
	query := `
		SELECT
			s.*,
			c.name as category_name
		FROM salons s
		LEFT JOIN categories c ON s.category_id = c.id
		WHERE s.id = $1
	`

	var salon domain.Salon
	if err := r.db.GetContext(ctx, &salon, query, id); err != nil {
		return nil, fmt.Errorf("failed to get salon: %w", err)
	}

	// Get services for this salon
	servicesQuery := `SELECT * FROM services WHERE salon_id = $1`
	if err := r.db.SelectContext(ctx, &salon.Services, servicesQuery, id); err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	// Get amenities for this salon
	amenitiesQuery := `
		SELECT a.name
		FROM amenities a
		JOIN salon_amenities sa ON a.id = sa.amenity_id
		WHERE sa.salon_id = $1
	`
	if err := r.db.SelectContext(ctx, &salon.Amenities, amenitiesQuery, id); err != nil {
		return nil, fmt.Errorf("failed to get amenities: %w", err)
	}

	return &salon, nil
}

// SearchSalons performs a search using PostgreSQL's full-text search.
func (r *PostgresRepository) SearchSalons(ctx context.Context, params domain.SalonSearchParams) ([]domain.Salon, int, error) {
	// Base query with full-text search
	query := `
		SELECT
			s.*,
			c.name as category_name,
			COUNT(*) OVER() as total_count
		FROM salons s
		LEFT JOIN categories c ON s.category_id = c.id
		WHERE s.is_active = true
	`

	args := []interface{}{}
	argNum := 1

	// Full-text search using PostgreSQL's to_tsvector
	if params.Query != "" {
		query += fmt.Sprintf(` AND to_tsvector('spanish', coalesce(s.name, '') || ' ' || coalesce(s.description, '')) @@ plainto_tsquery('spanish', $%d)`, argNum)
		args = append(args, params.Query)
		argNum++
	}

	// City filter
	if params.City != "" {
		query += fmt.Sprintf(` AND LOWER(s.city) = LOWER($%d)`, argNum)
		args = append(args, params.City)
		argNum++
	}

	// Category filter
	if params.CategoryID != nil {
		query += fmt.Sprintf(` AND s.category_id = $%d`, argNum)
		args = append(args, *params.CategoryID)
		argNum++
	}

	// Price range filter
	if params.PriceRange != nil {
		query += fmt.Sprintf(` AND s.price_range = $%d`, argNum)
		args = append(args, *params.PriceRange)
		argNum++
	}

	// Rating filter
	if params.MinRating != nil {
		query += fmt.Sprintf(` AND s.rating >= $%d`, argNum)
		args = append(args, *params.MinRating)
		argNum++
	}

	// Verified filter
	if params.IsVerified != nil && *params.IsVerified {
		query += ` AND s.is_verified = true`
	}

	// Order by rating
	query += ` ORDER BY s.rating DESC NULLS LAST, s.review_count DESC`

	// Pagination
	if params.PageSize <= 0 {
		params.PageSize = 10
	}
	if params.Page <= 0 {
		params.Page = 1
	}
	offset := (params.Page - 1) * params.PageSize

	query += fmt.Sprintf(` LIMIT $%d OFFSET $%d`, argNum, argNum+1)
	args = append(args, params.PageSize, offset)

	var results []struct {
		domain.Salon
		TotalCount int `db:"total_count"`
	}

	if err := r.db.SelectContext(ctx, &results, query, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to search salons: %w", err)
	}

	salons := make([]domain.Salon, len(results))
	totalCount := 0
	for i, r := range results {
		salons[i] = r.Salon
		totalCount = r.TotalCount
	}

	return salons, totalCount, nil
}

// GetCategories retrieves all categories
func (r *PostgresRepository) GetCategories(ctx context.Context) ([]domain.Category, error) {
	var categories []domain.Category
	if err := r.db.SelectContext(ctx, &categories, "SELECT * FROM categories ORDER BY name"); err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}
	return categories, nil
}
