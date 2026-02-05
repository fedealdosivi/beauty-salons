package domain

import (
	"testing"
	"time"
)

func TestGeoPoint_DistanceTo(t *testing.T) {
	tests := []struct {
		name     string
		from     GeoPoint
		to       GeoPoint
		wantMin  float64
		wantMax  float64
	}{
		{
			name:    "same point",
			from:    GeoPoint{Latitude: 40.7128, Longitude: -74.0060},
			to:      GeoPoint{Latitude: 40.7128, Longitude: -74.0060},
			wantMin: 0,
			wantMax: 0.001,
		},
		{
			name:    "NYC to LA approximately 3944km",
			from:    GeoPoint{Latitude: 40.7128, Longitude: -74.0060},
			to:      GeoPoint{Latitude: 34.0522, Longitude: -118.2437},
			wantMin: 3900,
			wantMax: 4000,
		},
		{
			name:    "short distance 1km",
			from:    GeoPoint{Latitude: 40.7128, Longitude: -74.0060},
			to:      GeoPoint{Latitude: 40.7218, Longitude: -74.0060},
			wantMin: 0.9,
			wantMax: 1.1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.from.DistanceTo(tt.to)
			if got < tt.wantMin || got > tt.wantMax {
				t.Errorf("DistanceTo() = %v, want between %v and %v", got, tt.wantMin, tt.wantMax)
			}
		})
	}
}

