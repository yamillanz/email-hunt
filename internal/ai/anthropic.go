package ai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const anthropicAPIURL = "https://api.anthropic.com/v1/messages"

type anthropicRequest struct {
	Model       string          `json:"model"`
	MaxTokens   int             `json:"max_tokens"`
	Temperature float64         `json:"temperature"`
	System      string          `json:"system"`
	Messages    []anthropicMsg  `json:"messages"`
}

type anthropicMsg struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type anthropicResponse struct {
	Content []struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"content"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type AnthropicProvider struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
}

func NewAnthropic() (*AnthropicProvider, error) {
	key := os.Getenv("EMAIL_HUNT_ANTHROPIC_KEY")
	if key == "" {
		return nil, fmt.Errorf("EMAIL_HUNT_ANTHROPIC_KEY environment variable not set")
	}
	return &AnthropicProvider{
		apiKey:     key,
		apiURL:     anthropicAPIURL,
		httpClient: &http.Client{},
	}, nil
}

func (p *AnthropicProvider) Generate(ctx context.Context, name, domain string) ([]string, error) {
	prompt := fmt.Sprintf(`Generate a list of possible professional email addresses for a person named "%s" at the company with domain "%s".

Rules:
- Generate at least 10 possible email combinations.
- Include creative and less common patterns beyond first.last@domain.com.
- Consider name variations (nicknames, initials, middle names if applicable).
- Include patterns with different separators (dot, dash, underscore, none).
- Return ONLY a valid JSON array of email address strings, with no additional text or explanation.`, name, domain)

	body := anthropicRequest{
		Model:       "claude-3-5-haiku-latest",
		MaxTokens:   500,
		Temperature: 0.3,
		System:      "You are an email address pattern generator. You respond only with JSON arrays of email strings. No explanations.",
		Messages: []anthropicMsg{
			{Role: "user", Content: prompt},
		},
	}

	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, p.apiURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", p.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")

	resp, err := p.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("api call: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("anthropic API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result anthropicResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if result.Error != nil {
		return nil, fmt.Errorf("anthropic error: %s — %s", result.Error.Type, result.Error.Message)
	}

	text := ""
	for _, c := range result.Content {
		if c.Type == "text" {
			text += c.Text
		}
	}
	if text == "" {
		return nil, fmt.Errorf("no text content in response")
	}

	text = strings.TrimSpace(text)
	text = stripMarkdownCodeBlock(text)

	var emails []string
	if err := json.Unmarshal([]byte(text), &emails); err != nil {
		return nil, fmt.Errorf("parse email list from AI response: %w\nresponse: %s", err, text)
	}

	return emails, nil
}
