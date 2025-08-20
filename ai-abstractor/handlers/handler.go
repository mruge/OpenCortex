package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"ai-abstractor/clients"
	"ai-abstractor/models"

	"github.com/sirupsen/logrus"
)

type AIHandler struct {
	openAI    *clients.OpenAIClient
	anthropic *clients.AnthropicClient
}

func NewAIHandler(openAI *clients.OpenAIClient, anthropic *clients.AnthropicClient) *AIHandler {
	return &AIHandler{
		openAI:    openAI,
		anthropic: anthropic,
	}
}

func (h *AIHandler) HandleRequest(ctx context.Context, data []byte) []byte {
	var req models.AIRequest
	if err := json.Unmarshal(data, &req); err != nil {
		logrus.WithError(err).Error("Failed to unmarshal AI request")
		response := models.NewErrorResponse("", "", fmt.Sprintf("Invalid request format: %v", err))
		responseData, _ := json.Marshal(response)
		return responseData
	}

	logrus.WithFields(logrus.Fields{
		"correlation_id":  req.CorrelationID,
		"provider":        req.Provider,
		"response_format": req.ResponseFormat,
	}).Info("Processing AI request")

	// Build the complete prompt with context and format instructions
	fullPrompt := h.buildPrompt(req)
	
	var response *models.AIResponse

	switch req.Provider {
	case models.ProviderOpenAI:
		if h.openAI == nil {
			response = models.NewErrorResponse(req.CorrelationID, req.Provider, "OpenAI client not configured")
		} else {
			response = h.handleOpenAIRequest(ctx, &req, fullPrompt)
		}
	case models.ProviderAnthropic:
		if h.anthropic == nil {
			response = models.NewErrorResponse(req.CorrelationID, req.Provider, "Anthropic client not configured")
		} else {
			response = h.handleAnthropicRequest(ctx, &req, fullPrompt)
		}
	default:
		response = models.NewErrorResponse(req.CorrelationID, req.Provider, fmt.Sprintf("Unknown provider: %s", req.Provider))
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		logrus.WithError(err).Error("Failed to marshal AI response")
		errorResponse := models.NewErrorResponse(req.CorrelationID, req.Provider, "Internal error")
		responseData, _ = json.Marshal(errorResponse)
	}

	return responseData
}

func (h *AIHandler) buildPrompt(req models.AIRequest) string {
	var promptParts []string

	// Add context if provided
	if len(req.Context) > 0 {
		promptParts = append(promptParts, "Context:")
		for i, context := range req.Context {
			promptParts = append(promptParts, fmt.Sprintf("%d. %s", i+1, context))
		}
		promptParts = append(promptParts, "")
	}

	// Add the main prompt
	promptParts = append(promptParts, req.Prompt)

	// Add response format instructions
	formatInstruction := h.getFormatInstruction(req.ResponseFormat)
	if formatInstruction != "" {
		promptParts = append(promptParts, "", formatInstruction)
	}

	return strings.Join(promptParts, "\n")
}

func (h *AIHandler) getFormatInstruction(format string) string {
	switch strings.ToLower(format) {
	case models.FormatJSON:
		return "Please provide your response in valid JSON format only. Do not include any additional text, explanations, or formatting outside the JSON structure."
	case models.FormatYAML:
		return "Please provide your response in valid YAML format only. Do not include any additional text, explanations, or formatting outside the YAML structure."
	case models.FormatMarkdown:
		return "Please provide your response in well-formatted Markdown. Use appropriate headings, lists, code blocks, and other Markdown formatting as needed."
	case models.FormatText:
		return "Please provide your response as plain text without any special formatting."
	default:
		return ""
	}
}

func (h *AIHandler) handleOpenAIRequest(ctx context.Context, req *models.AIRequest, fullPrompt string) *models.AIResponse {
	content, tokens, err := h.openAI.GenerateResponse(ctx, req.SystemMessage, fullPrompt)
	if err != nil {
		logrus.WithError(err).Error("OpenAI request failed")
		return models.NewErrorResponse(req.CorrelationID, req.Provider, fmt.Sprintf("OpenAI error: %v", err))
	}

	// Validate response format if needed
	if !h.validateResponseFormat(content, req.ResponseFormat) {
		logrus.WithFields(logrus.Fields{
			"format": req.ResponseFormat,
			"content_preview": content[:min(100, len(content))],
		}).Warn("Response format validation failed")
	}

	return models.NewSuccessResponse(req.CorrelationID, req.Provider, "openai", content, req.ResponseFormat, tokens)
}

func (h *AIHandler) handleAnthropicRequest(ctx context.Context, req *models.AIRequest, fullPrompt string) *models.AIResponse {
	content, tokens, err := h.anthropic.GenerateResponse(ctx, req.SystemMessage, fullPrompt)
	if err != nil {
		logrus.WithError(err).Error("Anthropic request failed")
		return models.NewErrorResponse(req.CorrelationID, req.Provider, fmt.Sprintf("Anthropic error: %v", err))
	}

	// Validate response format if needed
	if !h.validateResponseFormat(content, req.ResponseFormat) {
		logrus.WithFields(logrus.Fields{
			"format": req.ResponseFormat,
			"content_preview": content[:min(100, len(content))],
		}).Warn("Response format validation failed")
	}

	return models.NewSuccessResponse(req.CorrelationID, req.Provider, "anthropic", content, req.ResponseFormat, tokens)
}

func (h *AIHandler) validateResponseFormat(content, format string) bool {
	switch strings.ToLower(format) {
	case models.FormatJSON:
		var js json.RawMessage
		return json.Unmarshal([]byte(content), &js) == nil
	case models.FormatYAML:
		// Basic YAML validation - check for key-value structure
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line != "" && !strings.Contains(line, ":") && !strings.HasPrefix(line, "-") {
				return false
			}
		}
		return true
	case models.FormatMarkdown:
		// Basic markdown validation - check for common markdown elements
		return strings.Contains(content, "#") || strings.Contains(content, "*") || strings.Contains(content, "`")
	case models.FormatText:
		return true // Any text is valid
	default:
		return true
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}