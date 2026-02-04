package repository

import (
	"context"
	"fmt"

	"beauty-salons/internal/domain"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

// salonRow represents a salon as stored in the database (flat structure)
type salonRow struct {
	ID          int64    `db:"id"`
	Name        string   `db:"name"`
	Slug        string   `db:"slug"`
	Description *string  `db:"description"`
	Address     *string  `db:"address"`
	City        *string  `db:"city"`
	State       *string  `db:"state"`
	PostalCode  *string  `db:"postal_code"`
	Country     *string  `db:"country"`
	Latitude    *float64 `db:"latitude"`
	Longitude   *float64 `db:"longitude"`
	Phone       *string  `db:"phone"`
	Email       *string  `db:"email"`
	Website     *string  `db:"website"`
	CategoryID  *int64   `db:"category_id"`
	PriceRange  *int     `db:"price_range"`
	Rating      *float64 `db:"rating"`
	ReviewCount *int     `db:"review_count"`
	IsActive    bool     `db:"is_active"`
	IsVerified  bool     `db:"is_verified"`
	CreatedAt   string   `db:"created_at"`
	UpdatedAt   string   `db:"updated_at"`

	// Joined fields
	CategoryName *string `db:"category_name"`
	TotalCount   int     `db:"total_count"`
}

// toDomain converts a database row to a domain Salon
func (r *salonRow) toDomain() domain.Salon {
	salon := domain.Salon{
		ID:          r.ID,
		Name:        r.Name,
		Slug:        r.Slug,
		Description: r.Description,
		CategoryID:  r.CategoryID,
		IsActive:    r.IsActive,
		IsVerified:  r.IsVerified,
	}

	// Map Location
	if r.Address != nil {
		salon.Location.Address = *r.Address
	}
	if r.City != nil {
		salon.Location.City = *r.City
	}
	if r.State != nil {
		salon.Location.State = *r.State
	}
	if r.PostalCode != nil {
		salon.Location.PostalCode = *r.PostalCode
	}
	if r.Country != nil {
		salon.Location.Country = *r.Country
	}
	if r.Latitude != nil && r.Longitude != nil {
		salon.Location.GeoPoint = &domain.GeoPoint{
			Latitude:  *r.Latitude,
			Longitude: *r.Longitude,
		}
	}

	// Map Contact
	if r.Phone != nil {
		salon.Contact.Phone = *r.Phone
	}
	if r.Email != nil {
		salon.Contact.Email = *r.Email
	}
	if r.Website != nil {
		salon.Contact.Website = *r.Website
	}

	// Map business info
	if r.PriceRange != nil {
		salon.PriceRange = domain.PriceRange(*r.PriceRange)
	}
	salon.Rating = r.Rating
	if r.ReviewCount != nil {
		salon.ReviewCount = *r.ReviewCount
	}

	// Map category if joined
	if r.CategoryName != nil {
		salon.Category = &domain.Category{
			ID:   *r.CategoryID,
			Name: *r.CategoryName,
		}
	}

	return salon
}

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
			s.id, s.name, s.slug, s.description,
			s.address, s.city, s.state, s.postal_code, s.country,
			s.latitude, s.longitude,
			s.phone, s.email, s.website,
			s.category_id, s.price_range, s.rating, s.review_count,
			s.is_active, s.is_verified, s.created_at, s.updated_at,
			c.name as category_name,
			0 as total_count
		FROM salons s
		LEFT JOIN categories c ON s.category_id = c.id
		WHERE s.is_active = true
		ORDER BY s.id
	`

	var rows []salonRow
	if err := r.db.SelectContext(ctx, &rows, query); err != nil {
		return nil, fmt.Errorf("failed to get salons: %w", err)
	}

	salons := make([]domain.Salon, len(rows))
	for i, row := range rows {
		salons[i] = row.toDomain()
	}

	return salons, nil
}

// GetSalonByID retrieves a single salon by ID
func (r *PostgresRepository) GetSalonByID(ctx context.Context, id int64) (*domain.Salon, error) {
	query := `
		SELECT
			s.id, s.name, s.slug, s.description,
			s.address, s.city, s.state, s.postal_code, s.country,
			s.latitude, s.longitude,
			s.phone, s.email, s.website,
			s.category_id, s.price_range, s.rating, s.review_count,
			s.is_active, s.is_verified, s.created_at, s.updated_at,
			c.name as category_name,
			0 as total_count
		FROM salons s
		LEFT JOIN categories c ON s.category_id = c.id
		WHERE s.id = $1
	`

	var row salonRow
	if err := r.db.GetContext(ctx, &row, query, id); err != nil {
		return nil, fmt.Errorf("failed to get salon: %w", err)
	}

	salon := row.toDomain()

	// Get services for this salon
	servicesQuery := `SELECT id, salon_id, name, description, price_min, price_max, duration_minutes, created_at FROM services WHERE salon_id = $1`
	if err := r.db.SelectContext(ctx, &salon.Services, servicesQuery, id); err != nil {
		return nil, fmt.Errorf("failed to get services: %w", err)
	}

	// Get amenities for this salon
	amenitiesQuery := `
		SELECT a.id, a.name, a.icon
		FROM amenities a
		JOIN salon_amenities sa ON a.id = sa.amenity_id
		WHERE sa.salon_id = $1
	`
	if err := r.db.SelectContext(ctx, &salon.Amenities, amenitiesQuery, id); err != nil {
		return nil, fmt.Errorf("failed to get amenities: %w", err)
	}

	// Get operating hours for this salon
	hoursQuery := `
		SELECT id, salon_id, day_of_week, open_time, close_time, is_closed
		FROM operating_hours
		WHERE salon_id = $1
		ORDER BY day_of_week
	`
	if err := r.db.SelectContext(ctx, &salon.OperatingHours, hoursQuery, id); err != nil {
		return nil, fmt.Errorf("failed to get operating hours: %w", err)
	}

	return &salon, nil
}

// SearchSalons performs a search using PostgreSQL's full-text search.
func (r *PostgresRepository) SearchSalons(ctx context.Context, params domain.SalonSearchParams) ([]domain.Salon, int, error) {
	// Base query with full-text search
	query := `
		SELECT
			s.id, s.name, s.slug, s.description,
			s.address, s.city, s.state, s.postal_code, s.country,
			s.latitude, s.longitude,
			s.phone, s.email, s.website,
			s.category_id, s.price_range, s.rating, s.review_count,
			s.is_active, s.is_verified, s.created_at, s.updated_at,
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
	if params.PriceRange != 0 {
		query += fmt.Sprintf(` AND s.price_range = $%d`, argNum)
		args = append(args, params.PriceRange)
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

	// Geo search (within radius)
	if params.Location != nil && params.RadiusKm != nil {
		// Using Haversine formula approximation
		query += fmt.Sprintf(` AND (
			6371 * acos(
				cos(radians($%d)) * cos(radians(s.latitude)) *
				cos(radians(s.longitude) - radians($%d)) +
				sin(radians($%d)) * sin(radians(s.latitude))
			)
		) <= $%d`, argNum, argNum+1, argNum+2, argNum+3)
		args = append(args, params.Location.Latitude, params.Location.Longitude, params.Location.Latitude, *params.RadiusKm)
		argNum += 4
	}

	// Order by
	switch params.SortBy {
	case domain.SortByRating:
		query += ` ORDER BY s.rating DESC NULLS LAST, s.review_count DESC`
	case domain.SortByReviews:
		query += ` ORDER BY s.review_count DESC, s.rating DESC NULLS LAST`
	case domain.SortByNewest:
		query += ` ORDER BY s.created_at DESC`
	case domain.SortByDistance:
		if params.Location != nil {
			query += fmt.Sprintf(` ORDER BY (
				6371 * acos(
					cos(radians($%d)) * cos(radians(s.latitude)) *
					cos(radians(s.longitude) - radians($%d)) +
					sin(radians($%d)) * sin(radians(s.latitude))
				)
			) ASC`, argNum, argNum+1, argNum+2)
			args = append(args, params.Location.Latitude, params.Location.Longitude, params.Location.Latitude)
			argNum += 3
		} else {
			query += ` ORDER BY s.rating DESC NULLS LAST`
		}
	default:
		// Weighted ranking: rating*2 + log(1+reviews)*1.5 + verified bonus
		query += ` ORDER BY (
			COALESCE(s.rating, 0) * 2.0
			+ LN(1 + COALESCE(s.review_count, 0)) * 1.5
			+ CASE WHEN s.is_verified THEN 5.0 ELSE 0.0 END
		) DESC`
	}

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

	var rows []salonRow
	if err := r.db.SelectContext(ctx, &rows, query, args...); err != nil {
		return nil, 0, fmt.Errorf("failed to search salons: %w", err)
	}

	salons := make([]domain.Salon, len(rows))
	totalCount := 0
	for i, row := range rows {
		salons[i] = row.toDomain()
		totalCount = row.TotalCount
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
