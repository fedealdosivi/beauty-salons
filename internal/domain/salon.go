package domain

import "time"

// Salon represents a beauty salon/shop in our system.
// This struct is used both for database queries and API responses.
type Salon struct {
	ID          int64   `json:"id" db:"id"`
	Name        string  `json:"name" db:"name"`
	Slug        string  `json:"slug" db:"slug"`
	Description *string `json:"description,omitempty" db:"description"`

	// Location
	Address    *string  `json:"address,omitempty" db:"address"`
	City       *string  `json:"city,omitempty" db:"city"`
	State      *string  `json:"state,omitempty" db:"state"`
	PostalCode *string  `json:"postal_code,omitempty" db:"postal_code"`
	Country    *string  `json:"country,omitempty" db:"country"`
	Latitude   *float64 `json:"latitude,omitempty" db:"latitude"`
	Longitude  *float64 `json:"longitude,omitempty" db:"longitude"`

	// Contact
	Phone   *string `json:"phone,omitempty" db:"phone"`
	Email   *string `json:"email,omitempty" db:"email"`
	Website *string `json:"website,omitempty" db:"website"`

	// Business Info
	CategoryID  *int64   `json:"category_id,omitempty" db:"category_id"`
	PriceRange  *int     `json:"price_range,omitempty" db:"price_range"`
	Rating      *float64 `json:"rating,omitempty" db:"rating"`
	ReviewCount *int     `json:"review_count,omitempty" db:"review_count"`

	// Status
	IsActive   bool `json:"is_active" db:"is_active"`
	IsVerified bool `json:"is_verified" db:"is_verified"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Joined data (not always populated)
	CategoryName *string   `json:"category_name,omitempty" db:"category_name"`
	Services     []Service `json:"services,omitempty"`
	Amenities    []string  `json:"amenities,omitempty"`
}

// Category represents a type of beauty business
type Category struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Slug      string    `json:"slug" db:"slug"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Service represents a service offered by a salon
type Service struct {
	ID              int64    `json:"id" db:"id"`
	SalonID         int64    `json:"salon_id" db:"salon_id"`
	Name            string   `json:"name" db:"name"`
	Description     *string  `json:"description,omitempty" db:"description"`
	PriceMin        *float64 `json:"price_min,omitempty" db:"price_min"`
	PriceMax        *float64 `json:"price_max,omitempty" db:"price_max"`
	DurationMinutes *int     `json:"duration_minutes,omitempty" db:"duration_minutes"`
}

// SalonSearchParams contains all possible search/filter parameters
type SalonSearchParams struct {
	Query      string   // Full-text search query
	City       string   // Filter by city
	CategoryID *int64   // Filter by category
	PriceRange *int     // Filter by price range (1-4)
	MinRating  *float64 // Minimum rating filter
	IsVerified *bool    // Filter verified only
	Latitude   *float64 // For geo-search
	Longitude  *float64 // For geo-search
	RadiusKm   *float64 // Radius for geo-search
	Page       int      // Pagination
	PageSize   int      // Results per page
}
