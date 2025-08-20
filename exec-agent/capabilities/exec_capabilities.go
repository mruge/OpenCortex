package capabilities

// GetExecAgentCapabilities returns the capability definition for the exec agent service
func GetExecAgentCapabilities() *ServiceCapabilities {
	return &ServiceCapabilities{
		Operations: []Operation{
			{
				Name:        "execute_container",
				Description: "Execute Docker containers with custom configurations, data mounting, service access, and output collection",
				InputExample: map[string]interface{}{
					"correlation_id": "unique-request-id",
					"container": map[string]interface{}{
						"image":       "python:3.9-slim",
						"command":     []string{"python", "/workspace/input/script.py"},
						"working_dir": "/workspace",
						"ports":       map[string]string{"8080": "8080"},
					},
					"input": map[string]interface{}{
						"graph_data": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"id":         "node1",
									"labels":     []string{"Person"},
									"properties": map[string]interface{}{"name": "John", "age": 30},
								},
							},
							"relationships": []map[string]interface{}{},
						},
						"minio_objects": []map[string]interface{}{
							{
								"object_name": "input-data.csv",
								"local_path":  "data.csv",
							},
						},
						"files": []map[string]interface{}{
							{
								"name":    "script.py",
								"path":    "script.py",
								"content": "import json\nprint('Processing data...')\nwith open('output.json', 'w') as f:\n    json.dump({'result': 'success'}, f)",
							},
						},
						"config_data": map[string]interface{}{
							"processing_mode": "batch",
							"timeout":         300,
						},
					},
					"output": map[string]interface{}{
						"expected_files": []string{"output.json", "results.csv"},
						"minio_upload":   true,
						"graph_update":   true,
						"return_logs":    true,
					},
					"environment": map[string]string{
						"PROCESSING_MODE": "production",
						"LOG_LEVEL":       "INFO",
					},
					"timeout":        600,
					"service_access": []string{"data", "ai"},
				},
				OutputExample: map[string]interface{}{
					"correlation_id": "unique-request-id",
					"success":        true,
					"execution_id":   "exec_1234567890",
					"timestamp":      "2025-01-20T10:30:00Z",
					"duration":       "45.2s",
					"result": map[string]interface{}{
						"exit_code": 0,
						"output":    "Processing data...\nCompleted successfully",
						"logs":      "INFO: Starting processing\nINFO: Data loaded\nINFO: Processing complete",
						"graph_update": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"id":         "node1",
									"labels":     []string{"Person", "Processed"},
									"properties": map[string]interface{}{"name": "John", "age": 30, "processed": true},
								},
							},
							"relationships": []map[string]interface{}{},
						},
						"output_files": []map[string]interface{}{
							{
								"name":    "output.json",
								"path":    "/workspace/output/output.json",
								"content": `{"result": "success"}`,
								"size":    20,
							},
						},
						"minio_objects": []map[string]interface{}{
							{
								"object_name": "results-exec_1234567890.json",
								"size":        1024,
								"url":         "https://minio.example.com/bucket/results-exec_1234567890.json?presigned=true",
							},
						},
						"metadata": map[string]interface{}{
							"container_runtime": "docker",
							"resource_usage": map[string]interface{}{
								"cpu_time":    "2.1s",
								"memory_peak": "128MB",
							},
						},
					},
				},
				RetrySafe:         false, // Container execution may have side effects
				EstimatedDuration: "10s-10m",
			},
			{
				Name:        "data_processing",
				Description: "Execute data processing containers with graph data input/output and service access for complex analytics",
				InputExample: map[string]interface{}{
					"correlation_id": "data-proc-001",
					"container": map[string]interface{}{
						"image":   "data-processor:latest",
						"command": []string{"/app/process", "--input", "/workspace/input", "--output", "/workspace/output"},
					},
					"input": map[string]interface{}{
						"graph_data": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{"id": "dataset1", "labels": []string{"Dataset"}, "properties": map[string]interface{}{"type": "csv", "size": 1000000}},
							},
							"relationships": []map[string]interface{}{},
						},
					},
					"output": map[string]interface{}{
						"graph_update":   true,
						"expected_files": []string{"processed_data.json", "analytics_report.html"},
					},
					"service_access": []string{"data"},
					"timeout":        1800,
				},
				OutputExample: map[string]interface{}{
					"correlation_id": "data-proc-001",
					"success":        true,
					"execution_id":   "exec_data_proc_001",
					"result": map[string]interface{}{
						"exit_code": 0,
						"output":    "Data processing completed. 1,000,000 records processed.",
						"graph_update": map[string]interface{}{
							"nodes": []map[string]interface{}{
								{
									"id":     "dataset1",
									"labels": []string{"Dataset", "Processed"},
									"properties": map[string]interface{}{
										"type":             "csv",
										"size":             1000000,
										"processed_at":     "2025-01-20T10:30:00Z",
										"processing_stats": map[string]interface{}{"records_processed": 1000000, "errors": 0},
									},
								},
							},
						},
						"output_files": []map[string]interface{}{
							{
								"name": "processed_data.json",
								"path": "/workspace/output/processed_data.json",
								"size": 5242880,
							},
						},
					},
				},
				RetrySafe:         false,
				EstimatedDuration: "5m-30m",
			},
			{
				Name:        "ai_model_inference",
				Description: "Execute AI model inference containers with data mounting and AI service integration",
				InputExample: map[string]interface{}{
					"correlation_id": "inference-001",
					"container": map[string]interface{}{
						"image":   "ml-inference:pytorch-cpu",
						"command": []string{"python", "/app/inference.py"},
					},
					"input": map[string]interface{}{
						"minio_objects": []map[string]interface{}{
							{
								"object_name": "model.pkl",
								"local_path":  "model/model.pkl",
							},
							{
								"object_name": "input_data.json",
								"local_path":  "data/input.json",
							},
						},
						"config_data": map[string]interface{}{
							"model_config": map[string]interface{}{
								"batch_size":   32,
								"max_sequence": 512,
								"temperature":  0.7,
							},
						},
					},
					"output": map[string]interface{}{
						"expected_files": []string{"predictions.json", "confidence_scores.csv"},
						"minio_upload":   true,
					},
					"service_access": []string{"ai"},
					"timeout":        900,
				},
				OutputExample: map[string]interface{}{
					"correlation_id": "inference-001",
					"success":        true,
					"execution_id":   "exec_inference_001",
					"result": map[string]interface{}{
						"exit_code": 0,
						"output":    "Model inference completed. Processed 1,000 samples with average confidence 0.89",
						"output_files": []map[string]interface{}{
							{
								"name": "predictions.json",
								"path": "/workspace/output/predictions.json",
								"size": 102400,
							},
						},
						"minio_objects": []map[string]interface{}{
							{
								"object_name": "predictions-inference-001.json",
								"size":        102400,
							},
						},
						"metadata": map[string]interface{}{
							"inference_stats": map[string]interface{}{
								"samples_processed": 1000,
								"avg_confidence":    0.89,
								"processing_time":   "12.3s",
							},
						},
					},
				},
				RetrySafe:         false,
				EstimatedDuration: "30s-15m",
			},
		},
		MessagePatterns: MessagePatterns{
			RequestChannel:   "exec-requests",
			ResponseChannel:  "exec-responses",
			CorrelationField: "correlation_id",
		},
	}
}