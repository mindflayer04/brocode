// Package openrouter sends text chat requests to OpenRouter.
package openrouter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const chatURL = "https://openrouter.ai/api/v1/chat/completions"

// Message is one turn in a conversation. Role is system, user, assistant, or tool.
type Message struct {
	Role       string     `json:"role"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	Content    string     `json:"content"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

// ToolCall is a model's request to invoke a function with specific arguments.
type ToolCall struct {
	ID       string           `json:"id"`
	Type     string           `json:"type"`
	Function ToolCallFunction `json:"function"`
}

// ToolCallFunction holds the function name and a JSON-encoded arguments string.
type ToolCallFunction struct {
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ToolFunctionParams describes a JSON Schema for a function's input object.
// Property and array-item schemas reuse this type to support nested inputs.
type ToolFunctionParams struct {
	Type        string                        `json:"type"`
	Description string                        `json:"description,omitempty"`
	Properties  map[string]ToolFunctionParams `json:"properties,omitempty"`
	Required    []string                      `json:"required,omitempty"`
	Items       *ToolFunctionParams           `json:"items,omitempty"`
}

// ToolFunction defines the function offered to the model.
type ToolFunction struct {
	Name        string             `json:"name"`
	Description string             `json:"description"`
	Parameters  ToolFunctionParams `json:"parameters"`
}

// Tool wraps a function definition. Type is "function" for function tools.
type Tool struct {
	Type     string       `json:"type"`
	Function ToolFunction `json:"function"`
}

// Client holds the credentials and reuses an HTTP client across requests.
type Client struct {
	apiKey     string
	httpClient *http.Client
}

// NewClient creates a client. The caller supplies OPENROUTER_API_KEY from its environment.
func NewClient(apiKey string) (*Client, error) {
	apiKey = strings.TrimSpace(apiKey)
	if apiKey == "" {
		return nil, fmt.Errorf("OPENROUTER_API_KEY is not set")
	}
	return &Client{
		apiKey:     apiKey,
		httpClient: &http.Client{Timeout: 2 * time.Minute},
	}, nil
}

type chatRequest struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Tools    []Tool    `json:"tools,omitempty"`
	Stream   bool      `json:"stream"`
}

type apiError struct {
	Message string `json:"message"`
}

type chatResponse struct {
	Choices []struct {
		Message      Message   `json:"message"`
		FinishReason string    `json:"finish_reason"`
		Error        *apiError `json:"error"`
	} `json:"choices"`
	Error *apiError `json:"error"`
}

// Chat sends the conversation to the chosen model and returns its first text reply.
// Pass previous user and assistant messages to continue a conversation.
func (c *Client) Chat(ctx context.Context, model string, messages []Message) (string, error) {
	model = strings.TrimSpace(model)
	if model == "" {
		return "", fmt.Errorf("model must not be empty")
	}
	if len(messages) == 0 {
		return "", fmt.Errorf("at least one message is required")
	}

	body, err := json.Marshal(chatRequest{Model: model, Messages: messages, Stream: false})
	if err != nil {
		return "", fmt.Errorf("encode chat request: %w", err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, chatURL, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("create chat request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("send OpenRouter request: %w", err)
	}
	defer resp.Body.Close()

	// Bound how much response data we keep in memory.
	const maxResponseBytes = 4 << 20 // 4 MiB
	data, err := io.ReadAll(io.LimitReader(resp.Body, maxResponseBytes+1))
	if err != nil {
		return "", fmt.Errorf("read OpenRouter response: %w", err)
	}
	if len(data) > maxResponseBytes {
		return "", fmt.Errorf("OpenRouter response exceeds 4 MiB")
	}

	var result chatResponse
	decodeErr := json.Unmarshal(data, &result)
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if decodeErr == nil && result.Error != nil {
			return "", fmt.Errorf("OpenRouter HTTP %d: %s", resp.StatusCode, result.Error.Message)
		}
		return "", fmt.Errorf("OpenRouter HTTP %d: %s", resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	if decodeErr != nil {
		return "", fmt.Errorf("decode OpenRouter response: %w", decodeErr)
	}
	// OpenRouter can return a provider error in the body even with HTTP 200.
	if result.Error != nil {
		return "", fmt.Errorf("OpenRouter: %s", result.Error.Message)
	}
	if len(result.Choices) == 0 {
		return "", fmt.Errorf("OpenRouter returned no choices")
	}
	choice := result.Choices[0]
	if choice.Error != nil {
		return "", fmt.Errorf("OpenRouter: %s", choice.Error.Message)
	}
	if choice.FinishReason == "error" {
		return "", fmt.Errorf("OpenRouter model failed to complete the response")
	}
	if strings.TrimSpace(choice.Message.Content) == "" {
		return "", fmt.Errorf("OpenRouter returned no text content")
	}
	return choice.Message.Content, nil
}
