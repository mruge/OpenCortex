package capabilities

// GetAIAbstractorCapabilities returns the capability definition for the AI abstractor service
func GetAIAbstractorCapabilities() *ServiceCapabilities {
	return &ServiceCapabilities{
		Operations: []Operation{
			{
				Name:        "generate_content",
				Description: "Generate text content using OpenAI or Anthropic AI models with customizable prompts, context, and response formats",
				InputExample: map[string]interface{}{
					"provider":        "anthropic",
					"correlation_id":  "unique-request-id",
					"prompt":          "Write a technical summary of machine learning concepts",
					"system_message":  "You are a helpful technical writer who explains complex topics clearly",
					"context": []string{
						"Target audience: software developers",
						"Length: 500-1000 words",
					},
					"response_format": "markdown",
					"model":           "claude-3-sonnet",
					"max_tokens":      2000,
					"temperature":     0.7,
				},
				OutputExample: map[string]interface{}{
					"correlation_id":   "unique-request-id",
					"success":          true,
					"content":          "# Machine Learning Overview\n\nMachine learning is a subset of artificial intelligence...",
					"timestamp":        "2025-01-20T10:30:00Z",
					"provider":         "anthropic",
					"model":            "claude-3-sonnet",
					"tokens_used":      850,
					"response_format":  "markdown",
				},
				RetrySafe:         true,
				EstimatedDuration: "3-30s",
			},
			{
				Name:        "generate_structured_data",
				Description: "Generate structured data (JSON/YAML) using AI models for configuration, schemas, or data transformation",
				InputExample: map[string]interface{}{
					"provider":        "openai",
					"correlation_id":  "unique-request-id",
					"prompt":          "Create a JSON schema for a user profile with name, email, age, and preferences",
					"system_message":  "You are a data architect. Generate valid, well-structured schemas",
					"response_format": "json",
					"model":           "gpt-4",
					"max_tokens":      1000,
					"temperature":     0.3,
				},
				OutputExample: map[string]interface{}{
					"correlation_id":  "unique-request-id",
					"success":         true,
					"content": `{
  "type": "object",
  "properties": {
    "name": {"type": "string"},
    "email": {"type": "string", "format": "email"},
    "age": {"type": "integer", "minimum": 0},
    "preferences": {"type": "object"}
  },
  "required": ["name", "email"]
}`,
					"timestamp":       "2025-01-20T10:30:00Z",
					"provider":        "openai",
					"model":           "gpt-4",
					"tokens_used":     245,
					"response_format": "json",
				},
				RetrySafe:         true,
				EstimatedDuration: "2-15s",
			},
			{
				Name:        "text_analysis",
				Description: "Analyze and extract insights from text using AI models for classification, sentiment analysis, or content summarization",
				InputExample: map[string]interface{}{
					"provider":        "anthropic",
					"correlation_id":  "unique-request-id",
					"prompt":          "Analyze the sentiment and key themes in this customer feedback",
					"context": []string{
						"Customer feedback: The product is amazing and works perfectly, but the delivery was slow",
					},
					"response_format": "json",
					"model":           "claude-3-haiku",
					"max_tokens":      500,
					"temperature":     0.2,
				},
				OutputExample: map[string]interface{}{
					"correlation_id": "unique-request-id",
					"success":        true,
					"content": `{
  "sentiment": "mixed",
  "sentiment_score": 0.6,
  "themes": ["product_quality", "delivery_issues"],
  "positive_aspects": ["product functionality", "product quality"],
  "negative_aspects": ["delivery speed"],
  "overall_rating": "positive"
}`,
					"timestamp":       "2025-01-20T10:30:00Z",
					"provider":        "anthropic",
					"model":           "claude-3-haiku",
					"tokens_used":     125,
					"response_format": "json",
				},
				RetrySafe:         true,
				EstimatedDuration: "2-10s",
			},
		},
		MessagePatterns: MessagePatterns{
			RequestChannel:   "ai-requests",
			ResponseChannel:  "ai-responses",
			CorrelationField: "correlation_id",
		},
	}
}