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

const openaiAPIURL = "https://api.openai.com/v1/chat/completions"

type OpenAIProvider struct {
	apiKey     string
	apiURL     string
	httpClient *http.Client
}

func NewOpenAI() (*OpenAIProvider, error) {
	key := os.Getenv("EMAIL_HUNT_OPENAI_KEY")
	if key == "" {
		return nil, fmt.Errorf("EMAIL_HUNT_OPENAI_KEY environment variable not set")
	}
	return &OpenAIProvider{
		apiKey:     key,
		apiURL:     openaiAPIURL,
		httpClient: &http.Client{},
	}, nil
}

type openaiMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openaiRequest struct {
	Model       string          `json:"model"`
	Messages    []openaiMessage `json:"messages"`
	Temperature float64         `json:"temperature"`
	MaxTokens   int             `json:"max_tokens"`
}

type openaiResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}

func (p *OpenAIProvider) Generate(ctx context.Context, name, domain string) ([]string, error) {
	prompt := fmt.Sprintf(`Generate a list of possible professional email addresses for a person named "%s" at the company with domain "%s".

Rules:
- Generate at least 10 possible email combinations.
- Include creative and less common patterns beyond first.last@domain.com.
- Consider name variations (nicknames, initials, middle names if applicable).
- Include patterns with different separators (dot, dash, underscore, none).
- Return ONLY a valid JSON array of email address strings, with no additional text or explanation.

Example response format: ["jdoe@example.com","john.d@example.com"]

Return your answer as a JSON array of strings only.`, name, domain)

	body := openaiRequest{
		Model: "gpt-4o-mini",
		Messages: []openaiMessage{
			{Role: "system", Content: "You are an email address pattern generator. You respond only with JSON arrays of email strings. No explanations."},
			{Role: "user", Content: prompt},
		},
		Temperature: 0.3,
		MaxTokens:   500,
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
	req.Header.Set("Authorization", "Bearer "+p.apiKey)

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
		return nil, fmt.Errorf("openai API error (%d): %s", resp.StatusCode, string(respBody))
	}

	var result openaiResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return nil, fmt.Errorf("parse response: %w", err)
	}

	if len(result.Choices) == 0 {
		return nil, fmt.Errorf("no choices in response")
	}

	content := strings.TrimSpace(result.Choices[0].Message.Content)
	content = stripMarkdownCodeBlock(content)

	var emails []string
	if err := json.Unmarshal([]byte(content), &emails); err != nil {
		return nil, fmt.Errorf("parse email list from AI response: %w\nresponse: %s", err, content)
	}

	return emails, nil
}

func stripMarkdownCodeBlock(s string) string {
	s = strings.TrimPrefix(s, "```json")
	s = strings.TrimPrefix(s, "```")
	s = strings.TrimSuffix(s, "```")
	return strings.TrimSpace(s)
}
