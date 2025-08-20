package clients

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/sirupsen/logrus"
)

type AnthropicClient struct {
	apiKey    string
	baseURL   string
	model     string
	maxTokens int
	client    *http.Client
}

type anthropicRequest struct {
	Model     string                   `json:"model"`
	MaxTokens int                      `json:"max_tokens"`
	Messages  []anthropicMessage       `json:"messages"`
	System    string                   `json:"system,omitempty"`
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []anthropicContent `json:"content"`
	Usage   anthropicUsage     `json:"usage"`
	Error   *anthropicError    `json:"error,omitempty"`
}

type anthropicContent struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func NewAnthropicClient(apiKey, baseURL, model string, maxTokens int) (*AnthropicClient, error) {
	logrus.WithFields(logrus.Fields{
		"model":      model,
		"max_tokens": maxTokens,
		"base_url":   baseURL,
	}).Info("Anthropic client initialized")

	return &AnthropicClient{
		apiKey:    apiKey,
		baseURL:   baseURL,
		model:     model,
		maxTokens: maxTokens,
		client:    &http.Client{},
	}, nil
}

func (c *AnthropicClient) GenerateResponse(ctx context.Context, systemMessage, userPrompt string) (string, int, error) {
	reqBody := anthropicRequest{
		Model:     c.model,
		MaxTokens: c.maxTokens,
		Messages: []anthropicMessage{
			{
				Role:    "user",
				Content: userPrompt,
			},
		},
	}

	if systemMessage != "" {
		reqBody.System = systemMessage
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return "", 0, err
	}

	url := fmt.Sprintf("%s/v1/messages", c.baseURL)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewBuffer(jsonBody))
	if err != nil {
		return "", 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := c.client.Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()

	var anthropicResp anthropicResponse
	if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
		return "", 0, err
	}

	if anthropicResp.Error != nil {
		return "", 0, fmt.Errorf("anthropic error: %s", anthropicResp.Error.Message)
	}

	if len(anthropicResp.Content) == 0 {
		return "", anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens, nil
	}

	content := anthropicResp.Content[0].Text
	totalTokens := anthropicResp.Usage.InputTokens + anthropicResp.Usage.OutputTokens

	logrus.WithFields(logrus.Fields{
		"model":          c.model,
		"tokens_used":    totalTokens,
		"content_length": len(content),
	}).Debug("Anthropic response generated")

	return content, totalTokens, nil
}