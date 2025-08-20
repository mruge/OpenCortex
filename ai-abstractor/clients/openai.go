package clients

import (
	"context"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

type OpenAIClient struct {
	client      *openai.Client
	model       string
	maxTokens   int
	temperature float32
}

func NewOpenAIClient(apiKey, baseURL, model string, maxTokens int, temperature float32) (*OpenAIClient, error) {
	config := openai.DefaultConfig(apiKey)
	
	if baseURL != "" {
		config.BaseURL = baseURL
	}

	client := openai.NewClientWithConfig(config)

	logrus.WithFields(logrus.Fields{
		"model":       model,
		"max_tokens":  maxTokens,
		"temperature": temperature,
		"base_url":    config.BaseURL,
	}).Info("OpenAI client initialized")

	return &OpenAIClient{
		client:      client,
		model:       model,
		maxTokens:   maxTokens,
		temperature: temperature,
	}, nil
}

func (c *OpenAIClient) GenerateResponse(ctx context.Context, systemMessage, userPrompt string) (string, int, error) {
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleUser,
			Content: userPrompt,
		},
	}

	if systemMessage != "" {
		messages = []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleSystem,
				Content: systemMessage,
			},
			{
				Role:    openai.ChatMessageRoleUser,
				Content: userPrompt,
			},
		}
	}

	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Messages:    messages,
		MaxTokens:   c.maxTokens,
		Temperature: c.temperature,
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return "", 0, err
	}

	if len(resp.Choices) == 0 {
		return "", resp.Usage.TotalTokens, nil
	}

	content := resp.Choices[0].Message.Content
	tokens := resp.Usage.TotalTokens

	logrus.WithFields(logrus.Fields{
		"model":         c.model,
		"tokens_used":   tokens,
		"content_length": len(content),
	}).Debug("OpenAI response generated")

	return content, tokens, nil
}