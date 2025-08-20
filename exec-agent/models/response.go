package models

import "time"

type ExecutionResponse struct {
	CorrelationID string            `json:"correlation_id"`
	Success       bool              `json:"success"`
	Result        *ExecutionResult  `json:"result,omitempty"`
	Error         string            `json:"error,omitempty"`
	Timestamp     time.Time         `json:"timestamp"`
	Duration      time.Duration     `json:"duration"`
	ExecutionID   string            `json:"execution_id"`
}

type ExecutionResult struct {
	ExitCode      int                `json:"exit_code"`
	Output        string             `json:"output,omitempty"`
	Logs          string             `json:"logs,omitempty"`
	GraphUpdate   *GraphData         `json:"graph_update,omitempty"`
	OutputFiles   []OutputFile       `json:"output_files,omitempty"`
	MinioObjects  []MinioOutputObject `json:"minio_objects,omitempty"`
	Metadata      map[string]interface{} `json:"metadata,omitempty"`
}

type OutputFile struct {
	Name    string `json:"name"`
	Path    string `json:"path"`
	Content string `json:"content,omitempty"`
	Size    int64  `json:"size"`
}

type MinioOutputObject struct {
	ObjectName string `json:"object_name"`
	Size       int64  `json:"size"`
	URL        string `json:"url,omitempty"` // presigned URL if needed
}

func NewSuccessResponse(correlationID, executionID string, result *ExecutionResult, duration time.Duration) *ExecutionResponse {
	return &ExecutionResponse{
		CorrelationID: correlationID,
		Success:       true,
		Result:        result,
		Timestamp:     time.Now(),
		Duration:      duration,
		ExecutionID:   executionID,
	}
}

func NewErrorResponse(correlationID, executionID, error string, duration time.Duration) *ExecutionResponse {
	return &ExecutionResponse{
		CorrelationID: correlationID,
		Success:       false,
		Error:         error,
		Timestamp:     time.Now(),
		Duration:      duration,
		ExecutionID:   executionID,
	}
}