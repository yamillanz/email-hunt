package ai

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{"openai", true},
		{"anthropic", true},
		{"invalid", true},
		{"", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p, err := NewProvider(tt.name)
			if tt.wantErr && err == nil {
				t.Errorf("expected error for %q, got provider: %v", tt.name, p)
			}
		})
	}
}

func TestOpenAIGenerate(t *testing.T) {
	os.Setenv("EMAIL_HUNT_OPENAI_KEY", "test-key")
	defer os.Unsetenv("EMAIL_HUNT_OPENAI_KEY")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		resp := openaiResponse{
			Choices: []struct {
				Message struct {
					Content string `json:"content"`
				} `json:"message"`
			}{
				{Message: struct {
					Content string `json:"content"`
				}{Content: `["john.doe@example.com","jdoe@example.com","doe@example.com"]`}},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p, err := NewOpenAI()
	if err != nil {
		t.Fatalf("NewOpenAI failed: %v", err)
	}
	p.apiURL = srv.URL
	if err != nil {
		t.Fatalf("NewOpenAI failed: %v", err)
	}

	emails, err := p.Generate(t.Context(), "John Doe", "example.com")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if len(emails) != 3 {
		t.Errorf("expected 3 emails, got %d: %v", len(emails), emails)
	}
}

func TestOpenAIGenerateAuthError(t *testing.T) {
	os.Setenv("EMAIL_HUNT_OPENAI_KEY", "wrong-key")
	defer os.Unsetenv("EMAIL_HUNT_OPENAI_KEY")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, `{"error":{"message":"invalid api key"}}`, http.StatusUnauthorized)
	}))
	defer srv.Close()

	p, _ := NewOpenAI()
	p.apiURL = srv.URL
	_, err := p.Generate(t.Context(), "John Doe", "example.com")
	if err == nil {
		t.Error("expected error for auth failure")
	}
}

func TestOpenAIMissingAPIKey(t *testing.T) {
	os.Unsetenv("EMAIL_HUNT_OPENAI_KEY")
	_, err := NewOpenAI()
	if err == nil {
		t.Error("expected error when env var not set")
	}
}

func TestAnthropicGenerate(t *testing.T) {
	os.Setenv("EMAIL_HUNT_ANTHROPIC_KEY", "test-key")
	defer os.Unsetenv("EMAIL_HUNT_ANTHROPIC_KEY")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("x-api-key") != "test-key" {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}

		resp := anthropicResponse{
			Content: []struct {
				Type string `json:"type"`
				Text string `json:"text"`
			}{
				{Type: "text", Text: `["j.smith@example.com","smith.j@example.com"]`},
			},
		}
		json.NewEncoder(w).Encode(resp)
	}))
	defer srv.Close()

	p, err := NewAnthropic()
	if err != nil {
		t.Fatalf("NewAnthropic failed: %v", err)
	}
	p.apiURL = srv.URL
	if err != nil {
		t.Fatalf("NewAnthropic failed: %v", err)
	}

	emails, err := p.Generate(t.Context(), "Jane Smith", "example.com")
	if err != nil {
		t.Fatalf("Generate failed: %v", err)
	}
	if len(emails) != 2 {
		t.Errorf("expected 2 emails, got %d: %v", len(emails), emails)
	}
}

func TestAnthropicErrorResponse(t *testing.T) {
	os.Setenv("EMAIL_HUNT_ANTHROPIC_KEY", "test-key")
	defer os.Unsetenv("EMAIL_HUNT_ANTHROPIC_KEY")

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(anthropicResponse{
			Error: &struct {
				Type    string `json:"type"`
				Message string `json:"message"`
			}{Type: "invalid_request_error", Message: "model not found"},
		})
	}))
	defer srv.Close()

	p, _ := NewAnthropic()
	p.apiURL = srv.URL
	_, err := p.Generate(t.Context(), "Jane Smith", "example.com")
	if err == nil {
		t.Error("expected error for API error response")
	}
}

func TestStripMarkdownCodeBlock(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"```json\n[\"a@b.com\"]\n```", `["a@b.com"]`},
		{"```\n[\"a@b.com\"]\n```", `["a@b.com"]`},
		{"[\"a@b.com\"]", `["a@b.com"]`},
		{"  [\"a@b.com\"]  ", `["a@b.com"]`},
	}

	for _, tt := range tests {
		got := stripMarkdownCodeBlock(tt.input)
		if got != tt.want {
			t.Errorf("stripMarkdownCodeBlock(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
