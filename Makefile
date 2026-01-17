.PHONY: help up down logs api sync test-search health

# Default target
help:
	@echo "Beauty Salons Search Engine - Learning Project"
	@echo ""
	@echo "Infrastructure commands:"
	@echo "  make up          - Start PostgreSQL, Elasticsearch, and Kibana"
	@echo "  make down        - Stop all services"
	@echo "  make logs        - View container logs"
	@echo ""
	@echo "Application commands:"
	@echo "  make deps        - Download Go dependencies"
	@echo "  make api         - Run the API server"
	@echo "  make sync        - Sync data from PostgreSQL to Elasticsearch"
	@echo ""
	@echo "Testing commands:"
	@echo "  make test-search - Test search with sample queries"
	@echo "  make health      - Check cluster health"
	@echo ""
	@echo "Learning commands:"
	@echo "  make kibana      - Open Kibana in browser (port 5601)"
	@echo "  make es-info     - Show Elasticsearch cluster info"
	@echo "  make es-indices  - List all Elasticsearch indices"

# Start infrastructure
up:
	@echo "Starting infrastructure..."
	docker-compose up -d
	@echo ""
	@echo "Waiting for services to be ready..."
	@sleep 10
	@echo ""
	@echo "Services started!"
	@echo "  - PostgreSQL:    localhost:5432"
	@echo "  - Elasticsearch: localhost:9200"
	@echo "  - Kibana:        localhost:5601"

# Stop infrastructure
down:
	docker-compose down

# View logs
logs:
	docker-compose logs -f

# Download Go dependencies
deps:
	go mod download
	go mod tidy

# Run the API
api:
	go run cmd/api/main.go

# Sync data to Elasticsearch
sync:
	@echo "Syncing data from PostgreSQL to Elasticsearch..."
	curl -X POST http://localhost:8080/api/v1/admin/sync | jq .

# Test search queries
test-search:
	@echo "=== Search: 'barberia' ==="
	curl -s "http://localhost:8080/api/v1/search?q=barberia" | jq .
	@echo ""
	@echo "=== Search: 'spa' with min rating 4.5 ==="
	curl -s "http://localhost:8080/api/v1/search?q=spa&min_rating=4.5" | jq .
	@echo ""
	@echo "=== Search: verified salons ==="
	curl -s "http://localhost:8080/api/v1/search?verified=true" | jq .

# Check cluster health
health:
	@echo "=== Elasticsearch Cluster Health ==="
	curl -s http://localhost:9200/_cluster/health | jq .
	@echo ""
	@echo "=== API Health ==="
	curl -s http://localhost:8080/health | jq .

# Open Kibana
kibana:
	open http://localhost:5601

# Show Elasticsearch info
es-info:
	curl -s http://localhost:9200 | jq .

# List Elasticsearch indices
es-indices:
	curl -s http://localhost:9200/_cat/indices?v
