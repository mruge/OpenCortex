#!/usr/bin/env python3
"""
Example container script that demonstrates exec-agent capabilities.
This script shows how containers can:
1. Read input data (graph data, files, config)
2. Access other services via HTTP proxy
3. Generate output files and graph updates
"""

import json
import os
import requests
from pathlib import Path

def main():
    print("🚀 Starting exec-agent test container")
    
    # Check environment variables
    execution_id = os.environ.get('EXECUTION_ID', 'unknown')
    print(f"📋 Execution ID: {execution_id}")
    
    # Set up workspace paths
    input_dir = Path('/workspace/input')
    output_dir = Path('/workspace/output')
    config_dir = Path('/workspace/config')
    
    output_dir.mkdir(exist_ok=True)
    
    print(f"📁 Workspace structure:")
    print(f"   Input: {input_dir}")
    print(f"   Output: {output_dir}")
    print(f"   Config: {config_dir}")
    
    # Read input data
    results = {}
    
    # 1. Read graph data if available
    graph_file = input_dir / 'graph_data.json'
    if graph_file.exists():
        with open(graph_file) as f:
            graph_data = json.load(f)
        print(f"📊 Graph data: {len(graph_data.get('nodes', []))} nodes, {len(graph_data.get('relationships', []))} relationships")
        results['input_graph_stats'] = {
            'node_count': len(graph_data.get('nodes', [])),
            'relationship_count': len(graph_data.get('relationships', []))
        }
    
    # 2. Read configuration if available
    config_file = config_dir / 'config.json'
    if config_file.exists():
        with open(config_file) as f:
            config_data = json.load(f)
        print(f"⚙️  Configuration: {list(config_data.keys())}")
        results['config_keys'] = list(config_data.keys())
    
    # 3. List input files
    input_files = list(input_dir.glob('*'))
    print(f"📄 Input files: {[f.name for f in input_files if f.is_file()]}")
    results['input_files'] = [f.name for f in input_files if f.is_file()]
    
    # 4. Test service access if available
    service_proxy_url = os.environ.get('SERVICE_PROXY_URL')
    if service_proxy_url:
        try:
            print(f"🌐 Testing service proxy at {service_proxy_url}")
            health_response = requests.get(f"{service_proxy_url}/health", timeout=5)
            if health_response.status_code == 200:
                health_data = health_response.json()
                print(f"✅ Service proxy healthy: {health_data['status']}")
                results['service_proxy_status'] = 'healthy'
                results['available_services'] = list(health_data.get('services', {}).keys())
            else:
                print(f"⚠️  Service proxy unhealthy: {health_response.status_code}")
                results['service_proxy_status'] = 'unhealthy'
        except Exception as e:
            print(f"❌ Service proxy error: {e}")
            results['service_proxy_error'] = str(e)
    
    # 5. Generate output files
    print("📝 Generating output files...")
    
    # Main results file
    results_file = output_dir / 'results.json'
    results['execution_id'] = execution_id
    results['status'] = 'completed'
    results['message'] = 'Test container executed successfully'
    
    with open(results_file, 'w') as f:
        json.dump(results, f, indent=2)
    print(f"✅ Created {results_file}")
    
    # Example analysis file
    analysis_file = output_dir / 'analysis.txt'
    with open(analysis_file, 'w') as f:
        f.write("Container Analysis Report\n")
        f.write("========================\n\n")
        f.write(f"Execution ID: {execution_id}\n")
        f.write(f"Input files processed: {len(results.get('input_files', []))}\n")
        f.write(f"Service proxy available: {'service_proxy_status' in results}\n")
        f.write(f"Graph data available: {'input_graph_stats' in results}\n")
        f.write(f"Configuration available: {'config_keys' in results}\n")
    print(f"✅ Created {analysis_file}")
    
    # 6. Generate graph update (if we had graph input)
    if graph_file.exists():
        print("🔄 Generating graph update...")
        graph_update = {
            "nodes": [
                {
                    "id": f"analysis-{execution_id}",
                    "labels": ["Analysis"],
                    "properties": {
                        "type": "container_execution",
                        "execution_id": execution_id,
                        "timestamp": "2025-01-01T00:00:00Z",
                        "status": "completed"
                    }
                }
            ],
            "relationships": [
                {
                    "id": f"analyzed-{execution_id}",
                    "type": "ANALYZED_BY",
                    "start_node": "original-data",
                    "end_node": f"analysis-{execution_id}",
                    "properties": {
                        "timestamp": "2025-01-01T00:00:00Z"
                    }
                }
            ]
        }
        
        graph_update_file = output_dir / 'graph_update.json'
        with open(graph_update_file, 'w') as f:
            json.dump(graph_update, f, indent=2)
        print(f"✅ Created {graph_update_file}")
    
    print("🎉 Container execution completed successfully!")
    return 0

if __name__ == "__main__":
    exit(main())