func TestGeoPoint_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		point GeoPoint
		want  bool
	}{
		{"valid point", GeoPoint{Latitude: 40.7128, Longitude: -74.0060}, true},
		{"zero point", GeoPoint{Latitude: 0, Longitude: 0}, true},
		{"max bounds", GeoPoint{Latitude: 90, Longitude: 180}, true},
		{"min bounds", GeoPoint{Latitude: -90, Longitude: -180}, true},
		{"invalid latitude high", GeoPoint{Latitude: 91, Longitude: 0}, false},
		{"invalid latitude low", GeoPoint{Latitude: -91, Longitude: 0}, false},
		{"invalid longitude high", GeoPoint{Latitude: 0, Longitude: 181}, false},
		{"invalid longitude low", GeoPoint{Latitude: 0, Longitude: -181}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.point.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLocation_FullAddress(t *testing.T) {
	tests := []struct {
		name     string
		location Location
		want     string
	}{
		{
			name: "full address",
			location: Location{
				Address:    "123 Main St",
				City:       "New York",
				State:      "NY",
				PostalCode: "10001",
				Country:    "USA",
			},
			want: "123 Main St, New York, NY, 10001, USA",
		},
		{
			name: "partial address",
			location: Location{
				City:  "Miami",
				State: "FL",
			},
			want: "Miami, FL",
		},
		{
			name:     "empty address",
			location: Location{},
			want:     "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.location.FullAddress(); got != tt.want {
				t.Errorf("FullAddress() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriceRange_String(t *testing.T) {
	tests := []struct {
		pr   PriceRange
		want string
	}{
		{PriceBudget, "$"},
		{PriceModerate, "$$"},
		{PriceUpscale, "$$$"},
		{PriceLuxury, "$$$$"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			if got := tt.pr.String(); got != tt.want {
				t.Errorf("String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPriceRange_IsValid(t *testing.T) {
	tests := []struct {
		pr   PriceRange
		want bool
	}{
		{PriceRange(0), false},
		{PriceBudget, true},
		{PriceModerate, true},
		{PriceUpscale, true},
		{PriceLuxury, true},
		{PriceRange(5), false},
	}

	for _, tt := range tests {
		t.Run(tt.pr.String(), func(t *testing.T) {
			if got := tt.pr.IsValid(); got != tt.want {
				t.Errorf("IsValid() = %v, want %v for PriceRange(%d)", got, tt.want, tt.pr)
			}
		})
	}
}

func TestSalon_Validate(t *testing.T) {
	tests := []struct {
		name    string
		salon   Salon
		wantErr bool
	}{
		{
			name: "valid salon",
			salon: Salon{
				Name:       "Test Salon",
				Slug:       "test-salon",
				PriceRange: PriceModerate,
			},
			wantErr: false,
		},
		{
			name: "missing name",
			salon: Salon{
				Slug: "test-salon",
			},
			wantErr: true,
		},
		{
			name: "missing slug",
			salon: Salon{
				Name: "Test Salon",
			},
			wantErr: true,
		},
		{
			name: "invalid price range",
			salon: Salon{
				Name:       "Test Salon",
				Slug:       "test-salon",
				PriceRange: PriceRange(10),
			},
			wantErr: true,
		},
		{
			name: "invalid rating too high",
			salon: Salon{
				Name:   "Test Salon",
				Slug:   "test-salon",
				Rating: floatPtr(5.5),
			},
			wantErr: true,
		},
		{
			name: "invalid rating negative",
			salon: Salon{
				Name:   "Test Salon",
				Slug:   "test-salon",
				Rating: floatPtr(-1),
			},
			wantErr: true,
		},
		{
			name: "invalid geo coordinates",
			salon: Salon{
				Name: "Test Salon",
				Slug: "test-salon",
				Location: Location{
					GeoPoint: &GeoPoint{Latitude: 100, Longitude: 0},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.salon.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSalon_IsOpen(t *testing.T) {
	// Monday 10:00 AM
	monday10am := time.Date(2024, 1, 15, 10, 0, 0, 0, time.UTC)
	// Monday 8:00 PM
	monday8pm := time.Date(2024, 1, 15, 20, 0, 0, 0, time.UTC)
	// Sunday 10:00 AM
	sunday10am := time.Date(2024, 1, 14, 10, 0, 0, 0, time.UTC)

	salon := Salon{
		OperatingHours: []OperatingHours{
			{DayOfWeek: 1, OpenTime: "09:00:00", CloseTime: "18:00:00", IsClosed: false}, // Monday
			{DayOfWeek: 0, OpenTime: "00:00:00", CloseTime: "00:00:00", IsClosed: true},  // Sunday closed
		},
	}

	tests := []struct {
		name string
		time time.Time
		want bool
	}{
		{"Monday 10am - open", monday10am, true},
		{"Monday 8pm - closed", monday8pm, false},
		{"Sunday 10am - closed day", sunday10am, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := salon.IsOpen(tt.time); got != tt.want {
				t.Errorf("IsOpen() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSalon_DistanceTo(t *testing.T) {
	salonWithLocation := Salon{
		Location: Location{
			GeoPoint: &GeoPoint{Latitude: 40.7128, Longitude: -74.0060},
		},
	}
	salonWithoutLocation := Salon{}

	userLocation := GeoPoint{Latitude: 40.7218, Longitude: -74.0060}

	t.Run("salon with location", func(t *testing.T) {
		dist := salonWithLocation.DistanceTo(userLocation)
		if dist == nil {
			t.Error("DistanceTo() returned nil, expected distance")
		}
		if *dist < 0.9 || *dist > 1.1 {
			t.Errorf("DistanceTo() = %v, want ~1km", *dist)
		}
	})

	t.Run("salon without location", func(t *testing.T) {
		dist := salonWithoutLocation.DistanceTo(userLocation)
		if dist != nil {
			t.Error("DistanceTo() returned value, expected nil")
		}
	})
}

func TestService_PriceDisplay(t *testing.T) {
	tests := []struct {
		name    string
		service Service
		want    string
	}{
		{
			name:    "no prices",
			service: Service{},
			want:    "Price varies",
		},
		{
			name:    "same min and max",
			service: Service{PriceMin: floatPtr(50), PriceMax: floatPtr(50)},
			want:    "$50.00",
		},
		{
			name:    "price range",
			service: Service{PriceMin: floatPtr(30), PriceMax: floatPtr(50)},
			want:    "$30.00 - $50.00",
		},
		{
			name:    "only min",
			service: Service{PriceMin: floatPtr(30)},
			want:    "From $30.00",
		},
		{
			name:    "only max",
			service: Service{PriceMax: floatPtr(100)},
			want:    "Up to $100.00",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.service.PriceDisplay(); got != tt.want {
				t.Errorf("PriceDisplay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_DurationDisplay(t *testing.T) {
	tests := []struct {
		name    string
		service Service
		want    string
	}{
		{
			name:    "no duration",
			service: Service{},
			want:    "",
		},
		{
			name:    "30 minutes",
			service: Service{DurationMinutes: intPtr(30)},
			want:    "30 min",
		},
		{
			name:    "1 hour",
			service: Service{DurationMinutes: intPtr(60)},
			want:    "1 hr",
		},
		{
			name:    "1 hour 30 minutes",
			service: Service{DurationMinutes: intPtr(90)},
			want:    "1 hr 30 min",
		},
		{
			name:    "2 hours",
			service: Service{DurationMinutes: intPtr(120)},
			want:    "2 hr",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.service.DurationDisplay(); got != tt.want {
				t.Errorf("DurationDisplay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestService_Validate(t *testing.T) {
	tests := []struct {
		name    string
		service Service
		wantErr bool
	}{
		{
			name:    "valid service",
			service: Service{Name: "Haircut", SalonID: 1},
			wantErr: false,
		},
		{
			name:    "missing name",
			service: Service{SalonID: 1},
			wantErr: true,
		},
		{
			name:    "missing salon_id",
			service: Service{Name: "Haircut"},
			wantErr: true,
		},
		{
			name:    "negative price_min",
			service: Service{Name: "Haircut", SalonID: 1, PriceMin: floatPtr(-10)},
			wantErr: true,
		},
		{
			name:    "price_min greater than price_max",
			service: Service{Name: "Haircut", SalonID: 1, PriceMin: floatPtr(100), PriceMax: floatPtr(50)},
			wantErr: true,
		},
		{
			name:    "zero duration",
			service: Service{Name: "Haircut", SalonID: 1, DurationMinutes: intPtr(0)},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.service.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOperatingHours_DayName(t *testing.T) {
	tests := []struct {
		day  int
		want string
	}{
		{0, "Sunday"},
		{1, "Monday"},
		{2, "Tuesday"},
		{3, "Wednesday"},
		{4, "Thursday"},
		{5, "Friday"},
		{6, "Saturday"},
		{7, ""},
		{-1, ""},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			oh := OperatingHours{DayOfWeek: tt.day}
			if got := oh.DayName(); got != tt.want {
				t.Errorf("DayName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestOperatingHours_DisplayHours(t *testing.T) {
	tests := []struct {
		name string
		oh   OperatingHours
		want string
	}{
		{
			name: "open hours",
			oh:   OperatingHours{OpenTime: "09:00:00", CloseTime: "18:00:00", IsClosed: false},
			want: "09:00 - 18:00",
		},
		{
			name: "closed",
			oh:   OperatingHours{IsClosed: true},
			want: "Closed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.oh.DisplayHours(); got != tt.want {
				t.Errorf("DisplayHours() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSearchResponse(t *testing.T) {
	results := []SalonSearchResult{
		{Salon: Salon{ID: 1}},
		{Salon: Salon{ID: 2}},
	}
	params := SalonSearchParams{
		Query:    "test",
		Page:     1,
		PageSize: 10,
	}

	response := NewSearchResponse(results, 25, params)

	if response.Total != 25 {
		t.Errorf("Total = %v, want 25", response.Total)
	}
	if response.Page != 1 {
		t.Errorf("Page = %v, want 1", response.Page)
	}
	if response.PageSize != 10 {
		t.Errorf("PageSize = %v, want 10", response.PageSize)
	}
	if response.TotalPages != 3 {
		t.Errorf("TotalPages = %v, want 3", response.TotalPages)
	}
	if response.Query != "test" {
		t.Errorf("Query = %v, want test", response.Query)
	}
	if len(response.Results) != 2 {
		t.Errorf("len(Results) = %v, want 2", len(response.Results))
	}
}

// Helper functions
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}
