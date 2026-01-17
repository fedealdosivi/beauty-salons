-- ===========================================
-- Beauty Salons Database Schema
-- ===========================================
-- Elasticsearch will be a SECONDARY INDEX for search.

-- Categories of beauty services
CREATE TABLE IF NOT EXISTS categories (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    slug VARCHAR(100) NOT NULL UNIQUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Main salons/shops table
CREATE TABLE IF NOT EXISTS salons (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    slug VARCHAR(255) NOT NULL UNIQUE,
    description TEXT,

    -- Location data (important for geo-search later!)
    address VARCHAR(500),
    city VARCHAR(100),
    state VARCHAR(100),
    postal_code VARCHAR(20),
    country VARCHAR(100) DEFAULT 'USA',
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),

    -- Contact
    phone VARCHAR(50),
    email VARCHAR(255),
    website VARCHAR(500),

    -- Business info
    category_id INTEGER REFERENCES categories(id),
    price_range SMALLINT CHECK (price_range BETWEEN 1 AND 4), -- 1=$ to 4=$$$$
    rating DECIMAL(2, 1) CHECK (rating BETWEEN 0 AND 5),
    review_count INTEGER DEFAULT 0,

    -- Status
    is_active BOOLEAN DEFAULT true,
    is_verified BOOLEAN DEFAULT false,

    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Services offered by each salon
CREATE TABLE IF NOT EXISTS services (
    id SERIAL PRIMARY KEY,
    salon_id INTEGER REFERENCES salons(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    price_min DECIMAL(10, 2),
    price_max DECIMAL(10, 2),
    duration_minutes INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

-- Amenities/features (WiFi, Parking, etc.)
CREATE TABLE IF NOT EXISTS amenities (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL UNIQUE,
    icon VARCHAR(50)
);

-- Many-to-many: salons <-> amenities
CREATE TABLE IF NOT EXISTS salon_amenities (
    salon_id INTEGER REFERENCES salons(id) ON DELETE CASCADE,
    amenity_id INTEGER REFERENCES amenities(id) ON DELETE CASCADE,
    PRIMARY KEY (salon_id, amenity_id)
);

-- Operating hours
CREATE TABLE IF NOT EXISTS operating_hours (
    id SERIAL PRIMARY KEY,
    salon_id INTEGER REFERENCES salons(id) ON DELETE CASCADE,
    day_of_week SMALLINT CHECK (day_of_week BETWEEN 0 AND 6), -- 0=Sunday
    open_time TIME,
    close_time TIME,
    is_closed BOOLEAN DEFAULT false
);

-- ===========================================
-- Indexes for PostgreSQL queries
-- ===========================================
-- These are B-TREE indexes (different from Elasticsearch's inverted index!)
-- They speed up lookups but don't help with fulsl-text search.

CREATE INDEX idx_salons_city ON salons(city);
CREATE INDEX idx_salons_category ON salons(category_id);
CREATE INDEX idx_salons_rating ON salons(rating DESC);
CREATE INDEX idx_salons_location ON salons(latitude, longitude);
CREATE INDEX idx_services_salon ON services(salon_id);

-- Full-text search in PostgreSQL (for comparison with Elasticsearch)
-- This creates a GIN index with tsvector - PostgreSQL's way of doing text search
CREATE INDEX idx_salons_search ON salons USING GIN (
    to_tsvector('english', coalesce(name, '') || ' ' || coalesce(description, ''))
);
