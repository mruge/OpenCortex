# Orchestrator Service

A comprehensive workflow orchestration engine with AI-powered workflow generation, DAG-based execution, Redis state persistence, and automatic recovery capabilities.

## Features

### Core Capabilities
- **Workflow Engine**: DAG-based task execution with dependency resolution
- **AI Integration**: Generate workflows using natural language prompts
- **Template System**: Reusable workflow templates with variable substitution
- **State Management**: Redis-based persistence with TTL and cleanup
- **Message Coordination**: Inter-service communication via Redis pub/sub
- **Recovery System**: Automatic detection and recovery of failed workflows
- **HTTP API**: RESTful endpoints for workflow management

### Task Types
- **Data Tasks**: Integration with data-abstractor service
- **AI Tasks**: Integration with ai-abstractor service  
- **Exec Tasks**: Integration with exec-agent for container execution
- **Parallel Tasks**: Execute multiple sub-tasks concurrently
- **Condition Tasks**: Conditional workflow branching

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   HTTP API      │    │  Message Bus    │    │  State Manager  │
│                 │    │   (Redis)       │    │    (Redis)      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
┌─────────────────────────────────┼─────────────────────────────────┐
│                    Orchestrator Core                              │
├─────────────────┬─────────────────┬─────────────────┬─────────────┤
│ Workflow Engine │ AI Generator    │ Template Mgr    │ Recovery    │
│                 │                 │                 │ Manager     │
└─────────────────┴─────────────────┴─────────────────┴─────────────┘
         │                 │                 │                 │
         └─────────────────┼─────────────────┼─────────────────┘
                           │                 │
┌─────────────────────────┼─────────────────┼─────────────────────┐
│                External Services                                │
├─────────────────┬─────────────────┬─────────────────┬───────────┤
│ Data Abstractor │ AI Abstractor   │ Exec Agent      │ Templates │
│                 │                 │                 │ (Files)   │
└─────────────────┴─────────────────┴─────────────────┴───────────┘
```

## Quick Start

### Prerequisites
- Docker and Docker Compose
- Redis server
- Access to data-abstractor, ai-abstractor, and exec-agent services

### 1. Build and Run
```bash
# Build the service
cd orchestrator
docker build -t smart_data_abstractor-orchestrator .

# Run with Docker Compose (from project root)
cd ../
docker compose up -d orchestrator
```

### 2. Environment Variables
```bash
# Redis Configuration
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DATABASE=0

# Server Configuration  
ORCHESTRATOR_PORT=8080
ORCHESTRATOR_HOST=0.0.0.0

# Service Configuration
MAX_CONCURRENT_WORKFLOWS=10
DEFAULT_WORKFLOW_TIMEOUT=3600s
EXECUTION_TTL=24h
RECOVERY_ENABLED=true
RECOVERY_INTERVAL=5m

# Templates and Workspace
ORCHESTRATOR_TEMPLATES=./templates
ORCHESTRATOR_WORKSPACE=/tmp/orchestrator
```

## Usage

### HTTP API

#### Execute Workflow from Template
```bash
curl -X POST http://localhost:8080/api/v1/workflows \
  -H "Content-Type: application/json" \
  -d '{
    "correlation_id": "workflow-001",
    "workflow_template": "data-analysis-basic",
    "variables": {
      "query_limit": 50,
      "analysis_prompt": "Identify key trends in the data"
    }
  }'
```

#### Generate Workflow with AI
```bash
curl -X POST http://localhost:8080/api/v1/generate \
  -H "Content-Type: application/json" \
  -d '{
    "prompt": "Create a data analysis pipeline that fetches graph data, performs vector search, and generates an AI-powered summary report",
    "domain": "data-science", 
    "complexity": "medium",
    "required_services": ["data", "ai", "exec"]
  }'
```

#### Check Workflow Status
```bash
curl http://localhost:8080/api/v1/workflows/{execution_id}/status
```

#### List Templates
```bash
curl http://localhost:8080/api/v1/templates
```

### Redis Message Bus

#### Send Workflow Request
```bash
# Publish workflow request
docker exec redis redis-cli PUBLISH workflow-requests '{
  "correlation_id": "test-workflow",
  "workflow_template": "data-analysis-basic",
  "variables": {"query_limit": 100}
}'

# Listen for workflow responses
docker exec redis redis-cli SUBSCRIBE workflow-responses
```

## Workflow Templates

### Template Structure
```yaml
id: my-workflow
name: My Custom Workflow
description: Description of what this workflow does
category: data-science
version: "1.0"
variables:
  - name: input_param
    type: string
    description: Parameter description
    required: true
    default: "default_value"

workflow:
  id: my-workflow-execution
  name: My Workflow Execution
  timeout: 3600
  variables:
    input_param: ${input_param}
  
  tasks:
    - id: task1
      name: First Task
      type: data
      parameters:
        operation: traverse
        query:
          cypher: "MATCH (n) RETURN n LIMIT ${input_param}"
      retry_policy:
        max_retries: 2
        backoff_type: exponential
        initial_delay: 1s
      timeout: 60
      
    - id: task2
      name: Second Task  
      type: ai
      depends_on: ["task1"]
      parameters:
        provider: anthropic
        prompt: "Analyze this data: ${task1.output}"
        response_format: json
      timeout: 120
