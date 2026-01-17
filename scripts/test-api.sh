#!/bin/bash

# Test script for the Beauty Salons Search API

BASE_URL="http://localhost:8080/api/v1"

echo "========================================"
echo "Beauty Salons Search API - Test Suite"
echo "========================================"
echo ""

# Health check
echo "1. Health Check"
echo "----------------"
curl -s http://localhost:8080/health | jq .
echo ""

# Sync data
echo "2. Sync Data to Elasticsearch"
echo "------------------------------"
curl -s -X POST "$BASE_URL/admin/sync" | jq .
echo ""

# Wait for indexing
sleep 2

# Basic search
echo "3. Basic Search: 'peluqueria'"
echo "------------------------------"
curl -s "$BASE_URL/search?q=peluqueria" | jq '.data[] | {name, rating, city}'
echo ""

# Search with filters
echo "4. Search with Filters: barber shops with rating >= 4.5"
echo "--------------------------------------------------------"
curl -s "$BASE_URL/search?q=barberia&min_rating=4.5" | jq '.data[] | {name, rating}'
echo ""

# Fuzzy search (typo tolerance)
echo "5. Fuzzy Search: 'peluqeria' (with typo)"
echo "-----------------------------------------"
curl -s "$BASE_URL/search?q=peluqeria" | jq '.data[] | {name}'
echo ""

# Verified only
echo "6. Verified Salons Only"
echo "------------------------"
curl -s "$BASE_URL/search?verified=true" | jq '.total'
echo ""

# Compare PostgreSQL vs Elasticsearch
echo "7. PostgreSQL vs Elasticsearch Comparison"
echo "------------------------------------------"
echo "Elasticsearch:"
time curl -s "$BASE_URL/search?q=spa" > /dev/null
echo ""
echo "PostgreSQL:"
time curl -s "$BASE_URL/search/postgres?q=spa" > /dev/null
echo ""

# Get single salon
echo "8. Get Salon by ID"
echo "-------------------"
curl -s "$BASE_URL/salons/1" | jq '{name, description, services: .services[].name}'
echo ""

# Categories
echo "9. List Categories"
echo "-------------------"
curl -s "$BASE_URL/categories" | jq '.[] | .name'
echo ""

# Cluster health
echo "10. Elasticsearch Cluster Health"
echo "---------------------------------"
curl -s "$BASE_URL/admin/cluster/health" | jq '{cluster_name, status, number_of_nodes}'
echo ""

echo "========================================"
echo "Test Suite Complete!"
echo "========================================"
