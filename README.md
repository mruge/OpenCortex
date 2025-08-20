# Smart Data Abstractor

A comprehensive microservices architecture providing unified access to graph databases, vector databases, document stores, and AI APIs through Redis messaging.

## 🏗️ Architecture

### Services Overview

1. **Data Abstractor** (`/data-abstractor/`) - Graph data operations
2. **AI Abstractor** (`/ai-abstractor/`) - AI API operations
3. **Exec Agent** (`/exec-agent/`) - Container execution platform
4. **Shared Infrastructure** - Redis, Neo4j, MongoDB, Qdrant, Minio

### Message Channels

- **Data Operations**: `data-requests` → `data-responses`
- **AI Operations**: `ai-requests` → `ai-responses`
- **Container Execution**: `exec-requests` → `exec-responses`

## 📊 Data Abstractor

Provides unified access to:
- **Neo4j**: Graph traversal and Cypher queries
- **Qdrant**: Vector similarity search
- **MongoDB**: Data enrichment with metadata

### Supported Operations

#### 1. Graph Traversal (`traverse`)
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

#### 2. Similarity Search (`search`)
```json
{
  "operation": "search", 
  "correlation_id": "unique-id",
  "query": {
    "embedding": [0.1, 0.2, 0.3, 0.4, 0.5]
  },
  "enrich": ["metadata"],
  "limit": 100
}
```

#### 3. Node Enrichment (`enrich`)
```json
{
  "operation": "enrich",
  "correlation_id": "unique-id", 
  "query": {
    "node_ids": ["node1", "node2"]
  }
}
```

### Response Format
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

## 🤖 AI Abstractor

Provides unified access to:
- **OpenAI**: GPT models (GPT-4, GPT-3.5-turbo)  
- **Anthropic**: Claude models (Claude-3-Sonnet, Claude-3-Opus)

### Features

- **Context Management**: Include additional text context
- **Response Formatting**: Auto-format for JSON, YAML, Markdown, Text
- **Format Validation**: Verify response format compliance
- **Flexible Configuration**: Model-specific parameters

### Request Format
```json
{
  "provider": "openai",
  "correlation_id": "unique-id",
  "prompt": "Explain quantum computing",
  "system_message": "You are a helpful physics teacher",
  "context": [
    "Quantum computing uses quantum bits (qubits)",
    "Superposition allows multiple states simultaneously"
  ],
  "response_format": "markdown",
  "model": "gpt-4",
  "max_tokens": 2000,
  "temperature": 0.7
}
```

### Response Format
```json
{
  "correlation_id": "unique-id",
  "success": true,
  "content": "Quantum computing is a revolutionary...",
  "timestamp": "2025-01-01T00:00:00Z",
  "provider": "openai",
  "model": "gpt-4",
  "tokens_used": 150,
  "response_format": "markdown"
}
```

## 🚀 Quick Start

### Prerequisites
- Docker and Docker Compose
- API keys for AI services (optional)

### 1. Setup Environment
```bash
cd smart_data_abstractor
cp .env.example .env
# Add your API keys to .env (optional - only needed for AI services)
```

### 2. Start All Services
```bash
docker compose up -d
```

This starts:
- Redis (message bus)
- Neo4j (graph database) 
- MongoDB (document store)
- Qdrant (vector database)
- Data Abstractor (always)
- AI Abstractor (only if API keys provided)

### 3. Test Data Operations
```bash
# Send a graph query
docker exec smart_data_abstractor-redis-1 redis-cli PUBLISH data-requests '{
  "operation": "traverse",
  "correlation_id": "test-1", 
  "query": {"cypher": "RETURN 42 as answer"}
}'

# Listen for responses
docker exec smart_data_abstractor-redis-1 redis-cli SUBSCRIBE data-responses
```

### 4. Test AI Operations (requires API keys)
```bash
# Send an AI request
docker exec smart_data_abstractor-redis-1 redis-cli PUBLISH ai-requests '{
  "provider": "openai",
  "correlation_id": "ai-test-1",
  "prompt": "What is 2+2?",
  "response_format": "text"
}'

# Listen for AI responses  
docker exec smart_data_abstractor-redis-1 redis-cli SUBSCRIBE ai-responses
```

## 🔧 Configuration

### Environment Variables

```env
# AI Service API Keys (optional)
OPENAI_API_KEY=your_openai_key_here
ANTHROPIC_API_KEY=your_anthropic_key_here
```

### Service Ports
- **Redis**: 6379
- **Neo4j HTTP**: 7474
- **Neo4j Bolt**: 7687  
- **MongoDB**: 27017
- **Qdrant**: 6333-6334

