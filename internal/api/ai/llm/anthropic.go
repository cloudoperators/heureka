// SPDX-FileCopyrightText: 2026 SAP SE or an SAP affiliate company and Greenhouse contributors
// SPDX-License-Identifier: Apache-2.0

package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
)

type AnthropicClient struct {
	baseURL    string
	model      string
	token      string
	httpClient *http.Client
}

func NewAnthropicClient(baseURL, model, token string) *AnthropicClient {
	return &AnthropicClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		token:   token,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type anthropicMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicRequest struct {
	Model     string             `json:"model"`
	MaxTokens int                `json:"max_tokens"`
	System    string             `json:"system,omitempty"`
	Messages  []anthropicMessage `json:"messages"`
}

type anthropicContentBlock struct {
	Type string `json:"type"`
	Text string `json:"text"`
}

type anthropicResponse struct {
	Content []anthropicContentBlock `json:"content"`
	Error   *anthropicError         `json:"error,omitempty"`
}

type anthropicError struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

func (c *AnthropicClient) Complete(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	if c.baseURL == "" {
		return "", fmt.Errorf("anthropic client: AI_SAP_PROXY_URL is not configured")
	}

	if c.token == "" {
		return "", fmt.Errorf("anthropic client: AI_SAP_PROXY_TOKEN is not configured")
	}

	payload := anthropicRequest{
		Model:     c.model,
		MaxTokens: 1024,
		System:    systemPrompt,
		Messages: []anthropicMessage{
			{Role: "user", Content: userMessage},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("anthropic client: marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/messages", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("anthropic client: build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("anthropic client: http request: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.Warnf("error while closing response body: %s", err)
		}
	}()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("anthropic client: read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("anthropic client: unexpected status %d: %s", resp.StatusCode, raw)
	}

	var ar anthropicResponse
	if err := json.Unmarshal(raw, &ar); err != nil {
		return "", fmt.Errorf("anthropic client: unmarshal response: %w", err)
	}

	if ar.Error != nil {
		return "", fmt.Errorf("anthropic client: model error [%s]: %s", ar.Error.Type, ar.Error.Message)
	}

	for _, block := range ar.Content {
		if block.Type == "text" {
			return strings.TrimSpace(block.Text), nil
		}
	}

	return "", fmt.Errorf("anthropic client: no text content block in response")
}
