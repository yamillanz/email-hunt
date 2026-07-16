// Package ai provides optional AI-powered email pattern generation via LLM APIs.
package ai

import (
	"context"
	"fmt"
)

type Provider interface {
	Generate(ctx context.Context, name, domain string) ([]string, error)
}

func NewProvider(name string) (Provider, error) {
	switch name {
	case "openai":
		return NewOpenAI()
	case "anthropic":
		return NewAnthropic()
	default:
		return nil, fmt.Errorf("unknown AI provider: %s (valid: openai, anthropic)", name)
	}
}
