#!/bin/bash

echo "Testing Exec Agent Service"
echo "=========================="

# Function to send a request and display it
test_request() {
    local request="$1"
    local description="$2"
    
    echo ""
    echo "Test: $description"
    echo "Request: $request"
    echo ""
    
    # Send request
    docker exec smart_data_abstractor-redis-1 redis-cli PUBLISH exec-requests "$request"
    
    sleep 2
}

# Test 1: Simple container execution
test_request '{
  "correlation_id": "test-1",
  "container": {
    "image": "alpine:latest",
    "command": ["echo", "Hello from Exec Agent!"]
  },
  "input": {},
  "output": {
    "return_logs": true
  }
}' "Simple Alpine Container"

# Test 2: Python analysis with graph data
test_request '{
  "correlation_id": "test-2",
  "container": {
    "image": "python:3.9-slim"
  },
  "input": {
    "graph_data": {
      "nodes": [
        {"id": "node1", "labels": ["Person"], "properties": {"name": "Alice"}},
        {"id": "node2", "labels": ["Person"], "properties": {"name": "Bob"}}
      ],
      "relationships": [
        {"id": "rel1", "type": "KNOWS", "start_node": "node1", "end_node": "node2", "properties": {}}
      ]
    },
    "files": [
      {
        "name": "test_container.py",
        "path": "test_container.py", 
        "content": "import json\nwith open(\"/workspace/input/graph_data.json\") as f:\n    data = json.load(f)\nresult = {\"analysis\": \"Graph has {} nodes\".format(len(data[\"nodes\"]))}\nwith open(\"/workspace/output/analysis.json\", \"w\") as f:\n    json.dump(result, f)"
      }
    ]
  },
  "output": {
    "expected_files": ["analysis.json"],
    "return_logs": true
  }
}' "Python Graph Analysis"

# Test 3: Container with service access
test_request '{
  "correlation_id": "test-3",
  "container": {
    "image": "alpine:latest",
    "command": ["sh", "-c", "wget -qO- $SERVICE_PROXY_URL/health || echo \"Service proxy not accessible\""]
  },
  "input": {},
  "output": {
    "return_logs": true
  },
  "service_access": ["data", "ai"],
  "timeout": 60
}' "Service Access Test"

# Test 4: File processing with output
test_request '{
  "correlation_id": "test-4",
  "container": {
    "image": "alpine:latest",
    "command": ["sh", "-c", "cp /workspace/input/* /workspace/output/ && ls -la /workspace/output/"]
  },
  "input": {
    "files": [
      {"name": "input.txt", "path": "input.txt", "content": "Hello World"},
      {"name": "data.csv", "path": "data.csv", "content": "name,age\nAlice,25\nBob,30"}
    ]
  },
  "output": {
    "expected_files": ["input.txt", "data.csv"],
    "return_logs": true
  }
}' "File Processing"

echo ""
echo "Tests completed! Listen for responses:"
echo "docker exec smart_data_abstractor-redis-1 redis-cli SUBSCRIBE exec-responses"