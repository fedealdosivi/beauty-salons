# Beauty Salons Search Engines

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           BEAUTY SALONS SEARCH                          │
├─────────────────────────────────────────────────────────────────────────┤
│                                                                         │
│  ┌─────────────────┐         ┌─────────────────┐                       │
│  │   PostgreSQL    │         │  Elasticsearch  │                       │
│  │  (Source of     │ ──────► │    (Search      │                       │
│  │    Truth)       │  Sync   │     Cluster)    │                       │
│  └─────────────────┘         └─────────────────┘                       │
│         │                            │                                  │
│         │ CRUD                       │ Search                          │
│         ▼                            ▼                                  │
│  ┌─────────────────────────────────────────────────┐                   │
│  │                   Go API                        │                   │
│  │  - /api/v1/search (Elasticsearch)               │                   │
│  │  - /api/v1/search/postgres (PostgreSQL)         │                   │
│  │  - /api/v1/salons/:id                           │                   │
│  │  - /api/v1/admin/sync                           │                   │
│  └─────────────────────────────────────────────────┘                   │
│                                                                         │
└─────────────────────────────────────────────────────────────────────────┘
```

## Key Concepts You'll Learn

### 1. Clusters
Elasticsearch runs as a **scluster** - a collection of nodes that work together.

```bash
# Check cluster health
curl http://localhost:9200/_cluster/health | jq .
```

### 2. Storage & Source of Truth
PostgreSQL is the **source of truth**. All data originates here.
Elasticsearch is a **secondary index** optimized for search.

- PostgreSQL: ACID transactions, relational data, data integrity
- Elasticsearch: Full-text search, fuzzy matching, aggregations

### 3. Indexes

**PostgreSQL indexes** (B-Tree):
```sql
CREATE INDEX idx_salons_city ON salons(city);
```
Good for exact matches, range queries, sorting.

**Elasticsearch indexes** (Inverted Index):
```
"peluqueria" → [doc1, doc5, doc12]
"barberia"   → [doc2, doc7]
```
Good for full-text search, relevance scoring, fuzzy matching.

### 4. Shards & Replicas
Data in Elasticsearch is split into **shards** and copied into **replicas**:

```
Index: salons
├── Shard 0 (Primary) ──► Replica 0
├── Shard 1 (Primary) ──► Replica 1
└── Shard 2 (Primary) ──► Replica 2
```

- **Shards**: Allow data to be distributed across nodes
- **Replicas**: Provide fault tolerance and read scaling

## Getting Started

### Prerequisites
- Docker & Docker Compose
- Go 1.21+
- curl & jq (for testing)

### Quick Start

```bash
# 1. Start infrastructure (PostgreSQL, Elasticsearch, Kibana)
make up

# 2. Wait for services to start, then run the API
make api

# 3. In another terminal, sync data to Elasticsearch
make sync

# 4. Test the search
curl "http://localhost:8080/api/v1/search?q=barberia" | jq .
```

### Explore with Kibana

Open http://localhost:5601 and go to **Dev Tools** to run Elasticsearch queries:

```json
// Search for salons
GET /salons/_search
{
  "query": {
    "match": {
      "name": "peluqueria"
    }
  }
}

// Get index mapping
GET /salons/_mapping

// See cluster health
GET /_cluster/health
```

## API Endpoints

| Endpoint | Description |
|----------|-------------|
| `GET /api/v1/search?q=...` | Search using Elasticsearch |
| `GET /api/v1/search/postgres?q=...` | Search using PostgreSQL (for comparison) |
| `GET /api/v1/salons/:id` | Get salon by ID |
| `GET /api/v1/categories` | List all categories |
| `POST /api/v1/admin/sync` | Sync data to Elasticsearch |
| `GET /api/v1/admin/cluster/health` | Get cluster health |

### Search Parameters

| Parameter | Description | Example |
|-----------|-------------|---------|
| `q` | Search query | `?q=barberia` |
| `city` | Filter by city | `?city=Mar del Plata` |
| `category` | Filter by category ID | `?category=1` |
| `min_rating` | Minimum rating | `?min_rating=4.5` |
| `verified` | Verified only | `?verified=true` |
| `page` | Page number | `?page=2` |
| `page_size` | Results per page | `?page_size=20` |

