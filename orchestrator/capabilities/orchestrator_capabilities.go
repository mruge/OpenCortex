package capabilities

// GetOrchestratorCapabilities returns the capability definition for the orchestrator service
func GetOrchestratorCapabilities() *ServiceCapabilities {
	return &ServiceCapabilities{
		Operations: []Operation{
			{
				Name:        "execute_workflow",
				Description: "Execute complex workflows with DAG-based task dependencies, AI integration, and state persistence",
				InputExample: map[string]interface{}{
					"correlation_id":     "workflow-001",
					"workflow_template":  "data-analysis-basic",
					"variables": map[string]interface{}{
						"query_limit":      100,
						"analysis_prompt":  "Identify key trends and patterns in the data",
						"search_embedding": []float32{0.1, 0.2, 0.3, 0.4, 0.5},
					},
					"priority": 1,
				},
				OutputExample: map[string]interface{}{
					"correlation_id": "workflow-001",
					"execution_id":   "exec_1642678800_123",
					"status":         "completed",
					"success":        true,
					"duration":       "2m45.3s",
					"timestamp":      "2025-01-20T10:30:00Z",
					"results": map[string]interface{}{
						"graph_data_processed": true,
						"analysis_generated":   true,
						"report_created":       true,
					},
					"task_results": map[string]interface{}{
						"graph_traversal": map[string]interface{}{
							"nodes_found":         150,
							"relationships_found": 245,
						},
						"ai_analysis": map[string]interface{}{
							"insights_count": 5,
							"confidence":     0.92,
						},
						"generate_report": map[string]interface{}{
							"report_generated": true,
							"file_size":        "2.1MB",
						},
					},
				},
				RetrySafe:         false, // Workflows may have side effects
				EstimatedDuration: "30s-60m",
			},
			{
				Name:        "generate_ai_workflow",
				Description: "Generate workflow definitions using AI based on natural language descriptions and requirements",
				InputExample: map[string]interface{}{
					"prompt":            "Create a data analysis pipeline that fetches graph data, performs vector search, and generates an AI-powered summary report",
					"domain":            "data-science",
					"complexity":        "medium",
					"required_services": []string{"data", "ai", "exec"},
					"output_format":     "yaml",
				},
				OutputExample: map[string]interface{}{
					"id":          "ai-generated-12345",
					"name":        "Data Analysis Pipeline",
					"description": "AI-generated workflow for comprehensive data analysis",
					"version":     "1.0",
					"tasks": []map[string]interface{}{
						{
							"id":   "fetch_data",
							"name": "Fetch Graph Data",
							"type": "data",
							"parameters": map[string]interface{}{
								"operation": "traverse",
								"query": map[string]interface{}{
									"cypher": "MATCH (n)-[r]-(m) RETURN n,r,m LIMIT ${query_limit}",
								},
							},
						},
						{
							"id":         "analyze_data",
							"name":       "AI Analysis",
							"type":       "ai",
							"depends_on": []string{"fetch_data"},
							"parameters": map[string]interface{}{
								"provider": "anthropic",
								"prompt":   "Analyze the data and generate insights",
							},
						},
					},
				},
				RetrySafe:         true,
				EstimatedDuration: "5-30s",
			},
			{
				Name:        "get_execution_status",
				Description: "Retrieve the current status and progress of a running or completed workflow execution",
				InputExample: map[string]interface{}{
					"execution_id": "exec_1642678800_123",
				},
				OutputExample: map[string]interface{}{
					"execution_id": "exec_1642678800_123",
					"workflow_id":  "data-analysis-basic",
					"status":       "running",
					"start_time":   "2025-01-20T10:25:00Z",
					"end_time":     null,
					"progress": map[string]interface{}{
						"total_tasks":     5,
						"completed_tasks": 3,
						"failed_tasks":    0,
						"current_tasks":   []string{"ai_analysis"},
					},
					"task_states": map[string]interface{}{
						"graph_traversal": map[string]interface{}{
							"status":     "completed",
							"start_time": "2025-01-20T10:25:01Z",
							"end_time":   "2025-01-20T10:25:15Z",
							"duration":   "14s",
						},
						"ai_analysis": map[string]interface{}{
							"status":      "running",
							"start_time":  "2025-01-20T10:27:30Z",
							"retry_count": 0,
						},
					},
				},
				RetrySafe:         true,
				EstimatedDuration: "100-500ms",
			},
			{
				Name:        "cancel_workflow",
				Description: "Cancel a running workflow execution and clean up resources",
				InputExample: map[string]interface{}{
					"execution_id": "exec_1642678800_123",
					"reason":       "User requested cancellation",
					"force":        false,
				},
				OutputExample: map[string]interface{}{
					"execution_id": "exec_1642678800_123",
					"cancelled":    true,
					"status":       "cancelled",
					"cancelled_at": "2025-01-20T10:30:00Z",
					"cleanup": map[string]interface{}{
						"tasks_stopped":     2,
						"resources_cleaned": true,
						"state_preserved":   true,
					},
				},
				RetrySafe:         false, // Cancellation is a one-time operation
				EstimatedDuration: "1-10s",
			},
			{
				Name:        "list_templates",
				Description: "Retrieve available workflow templates with filtering and categorization options",
				InputExample: map[string]interface{}{
					"category": "data-science",
					"search":   "analysis",
					"limit":    20,
				},
				OutputExample: map[string]interface{}{
					"templates": []map[string]interface{}{
						{
							"id":          "data-analysis-basic",
							"name":        "Basic Data Analysis Workflow",
							"description": "Perform graph traversal, vector search, and AI-powered analysis",
							"category":    "data-science",
							"version":     "1.0",
							"variables": []map[string]interface{}{
								{
									"name":        "query_limit",
									"type":        "int",
									"required":    true,
									"default":     100,
									"description": "Maximum number of results to return",
								},
							},
						},
						{
							"id":          "ai-content-pipeline",
							"name":        "AI Content Generation Pipeline",
							"description": "Multi-stage AI content creation with review and optimization",
							"category":    "content-generation",
							"version":     "1.0",
						},
					},
					"total_count": 15,
					"categories":  []string{"data-science", "content-generation", "automation"},
				},
				RetrySafe:         true,
				EstimatedDuration: "100-500ms",
			},
		},
		MessagePatterns: MessagePatterns{
			RequestChannel:   "workflow-requests",
			ResponseChannel:  "workflow-responses",
			CorrelationField: "correlation_id",
		},
	}
}