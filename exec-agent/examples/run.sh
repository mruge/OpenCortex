#!/bin/bash

# Example run script for worker container
# This script demonstrates how to structure a worker container for exec-agent

echo "=== Worker Container Starting ==="
echo "Container: $(hostname)"
echo "Working Directory: $(pwd)"
echo "Available files:"
ls -la /workspace/ 2>/dev/null || echo "No workspace directory"

# Show environment variables
echo "Environment variables:"
env | grep -E "(OPERATION|CONFIG|INPUT)" || echo "No relevant environment variables"

# Check for input files
if [ -d "/workspace/input" ]; then
    echo "Input files:"
    ls -la /workspace/input/
fi

# Execute the main worker script
echo "=== Starting Main Processing ==="
python3 /app/worker_script.py

# Check exit code
if [ $? -eq 0 ]; then
    echo "=== Processing Completed Successfully ==="
    
    # Show output files
    if [ -d "/workspace/output" ]; then
        echo "Output files:"
        ls -la /workspace/output/
    fi
else
    echo "=== Processing Failed ==="
    exit 1
fi

echo "=== Worker Container Finished ==="