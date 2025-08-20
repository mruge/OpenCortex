package capabilities

// GetDataAbstractorCapabilities returns the capability definition for the data abstractor service
func GetDataAbstractorCapabilities() *ServiceCapabilities {
	return &ServiceCapabilities{
		Operations: []Operation{
			{
				Name:        "traverse",
				Description: "Execute Cypher queries against Neo4j graph database to traverse relationships and retrieve connected nodes",
				InputExample: map[string]interface{}{
					"operation":      "traverse",
					"correlation_id": "unique-request-id",
					"query": map[string]interface{}{
						"cypher": "MATCH (n:Person)-[r:KNOWS]->(m:Person) RETURN n, r, m LIMIT 10",
					},
					"enrich": []string{"metadata"},
					"limit":  100,
				},
				OutputExample: map[string]interface{}{
					"correlation_id": "unique-request-id",
					"success":        true,
					"operation":      "traverse",
					"timestamp":      "2025-01-20T10:30:00Z",
					"data": map[string]interface{}{
						"nodes": []map[string]interface{}{
							{
								"id":         "node1",
								"labels":     []string{"Person"},
								"properties": map[string]interface{}{"name": "John", "age": 30},
								"metadata":   map[string]interface{}{"enrichment_data": "additional_info"},
							},
						},
						"relationships": []map[string]interface{}{
							{
								"id":         "rel1",
								"type":       "KNOWS",
								"start_node": "node1",
								"end_node":   "node2",
								"properties": map[string]interface{}{"since": "2020"},
							},
						},
					},
				},
				RetrySafe:         true,
				EstimatedDuration: "1-5s",
			},
			{
				Name:        "search",
				Description: "Perform vector similarity search using Qdrant or text search, then retrieve matching nodes from Neo4j",
				InputExample: map[string]interface{}{
					"operation":      "search",
					"correlation_id": "unique-request-id",
					"query": map[string]interface{}{
						"embedding": []float32{0.1, 0.2, 0.3, 0.4, 0.5},
						// Alternative: "text": "search query text"
					},
					"enrich": []string{"metadata"},
					"limit":  50,
				},
				OutputExample: map[string]interface{}{
					"correlation_id": "unique-request-id",
					"success":        true,
					"operation":      "search",
					"timestamp":      "2025-01-20T10:30:00Z",
					"data": map[string]interface{}{
						"nodes": []map[string]interface{}{
							{
								"id":         "node1",
								"labels":     []string{"Document"],
								"properties": map[string]interface{}{"title": "Document Title", "content": "..."},
								"metadata":   map[string]interface{}{"source": "system1"},
								"score":      0.95,
							},
						},
						"relationships": []map[string]interface{}{},
					},
				},
				RetrySafe:         true,
				EstimatedDuration: "2-10s",
			},
			{
				Name:        "enrich",
				Description: "Retrieve specific nodes by IDs from Neo4j and enrich them with metadata from MongoDB",
				InputExample: map[string]interface{}{
					"operation":      "enrich",
					"correlation_id": "unique-request-id",
					"query": map[string]interface{}{
						"node_ids": []string{"node1", "node2", "node3"},
					},
				},
				OutputExample: map[string]interface{}{
					"correlation_id": "unique-request-id",
					"success":        true,
					"operation":      "enrich",
					"timestamp":      "2025-01-20T10:30:00Z",
					"data": map[string]interface{}{
						"nodes": []map[string]interface{}{
							{
								"id":         "node1",
								"labels":     []string{"Person"},
								"properties": map[string]interface{}{"name": "John"},
								"metadata": map[string]interface{}{
									"last_updated": "2025-01-20T10:00:00Z",
									"source":       "external_api",
									"confidence":   0.9,
								},
							},
						},
						"relationships": []map[string]interface{}{},
					},
				},
				RetrySafe:         true,
				EstimatedDuration: "1-3s",
			},
		},
		MessagePatterns: MessagePatterns{
			RequestChannel:   "data-requests",
			ResponseChannel:  "data-responses",
			CorrelationField: "correlation_id",
		},
	}
}