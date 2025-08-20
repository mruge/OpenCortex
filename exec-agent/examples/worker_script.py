#!/usr/bin/env python3
"""
Example worker script that demonstrates capability-aware processing.
This script would be embedded in a worker container image.
"""

import json
import sys
import os
from pathlib import Path

def main():
    """Main entry point for the worker script."""
    print("Starting worker script execution...")
    
    # Read input configuration
    config_path = "/workspace/input/config.json"
    if os.path.exists(config_path):
        with open(config_path, 'r') as f:
            config = json.load(f)
        print(f"Loaded configuration: {config}")
    else:
        print("No configuration found, using defaults")
        config = {"operation": "default"}
    
    # Process based on operation type
    operation = config.get("operation", "default")
    
    if operation == "data_processing":
        result = process_data(config)
    elif operation == "ml_training":
        result = train_model(config)
    else:
        result = default_processing(config)
    
    # Write output
    output_dir = Path("/workspace/output")
    output_dir.mkdir(exist_ok=True)
    
    with open(output_dir / "result.json", 'w') as f:
        json.dump(result, f, indent=2)
    
    print(f"Worker script completed successfully. Result: {result}")
    return 0

def process_data(config):
    """Process data files."""
    print("Processing data files...")
    
    # Simulate data processing
    input_files = config.get("input_files", [])
    processed_count = len(input_files)
    
    return {
        "operation": "data_processing",
        "files_processed": processed_count,
        "status": "completed",
        "output_files": ["processed_data.json"]
    }

def train_model(config):
    """Train a machine learning model."""
    print("Training ML model...")
    
    # Simulate model training
    model_type = config.get("model_type", "random_forest")
    
    return {
        "operation": "ml_training",
        "model_type": model_type,
        "accuracy": 0.95,
        "status": "completed",
        "output_files": ["model.pkl", "metrics.json"]
    }

def default_processing(config):
    """Default processing operation."""
    print("Performing default processing...")
    
    return {
        "operation": "default",
        "status": "completed",
        "message": "Default processing completed successfully"
    }

if __name__ == "__main__":
    sys.exit(main())