package domain

import (
	"errors"
	"fmt"
	"math"
	"strings"
	"time"
)

// ===========================================
// Value Objects
// ===========================================

// GeoPoint represents a geographic coordinate
type GeoPoint struct {
	Latitude  float64 `json:"lat" db:"latitude"`
	Longitude float64 `json:"lon" db:"longitude"`
}

// DistanceTo calculates the distance in kilometers to another point using Haversine formula
func (g GeoPoint) DistanceTo(other GeoPoint) float64 {
	const earthRadiusKm = 6371.0

	lat1Rad := g.Latitude * math.Pi / 180
	lat2Rad := other.Latitude * math.Pi / 180
	deltaLat := (other.Latitude - g.Latitude) * math.Pi / 180
	deltaLon := (other.Longitude - g.Longitude) * math.Pi / 180

	a := math.Sin(deltaLat/2)*math.Sin(deltaLat/2) +
		math.Cos(lat1Rad)*math.Cos(lat2Rad)*math.Sin(deltaLon/2)*math.Sin(deltaLon/2)
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))

	return earthRadiusKm * c
}

// IsValid checks if the coordinates are within valid ranges
func (g GeoPoint) IsValid() bool {
	return g.Latitude >= -90 && g.Latitude <= 90 &&
		g.Longitude >= -180 && g.Longitude <= 180
}

// Location represents a physical address
type Location struct {
	Address    string    `json:"address,omitempty" db:"address"`
	City       string    `json:"city,omitempty" db:"city"`
	State      string    `json:"state,omitempty" db:"state"`
	PostalCode string    `json:"postal_code,omitempty" db:"postal_code"`
	Country    string    `json:"country,omitempty" db:"country"`
	GeoPoint   *GeoPoint `json:"geo_point,omitempty"`
}

// FullAddress returns a formatted complete address
func (l Location) FullAddress() string {
	parts := []string{}
	if l.Address != "" {
		parts = append(parts, l.Address)
	}
	if l.City != "" {
		parts = append(parts, l.City)
	}
	if l.State != "" {
		parts = append(parts, l.State)
	}
	if l.PostalCode != "" {
		parts = append(parts, l.PostalCode)
	}
	if l.Country != "" {
		parts = append(parts, l.Country)
	}
	return strings.Join(parts, ", ")
}

// Contact represents contact information
type Contact struct {
	Phone   string `json:"phone,omitempty" db:"phone"`
	Email   string `json:"email,omitempty" db:"email"`
	Website string `json:"website,omitempty" db:"website"`
}

// PriceRange represents pricing tier (1-4)
type PriceRange int

const (
	PriceBudget    PriceRange = 1 // $
	PriceModerate  PriceRange = 2 // $$
	PriceUpscale   PriceRange = 3 // $$$
	PriceLuxury    PriceRange = 4 // $$$$
)

// String returns the dollar sign representation
func (p PriceRange) String() string {
	return strings.Repeat("$", int(p))
}

// IsValid checks if the price range is within bounds
func (p PriceRange) IsValid() bool {
	return p >= PriceBudget && p <= PriceLuxury
}

// ===========================================
// Core Entities
// ===========================================

