# OpenCortex

A comprehensive distributed intelligence platform providing unified access to graph databases, vector databases, document stores, AI APIs, and dynamic container execution through Redis messaging. OpenCortex enables seamless orchestration of data operations, AI processing, and containerized workloads with real-time capability discovery and management.

## 🏗️ Architecture

### Services Overview

1. **Data Abstractor** (`/data-abstractor/`) - Graph data operations with Neo4j, Qdrant, MongoDB
2. **AI Abstractor** (`/ai-abstractor/`) - Multi-provider AI API operations
3. **Exec Agent** (`/exec-agent/`) - Dynamic container execution with OCI image capability discovery
4. **Orchestrator** (`/orchestrator/`) - Workflow orchestration and capability management
5. **Admin Interface** (`/admin-ui/`) - NextJS web interface for monitoring and management
6. **Shared Infrastructure** - Redis, Neo4j, MongoDB, Qdrant, Minio

### Message Channels

- **Data Operations**: `data-requests` → `data-responses`
- **AI Operations**: `ai-requests` → `ai-responses`
- **Container Execution**: `exec-requests` → `exec-responses`
- **Capability Announcements**: `capability-announcements`
- **Workflow Events**: `workflow-events`
- **System Status**: `system-status`

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
- Exec Agent (container execution)
- Orchestrator (workflow management)
- Admin Interface (web dashboard)

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

# Monitor capability announcements
docker exec smart_data_abstractor-redis-1 redis-cli SUBSCRIBE capability-announcements
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

Provides dynamic containerized execution platform:
- **Docker Orchestration**: Execute any Docker container with custom configurations
- **OCI Image Capability Discovery**: Automatically scan Docker images for embedded capabilities
- **Dynamic Capability Management**: Real-time discovery and announcement of worker image capabilities
- **Data Management**: Mount JSON graph data, Minio blobs, and files 
- **Service Integration**: HTTP proxy for containers to access other services
- **Output Collection**: Extract files, graph updates, and Minio uploads

### Capability Discovery Methods

1. **Docker Labels**: Capabilities defined in image labels
   ```dockerfile
   LABEL capability.name="data-processor"
   LABEL capability.description="Processes CSV data"
   LABEL capability.inputs='["csv_file"]'
   LABEL capability.outputs='["processed_data"]'
   ```

2. **Embedded JSON**: Capability files within the image
   ```json
   {
     "name": "data-processor",
     "description": "Advanced data processing capabilities",
     "operations": ["transform", "analyze", "export"]
   }
   ```

3. **Automatic Inference**: Capability detection based on image analysis
   - Scans installed packages, binaries, and frameworks
   - Generates capability descriptions automatically

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

## 🎛️ Admin Interface

OpenCortex includes a comprehensive NextJS-based admin interface for monitoring and management:

### Features
- **Dashboard**: System health monitoring and real-time metrics
- **Message Bus Monitor**: Live Redis pub/sub message visualization with filtering
- **Graph Visualization**: Interactive Neo4j graph exploration with Cypher queries
- **AI Query Interface**: Multi-model AI chat with conversation history
- **Capabilities Management**: Real-time capability discovery, testing, and monitoring

### Access
```bash
# Start the admin interface
cd admin-ui
npm install && npm run dev
# Open http://localhost:3000
```

### Docker Deployment
```bash
# Include admin interface in main stack
docker compose -f docker-compose.yml -f admin-ui/docker-compose.yml up -d
```

## 🎼 Orchestrator

Manages complex workflows and capability coordination:
- **Workflow Engine**: Define and execute multi-step operations
- **Capability Registry**: Central registry of all system capabilities
- **Dynamic Routing**: Route requests based on available capabilities
- **Dependency Resolution**: Handle complex multi-service workflows

## 🔍 Monitoring & Logs

### View Service Logs
```bash
docker compose logs data-abstractor
docker compose logs ai-abstractor
docker compose logs exec-agent
docker compose logs orchestrator
docker compose logs admin-ui
docker compose logs redis
```

### Service Status
```bash
docker compose ps
docker compose logs -f  # Follow all logs
```

### Real-time Monitoring
Use the admin interface at `http://localhost:3000` for:
- Live message bus monitoring
- Service health dashboards
- Capability status tracking
- System performance metrics

## 📁 Project Structure

```
opencortex/
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
├── exec-agent/            # Container execution service
│   ├── main.go           # Application entry
│   ├── config/           # Configuration management
│   ├── clients/          # Docker, Minio, Redis clients
│   ├── handlers/         # Execution and data management
│   ├── capabilities/     # OCI image capability discovery
│   ├── models/           # Request/response structures
│   ├── examples/         # Test containers and scripts
│   ├── Dockerfile        # Container definition
│   └── README.md        # Service-specific docs
│
├── orchestrator/          # Workflow orchestration service
│   ├── main.go           # Application entry
│   ├── config/           # Configuration management
│   ├── workflow/         # Workflow engine
│   ├── registry/         # Capability registry
│   ├── models/           # Data structures
│   ├── Dockerfile        # Container definition
│   └── README.md        # Service-specific docs
│
└── admin-ui/             # NextJS admin interface
    ├── app/              # Next.js app directory
    ├── components/       # React components
    ├── public/           # Static assets
    ├── package.json      # Dependencies
    ├── Dockerfile        # Container definition
    ├── docker-compose.yml # Admin interface deployment
    └── README.md        # Interface documentation
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

OpenCortex provides a solid foundation for production distributed intelligence operations at scale, with dynamic capability discovery, real-time monitoring, and comprehensive workflow orchestration.