## 📋 Example Use Cases

### 1. Knowledge Graph Query with AI Analysis
```bash
# 1. Query graph data
PUBLISH data-requests '{
  "operation": "traverse",
  "correlation_id": "kb-1",
  "query": {"cypher": "MATCH (p:Person)-[r:WORKS_AT]->(c:Company) RETURN p,r,c LIMIT 5"}
}'

# 2. Analyze results with AI
PUBLISH ai-requests '{
  "provider": "anthropic", 
  "correlation_id": "analysis-1",
  "prompt": "Analyze this employment data and identify patterns",
  "context": ["Results from previous query..."],
  "response_format": "json"
}'
```

### 2. Semantic Search with Enrichment
```bash
# Search by vector similarity and enrich with metadata
PUBLISH data-requests '{
  "operation": "search",
  "correlation_id": "semantic-1",
  "query": {"embedding": [0.1, 0.2, 0.3]},
  "enrich": ["metadata"],
  "limit": 10
}'
```

### 3. Multi-format AI Content Generation
```bash
# Generate structured YAML configuration  
PUBLISH ai-requests '{
  "provider": "openai",
  "correlation_id": "yaml-gen",
  "prompt": "Create a microservices deployment config",
  "response_format": "yaml"
}'
```

## 🤖 Exec Agent

Provides containerized execution platform:
- **Docker Orchestration**: Execute any Docker container with custom configurations
- **Data Management**: Mount JSON graph data, Minio blobs, and files 
- **Service Integration**: HTTP proxy for containers to access other services
- **Output Collection**: Extract files, graph updates, and Minio uploads

### Request Format
```json
{
  "correlation_id": "unique-id",
  "container": {
    "image": "python:3.9-slim",
    "command": ["python", "/workspace/input/script.py"]
  },
  "input": {
    "graph_data": {"nodes": [...], "relationships": [...]},
    "minio_objects": [{"object_name": "data.csv", "local_path": "input.csv"}],
    "files": [{"name": "script.py", "path": "script.py", "content": "import json\n..."}]
  },
  "output": {
    "expected_files": ["results.json"],
    "minio_upload": true,
    "graph_update": true
  },
  "service_access": ["data", "ai"]
}
```

## 🔍 Monitoring & Logs

### View Service Logs
```bash
docker compose logs data-abstractor
docker compose logs ai-abstractor
docker compose logs exec-agent
docker compose logs redis
```

### Service Status
```bash
docker compose ps
docker compose logs -f  # Follow all logs
```

## 📁 Project Structure

```
smart_data_abstractor/
├── docker-compose.yml          # Master orchestration
├── .env.example               # Environment template
├── README.md                 # This file
│
├── data-abstractor/          # Graph/Vector/Document service
│   ├── main.go              # Application entry
│   ├── config/              # Configuration management  
│   ├── clients/             # Database clients
│   ├── handlers/            # Business logic
│   ├── models/              # Data structures
│   ├── Dockerfile           # Container definition
│   └── README.md           # Service-specific docs
│
├── ai-abstractor/           # AI service
│   ├── main.go             # Application entry
│   ├── config/             # Configuration management
│   ├── clients/            # AI API clients  
│   ├── handlers/           # Business logic
│   ├── models/             # Data structures
│   ├── Dockerfile          # Container definition
│   └── README.md          # Service-specific docs
│
└── exec-agent/            # Container execution service
    ├── main.go           # Application entry
    ├── config/           # Configuration management
    ├── clients/          # Docker, Minio, Redis clients
    ├── handlers/         # Execution and data management
    ├── models/           # Request/response structures
    ├── examples/         # Test containers and scripts
    ├── Dockerfile        # Container definition
    └── README.md        # Service-specific docs
```

## 🛠️ Development

### Run Individual Services
```bash
# Data abstractor only
cd data-abstractor && docker compose up -d

# AI abstractor only  
cd ai-abstractor && docker compose up -d
```

### Build from Source
```bash
# Build specific service
docker compose build data-abstractor
docker compose build ai-abstractor

# Rebuild and restart
docker compose up --build -d
```

## 🔐 Security Notes

- API keys are passed as environment variables
- Services communicate through internal Docker network
- Only necessary ports are exposed to host
- No hardcoded secrets in code or containers

## ✅ Production Readiness

Both services include:
- ✅ Structured JSON logging
- ✅ Graceful shutdown handling  
- ✅ Environment-based configuration
- ✅ Health check endpoints
- ✅ Error handling and recovery
- ✅ Resource cleanup
- ✅ Container optimization

This architecture provides a solid foundation for production graph data and AI operations at scale.