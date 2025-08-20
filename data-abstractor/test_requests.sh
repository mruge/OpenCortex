#!/bin/bash

echo "Testing Data Abstractor Service"
echo "================================"

# Function to send a request and wait for response
test_request() {
    local request="$1"
    local description="$2"
    
    echo ""
    echo "Test: $description"
    echo "Request: $request"
    echo ""
    
    # Send request
    docker exec data-abstractor-redis-1 redis-cli PUBLISH data-requests "$request"
    
    # Listen for response (with timeout)
    timeout 5s docker exec data-abstractor-redis-1 redis-cli SUBSCRIBE data-responses &
    LISTENER_PID=$!
    
    sleep 2
    kill $LISTENER_PID 2>/dev/null
}

# Test 1: Simple traversal
test_request '{
  "operation": "traverse", 
  "correlation_id": "test-1",
  "query": {
    "cypher": "MATCH (n) RETURN n LIMIT 5"
  }
}' "Simple Node Traversal"

# Test 2: Traversal with relationships  
test_request '{
  "operation": "traverse",
  "correlation_id": "test-2", 
  "query": {
    "cypher": "MATCH (n)-[r]-(m) RETURN n,r,m LIMIT 3"
  },
  "enrich": ["metadata"]
}' "Graph Traversal with Enrichment"

# Test 3: Similarity search
test_request '{
  "operation": "search",
  "correlation_id": "test-3",
  "query": {
    "embedding": [0.1, 0.2, 0.3, 0.4, 0.5]
  },
  "limit": 10
}' "Vector Similarity Search"

# Test 4: Node enrichment
test_request '{
  "operation": "enrich",
  "correlation_id": "test-4", 
  "query": {
    "node_ids": ["node1", "node2", "node3"]
  }
}' "Node Enrichment"

echo ""
echo "Tests completed! Check application logs for processing details."
echo "To see logs: docker compose logs data-abstractor"