# Exec Agent

A containerized execution platform that orchestrates Docker containers through Redis messaging, with built-in data mounting, service access, and output management.

## 🏗️ Features

- **Container Orchestration**: Execute any Docker container with custom configurations
- **Data Management**: Mount JSON graph data, Minio blobs, and files into containers
- **Service Integration**: HTTP proxy for containers to access Data and AI abstractors
- **Output Extraction**: Collect files, graph updates, and Minio uploads from containers
- **Redis Messaging**: Listen on `exec-requests`, respond on `exec-responses`
- **Workspace Isolation**: Each execution gets isolated workspace with cleanup

## 🚀 Architecture

### Container Lifecycle
1. **Request Processing**: Parse execution request from Redis
2. **Workspace Creation**: Create isolated workspace with input/output/config directories
3. **Data Mounting**: Mount graph data, Minio objects, and files into workspace
4. **Container Execution**: Run specified Docker container with mounted workspace
5. **Output Collection**: Extract results, files, and graph updates
6. **Cleanup**: Remove workspace and temporary resources

### Service Access Methods
Containers can access other services through:
- **HTTP Proxy**: Service proxy at `http://host.docker.internal:9000`
- **Environment Variables**: Pre-configured service URLs
- **Direct Mounting**: Configuration files with service endpoints

## 📋 Request Format

```json
{
  "correlation_id": "unique-id",
  "container": {
    "image": "python:3.9-slim",
    "command": ["python", "/workspace/input/script.py"],
    "working_dir": "/workspace",
    "ports": {"8000": "8001"}
  },
  "input": {
    "graph_data": {
      "nodes": [...],
      "relationships": [...]
    },
    "minio_objects": [
      {
        "object_name": "data/input.csv",
        "local_path": "input.csv"
      }
    ],
    "files": [
      {
        "name": "script.py",
        "path": "script.py",
        "content": "import json\n..."
      }
    ],
    "config_data": {
      "param1": "value1",
      "param2": 42
    }
  },
  "output": {
    "expected_files": ["results.json", "analysis.csv"],
    "minio_upload": true,
    "graph_update": true,
    "return_logs": true
  },
  "environment": {
    "CUSTOM_VAR": "value"
  },
  "timeout": 300,
  "service_access": ["data", "ai"]
}
```

### Request Fields

- **`container`**: Docker container specification
  - `image`: Container image name
  - `command`: Command to execute (optional)
  - `working_dir`: Working directory (default: `/workspace`)
  - `ports`: Port mappings (optional)

- **`input`**: Input data specification
  - `graph_data`: JSON graph data to mount as `/workspace/input/graph_data.json`
  - `minio_objects`: Minio blobs to download and mount
  - `files`: Direct file content to create
  - `config_data`: Configuration data as `/workspace/config/config.json`

- **`output`**: Output collection specification
  - `expected_files`: Files to collect from `/workspace/output/`
  - `minio_upload`: Upload output directory to Minio
  - `graph_update`: Look for graph updates in `output/graph_update.json`
  - `return_logs`: Include execution logs in response

- **`environment`**: Custom environment variables
- **`timeout`**: Execution timeout in seconds (default: 300)
- **`service_access`**: Services to make available (["data", "ai"])

## 📤 Response Format

```json
{
  "correlation_id": "unique-id",
  "success": true,
  "result": {
    "exit_code": 0,
    "output": "Container stdout/stderr",
    "logs": "Execution logs",
    "graph_update": {
      "nodes": [...],
      "relationships": [...]
    },
    "output_files": [
      {
        "name": "results.json",
        "path": "/workspace/output/results.json", 
        "content": "{...}",
        "size": 1024
      }
    ],
    "minio_objects": [
      {
        "object_name": "executions/uuid/output/results.json",
        "size": 1024
      }
    ],
    "metadata": {
      "execution_time": "5.2s"
    }
  },
  "timestamp": "2025-01-01T00:00:00Z",
  "duration": "5.2s",
  "execution_id": "uuid"
}
```

## 🔧 Container Environment

Containers receive these environment variables:
- `EXECUTION_ID`: Unique execution identifier
- `WORKSPACE_INPUT`: Input directory path (`/workspace/input`)
- `WORKSPACE_OUTPUT`: Output directory path (`/workspace/output`)  
- `WORKSPACE_CONFIG`: Config directory path (`/workspace/config`)
- `SERVICE_PROXY_URL`: Service proxy endpoint (if service access enabled)
- `DATA_SERVICE_URL`: Data abstractor endpoint (if data service enabled)
- `AI_SERVICE_URL`: AI abstractor endpoint (if AI service enabled)

## 🌐 Service Access

### HTTP Proxy Endpoints

When `service_access` includes `["data", "ai"]`, containers can access:

