package models

type AIRequest struct {
	Provider      string   `json:"provider"`      // "openai" or "anthropic"
	CorrelationID string   `json:"correlation_id"`
	Prompt        string   `json:"prompt"`
	SystemMessage string   `json:"system_message,omitempty"`
	Context       []string `json:"context,omitempty"`
	ResponseFormat string  `json:"response_format"` // "json", "yaml", "text", "markdown"
	Model         string   `json:"model,omitempty"`
	MaxTokens     int      `json:"max_tokens,omitempty"`
	Temperature   float32  `json:"temperature,omitempty"`
}

const (
	ProviderOpenAI    = "openai"
	ProviderAnthropic = "anthropic"
	
	FormatJSON     = "json"
	FormatYAML     = "yaml"
	FormatText     = "text"
	FormatMarkdown = "markdown"
)