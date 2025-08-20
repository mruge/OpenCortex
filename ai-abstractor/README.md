# AI Abstractor

A Go application that provides a unified interface to OpenAI and Anthropic APIs through Redis messaging, with support for prompt formatting, context inclusion, and response format enforcement.

## Features

- **Redis Message Bus**: Listen on `ai-requests` channel, respond on `ai-responses`
- **Multiple AI Providers**: Support for OpenAI GPT models and Anthropic Claude models
- **Context Management**: Include additional text context in prompts
- **Response Formatting**: Automatic prompt enhancement for JSON, YAML, Markdown, and text responses
- **Format Validation**: Basic validation of response formats
- **Flexible Configuration**: Environment-based configuration for all settings

## Request Format

Send AI requests via Redis:

```json
{
  "provider": "openai",
  "correlation_id": "unique-id",
  "prompt": "Explain quantum computing",
  "system_message": "You are a helpful physics teacher",
  "context": [
    "Quantum computing uses quantum bits (qubits)",
    "Superposition allows multiple states simultaneously"
  ],
  "response_format": "markdown",
  "model": "gpt-4",
  "max_tokens": 2000,
  "temperature": 0.7
}
```

### Request Fields

- `provider`: `"openai"` or `"anthropic"`
- `correlation_id`: Unique identifier for request tracking
- `prompt`: Main prompt/question for the AI
- `system_message` (optional): System message for persona/behavior
- `context` (optional): Array of context strings to include
- `response_format`: `"json"`, `"yaml"`, `"markdown"`, or `"text"`
- `model` (optional): Override default model
- `max_tokens` (optional): Override default token limit
- `temperature` (optional): Override default temperature (OpenAI only)

## Response Format

All responses return structured data:

```json
{
  "correlation_id": "unique-id",
  "success": true,
  "content": "Quantum computing is a revolutionary...",
  "timestamp": "2025-01-01T00:00:00Z",
  "provider": "openai",
  "model": "gpt-4",
  "tokens_used": 150,
  "response_format": "markdown"
}
```

Error responses:

```json
{
  "correlation_id": "unique-id",
  "success": false,
  "error": "OpenAI error: insufficient quota",
  "timestamp": "2025-01-01T00:00:00Z",
  "provider": "openai"
}
```

## Supported Response Formats

### JSON
Automatically adds format instruction: *"Please provide your response in valid JSON format only..."*

### YAML
Automatically adds format instruction: *"Please provide your response in valid YAML format only..."*

### Markdown
Automatically adds format instruction: *"Please provide your response in well-formatted Markdown..."*

### Text
Plain text response without special formatting instructions.

## Configuration

Environment variables:

```env
# Redis
REDIS_URL=redis://localhost:6379
AI_REQUEST_CHANNEL=ai-requests
AI_RESPONSE_CHANNEL=ai-responses

# OpenAI
OPENAI_API_KEY=your_key_here
OPENAI_MODEL=gpt-4
OPENAI_MAX_TOKENS=4000
OPENAI_TEMPERATURE=0.7

# Anthropic
ANTHROPIC_API_KEY=your_key_here
ANTHROPIC_MODEL=claude-3-sonnet-20240229
ANTHROPIC_MAX_TOKENS=4000

# App
LOG_LEVEL=info
```

## Running

### With Docker Compose (Recommended)

1. Copy environment file:
   ```bash
   cp .env.example .env
   ```

2. Add your API keys to `.env`

3. Start the service:
   ```bash
   docker-compose up -d
   ```

### Local Development

1. Install Go 1.21+
2. Install dependencies:
   ```bash
   go mod download
   ```
3. Set environment variables
4. Start Redis server
5. Run the application:
   ```bash
   go run main.go
   ```

## Testing

Send a test message via Redis CLI:

```bash
redis-cli
PUBLISH ai-requests '{"provider":"openai","correlation_id":"test-1","prompt":"What is 2+2?","response_format":"text"}'
```

Listen for responses:
```bash
redis-cli
SUBSCRIBE ai-responses
```

## Example Requests

### Simple Question
```json
{
  "provider": "openai",
  "correlation_id": "simple-1",
  "prompt": "What is the capital of France?",
  "response_format": "text"
}
```

### JSON Analysis with Context
```json
{
  "provider": "anthropic", 
  "correlation_id": "analysis-1",
  "prompt": "Analyze the sales data and provide insights",
  "context": [
    "Q1 sales: $100k",
    "Q2 sales: $150k", 
    "Q3 sales: $120k"
  ],
  "response_format": "json"
}
```

### Code Review with Persona
```json
{
  "provider": "openai",
  "correlation_id": "review-1", 
  "system_message": "You are a senior software engineer conducting a code review",
  "prompt": "Review this Python function for best practices",
  "context": ["def fibonacci(n): return n if n <= 1 else fibonacci(n-1) + fibonacci(n-2)"],
  "response_format": "markdown"
}
```

## Architecture

- `config/`: Environment configuration management
- `clients/`: OpenAI, Anthropic, and Redis client implementations
- `models/`: Request/response data structures
- `handlers/`: Business logic and prompt processing
- `main.go`: Application entry point with graceful shutdown