package models

import "time"

type AIResponse struct {
	CorrelationID  string    `json:"correlation_id"`
	Success        bool      `json:"success"`
	Content        string    `json:"content,omitempty"`
	Error          string    `json:"error,omitempty"`
	Timestamp      time.Time `json:"timestamp"`
	Provider       string    `json:"provider"`
	Model          string    `json:"model"`
	TokensUsed     int       `json:"tokens_used,omitempty"`
	ResponseFormat string    `json:"response_format"`
}

func NewSuccessResponse(correlationID, provider, model, content, format string, tokensUsed int) *AIResponse {
	return &AIResponse{
		CorrelationID:  correlationID,
		Success:        true,
		Content:        content,
		Timestamp:      time.Now(),
		Provider:       provider,
		Model:          model,
		TokensUsed:     tokensUsed,
		ResponseFormat: format,
	}
}

func NewErrorResponse(correlationID, provider, error string) *AIResponse {
	return &AIResponse{
		CorrelationID: correlationID,
		Success:       false,
		Error:         error,
		Timestamp:     time.Now(),
		Provider:      provider,
	}
}