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

type SAPProxyClient struct {
	baseURL    string
	model      string
	token      string
	httpClient *http.Client
}

func NewSAPProxyClient(baseURL, model, token string) *SAPProxyClient {
	return &SAPProxyClient{
		baseURL: strings.TrimRight(baseURL, "/"),
		model:   model,
		token:   token,
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatRequest struct {
	Model    string        `json:"model"`
	Messages []chatMessage `json:"messages"`
}

type chatChoice struct {
	Message chatMessage `json:"message"`
}

type chatResponse struct {
	Choices []chatChoice `json:"choices"`
	Error   *chatError   `json:"error,omitempty"`
}

type chatError struct {
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

func (c *SAPProxyClient) Complete(ctx context.Context, systemPrompt, userMessage string) (string, error) {
	if c.baseURL == "" {
		return "", fmt.Errorf("sap proxy: AI_SAP_PROXY_URL is not configured")
	}

	if c.token == "" {
		return "", fmt.Errorf("sap proxy: AI_SAP_PROXY_TOKEN is not configured")
	}

	payload := chatRequest{
		Model: c.model,
		Messages: []chatMessage{
			{Role: "system", Content: systemPrompt},
			{Role: "user", Content: userMessage},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("sap proxy: marshal request: %w", err)
	}

	url := c.baseURL + "/chat/completions"

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("sap proxy: build request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.token)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("sap proxy: http request: %w", err)
	}

	defer func() {
		if err := resp.Body.Close(); err != nil {
			logrus.Warnf("error while closing response body: %s", err)
		}
	}()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("sap proxy: read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("sap proxy: unexpected status %d: %s", resp.StatusCode, raw)
	}

	var chatResp chatResponse
	if err := json.Unmarshal(raw, &chatResp); err != nil {
		return "", fmt.Errorf("sap proxy: unmarshal response: %w", err)
	}

	if chatResp.Error != nil {
		return "", fmt.Errorf("sap proxy: model error [%s]: %s", chatResp.Error.Code, chatResp.Error.Message)
	}

	if len(chatResp.Choices) == 0 {
		return "", fmt.Errorf("sap proxy: no choices in response")
	}

	return strings.TrimSpace(chatResp.Choices[0].Message.Content), nil
}
