#!/bin/bash

echo "Testing AI Abstractor Service"
echo "============================="

# Function to send a request and display it
test_request() {
    local request="$1"
    local description="$2"
    
    echo ""
    echo "Test: $description"
    echo "Request: $request"
    echo ""
    
    # Send request
    docker exec data-abstractor-redis-1 redis-cli PUBLISH ai-requests "$request"
    
    sleep 2
}

# Test 1: Simple OpenAI text request
test_request '{
  "provider": "openai",
  "correlation_id": "test-1",
  "prompt": "What is 2+2?",
  "response_format": "text"
}' "Simple OpenAI Question"

# Test 2: Anthropic JSON response
test_request '{
  "provider": "anthropic", 
  "correlation_id": "test-2",
  "prompt": "Create a simple user profile object",
  "response_format": "json"
}' "Anthropic JSON Response"

# Test 3: OpenAI with system message and context
test_request '{
  "provider": "openai",
  "correlation_id": "test-3",
  "system_message": "You are a helpful data analyst",
  "prompt": "Analyze the sales trend and provide insights",
  "context": [
    "January: $10,000",
    "February: $15,000", 
    "March: $12,000"
  ],
  "response_format": "markdown"
}' "OpenAI with Context and System Message"

# Test 4: YAML format request
test_request '{
  "provider": "openai",
  "correlation_id": "test-4", 
  "prompt": "Create a configuration file for a web application",
  "response_format": "yaml"
}' "YAML Configuration Example"

echo ""
echo "Tests completed! Listen for responses:"
echo "docker exec data-abstractor-redis-1 redis-cli SUBSCRIBE ai-responses"