// Salon represents a beauty salon/shop in our system.
type Salon struct {
	ID          int64   `json:"id" db:"id"`
	Name        string  `json:"name" db:"name"`
	Slug        string  `json:"slug" db:"slug"`
	Description *string `json:"description,omitempty" db:"description"`

	// Composed value objects (flattened for DB compatibility)
	Location Location `json:"location"`
	Contact  Contact  `json:"contact"`

	// Business Info
	CategoryID  *int64     `json:"category_id,omitempty" db:"category_id"`
	PriceRange  PriceRange `json:"price_range,omitempty" db:"price_range"`
	Rating      *float64   `json:"rating,omitempty" db:"rating"`
	ReviewCount int        `json:"review_count" db:"review_count"`

	// Status
	IsActive   bool `json:"is_active" db:"is_active"`
	IsVerified bool `json:"is_verified" db:"is_verified"`

	// Timestamps
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`

	// Related data (populated via joins or separate queries)
	Category       *Category        `json:"category,omitempty"`
	Services       []Service        `json:"services,omitempty"`
	Amenities      []Amenity        `json:"amenities,omitempty"`
	OperatingHours []OperatingHours `json:"operating_hours,omitempty"`
}

// Validate checks if the salon data is valid
func (s *Salon) Validate() error {
	var errs []string

	if strings.TrimSpace(s.Name) == "" {
		errs = append(errs, "name is required")
	}
	if len(s.Name) > 255 {
		errs = append(errs, "name must be less than 255 characters")
	}
	if strings.TrimSpace(s.Slug) == "" {
		errs = append(errs, "slug is required")
	}
	if s.PriceRange != 0 && !s.PriceRange.IsValid() {
		errs = append(errs, "price_range must be between 1 and 4")
	}
	if s.Rating != nil && (*s.Rating < 0 || *s.Rating > 5) {
		errs = append(errs, "rating must be between 0 and 5")
	}
	if s.Location.GeoPoint != nil && !s.Location.GeoPoint.IsValid() {
		errs = append(errs, "invalid geo coordinates")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// DistanceTo calculates distance to a geographic point (returns nil if no coordinates)
func (s *Salon) DistanceTo(point GeoPoint) *float64 {
	if s.Location.GeoPoint == nil {
		return nil
	}
	dist := s.Location.GeoPoint.DistanceTo(point)
	return &dist
}

// IsOpen checks if the salon is currently open based on operating hours
func (s *Salon) IsOpen(t time.Time) bool {
	if len(s.OperatingHours) == 0 {
		return false // Unknown, assume closed
	}

	dayOfWeek := int(t.Weekday())
	currentTime := t.Format("15:04:05")

	for _, oh := range s.OperatingHours {
		if oh.DayOfWeek == dayOfWeek && !oh.IsClosed {
			if currentTime >= oh.OpenTime && currentTime <= oh.CloseTime {
				return true
			}
		}
	}
	return false
}

// IsOpenNow checks if the salon is currently open
func (s *Salon) IsOpenNow() bool {
	return s.IsOpen(time.Now())
}

// Category represents a type of beauty business
type Category struct {
	ID        int64     `json:"id" db:"id"`
	Name      string    `json:"name" db:"name"`
	Slug      string    `json:"slug" db:"slug"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

// Validate checks if the category data is valid
func (c *Category) Validate() error {
	var errs []string

	if strings.TrimSpace(c.Name) == "" {
		errs = append(errs, "name is required")
	}
	if len(c.Name) > 100 {
		errs = append(errs, "name must be less than 100 characters")
	}
	if strings.TrimSpace(c.Slug) == "" {
		errs = append(errs, "slug is required")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// Service represents a service offered by a salon
type Service struct {
	ID              int64     `json:"id" db:"id"`
	SalonID         int64     `json:"salon_id" db:"salon_id"`
	Name            string    `json:"name" db:"name"`
	Description     *string   `json:"description,omitempty" db:"description"`
	PriceMin        *float64  `json:"price_min,omitempty" db:"price_min"`
	PriceMax        *float64  `json:"price_max,omitempty" db:"price_max"`
	DurationMinutes *int      `json:"duration_minutes,omitempty" db:"duration_minutes"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
}

// PriceDisplay returns a formatted price string
func (s *Service) PriceDisplay() string {
	if s.PriceMin == nil && s.PriceMax == nil {
		return "Price varies"
	}
	if s.PriceMin != nil && s.PriceMax != nil {
		if *s.PriceMin == *s.PriceMax {
			return fmt.Sprintf("$%.2f", *s.PriceMin)
		}
		return fmt.Sprintf("$%.2f - $%.2f", *s.PriceMin, *s.PriceMax)
	}
	if s.PriceMin != nil {
		return fmt.Sprintf("From $%.2f", *s.PriceMin)
	}
	return fmt.Sprintf("Up to $%.2f", *s.PriceMax)
}

// DurationDisplay returns a formatted duration string
func (s *Service) DurationDisplay() string {
	if s.DurationMinutes == nil {
		return ""
	}
	if *s.DurationMinutes < 60 {
		return fmt.Sprintf("%d min", *s.DurationMinutes)
	}
	hours := *s.DurationMinutes / 60
	mins := *s.DurationMinutes % 60
	if mins == 0 {
		return fmt.Sprintf("%d hr", hours)
	}
	return fmt.Sprintf("%d hr %d min", hours, mins)
}

// Validate checks if the service data is valid
func (s *Service) Validate() error {
	var errs []string

	if strings.TrimSpace(s.Name) == "" {
		errs = append(errs, "name is required")
	}
	if s.SalonID <= 0 {
		errs = append(errs, "salon_id is required")
	}
	if s.PriceMin != nil && *s.PriceMin < 0 {
		errs = append(errs, "price_min cannot be negative")
	}
	if s.PriceMax != nil && *s.PriceMax < 0 {
		errs = append(errs, "price_max cannot be negative")
	}
	if s.PriceMin != nil && s.PriceMax != nil && *s.PriceMin > *s.PriceMax {
		errs = append(errs, "price_min cannot be greater than price_max")
	}
	if s.DurationMinutes != nil && *s.DurationMinutes <= 0 {
		errs = append(errs, "duration_minutes must be positive")
	}

	if len(errs) > 0 {
		return errors.New(strings.Join(errs, "; "))
	}
	return nil
}

// Amenity represents a feature/amenity (WiFi, Parking, etc.)
type Amenity struct {
	ID   int64  `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
	Icon string `json:"icon,omitempty" db:"icon"`
}

// OperatingHours represents business hours for a specific day
type OperatingHours struct {
	ID        int64  `json:"id" db:"id"`
	SalonID   int64  `json:"salon_id" db:"salon_id"`
	DayOfWeek int    `json:"day_of_week" db:"day_of_week"` // 0=Sunday, 6=Saturday
	OpenTime  string `json:"open_time" db:"open_time"`     // "09:00:00"
	CloseTime string `json:"close_time" db:"close_time"`   // "18:00:00"
	IsClosed  bool   `json:"is_closed" db:"is_closed"`
}

// DayName returns the name of the day
func (oh OperatingHours) DayName() string {
	days := []string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	if oh.DayOfWeek >= 0 && oh.DayOfWeek < len(days) {
		return days[oh.DayOfWeek]
	}
	return ""
}

// DisplayHours returns formatted hours string
func (oh OperatingHours) DisplayHours() string {
	if oh.IsClosed {
		return "Closed"
	}
	return fmt.Sprintf("%s - %s", oh.OpenTime[:5], oh.CloseTime[:5])
}

// ===========================================
// Search Types
// ===========================================

// SalonSearchParams contains all possible search/filter parameters
type SalonSearchParams struct {
	Query      string     // Full-text search query
	City       string     // Filter by city
	CategoryID *int64     // Filter by category
	PriceRange PriceRange // Filter by price range (1-4)
	MinRating  *float64   // Minimum rating filter
	IsVerified *bool      // Filter verified only
	Location   *GeoPoint  // For geo-search
	RadiusKm   *float64   // Radius for geo-search
	Page       int        // Pagination
	PageSize   int        // Results per page
	SortBy     SortOption // Sort field
}

// SortOption defines how results should be sorted
type SortOption string

const (
	SortByRelevance SortOption = "relevance"
	SortByRating    SortOption = "rating"
	SortByDistance  SortOption = "distance"
	SortByNewest    SortOption = "newest"
	SortByReviews   SortOption = "reviews"
)

// SalonSearchResult wraps a salon with search metadata
type SalonSearchResult struct {
	Salon      Salon              `json:"salon"`
	Score      float64            `json:"score,omitempty"`       // Search relevance score
	Distance   *float64           `json:"distance_km,omitempty"` // Distance from search point
	Highlights map[string]string  `json:"highlights,omitempty"`  // Matched text highlights
}

// SearchResponse contains paginated search results
type SearchResponse struct {
	Results    []SalonSearchResult `json:"results"`
	Total      int64               `json:"total"`
	Page       int                 `json:"page"`
	PageSize   int                 `json:"page_size"`
	TotalPages int                 `json:"total_pages"`
	Query      string              `json:"query,omitempty"`
	Source     string              `json:"source,omitempty"`
}

// NewSearchResponse creates a SearchResponse with calculated pagination
func NewSearchResponse(results []SalonSearchResult, total int64, params SalonSearchParams) SearchResponse {
	totalPages := int(total) / params.PageSize
	if int(total)%params.PageSize > 0 {
		totalPages++
	}

	return SearchResponse{
		Results:    results,
		Total:      total,
		Page:       params.Page,
		PageSize:   params.PageSize,
		TotalPages: totalPages,
		Query:      params.Query,
	}
}