```

### Built-in Templates

#### Data Analysis Pipeline (`data-analysis-basic`)
- Graph traversal with Neo4j
- Vector similarity search with Qdrant
- Data combination and processing
- AI-powered analysis
- Report generation

#### AI Content Pipeline (`ai-content-pipeline`)
- Multi-stage content generation
- Content review and optimization
- SEO analysis
- Quality gate validation
- Final packaging

## Task Types

### Data Tasks
Execute operations on the data-abstractor service:
```yaml
- id: graph_query
  type: data
  parameters:
    operation: traverse
    query:
      cypher: "MATCH (n)-[r]-(m) RETURN n,r,m LIMIT 100"
    enrich: ["metadata"]
```

### AI Tasks
Execute AI operations:
```yaml
- id: ai_analysis
  type: ai
  parameters:
    provider: anthropic
    prompt: "Analyze this data and identify patterns"
    model: claude-3-sonnet
    response_format: json
    max_tokens: 2000
```

### Exec Tasks
Execute containers with the exec-agent:
```yaml
- id: data_processing
  type: exec
  parameters:
    image: "python:3.9-slim"
    command: ["python", "/workspace/input/process.py"]
    files:
      - name: process.py
        content: |
          import json
          # Processing logic here
```

### Parallel Tasks
Execute multiple tasks concurrently:
```yaml
- id: parallel_processing
  type: parallel
  parameters:
    tasks:
      - type: data
        parameters:
          operation: search
      - type: ai
        parameters:
          prompt: "Generate summary"
```

### Condition Tasks
Conditional workflow branching:
```yaml
- id: quality_check
  type: condition
  condition: "${data_quality_score > 0.8}"
  on_success: ["publish_results"]
  on_failure: ["retry_processing"]
```

## Recovery System

The orchestrator includes automatic recovery capabilities:

### Features
- **Timeout Detection**: Identifies workflows and tasks that have exceeded time limits
- **Orphan Recovery**: Detects workflows with no recent activity
- **Checkpoint System**: Saves intermediate task states for recovery
- **Automatic Retry**: Configurable retry policies with backoff strategies
- **Cleanup**: Removes expired execution data

### Recovery Strategies
- **Restart**: Start workflow from beginning
- **Resume**: Continue from last checkpoint
- **Fail**: Mark as failed and stop execution

## Monitoring

### Health Check
```bash
curl http://localhost:8080/health
```

### Service Status
```bash
curl http://localhost:8080/status
```

### Metrics
- Active workflow count
- Template statistics  
- Recovery operations
- Redis connection status

## Development

### Project Structure
```
orchestrator/
├── clients/              # External service clients
│   ├── redis_state.go   # State persistence
│   └── message_coordinator.go # Service communication
├── config/              # Configuration management
├── engine/              # Workflow execution engine
│   ├── dag.go          # Dependency analysis
│   └── executor.go     # Workflow orchestration
├── handlers/            # Business logic handlers
│   ├── ai_generator.go # AI workflow generation
│   ├── template_manager.go # Template management
│   ├── task_executor.go # Task execution
│   └── recovery_manager.go # Recovery system
├── models/              # Data structures
├── templates/           # Workflow templates
├── examples/           # Example workflows and usage
├── main.go             # Application entry point
├── Dockerfile          # Container definition
└── README.md          # This file
```

### Building from Source
```bash
# Install dependencies
go mod download

# Run tests
go test ./...

# Build binary
go build -o orchestrator main.go

# Run locally
./orchestrator
```

### Adding New Task Types
1. Define task parameters in `models/workflow.go`
2. Add execution logic in `handlers/task_executor.go`
3. Update validation in `ValidateTask()`
4. Add examples and documentation

### Creating Templates
1. Create YAML file in `templates/` directory
2. Define variables and workflow structure  
3. Test with HTTP API or message bus
4. Add to version control

## Integration

### With Data Abstractor
- Uses `data-requests` / `data-responses` channels
- Supports traverse, search, and enrich operations
- Automatic retry and error handling

### With AI Abstractor  
- Uses `ai-requests` / `ai-responses` channels
- Supports multiple providers (OpenAI, Anthropic)
- Response format validation

### With Exec Agent
- Uses `exec-requests` / `exec-responses` channels
- Container orchestration and data mounting
- Service proxy for inter-service communication

## Security

- No hardcoded secrets or API keys
- Environment-based configuration
- Redis authentication support
- Input validation and sanitization
- CORS configuration for web access

## Production Considerations

- **Scalability**: Horizontal scaling through Redis clustering
- **High Availability**: Multiple orchestrator instances
- **Monitoring**: Structured JSON logging with correlation IDs
- **Resource Limits**: Configurable concurrency limits
- **Data Persistence**: Redis persistence and backup strategies
- **Network Security**: Internal service communication
- **Error Handling**: Comprehensive error reporting and recovery

This orchestrator provides a robust foundation for complex workflow orchestration with AI integration, making it suitable for production data processing and automation pipelines.