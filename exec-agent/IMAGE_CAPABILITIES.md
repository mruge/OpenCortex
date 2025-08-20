# Exec Agent Image Capability Scanning

The exec agent now supports automatic discovery and announcement of capabilities from OCI (Docker) container images. This allows the exec agent to dynamically advertise the capabilities of all known worker images, making the system more intelligent and self-documenting.

## Features

- **Automatic Image Scanning**: Scans known worker images for capability metadata
- **Multiple Discovery Methods**: Supports Docker labels, embedded files, and inference
- **Dynamic Updates**: Periodically rescans images and updates capability announcements
- **Integration with Service Registry**: Worker image capabilities are announced to other services

## Configuration

### Environment Variables

```bash
# Enable image scanning (default: true)
IMAGE_SCAN_ENABLED=true

# How often to scan images for changes (default: 30m)
IMAGE_SCAN_INTERVAL=30m

# Comma-separated list of known worker images to scan
KNOWN_WORKER_IMAGES="python:3.9-slim,node:16-alpine,tensorflow/tensorflow:latest-py3,pytorch/pytorch:latest,data-processor:latest,etl-worker:latest"

# Capability announcement settings
CAPABILITY_ANNOUNCEMENTS_ENABLED=true
CAPABILITY_REFRESH_INTERVAL=5m
```

## How Images Declare Capabilities

### Method 1: Docker Labels

Add capability metadata directly to your Dockerfile using labels:

```dockerfile
FROM python:3.9-slim

# Declare capabilities via labels
LABEL exec-agent.capabilities='[
  {
    "name": "data_processing",
    "description": "Process CSV and JSON data with pandas",
    "retry_safe": true,
    "estimated_duration": "1m-10m"
  }
]'

# Additional metadata
LABEL version="1.0.0"
LABEL description="Python data processing worker"
LABEL framework="python"
```

### Method 2: Embedded Capabilities File

Include a `/app/capabilities.json` file in your image:

```json
[
  {
    "name": "ml_training",
    "description": "Train machine learning models using scikit-learn",
    "input_example": {
      "files": [{"name": "training_data.csv", "path": "training_data.csv"}],
      "config_data": {"model_type": "random_forest"}
    },
    "output_example": {
      "exit_code": 0,
      "output": "Model trained with accuracy: 0.95",
      "output_files": [{"name": "model.pkl", "path": "/workspace/output/model.pkl"}]
    },
    "retry_safe": false,
    "estimated_duration": "5m-60m"
  }
]
```

### Method 3: Automatic Inference

The exec agent can infer capabilities based on image names and labels:

- Images containing "python" → `python_execution` capability
- Images containing "tensorflow" → `ml_inference` capability  
- Images containing "pytorch" → `pytorch_inference` capability
- Images containing "node" → `node_execution` capability
- Images containing "data-processor" → `data_processing` capability
- Images containing "etl" → `etl_processing` capability

## Capability Announcement Format

Discovered image capabilities are announced with this naming pattern:
`image_{sanitized_image_name}_{operation_name}`

For example:
- Image: `my-registry/data-processor:v1.0`
- Operation: `csv_processing`
- Announced as: `image_data_processor_csv_processing`

## Example Worker Image Structure

```
worker-image/
├── Dockerfile
├── capabilities.json      # Optional: Embedded capabilities
├── worker_script.py      # Main processing script
├── run.sh               # Entry point script
└── requirements.txt     # Dependencies
```

### Sample Dockerfile

```dockerfile
FROM python:3.9-slim

# Install dependencies
COPY requirements.txt /app/
RUN pip install -r /app/requirements.txt

# Add capability metadata
LABEL exec-agent.capabilities='[{"name":"data_processing","description":"Process data files","retry_safe":true,"estimated_duration":"1m-10m"}]'
LABEL framework="python"
LABEL version="1.0.0"

# Copy application
COPY worker_script.py /app/
COPY run.sh /app/
COPY capabilities.json /app/  # Optional

RUN chmod +x /app/run.sh

WORKDIR /app
CMD ["/app/run.sh"]
```

## How It Works

1. **Startup Scan**: On startup, the exec agent scans all configured worker images
2. **Capability Extraction**: For each image, it:
   - Inspects Docker labels for `exec-agent.capabilities`
   - Attempts to extract `/app/capabilities.json` from the container
   - Infers capabilities based on image name and framework labels
3. **Dynamic Updates**: Periodically rescans images and announces changes
4. **Service Integration**: All discovered capabilities are announced via Redis to other services

## Monitoring

The exec agent logs detailed information about image scanning:

```json
{
  "level": "info",
  "msg": "Image capabilities scanned successfully",
  "image": "data-processor:latest",
  "operations": 3,
  "component": "image_scanner"
}
```

## Benefits

1. **Self-Documenting**: Worker images declare their own capabilities
2. **Dynamic Discovery**: New images are automatically discovered and announced
3. **Centralized Management**: All worker capabilities are visible to the orchestrator
4. **Intelligent Workflow Generation**: AI can use actual worker capabilities when generating workflows
5. **Service Awareness**: Other services know exactly what worker operations are available

## Troubleshooting

### Images Not Found
- Ensure images are available locally or can be pulled
- Check `KNOWN_WORKER_IMAGES` configuration
- Verify Docker daemon is accessible

### Capabilities Not Detected
- Check image labels: `docker inspect <image> | jq '.[0].Config.Labels'`
- Verify capabilities.json format and location
- Enable debug logging: `LOG_LEVEL=debug`

### Performance Issues
- Reduce `IMAGE_SCAN_INTERVAL` for less frequent scanning
- Minimize the number of images in `KNOWN_WORKER_IMAGES`
- Consider using image labels instead of file extraction

## Example Usage

1. Build a worker image with capabilities:
```bash
docker build -t my-data-processor:v1.0 -f worker-image.Dockerfile .
```

2. Add it to the known images list:
```bash
export KNOWN_WORKER_IMAGES="python:3.9-slim,my-data-processor:v1.0"
```

3. Start the exec agent:
```bash
./exec-agent
```

4. The exec agent will discover and announce the image capabilities automatically.

5. Other services (like the orchestrator) can now use these capabilities for workflow generation and task routing.