```bash
# Data abstractor queries
curl http://host.docker.internal:9000/data/query \
  -d '{"operation":"traverse","query":{"cypher":"MATCH (n) RETURN n"}}'

# AI abstractor queries  
curl http://host.docker.internal:9000/ai/query \
  -d '{"provider":"openai","prompt":"Hello world"}'

# Health check
curl http://host.docker.internal:9000/health
```

## 💾 Data Mounting Structure

```
/workspace/
├── input/           # Input data
│   ├── graph_data.json    # Graph data (if provided)
│   ├── input.csv          # Minio objects (if provided)
│   └── script.py          # Direct files (if provided)
├── output/          # Output collection
│   ├── results.json       # Expected output files
│   └── graph_update.json  # Graph updates (if any)
└── config/          # Configuration
    └── config.json        # Config data (if provided)
```

## 🚀 Usage Examples

### 1. Python Data Analysis

```json
{
  "correlation_id": "analysis-1",
  "container": {
    "image": "python:3.9-slim",
    "command": ["python", "-c", "
      import json, os
      with open('/workspace/input/graph_data.json') as f:
          data = json.load(f)
      result = {'node_count': len(data['nodes'])}
      with open('/workspace/output/analysis.json', 'w') as f:
          json.dump(result, f)
    "]
  },
  "input": {
    "graph_data": {
      "nodes": [{"id": "1"}, {"id": "2"}],
      "relationships": []
    }
  },
  "output": {
    "expected_files": ["analysis.json"]
  }
}
```

### 2. R Statistical Analysis with Minio

```json
{
  "correlation_id": "stats-1", 
  "container": {
    "image": "r-base:latest",
    "command": ["Rscript", "/workspace/input/analysis.R"]
  },
  "input": {
    "minio_objects": [
      {"object_name": "datasets/sales.csv", "local_path": "sales.csv"}
    ],
    "files": [
      {
        "name": "analysis.R",
        "path": "analysis.R",
        "content": "data <- read.csv('/workspace/input/sales.csv')\nsummary(data)\nwrite.csv(summary(data), '/workspace/output/summary.csv')"
      }
    ]
  },
  "output": {
    "expected_files": ["summary.csv"],
    "minio_upload": true
  }
}
```

### 3. AI-Powered Analysis

```json
{
  "correlation_id": "ai-analysis-1",
  "container": {
    "image": "python:3.9-slim", 
    "command": ["python", "/workspace/input/ai_analysis.py"]
  },
  "input": {
    "files": [
      {
        "name": "ai_analysis.py",
        "path": "ai_analysis.py",
        "content": "
import requests, json, os
data = json.load(open('/workspace/input/graph_data.json'))
ai_request = {
  'provider': 'openai',
  'prompt': f'Analyze this graph data: {data}',
  'response_format': 'json'
}
response = requests.post(
  os.environ['AI_SERVICE_URL'] + '/query',
  json=ai_request
)
with open('/workspace/output/ai_analysis.json', 'w') as f:
    json.dump(response.json(), f)
        "
      }
    ],
    "graph_data": {"nodes": [...], "relationships": [...]}
  },
  "output": {
    "expected_files": ["ai_analysis.json"]
  },
  "service_access": ["ai"]
}
```

## 🐳 Running

### With Docker Compose (Standalone)

```bash
cd exec-agent
docker-compose up -d
```

### With Main System

```bash
# From smart_data_abstractor root
docker-compose up -d
```

## 🧪 Testing

Send execution request via Redis:

```bash
redis-cli PUBLISH exec-requests '{
  "correlation_id": "test-1",
  "container": {
    "image": "alpine:latest",
    "command": ["echo", "Hello from container!"]
  },
  "input": {},
  "output": {
    "return_logs": true
  }
}'

# Listen for response
redis-cli SUBSCRIBE exec-responses
```

## 🔒 Security Considerations

- **Docker Socket Access**: Agent requires Docker socket mount for container management
- **Network Isolation**: Containers run on isolated networks with controlled service access
- **Workspace Cleanup**: Automatic cleanup of workspaces and temporary files
- **Resource Limits**: Configure Docker container resource limits as needed
- **Image Validation**: Consider implementing image allowlists for production

## 🛠️ Configuration

Environment variables in `.env`:

```env
REDIS_URL=redis://localhost:6379
DOCKER_HOST=unix:///var/run/docker.sock
MINIO_ENDPOINT=localhost:9000
SERVICE_PROXY_PORT=9000
```

## 🔮 Future Extensions

- **Kubernetes Support**: K8s job execution for cloud-native deployments
- **Resource Limits**: CPU/memory limits for container execution
- **Image Registry**: Private registry support with authentication
- **Execution Queuing**: Job queue management with priority scheduling
- **Monitoring**: Metrics collection and execution tracking
- **Security Scanning**: Container image vulnerability scanning