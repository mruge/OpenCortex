package models

import "time"

type Response struct {
	CorrelationID string      `json:"correlation_id"`
	Success       bool        `json:"success"`
	Data          *GraphData  `json:"data,omitempty"`
	Error         string      `json:"error,omitempty"`
	Timestamp     time.Time   `json:"timestamp"`
	Operation     string      `json:"operation"`
}

type GraphData struct {
	Nodes         []GraphNode         `json:"nodes"`
	Relationships []GraphRelationship `json:"relationships"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type GraphNode struct {
	ID         string                 `json:"id"`
	Labels     []string               `json:"labels"`
	Properties map[string]interface{} `json:"properties"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	Score      float32                `json:"score,omitempty"`
}

type GraphRelationship struct {
	ID         string                 `json:"id"`
	Type       string                 `json:"type"`
	StartNode  string                 `json:"start_node"`
	EndNode    string                 `json:"end_node"`
	Properties map[string]interface{} `json:"properties"`
}

func NewSuccessResponse(correlationID, operation string, data *GraphData) *Response {
	return &Response{
		CorrelationID: correlationID,
		Success:       true,
		Data:          data,
		Timestamp:     time.Now(),
		Operation:     operation,
	}
}

func NewErrorResponse(correlationID, operation, error string) *Response {
	return &Response{
		CorrelationID: correlationID,
		Success:       false,
		Error:         error,
		Timestamp:     time.Now(),
		Operation:     operation,
	}
}