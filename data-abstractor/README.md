# Data Abstractor

A Go application that implements a data abstraction layer with Redis messaging, Neo4j graph queries, Qdrant similarity search, and MongoDB enrichment.

## Features

- **Redis Message Bus**: Listen on `data-requests` channel, respond on `data-responses`
- **Neo4j Graph Queries**: Execute Cypher queries and return structured graph data
- **Qdrant Similarity Search**: Find nodes by embedding similarity
- **MongoDB Enrichment**: Add metadata to graph nodes when requested

## Supported Operations

### 1. Graph Traversal (`traverse`)
Execute Cypher queries on Neo4j and return subgraphs.

```json
{
  "operation": "traverse",
  "correlation_id": "unique-id",
  "query": {
    "cypher": "MATCH (n)-[r]-(m) RETURN n,r,m LIMIT 100"
  },
  "enrich": ["metadata"],
  "limit": 100
}
```

### 2. Similarity Search (`search`)
Find nodes by embedding similarity via Qdrant, then fetch from Neo4j.

```json
{
  "operation": "search", 
  "correlation_id": "unique-id",
  "query": {
    "text": "web server vulnerability",
    "embedding": [0.1, 0.2, ...]
  },
  "enrich": ["metadata"],
  "limit": 100
}
```

### 3. Node Enrichment (`enrich`)
Add MongoDB document data to graph nodes.

```json
{
  "operation": "enrich",
  "correlation_id": "unique-id", 
  "query": {
    "node_ids": ["node1", "node2"]
  }
}
```

## Configuration

Environment variables:

```env
REDIS_URL=redis://localhost:6379
NEO4J_URL=bolt://localhost:7687
NEO4J_USER=neo4j
NEO4J_PASSWORD=password
MONGODB_URL=mongodb://localhost:27017
MONGODB_DATABASE=enrichment
QDRANT_URL=http://localhost:6333
QDRANT_COLLECTION=embeddings
LOG_LEVEL=info
```

## Response Format

All responses return structured graph data as JSON:

```json
{
  "correlation_id": "unique-id",
  "success": true,
  "data": {
    "nodes": [
      {
        "id": "node1",
        "labels": ["Person"],
        "properties": {"name": "John"},
        "metadata": {"enrichment": "data"},
        "score": 0.95
      }
    ],
    "relationships": [
      {
        "id": "rel1", 
        "type": "KNOWS",
        "start_node": "node1",
        "end_node": "node2",
        "properties": {}
      }
    ]
  },
  "timestamp": "2025-01-01T00:00:00Z",
  "operation": "search"
}
```

## Running

### With Docker Compose (Recommended)

```bash
docker-compose up -d
```

This starts the data abstractor along with Redis, Neo4j, MongoDB, and Qdrant.

### Local Development

1. Install Go 1.21+
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Copy environment file:
   ```bash
   cp .env.example .env
   ```
4. Start the databases (Redis, Neo4j, MongoDB, Qdrant)
5. Run the application:
   ```bash
   go run main.go
   ```

## Testing

Send a test message via Redis CLI:

```bash
redis-cli
PUBLISH data-requests '{"operation":"traverse","correlation_id":"test-1","query":{"cypher":"MATCH (n) RETURN n LIMIT 5"}}'
```

Listen for responses:
```bash
redis-cli
SUBSCRIBE data-responses
```

## Architecture

- `config/`: Configuration management
- `clients/`: Database client implementations
- `models/`: Request/response data structures  
- `handlers/`: Business logic and request processing
- `main.go`: Application entry point with graceful shutdown